package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/http/htcore"
	"github.com/open-control-systems/device-hub/components/pipeline/pipdevice"
	"github.com/open-control-systems/device-hub/components/pipeline/piphttp"
	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/open-control-systems/device-hub/components/storage/stinfluxdb"
	"github.com/open-control-systems/device-hub/components/system/syscore"
	"github.com/open-control-systems/device-hub/components/system/sysmdns"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

type envContext struct {
	dbParams stinfluxdb.DBParams
	logPath  string
	cacheDir string
	port     int

	device struct {
		HTTP struct {
			fetchTimeout  string
			fetchInterval string
		}

		monitor struct {
			inactive struct {
				disable        bool
				maxInterval    string
				updateInterval string
			}
		}
	}

	mdns struct {
		browseInterval       string
		browseTimeout        string
		disableAutodiscovery bool
	}
}

type appPipeline struct {
	closer      *core.FanoutCloser
	starter     *syssched.FanoutStarter
	systemClock syscore.SystemClock
}

func (p *appPipeline) start(ec *envContext) error {
	appContext, cancelFunc := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancelFunc()
	storagePipeline := stinfluxdb.NewPipeline(appContext, ec.dbParams)
	p.closer.Add("storage-influxdb-pipeline", storagePipeline)
	p.starter.Add(storagePipeline)

	var db stcore.DB

	if ec.cacheDir != "" {
		ec.cacheDir = path.Join(ec.cacheDir, "bbolt.db")

		bboltDB, err := stcore.NewBboltDB(ec.cacheDir, &bbolt.Options{
			Timeout: time.Second * 5,
		})
		if err != nil {
			return err
		}
		p.closer.Add("bbolt-database", bboltDB)

		db = stcore.NewBboltDBBucket(bboltDB, "device_bucket")
	} else {
		db = &stcore.NoopDB{}
	}

	resolveStore := sysnet.NewResolveStore()

	mdnsBrowseInterval, err := time.ParseDuration(ec.mdns.browseInterval)
	if err != nil {
		return err
	}
	if mdnsBrowseInterval < time.Second {
		return errors.New("mDNS browse interval can't be less than 1s")
	}

	mdnsBrowseTimeout, err := time.ParseDuration(ec.mdns.browseTimeout)
	if err != nil {
		return err
	}
	if mdnsBrowseTimeout < time.Second {
		return errors.New("mDNS browse timeout can't be less than 1s")
	}

	fanoutServiceHandler := &sysmdns.FanoutServiceHandler{}

	resolveServiceHandler := sysmdns.NewResolveServiceHandler(resolveStore)
	fanoutServiceHandler.Add(resolveServiceHandler)

	mdnsBrowser := sysmdns.NewZeroconfBrowser(
		appContext,
		fanoutServiceHandler,
		sysmdns.ZeroconfBrowserParams{
			Service: sysmdns.ServiceName(sysmdns.ServiceTypeHTTP, sysmdns.ProtoTCP),
			Domain:  "local",
			Timeout: mdnsBrowseTimeout,
		},
	)
	p.closer.Add("mdns-zeroconf-browser", mdnsBrowser)

	mdnsBrowserRunner := syssched.NewAsyncTaskRunner(
		appContext,
		mdnsBrowser,
		mdnsBrowser,
		mdnsBrowseInterval,
	)
	p.closer.Add("mdns-zeroconf-browser-runner", mdnsBrowserRunner)
	p.starter.Add(mdnsBrowserRunner)

	fetchInterval, err := time.ParseDuration(ec.device.HTTP.fetchInterval)
	if err != nil {
		return err
	}
	if fetchInterval < time.Millisecond {
		return errors.New("HTTP device fetch interval can't be less than 1ms")
	}

	fetchTimeout, err := time.ParseDuration(ec.device.HTTP.fetchTimeout)
	if err != nil {
		return err
	}
	if fetchTimeout < time.Millisecond {
		return errors.New("HTTP device fetch timeout can't be less than 1ms")
	}

	cacheStoreParams := pipdevice.CacheStoreParams{}
	cacheStoreParams.HTTP.FetchInterval = fetchInterval
	cacheStoreParams.HTTP.FetchTimeout = fetchTimeout

	cacheStore := pipdevice.NewCacheStore(
		appContext,
		p.systemClock,
		storagePipeline.GetSystemClock(),
		storagePipeline.GetDataHandler(),
		db,
		resolveStore,
		cacheStoreParams,
	)
	p.closer.Add("device-cache-store", cacheStore)
	p.starter.Add(cacheStore)

	var deviceStore pipdevice.Store

	deviceStore = pipdevice.NewStoreAwakener(mdnsBrowserRunner, cacheStore)

	if !ec.device.monitor.inactive.disable {
		inactiveMaxInterval, err :=
			time.ParseDuration(ec.device.monitor.inactive.maxInterval)
		if err != nil {
			return err
		}

		if inactiveMaxInterval < time.Millisecond {
			return errors.New("device-monitor-inactive-max-interval can't be" +
				" less than 1ms")
		}

		if !ec.mdns.disableAutodiscovery {
			if inactiveMaxInterval < mdnsBrowseInterval {
				return errors.New("device-monitor-inactive-max-interval can't be" +
					" less than mdns-browse-interval")
			}
		}

		inactiveUpdateInterval, err :=
			time.ParseDuration(ec.device.monitor.inactive.updateInterval)
		if err != nil {
			return err
		}

		if inactiveUpdateInterval < time.Millisecond {
			return errors.New("device-monitor-inactive-update-interval can't be" +
				" less than 1ms")
		}

		aliveMonitor := pipdevice.NewStoreAliveMonitor(
			&syscore.LocalMonotonicClock{},
			deviceStore,
			inactiveMaxInterval,
		)
		cacheStore.SetAliveMonitor(aliveMonitor)

		deviceStore = aliveMonitor

		aliveMonitorRunner := syssched.NewAsyncTaskRunner(
			appContext,
			aliveMonitor,
			aliveMonitor,
			inactiveUpdateInterval,
		)

		p.closer.Add("device-alive-monitor-runner", aliveMonitorRunner)
		p.starter.Add(aliveMonitorRunner)
	}

	if !ec.mdns.disableAutodiscovery {
		storeMdnsHandler := pipdevice.NewStoreMdnsHandler(deviceStore)
		fanoutServiceHandler.Add(storeMdnsHandler)
	}

	mux := http.NewServeMux()

	server, err := htcore.NewServer(mux, htcore.ServerParams{
		Port: ec.port,
	})
	if err != nil {
		return err
	}
	p.closer.Add("http-server", server)
	p.starter.Add(server)

	registerHTTPRoutes(
		mux,
		// Time valid since 2024/12/03.
		piphttp.NewSystemTimeHandler(p.systemClock, time.Unix(1733215816, 0)),
		pipdevice.NewStoreHTTPHandler(deviceStore),
	)

	p.starter.Start()

	<-appContext.Done()

	return nil
}

func (p *appPipeline) close() error {
	return p.closer.Close()
}

func registerHTTPRoutes(
	mux *http.ServeMux,
	timeHandler *piphttp.SystemTimeHandler,
	storeHTTPHandler *pipdevice.StoreHTTPHandler,
) {
	mux.Handle("/api/v1/system/time", timeHandler)

	mux.HandleFunc("/api/v1/device/add", func(w http.ResponseWriter, r *http.Request) {
		storeHTTPHandler.HandleAdd(w, r)
	})
	mux.HandleFunc("/api/v1/device/remove", func(w http.ResponseWriter, r *http.Request) {
		storeHTTPHandler.HandleRemove(w, r)
	})
	mux.HandleFunc("/api/v1/device/list", func(w http.ResponseWriter, r *http.Request) {
		storeHTTPHandler.HandleList(w, r)
	})
}

func newAppPipeline() *appPipeline {
	return &appPipeline{
		systemClock: &syscore.LocalSystemClock{},
		closer:      &core.FanoutCloser{},
		starter:     &syssched.FanoutStarter{},
	}
}

func prepareEnvironment(ec *envContext) error {
	if ec.dbParams.URL == "" {
		return fmt.Errorf("influxdb URL is required")
	}
	if ec.dbParams.Org == "" {
		return fmt.Errorf("influxdb org is required")
	}
	if ec.dbParams.Bucket == "" {
		return fmt.Errorf("influxdb bucket is required")
	}
	if ec.dbParams.Token == "" {
		return fmt.Errorf("influxdb token is required")
	}

	if ec.cacheDir != "" {
		fi, err := os.Stat(ec.cacheDir)
		if err != nil {
			return err
		}

		if !fi.Mode().IsDir() {
			return errors.New("cache path should be a directory")
		}
	}

	if ec.logPath == "" {
		return fmt.Errorf("log path is required")
	}
	if err := core.SetLogFile(ec.logPath); err != nil {
		return err
	}

	return nil
}

func main() {
	appPipeline := newAppPipeline()
	envContext := &envContext{}

	cmd := &cobra.Command{
		Use:           "device-hub",
		Short:         "device-hub CLI",
		Long:          "device-hub collects and stores various data from IoT devices",
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return prepareEnvironment(envContext)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := appPipeline.start(envContext); err != nil {
				return err
			}

			return appPipeline.close()
		},
	}

	cmd.Flags().IntVar(&envContext.port, "http-port", 0,
		"HTTP server port (0 for random port)")

	cmd.Flags().StringVar(&envContext.cacheDir, "cache-dir", "", "device-hub cache directory")
	cmd.Flags().StringVar(&envContext.logPath, "log-path", "", "device-hub log file path")

	cmd.Flags().StringVar(&envContext.dbParams.URL, "influxdb-url", "", "influxdb URL")
	cmd.Flags().StringVar(&envContext.dbParams.Org, "influxdb-org", "", "influxdb Org")

	cmd.Flags().StringVar(&envContext.dbParams.Token, "influxdb-api-token", "",
		"influxdb API token")

	cmd.Flags().StringVar(&envContext.dbParams.Bucket, "influxdb-bucket", "",
		"influxdb bucket")

	cmd.Flags().StringVar(
		&envContext.device.HTTP.fetchInterval,
		"device-http-fetch-interval", "5s",
		"HTTP device data fetch interval",
	)
	cmd.Flags().StringVar(
		&envContext.device.HTTP.fetchTimeout,
		"device-http-fetch-timeout", "5s",
		"HTTP device data fetch timeout",
	)

	cmd.Flags().StringVar(
		&envContext.device.monitor.inactive.maxInterval,
		"device-monitor-inactive-max-interval", "2m",
		"How long it's allowed for a device to be inactive",
	)

	cmd.Flags().StringVar(
		&envContext.device.monitor.inactive.updateInterval,
		"device-monitor-inactive-update-interval", "10s",
		"How often to check for a device inactivity",
	)

	cmd.Flags().BoolVar(
		&envContext.device.monitor.inactive.disable,
		"device-monitor-inactive-disable", false,
		"Disable device inactivity monitoring",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.browseInterval,
		"mdns-browse-interval", "1m",
		"How often to perform mDNS lookup over local network",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.browseTimeout,
		"mdns-browse-timeout", "30s",
		"How long to perform a single mDNS lookup over local network",
	)

	cmd.Flags().BoolVar(
		&envContext.mdns.disableAutodiscovery,
		"mdns-autodiscovery-disable", false,
		"Disable automatic device discovery on the local network",
	)

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
