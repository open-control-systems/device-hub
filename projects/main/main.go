package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.etcd.io/bbolt"

	"github.com/open-control-systems/zeroconf"

	"github.com/open-control-systems/device-hub/components/device/devstore"
	"github.com/open-control-systems/device-hub/components/http/htcore"
	"github.com/open-control-systems/device-hub/components/http/hthandler"
	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/open-control-systems/device-hub/components/storage/stinfluxdb"
	"github.com/open-control-systems/device-hub/components/system/syscore"
	"github.com/open-control-systems/device-hub/components/system/sysmdns"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

type envContext struct {
	logPath  string
	cacheDir string
	port     int

	storage struct {
		influxdb stinfluxdb.DBParams
	}

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
		browse struct {
			interval string
			timeout  string
			iface    string
		}

		autodiscovery struct {
			disable bool
		}

		server struct {
			disable      bool
			hostname     string
			instanceName string
			iface        string
		}
	}
}

type appPipeline struct {
	stopper     *syssched.FanoutStopper
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

	resolveStore := sysnet.NewResolveStore()
	resolveServiceHandler := sysmdns.NewResolveServiceHandler(resolveStore)

	fanoutServiceHandler := &sysmdns.FanoutServiceHandler{}
	fanoutServiceHandler.Add(resolveServiceHandler)

	mdnsBrowseAwakener, err := p.createMdnsBrowser(appContext, fanoutServiceHandler, ec)
	if err != nil {
		return err
	}

	deviceStore, err := p.createDeviceStore(appContext, resolveStore, mdnsBrowseAwakener, ec)
	if err != nil {
		return err
	}

	if !ec.mdns.autodiscovery.disable {
		storeMdnsHandler := devstore.NewStoreMdnsHandler(deviceStore)
		fanoutServiceHandler.Add(storeMdnsHandler)
	}

	mux := http.NewServeMux()

	server, err := htcore.NewServer(mux, htcore.ServerParams{
		Port: ec.port,
	})
	if err != nil {
		return err
	}
	p.stopper.Add("http-server", server)
	p.starter.Add(server)

	registerHTTPRoutes(
		mux,
		// Time valid since 2024/12/03.
		hthandler.NewSystemTimeHandler(p.systemClock, time.Unix(1733215816, 0)),
		devstore.NewStoreHTTPHandler(deviceStore),
	)

	if !ec.mdns.server.disable {
		if err := p.configureMdnsServer(server, ec); err != nil {
			return err
		}
	}

	if err := p.starter.Start(); err != nil {
		return err
	}

	<-appContext.Done()

	return nil
}

func (p *appPipeline) stop() error {
	return p.stopper.Stop()
}

func (p *appPipeline) createDeviceStore(
	ctx context.Context,
	resolveStore *sysnet.ResolveStore,
	awakener syssched.Awakener,
	ec *envContext,
) (devstore.Store, error) {
	cacheStore, err := p.createCacheStore(ctx, resolveStore, ec)
	if err != nil {
		return nil, err
	}

	awakeStore := devstore.NewAwakeStore(awakener, cacheStore)

	if ec.device.monitor.inactive.disable {
		return awakeStore, nil
	}

	inactiveMaxInterval, err :=
		time.ParseDuration(ec.device.monitor.inactive.maxInterval)
	if err != nil {
		return nil, err
	}

	if inactiveMaxInterval < time.Millisecond {
		return nil, errors.New("device-monitor-inactive-max-interval can't be" +
			" less than 1ms")
	}

	inactiveUpdateInterval, err :=
		time.ParseDuration(ec.device.monitor.inactive.updateInterval)
	if err != nil {
		return nil, err
	}

	if inactiveUpdateInterval < time.Millisecond {
		return nil, errors.New("device-monitor-inactive-update-interval can't be" +
			" less than 1ms")
	}

	aliveMonitor := devstore.NewStoreAliveMonitor(
		&syscore.LocalMonotonicClock{},
		awakeStore,
		inactiveMaxInterval,
	)
	cacheStore.SetAliveMonitor(aliveMonitor)

	aliveMonitorRunner := syssched.NewAsyncTaskRunner(
		ctx,
		aliveMonitor,
		aliveMonitor,
		inactiveUpdateInterval,
	)

	p.stopper.Add("device-alive-monitor-runner", aliveMonitorRunner)
	p.starter.Add(aliveMonitorRunner)

	return aliveMonitor, nil
}

func (p *appPipeline) createMdnsBrowser(
	ctx context.Context,
	fanoutServiceHandler *sysmdns.FanoutServiceHandler,
	ec *envContext,
) (syssched.Awakener, error) {
	mdnsBrowseInterval, err := time.ParseDuration(ec.mdns.browse.interval)
	if err != nil {
		return nil, err
	}
	if mdnsBrowseInterval < time.Second {
		return nil, errors.New("mDNS browse interval can't be less than 1s")
	}

	mdnsBrowseTimeout, err := time.ParseDuration(ec.mdns.browse.timeout)
	if err != nil {
		return nil, err
	}
	if mdnsBrowseTimeout < time.Second {
		return nil, errors.New("mDNS browse timeout can't be less than 1s")
	}

	filteredIfaces, err := parseIfaceOption(ec.mdns.browse.iface)
	if err != nil {
		return nil, err
	}

	mdnsBrowser := sysmdns.NewZeroconfBrowser(
		ctx,
		fanoutServiceHandler,
		sysmdns.ZeroconfBrowserParams{
			Service: sysmdns.ServiceName(sysmdns.ServiceTypeHTTP, sysmdns.ProtoTCP),
			Domain:  "local",
			Timeout: mdnsBrowseTimeout,
			Opts: []zeroconf.ClientOption{
				zeroconf.SelectIfaces(filteredIfaces),
			},
		},
	)
	p.stopper.Add("mdns-zeroconf-browser", mdnsBrowser)

	mdnsBrowserRunner := syssched.NewAsyncTaskRunner(
		ctx,
		mdnsBrowser,
		mdnsBrowser,
		mdnsBrowseInterval,
	)
	p.stopper.Add("mdns-zeroconf-browser-runner", mdnsBrowserRunner)
	p.starter.Add(mdnsBrowserRunner)

	return mdnsBrowserRunner, nil
}

func (p *appPipeline) createCacheStore(
	ctx context.Context,
	resolveStore *sysnet.ResolveStore,
	ec *envContext,
) (*devstore.CacheStore, error) {
	fetchInterval, err := time.ParseDuration(ec.device.HTTP.fetchInterval)
	if err != nil {
		return nil, err
	}
	if fetchInterval < time.Millisecond {
		return nil, errors.New("HTTP device fetch interval can't be less than 1ms")
	}

	fetchTimeout, err := time.ParseDuration(ec.device.HTTP.fetchTimeout)
	if err != nil {
		return nil, err
	}
	if fetchTimeout < time.Millisecond {
		return nil, errors.New("HTTP device fetch timeout can't be less than 1ms")
	}

	cacheStoreParams := devstore.CacheStoreParams{}
	cacheStoreParams.HTTP.FetchInterval = fetchInterval
	cacheStoreParams.HTTP.FetchTimeout = fetchTimeout

	db, err := p.createDB(ec)
	if err != nil {
		return nil, err
	}

	storagePipeline := stinfluxdb.NewPipeline(ctx, ec.storage.influxdb)
	p.stopper.Add("storage-influxdb-pipeline", storagePipeline)
	p.starter.Add(storagePipeline)

	cacheStore := devstore.NewCacheStore(
		ctx,
		p.systemClock,
		storagePipeline.GetSystemClock(),
		storagePipeline.GetDataHandler(),
		db,
		resolveStore,
		cacheStoreParams,
	)
	p.stopper.Add("device-cache-store", cacheStore)
	p.starter.Add(cacheStore)

	return cacheStore, nil
}

func (p *appPipeline) createDB(ec *envContext) (stcore.DB, error) {
	if ec.cacheDir == "" {
		return &stcore.NoopDB{}, nil
	}

	ec.cacheDir = path.Join(ec.cacheDir, "bbolt.db")

	bboltDB, err := stcore.NewBboltDB(ec.cacheDir, &bbolt.Options{
		Timeout: time.Second * 5,
	})
	if err != nil {
		return nil, err
	}

	p.stopper.Add("bbolt-database", syssched.FuncStopper(func() error {
		return bboltDB.Close()
	}))

	return stcore.NewBboltDBBucket(bboltDB, "device_bucket"), nil
}

func (p *appPipeline) configureMdnsServer(server *htcore.Server, ec *envContext) error {
	services := []*sysmdns.Service{
		{
			Instance:   ec.mdns.server.instanceName,
			Name:       sysmdns.ServiceName(sysmdns.ServiceTypeHTTP, sysmdns.ProtoTCP),
			Hostname:   ec.mdns.server.hostname,
			Port:       server.Port(),
			TxtRecords: []string{"api=/api/v1"},
		},
	}

	filteredIfaces, err := parseIfaceOption(ec.mdns.server.iface)
	if err != nil {
		return err
	}

	zeroconfServer := sysmdns.NewZeroconfServer(services, filteredIfaces)
	p.stopper.Add("mdns-server", zeroconfServer)
	p.starter.Add(zeroconfServer)

	return nil
}

func parseIfaceOption(opt string) ([]net.Interface, error) {
	var allowedIfaces []string

	if opt != "" {
		allowedIfaces = strings.Split(opt, ",")
		if len(allowedIfaces) < 1 {
			return nil, errors.New("mDNS network interface list has invalid format")
		}
	}

	filteredIfaces, err := sysnet.FilterInterfaces(func(iface net.Interface) bool {
		if iface.Flags&net.FlagMulticast == 0 {
			return false
		}

		if allowedIfaces == nil {
			return true
		}

		for _, allowedIface := range allowedIfaces {
			if allowedIface == iface.Name {
				return true
			}
		}

		return false
	})
	if err != nil {
		return nil, err
	}

	return filteredIfaces, nil
}

func registerHTTPRoutes(
	mux *http.ServeMux,
	timeHandler http.Handler,
	storeHTTPHandler *devstore.StoreHTTPHandler,
) {
	mux.Handle("/api/v1/system/time", timeHandler)

	mux.HandleFunc("/api/v1/device/add", storeHTTPHandler.HandleAdd)
	mux.HandleFunc("/api/v1/device/remove", storeHTTPHandler.HandleRemove)
	mux.HandleFunc("/api/v1/device/list", storeHTTPHandler.HandleList)
}

func newAppPipeline() *appPipeline {
	return &appPipeline{
		systemClock: &syscore.LocalSystemClock{},
		stopper:     &syssched.FanoutStopper{},
		starter:     &syssched.FanoutStarter{},
	}
}

func prepareEnvironment(ec *envContext) error {
	if ec.storage.influxdb.URL == "" {
		return fmt.Errorf("influxdb URL is required")
	}
	if ec.storage.influxdb.Org == "" {
		return fmt.Errorf("influxdb org is required")
	}
	if ec.storage.influxdb.Bucket == "" {
		return fmt.Errorf("influxdb bucket is required")
	}
	if ec.storage.influxdb.Token == "" {
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
	if err := syscore.SetLogFile(ec.logPath); err != nil {
		return err
	}

	if !ec.mdns.server.disable {
		if ec.mdns.server.hostname == "" {
			return errors.New("mDNS server hostname can't be empty")
		}
		if ec.mdns.server.instanceName == "" {
			return errors.New("mDNS server instance name can't be empty")
		}
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

			return appPipeline.stop()
		},
	}

	cmd.Flags().IntVar(&envContext.port, "http-port", 0,
		"HTTP server port (0 for random port)")

	cmd.Flags().StringVar(&envContext.cacheDir, "cache-dir", "", "device-hub cache directory")
	cmd.Flags().StringVar(&envContext.logPath, "log-path", "", "device-hub log file path")

	cmd.Flags().StringVar(&envContext.storage.influxdb.URL, "storage-influxdb-url", "",
		"influxdb URL")
	cmd.Flags().StringVar(&envContext.storage.influxdb.Org, "storage-influxdb-org", "",
		"influxdb Org")
	cmd.Flags().StringVar(&envContext.storage.influxdb.Token, "storage-influxdb-api-token", "",
		"influxdb API token")
	cmd.Flags().StringVar(&envContext.storage.influxdb.Bucket, "storage-influxdb-bucket", "",
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
		&envContext.mdns.browse.interval,
		"mdns-browse-interval", "40s",
		"How often to perform mDNS lookup over local network",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.browse.timeout,
		"mdns-browse-timeout", "10s",
		"How long to perform a single mDNS lookup over local network",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.browse.iface,
		"mdns-browse-iface", "",
		"Comma-separated list of network interfaces for the mDNS lookup"+
			" (empty for all interfaces)",
	)

	cmd.Flags().BoolVar(
		&envContext.mdns.autodiscovery.disable,
		"mdns-autodiscovery-disable", false,
		"Disable automatic device discovery on the local network",
	)

	cmd.Flags().BoolVar(
		&envContext.mdns.server.disable,
		"mdns-server-disable", false,
		"Disable mDNS server",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.server.hostname,
		"mdns-server-hostname", "device-hub",
		"mDNS server hostname",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.server.instanceName,
		"mdns-server-instance", "Device Hub Software",
		"mDNS server instance name",
	)

	cmd.Flags().StringVar(
		&envContext.mdns.server.iface,
		"mdns-server-iface", "",
		"Comma-separated list of network interfaces for the mDNS server"+
			" (empty for all interfaces)",
	)

	if err := cmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
