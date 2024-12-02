package stinfluxdb

import (
	"context"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/device"
	"github.com/open-control-systems/device-hub/components/http/htclient"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
)

// Fetch device data over HTTP and store it in the influxDB database.
type HttpPipeline struct {
	dbParams      DbParams
	fetchInterval time.Duration
	ctx           context.Context
	dataHandler   device.DataHandler
	device        device.Device
	doneCh        chan struct{}
}

type HttpPipelineParams struct {
	// Various InfluxDB parameters.
	DbParams DbParams

	// Device API base URL.
	BaseUrl string

	// How often to fetch data from the device.
	FetchInterval time.Duration

	// How long to wait for the response from the device.
	FetchTimeout time.Duration
}

// Initialize HTTP pipeline.
//
// Parameters:
//   - ctx - parent context.
//   - closer - to register all resources that should be closed.
//   - params - various pipeline parameters.
func NewHttpPipeline(
	ctx context.Context,
	closer *core.FanoutCloser,
	params HttpPipelineParams,
) *HttpPipeline {
	dataHandler := NewDataHandler(ctx, params.DbParams)
	closer.Add("influxdb-data-handler", dataHandler)

	resolver := &sysnet.PionMdnsResolver{}
	closer.Add("pion-mdns-resolver", resolver)

	pollDevice := device.NewPollDevice(
		htclient.NewUrlFetcher(
			ctx,
			htclient.NewResolveClient(resolver),
			params.BaseUrl+"/registration",
			params.FetchTimeout,
		),
		htclient.NewUrlFetcher(
			ctx,
			htclient.NewResolveClient(resolver),
			params.BaseUrl+"/telemetry",
			params.FetchTimeout,
		),
		dataHandler,
	)

	pipeline := &HttpPipeline{
		dbParams:      params.DbParams,
		fetchInterval: params.FetchInterval,
		ctx:           ctx,
		device:        pollDevice,
		doneCh:        make(chan struct{}),
	}
	closer.Add("influxdb-http-pipeline", pipeline)

	return pipeline
}

// Start asynchronous data processing.
func (p *HttpPipeline) Start() {
	core.LogInf.Printf("influxdb-http-pipeline: starting: url=%s org=%s bucket=%s\n",
		p.dbParams.Url, p.dbParams.Org, p.dbParams.Bucket)

	go p.run()
}

// Stop device data processing.
func (p *HttpPipeline) Close() error {
	core.LogInf.Println("influxdb-http-pipeline: stopping")
	<-p.doneCh
	core.LogInf.Println("influxdb-http-pipeline: stopped")

	return nil
}

func (p *HttpPipeline) run() {
	ticker := time.NewTicker(p.fetchInterval)
	defer ticker.Stop()
	defer close(p.doneCh)

	for {
		select {
		case <-ticker.C:
			if err := p.device.Update(); err != nil {
				core.LogErr.Printf(
					"influxdb-http-pipeline: failed to handle device data: %v\n", err)
			}

		case <-p.ctx.Done():
			return
		}
	}
}