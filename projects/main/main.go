package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/pipeline/pipdevice"
	"github.com/open-control-systems/device-hub/components/storage/stinfluxdb"
)

func main() {
	if err := core.SetLogFile(os.Getenv("DEVICE_HUB_LOG_PATH")); err != nil {
		core.LogErr.Println("main: failed to setup log file: ", err)
	}

	appContext, cancelFunc := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancelFunc()

	fanoutCloser := &core.FanoutCloser{}
	defer func() {
		if err := fanoutCloser.Close(); err != nil {
			core.LogErr.Println("main: failed to close resources: ", err)
		}
	}()

	storagePipeline := stinfluxdb.NewPipeline(
		appContext,
		fanoutCloser,
		stinfluxdb.DbParams{
			URL:    os.Getenv("INFLUXDB_URL"),
			Org:    os.Getenv("INFLUXDB_ORG"),
			Bucket: os.Getenv("INFLUXDB_BUCKET"),
			Token:  os.Getenv("INFLUXDB_API_TOKEN"),
		})

	devicePipeline := pipdevice.NewHTTPPipeline(
		appContext,
		fanoutCloser,
		storagePipeline.GetDataHandler(),
		storagePipeline.GetSystemClock(),
		pipdevice.HTTPPipelineParams{
			ID:            "influxdb-http-pipeline",
			BaseURL:       os.Getenv("DEVICE_HUB_API_BASE_URL"),
			FetchInterval: time.Second * 5,
			FetchTimeout:  time.Second * 10,
		})

	storagePipeline.Start()
	devicePipeline.Start()

	<-appContext.Done()
}
