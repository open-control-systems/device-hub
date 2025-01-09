package stinfluxdb

import (
	"context"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/device"
	"github.com/open-control-systems/device-hub/components/http/htclient"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

// HTTPPipeline fetches device data over HTTP and store it in the influxDB database.
type HTTPPipeline struct {
	dbParams      DbParams
	fetchInterval time.Duration
	ctx           context.Context
	task          syssched.Task
	doneCh        chan struct{}
}

// HTTPPipelineParams provides various configuration options for HttpPipeline.
type HTTPPipelineParams struct {
	// DbParams provides various configuration options for influxDB.
	DbParams DbParams

	// BaseURL - device API base URL.
	BaseURL string

	// FetchInterval - how often to fetch data from the device.
	FetchInterval time.Duration

	// FetchTimeout - how long to wait for the response from the device.
	FetchTimeout time.Duration
}

// NewHTTPPipeline initializes HTTP pipeline.
//
// Parameters:
//   - ctx - parent context.
//   - closer - to register all resources that should be closed.
//   - params - various pipeline parameters.
func NewHTTPPipeline(
	ctx context.Context,
	closer *core.FanoutCloser,
	params HTTPPipelineParams,
) *HTTPPipeline {
	dataHandler := newDataHandler(ctx, params.DbParams)
	closer.Add("influxdb-data-handler", dataHandler)

	resolver := &sysnet.PionMdnsResolver{}
	closer.Add("pion-mdns-resolver", resolver)

	pollDevice := device.NewPollDevice(
		htclient.NewURLFetcher(
			ctx,
			htclient.NewResolveClient(resolver),
			params.BaseURL+"/registration",
			params.FetchTimeout,
		),
		htclient.NewURLFetcher(
			ctx,
			htclient.NewResolveClient(resolver),
			params.BaseURL+"/telemetry",
			params.FetchTimeout,
		),
		dataHandler,
	)

	pipeline := &HTTPPipeline{
		dbParams:      params.DbParams,
		fetchInterval: params.FetchInterval,
		ctx:           ctx,
		task:          pollDevice,
		doneCh:        make(chan struct{}),
	}
	closer.Add("influxdb-http-pipeline", pipeline)

	return pipeline
}

// Start begins asynchronous data processing.
func (p *HTTPPipeline) Start() {
	core.LogInf.Printf("influxdb-http-pipeline: starting: url=%s org=%s bucket=%s\n",
		p.dbParams.URL, p.dbParams.Org, p.dbParams.Bucket)

	go p.run()
}

// Close ends device data processing.
func (p *HTTPPipeline) Close() error {
	core.LogInf.Println("influxdb-http-pipeline: stopping")
	<-p.doneCh
	core.LogInf.Println("influxdb-http-pipeline: stopped")

	return nil
}

func (p *HTTPPipeline) run() {
	ticker := time.NewTicker(p.fetchInterval)
	defer ticker.Stop()
	defer close(p.doneCh)

	for {
		select {
		case <-ticker.C:
			if err := p.task.Run(); err != nil {
				core.LogErr.Printf(
					"influxdb-http-pipeline: failed to handle device data: %v\n", err)
			}

		case <-p.ctx.Done():
			return
		}
	}
}
