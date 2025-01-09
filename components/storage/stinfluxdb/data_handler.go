package stinfluxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/device"
)

// DataHandler stores incoming data in influxDB.
//
// References:
//   - https://docs.influxdata.com/influxdb/cloud/get-started
//   - https://docs.influxdata.com/influxdb/cloud/api-guide/client-libraries/go/
type DataHandler struct {
	ctx         context.Context
	dbClient    influxdb2.Client
	writeClient api.WriteAPIBlocking
}

// NewDataHandler initializes influxDB handler.
//
// Parameters:
//   - ctx - parent context.
//   - closer - to register the handler for the underlying resource deallocation.
//   - params - various influxDB configuration parameters.
func NewDataHandler(
	ctx context.Context,
	closer *core.FanoutCloser,
	params DbParams,
) *DataHandler {
	dbClient := influxdb2.NewClient(params.URL, params.Token)
	writeClient := dbClient.WriteAPIBlocking(params.Org, params.Bucket)

	handler := &DataHandler{
		ctx:         ctx,
		dbClient:    dbClient,
		writeClient: writeClient,
	}

	closer.Add("influxdb-data-handler", handler)

	return handler
}

// HandleTelemetry stores telemetry data in influxDB.
func (h *DataHandler) HandleTelemetry(deviceID string, js device.JSON) error {
	return h.handleData("telemetry", deviceID, js)
}

// HandleRegistration stores registration data in influxDB.
func (h *DataHandler) HandleRegistration(deviceID string, js device.JSON) error {
	return h.handleData("registration", deviceID, js)
}

// Close stops writing data to the DB.
func (h *DataHandler) Close() error {
	h.dbClient.Close()

	return nil
}

func (h *DataHandler) handleData(dataID string, deviceID string, js device.JSON) error {
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
