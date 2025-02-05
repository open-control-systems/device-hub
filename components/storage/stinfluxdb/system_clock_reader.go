package stinfluxdb

import (
	"context"
	"fmt"

	"github.com/influxdata/influxdb-client-go/v2/api"
)

// SystemClockReader reads the UNIX timestamp from the influxdb.
type SystemClockReader struct {
	bucket string
	client api.QueryAPI
}

// NewSystemClockReader is an initialization of SystemClockReader.
func NewSystemClockReader(client api.QueryAPI, bucket string) *SystemClockReader {
	return &SystemClockReader{
		bucket: bucket,
		client: client,
	}
}

// ReadTimestamp reads the most recent UNIX timestamp from the influxdb.
func (r *SystemClockReader) ReadTimestamp(ctx context.Context) (int64, error) {
	query := fmt.Sprintf(`
	from(bucket: "%s")
	  |> range(start: -30d)
	  |> filter(fn: (r) => r["_measurement"] == "%s")
	  |> aggregateWindow(every: 10m, fn: last, createEmpty: false)
	  |> keep(columns: ["_time"])
	  |> sort(columns: ["_time"], desc: true)
	  |> limit(n: 1)`, r.bucket, "telemetry")

	result, err := r.client.Query(ctx, query)
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
