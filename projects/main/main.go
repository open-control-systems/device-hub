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
)

type envContext struct {
	dbParams stinfluxdb.DbParams
	cacheDir string
	port     int
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

	deviceStoreParams := pipdevice.StoreParams{}
	deviceStoreParams.HTTP.FetchInterval = time.Second * 5
	deviceStoreParams.HTTP.FetchTimeout = time.Second * 5

	deviceStore := pipdevice.NewStore(
		appContext,
		p.systemClock,
		storagePipeline.GetSystemClock(),
		storagePipeline.GetDataHandler(),
		db,
		deviceStoreParams,
	)
	p.closer.Add("device-pipeline-store", deviceStore)

	registerHTTPRoutes(
		serverPipeline.GetServeMux(),
		// Time valid since 2024/12/03.
		piphttp.NewSystemTimeHandler(p.systemClock, time.Unix(1733215816, 0)),
		pipdevice.NewStoreHTTPHandler(deviceStore),
	)

	deviceStore.Start()
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
	dbParams := stinfluxdb.DbParams{
		URL:    os.Getenv("INFLUXDB_URL"),
		Org:    os.Getenv("INFLUXDB_ORG"),
		Bucket: os.Getenv("INFLUXDB_BUCKET"),
		Token:  os.Getenv("INFLUXDB_API_TOKEN"),
	}

	if dbParams.URL == "" {
		return fmt.Errorf("environment variable INFLUXDB_URL is required")
	}
	if dbParams.Org == "" {
		return fmt.Errorf("environment variable INFLUXDB_ORG is required")
	}
	if dbParams.Bucket == "" {
		return fmt.Errorf("environment variable INFLUXDB_BUCKET is required")
	}
	if dbParams.Token == "" {
		return fmt.Errorf("environment variable INFLUXDB_API_TOKEN is required")
	}

	logPath := os.Getenv("DEVICE_HUB_LOG_PATH")
	if logPath == "" {
		return fmt.Errorf("environment variable DEVICE_HUB_LOG_PATH is required")
	}
	if err := core.SetLogFile(logPath); err != nil {
		return err
	}

	if ec.cacheDir == "" {
		ec.cacheDir = os.Getenv("DEVICE_HUB_CACHE_DIR")
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

	ec.dbParams = dbParams

	return nil
}

func main() {
	appPipeline := newAppPipeline()
	envContext := &envContext{}

	cmd := &cobra.Command{
		Use:   "device-hub",
		Short: "device-hub CLI",
		Long: `device-hub collects and stores various data from IoT devices.

Required environment variables:
- INFLUXDB_URL
- INFLUXDB_ORG
- INFLUXDB_BUCKET
- INFLUXDB_API_TOKEN

- DEVICE_HUB_LOG_PATH`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return prepareEnvironment(envContext)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return appPipeline.start(envContext)
		},
	}

	cmd.Flags().IntVar(&envContext.port, "port", 0, "HTTP server port (0 for random port)")
	cmd.Flags().StringVar(&envContext.cacheDir, "cache-dir", "", "device-hub cache directory")

	if err := cmd.Execute(); err != nil {
		core.LogErr.Printf("main: failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
