package stinfluxdb

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/open-control-systems/device-hub/components/system/syscore"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

// DBParams provides various configuration options for influxDB.
type DBParams struct {
	URL    string
	Org    string
	Token  string
	Bucket string
}

// Pipeline contains various building blocks for persisting data in influxdb.
type Pipeline struct {
	dbClient influxdb2.Client
	restorer *stcore.SystemClockRestorer
	runner   *syssched.AsyncTaskRunner
	handler  *DataHandler
}

// NewPipeline initializes all components associated with the influxdb subsystem.
//
// Parameters:
//   - ctx - parent context.
//   - params - various influxDB configuration parameters.
func NewPipeline(ctx context.Context, params DBParams) *Pipeline {
	dbClient := influxdb2.NewClient(params.URL, params.Token)
	writeClient := dbClient.WriteAPIBlocking(params.Org, params.Bucket)
	queryClient := dbClient.QueryAPI(params.Org)

	reader := NewSystemClockReader(queryClient, params.Bucket)
	restorer := stcore.NewSystemClockRestorer(ctx, reader)
	runner := syssched.NewAsyncTaskRunner(
		ctx,
		restorer,
		restorer,
		syssched.AsyncTaskRunnerParams{
			UpdateInterval: time.Second * 5,
			ExitOnSuccess:  true,
		},
	)

	return &Pipeline{
		dbClient: dbClient,
		restorer: restorer,
		runner:   runner,
		handler:  NewDataHandler(ctx, restorer, writeClient),
	}
}

// GetDataHandler returns the underlying influxdb data handler.
func (p *Pipeline) GetDataHandler() *DataHandler {
	return p.handler
}

// GetSystemClock returns the clock to get last persisted UNIX time.
func (p *Pipeline) GetSystemClock() syscore.SystemClock {
	return p.restorer
}

// Start starts the asynchronous UNIX time restoring.
func (p *Pipeline) Start() error {
	return p.runner.Start()
}

// Stop stops writing data to the DB.
func (p *Pipeline) Stop() error {
	p.dbClient.Close()

	return p.runner.Stop()
}
