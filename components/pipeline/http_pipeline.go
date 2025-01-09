package pipeline

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
	id            string
	fetchInterval time.Duration
	ctx           context.Context
	task          syssched.Task
	doneCh        chan struct{}
}

// HTTPPipelineParams provides various configuration options for HttpPipeline.
type HTTPPipelineParams struct {
	// ID - unique pipeline identifier, to distinguish one pipeline from another.
	ID string

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
//   - dataHandler - to handle device data.
//   - params - various pipeline parameters.
func NewHTTPPipeline(
	ctx context.Context,
	closer *core.FanoutCloser,
	dataHandler device.DataHandler,
	params HTTPPipelineParams,
) *HTTPPipeline {
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
		id:            params.ID,
		fetchInterval: params.FetchInterval,
		ctx:           ctx,
		task:          pollDevice,
		doneCh:        make(chan struct{}),
	}
	closer.Add(params.ID, pipeline)

	return pipeline
}

// Start begins asynchronous data processing.
func (p *HTTPPipeline) Start() {
	go p.run()
}

// Close ends device data processing.
func (p *HTTPPipeline) Close() error {
	core.LogInf.Printf(p.id, ": stopping")
	<-p.doneCh
	core.LogInf.Println(p.id, ": stopped")

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
				core.LogErr.Printf("%s: failed to handle device data: %v\n", p.id, err)
			}

		case <-p.ctx.Done():
			return
		}
	}
}
