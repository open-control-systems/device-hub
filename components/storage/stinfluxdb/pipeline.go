package stinfluxdb

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// Pipeline contains various building blocks for persisting data in influxdb.
type Pipeline struct {
	dbClient influxdb2.Client
	clock    *systemClock
	handler  *DataHandler
}

// NewPipeline initializes all components associated with the influxdb subsystem.
//
// Parameters:
//   - ctx - parent context.
//   - closer - to register the handler for the underlying resource deallocation.
//   - params - various influxDB configuration parameters.
func NewPipeline(
	ctx context.Context,
	closer *core.FanoutCloser,
	params DbParams,
) *Pipeline {
	dbClient := influxdb2.NewClient(params.URL, params.Token)
	writeClient := dbClient.WriteAPIBlocking(params.Org, params.Bucket)
	queryClient := dbClient.QueryAPI(params.Org)

	clock := newSystemClock(ctx, queryClient, time.Second*5, params)
	closer.Add("influxdb-system-clock", clock)

	pipeline := &Pipeline{
		dbClient: dbClient,
		clock:    clock,
		handler:  NewDataHandler(ctx, clock, writeClient),
	}

	closer.Add("influxdb-pipeline", pipeline)

	return pipeline
}

// GetDataHandler returns the underlying influxdb data handler.
func (p *Pipeline) GetDataHandler() *DataHandler {
	return p.handler
}

// GetSystemClock returns the clock to get last persisted UNIX time.
func (p *Pipeline) GetSystemClock() syscore.SystemClock {
	return p.clock
}

// Start starts the asynchronous UNIX time restoring.
func (p *Pipeline) Start() {
	go p.clock.run()
}

// Close stops writing data to the DB.
func (p *Pipeline) Close() error {
	p.dbClient.Close()

	return nil
}
