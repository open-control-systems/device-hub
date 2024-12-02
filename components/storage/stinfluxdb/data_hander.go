package stinfluxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/open-control-systems/device-hub/components/device"
)

// Store incoming data in influxDB.
//
// References:
//   - InfluxDB quick start guide: https://docs.influxdata.com/influxdb/cloud/get-started
//   - InfluxDB Go client library: https://docs.influxdata.com/influxdb/cloud/api-guide/client-libraries/go/
type DataHandler struct {
	ctx         context.Context
	dbClient    influxdb2.Client
	writeClient api.WriteAPIBlocking
}

// Initialize handler.
//
// Parameters:
//   - ctx - parent context.
//   - params - various influxDB configuration parameters.
func NewDataHandler(ctx context.Context, params DbParams) *DataHandler {
	dbClient := influxdb2.NewClient(params.Url, params.Token)
	writeClient := dbClient.WriteAPIBlocking(params.Org, params.Bucket)

	return &DataHandler{
		ctx:         ctx,
		dbClient:    dbClient,
		writeClient: writeClient,
	}
}

// Store telemetry data in influxDB.
func (h *DataHandler) HandleTelemetry(deviceId string, js device.Json) error {
	return h.handleData("telemetry", deviceId, js)
}

// Store registration data in influxDB.
func (h *DataHandler) HandleRegistration(deviceId string, js device.Json) error {
	return h.handleData("registration", deviceId, js)
}

// Stop writing data to the DB.
func (h *DataHandler) Close() error {
	h.dbClient.Close()

	return nil
}

func (h *DataHandler) handleData(dataId string, deviceId string, js device.Json) error {
	ts, ok := js["timestamp"]
	if !ok {
		return fmt.Errorf("influxdb-data-handler: missed timestamp field")
	}

	timestamp, ok := ts.(float64)
	if !ok {
		return fmt.Errorf("influxdb-data-handler: invalid type for timestamp")
	}

	unixTimestamp := time.Unix(int64(timestamp), 0)

	point := influxdb2.NewPoint(dataId,
		map[string]string{"device_id": deviceId},
		js,
		unixTimestamp)

	if err := h.writeClient.WritePoint(h.ctx, point); err != nil {
		return fmt.Errorf("influxdb-data-handler: failed to write to DB: %w", err)
	}

	return nil
}
