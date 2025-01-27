package stinfluxdb

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

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
//   - params - various influxDB configuration parameters.
func NewPipeline(
	ctx context.Context,
	params DBParams,
) *Pipeline {
	dbClient := influxdb2.NewClient(params.URL, params.Token)
	writeClient := dbClient.WriteAPIBlocking(params.Org, params.Bucket)
	queryClient := dbClient.QueryAPI(params.Org)

	clock := newSystemClock(ctx, queryClient, time.Second*5, params)

	return &Pipeline{
		dbClient: dbClient,
		clock:    clock,
		handler:  NewDataHandler(ctx, clock, writeClient),
	}
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

// Stop stops writing data to the DB.
func (p *Pipeline) Stop() error {
	p.dbClient.Close()

	return p.clock.Stop()
}
