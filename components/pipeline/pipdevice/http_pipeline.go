package pipdevice

import (
	"context"
	"strings"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/device"
	"github.com/open-control-systems/device-hub/components/http/htcore"
	"github.com/open-control-systems/device-hub/components/pipeline/piphttp"
	"github.com/open-control-systems/device-hub/components/system/syscore"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

// HTTPPipeline fetches device data over HTTP.
type HTTPPipeline struct {
	baseURL       string
	desc          string
	fetchInterval time.Duration
	ctx           context.Context
	task          syssched.Task
	doneCh        chan struct{}
	holder        *device.IDHolder
	errorReporter device.ErrorReporter
}

// HTTPPipelineParams provides various configuration options for HTTPPipeline.
type HTTPPipelineParams struct {
	// BaseURL - device API base URL.
	BaseURL string

	// Desc is the human readable device description.
	Desc string

	// FetchInterval - how often to fetch data from the device.
	FetchInterval time.Duration

	// FetchTimeout - how long to wait for the response from the device.
	FetchTimeout time.Duration
}

// NewHTTPPipeline initializes HTTP pipeline.
//
// Parameters:
//   - ctx - parent context.
//   - closer to register all resources that should be closed.
//   - dataHandler to handle device data.
//   - localClock to handle local UNIX time.
//   - remoteLastClock to get the last persisted UNIX time.
//   - params - various pipeline parameters.
func NewHTTPPipeline(
	ctx context.Context,
	closer *core.FanoutCloser,
	dataHandler device.DataHandler,
	errorReporter device.ErrorReporter,
	localClock syscore.SystemClock,
	remoteLastClock syscore.SystemClock,
	params HTTPPipelineParams,
) *HTTPPipeline {
	var resolver sysnet.Resolver

	if strings.Contains(params.BaseURL, ".local") {
		mdnsResolver := &sysnet.PionMdnsResolver{}
		closer.Add("pion-mdns-resolver", mdnsResolver)

		resolver = mdnsResolver
	}

	makeHTTPClient := func(r sysnet.Resolver) *htcore.HTTPClient {
		if r != nil {
			return htcore.NewResolveClient(r)
		}

		return htcore.NewDefaultClient()
	}

	remoteCurrClock := piphttp.NewSystemClock(
		ctx,
		makeHTTPClient(resolver),
		params.BaseURL+"/system/time",
		params.FetchTimeout,
	)

	clockSynchronizer := syscore.NewSystemClockSynchronizer(
		localClock, remoteLastClock, remoteCurrClock)

	holder := device.NewIDHolder(dataHandler)

	pollDevice := device.NewPollDevice(
		htcore.NewURLFetcher(
			ctx,
			makeHTTPClient(resolver),
			params.BaseURL+"/registration",
			params.FetchTimeout,
		),
		htcore.NewURLFetcher(
			ctx,
			makeHTTPClient(resolver),
			params.BaseURL+"/telemetry",
			params.FetchTimeout,
		),
		holder,
		clockSynchronizer,
	)

	pipeline := &HTTPPipeline{
		baseURL:       params.BaseURL,
		desc:          params.Desc,
		fetchInterval: params.FetchInterval,
		ctx:           ctx,
		task:          pollDevice,
		doneCh:        make(chan struct{}),
		holder:        holder,
		errorReporter: errorReporter,
	}

	return pipeline
}

// Start begins asynchronous data processing.
func (p *HTTPPipeline) Start() {
	go p.run()
}

// Close ends device data processing.
func (p *HTTPPipeline) Close() error {
	<-p.doneCh

	return nil
}

// GetDeviceID returns the unique identifier of the device.
func (p *HTTPPipeline) GetDeviceID() string {
	return p.holder.Get()
}

func (p *HTTPPipeline) run() {
	defer close(p.doneCh)

	ticker := time.NewTicker(p.fetchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.task.Run(); err != nil {
				p.errorReporter.ReportError(p.baseURL, p.desc, err)
			}

		case <-p.ctx.Done():
			return
		}
	}
}
