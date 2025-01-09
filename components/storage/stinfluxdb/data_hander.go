package stinfluxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/open-control-systems/device-hub/components/device"
)

// dataHandler stores incoming data in influxDB.
//
// References:
//   - https://docs.influxdata.com/influxdb/cloud/get-started
//   - https://docs.influxdata.com/influxdb/cloud/api-guide/client-libraries/go/
type dataHandler struct {
	ctx         context.Context
	dbClient    influxdb2.Client
	writeClient api.WriteAPIBlocking
}

// newDataHandler initializes influxDB handler.
//
// Parameters:
//   - ctx - parent context.
//   - params - various influxDB configuration parameters.
func newDataHandler(ctx context.Context, params DbParams) *dataHandler {
	dbClient := influxdb2.NewClient(params.URL, params.Token)
	writeClient := dbClient.WriteAPIBlocking(params.Org, params.Bucket)

	return &dataHandler{
		ctx:         ctx,
		dbClient:    dbClient,
		writeClient: writeClient,
	}
}

// HandleTelemetry stores telemetry data in influxDB.
func (h *dataHandler) HandleTelemetry(deviceID string, js device.JSON) error {
	return h.handleData("telemetry", deviceID, js)
}

// HandleRegistration stores registration data in influxDB.
func (h *dataHandler) HandleRegistration(deviceID string, js device.JSON) error {
	return h.handleData("registration", deviceID, js)
}

// Close stops writing data to the DB.
func (h *dataHandler) Close() error {
	h.dbClient.Close()

	return nil
}

func (h *dataHandler) handleData(dataID string, deviceID string, js device.JSON) error {
	ts, ok := js["timestamp"]
	if !ok {
		return fmt.Errorf("influxdb-data-handler: missed timestamp field")
	}

	timestamp, ok := ts.(float64)
	if !ok {
		return fmt.Errorf("influxdb-data-handler: invalid type for timestamp")
	}

	unixTimestamp := time.Unix(int64(timestamp), 0)

	point := influxdb2.NewPoint(dataID,
		map[string]string{"device_id": deviceID},
		js,
		unixTimestamp)

	if err := h.writeClient.WritePoint(h.ctx, point); err != nil {
		return fmt.Errorf("influxdb-data-handler: failed to write to DB: %w", err)
	}

	return nil
}
