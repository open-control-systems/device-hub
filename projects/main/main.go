package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/storage/stinfluxdb"
)

func main() {
	if err := core.SetLogFile(os.Getenv("DEVICE_HUB_LOG_PATH")); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to setup log file: ", err)
	}

	appContext, cancelFunc := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	defer cancelFunc()

	fanoutCloser := &core.FanoutCloser{}
	defer fanoutCloser.Close()

	devicePipeline := stinfluxdb.NewHttpPipeline(
		appContext,
		fanoutCloser,
		stinfluxdb.HttpPipelineParams{
			DbParams: stinfluxdb.DbParams{
				Url:    os.Getenv("INFLUXDB_URL"),
				Org:    os.Getenv("INFLUXDB_ORG"),
				Bucket: os.Getenv("INFLUXDB_BUCKET"),
				Token:  os.Getenv("INFLUXDB_API_TOKEN"),
			},
			BaseUrl:       os.Getenv("DEVICE_HUB_API_BASE_URL"),
			FetchInterval: time.Second * 5,
			FetchTimeout:  time.Second * 10,
		})
	devicePipeline.Start()

	<-appContext.Done()
}
