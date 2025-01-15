package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/http/htcore"
	"github.com/open-control-systems/device-hub/components/pipeline/pipdevice"
	"github.com/open-control-systems/device-hub/components/pipeline/piphttp"
	"github.com/open-control-systems/device-hub/components/storage/stinfluxdb"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

type envContext struct {
	dbParams stinfluxdb.DbParams
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
		p.systemClock,
		htcore.ServerParams{
			Port: ec.port,
		},
	)
	if err != nil {
		return err
	}

	storagePipeline := stinfluxdb.NewPipeline(appContext, p.closer, ec.dbParams)

	deviceStore := pipdevice.NewPipelineStore(
		appContext,
		p.systemClock,
		storagePipeline.GetSystemClock(),
		storagePipeline.GetDataHandler(),
	)
	p.closer.Add("device-pipeline-store", deviceStore)

	deviceHandler := piphttp.NewDeviceHandler(deviceStore)

	registerHTTPRoutes(serverPipeline.GetServeMux(), deviceHandler)

	storagePipeline.Start()
	serverPipeline.Start()

	<-appContext.Done()

	return nil
}

func registerHTTPRoutes(
	mux *http.ServeMux,
	deviceHandler *piphttp.DeviceHandler,
) {
	mux.HandleFunc("/api/v1/device/add", func(w http.ResponseWriter, r *http.Request) {
		deviceHandler.HandleAdd(w, r)
	})
	mux.HandleFunc("/api/v1/device/remove", func(w http.ResponseWriter, r *http.Request) {
		deviceHandler.HandleRemove(w, r)
	})
	mux.HandleFunc("/api/v1/device/list", func(w http.ResponseWriter, r *http.Request) {
		deviceHandler.HandleList(w, r)
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

	if err := cmd.Execute(); err != nil {
		core.LogErr.Printf("main: failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
