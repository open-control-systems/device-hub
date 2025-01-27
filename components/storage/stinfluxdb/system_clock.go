package stinfluxdb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

type systemClock struct {
	restoreUpdateInterval time.Duration
	params                DBParams
	ctx                   context.Context
	client                api.QueryAPI

	doneCh chan struct{}

	mu        sync.Mutex
	restored  bool
	timestamp int64
}

func newSystemClock(
	ctx context.Context,
	client api.QueryAPI,
	restoreUpdateInterval time.Duration,
	params DBParams,
) *systemClock {
	return &systemClock{
		restoreUpdateInterval: restoreUpdateInterval,
		params:                params,
		ctx:                   ctx,
		client:                client,
		doneCh:                make(chan struct{}),
		timestamp:             int64(-1),
	}
}

func (c *systemClock) run() {
	defer close(c.doneCh)

	ticker := time.NewTicker(c.restoreUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return

		case <-ticker.C:
			if c.tryRestoreTimestamp() {
				return
			}
		}
	}
}

// Stop ends asynchronous time restoring.
func (c *systemClock) Stop() error {
	<-c.doneCh
	return nil
}

// SetTimestamp sets the most recent UNIX time.
func (c *systemClock) SetTimestamp(timestamp int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if timestamp > c.timestamp {
		c.timestamp = timestamp
	}

	if !c.restored {
		c.restored = true

		syscore.LogInf.Printf("influxdb-system-clock: skip timestamp restoring: value=%v\n",
			timestamp)
	}

	return nil
}

// GetTimestamp returns the most recent UNIX time.
func (c *systemClock) GetTimestamp() (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.restored {
		return -1, status.StatusInvalidState
	}

	return c.timestamp, nil
}

func (c *systemClock) tryRestoreTimestamp() bool {
	timestamp, err := c.readTimestamp()
	if err != nil {
		syscore.LogErr.Printf(
			"influxdb-system-clock: failed to restore timestamp: err=%v\n", err)

		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.restored {
		syscore.LogInf.Printf(
			"influxdb-system-clock: timestamp already restored: restored=%v persisted=%v\n",
			c.timestamp, timestamp)
	} else {
		c.restored = true
		c.timestamp = timestamp

		syscore.LogInf.Printf("influxdb-system-clock: timestamp restored: value=%v\n",
			c.timestamp)
	}

	return true
}

func (c *systemClock) readTimestamp() (int64, error) {
	query := fmt.Sprintf(`
	from(bucket: "%s")
	  |> range(start: -30d)
	  |> filter(fn: (r) => r["_measurement"] == "%s")
	  |> aggregateWindow(every: 10m, fn: last, createEmpty: false)
	  |> keep(columns: ["_time"])
	  |> sort(columns: ["_time"], desc: true)
	  |> limit(n: 1)`, c.params.Bucket, "telemetry")

	result, err := c.client.Query(c.ctx, query)
	if err != nil {
		return -1, fmt.Errorf("influxdb: failed to query: %w", err)
	}
	defer result.Close()

	if result.Err() != nil {
		return -1, result.Err()
	}

	if !result.Next() {
		if result.Err() != nil {
			return -1, fmt.Errorf("influxdb: query error: %w", result.Err())
		}

		return -1, fmt.Errorf("influxdb: no records found in query result")
	}

	record := result.Record()
	if record == nil {
		return -1, fmt.Errorf("influxdb: no valid record returned")
	}

	return record.Time().Unix(), nil
}
