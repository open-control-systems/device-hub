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

	deviceHTTP struct {
		fetchTimeout  string
		fetchInterval string
	}

	mdns struct {
		browseInterval string
		browseTimeout  string
	}
}

type appPipeline struct {
	closer      *core.FanoutCloser
	systemClock syscore.SystemClock
}

func (p *appPipeline) start(ec *envContext) error {
	appContext, cancelFunc := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancelFunc()
	defer func() {
		if err := p.closer.Close(); err != nil {
			core.LogErr.Printf("main: failed to close resources: %v\n", err)
		}
	}()

	serverPipeline, err := piphttp.NewServerPipeline(
		p.closer,
		htcore.ServerParams{
			Port: ec.port,
		},
	)
	if err != nil {
		return err
	}

	storagePipeline := stinfluxdb.NewPipeline(appContext, p.closer, ec.dbParams)

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

	mdnsBrowser := sysmdns.NewZeroconfBrowser(
		appContext,
		resolveStore,
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

	fetchInterval, err := time.ParseDuration(ec.deviceHTTP.fetchInterval)
	if err != nil {
		return err
	}
	if fetchInterval < time.Millisecond {
		return errors.New("HTTP device fetch interval can't be less than 1ms")
	}

	fetchTimeout, err := time.ParseDuration(ec.deviceHTTP.fetchTimeout)
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

	storeAwakener := pipdevice.NewStoreAwakener(mdnsBrowserRunner, cacheStore)

	registerHTTPRoutes(
		serverPipeline.GetServeMux(),
		// Time valid since 2024/12/03.
		piphttp.NewSystemTimeHandler(p.systemClock, time.Unix(1733215816, 0)),
		pipdevice.NewStoreHTTPHandler(storeAwakener),
	)

	mdnsBrowserRunner.Start()
	cacheStore.Start()
	storagePipeline.Start()
	serverPipeline.Start()

	<-appContext.Done()

	return nil
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
			return appPipeline.start(envContext)
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
		&envContext.deviceHTTP.fetchInterval,
		"device-http-fetch-interval", "5s",
		"HTTP device data fetch interval, in form of: 1h35m10s12ms"+
			" (valid time units are ms, s, m, h)",
	)
	cmd.Flags().StringVar(
		&envContext.deviceHTTP.fetchTimeout,
		"device-http-fetch-timeout", "5s",
		"HTTP device data fetch timeout, in form of: 1h35m10s12ms"+
			" (valid time units are ms, s, m, h)",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.browseInterval,
		"mdns-browse-interval", "1m",
		"How often to perform mDNS lookup over local network, in form of: 1h35m10s"+
			" (valid time units are s, m, h)",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.browseTimeout,
		"mdns-browse-timeout", "30s",
		"How long to perform a single mDNS lookup over local network, in form of: 1h35m10s"+
			" (valid time units are s, m, h)",
	)

	if err := cmd.Execute(); err != nil {
		core.LogErr.Printf("main: failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
