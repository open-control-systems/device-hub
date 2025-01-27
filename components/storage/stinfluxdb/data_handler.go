package stinfluxdb

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/open-control-systems/device-hub/components/device/devcore"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// DataHandler stores incoming data in influxDB.
//
// References:
//   - https://docs.influxdata.com/influxdb/cloud/get-started
//   - https://docs.influxdata.com/influxdb/cloud/api-guide/client-libraries/go/
type DataHandler struct {
	ctx    context.Context
	clock  syscore.SystemClock
	client api.WriteAPIBlocking
}

// NewDataHandler initializes influxDB handler.
//
// Parameters:
//   - ctx - parent context.
//   - clock to update the most recent UNIX time.
//   - client to write data to the influxdb.
func NewDataHandler(
	ctx context.Context,
	clock syscore.SystemClock,
	client api.WriteAPIBlocking,
) *DataHandler {
	return &DataHandler{
		ctx:    ctx,
		clock:  clock,
		client: client,
	}
}

// HandleTelemetry stores telemetry data in influxDB.
func (h *DataHandler) HandleTelemetry(deviceID string, js devcore.JSON) error {
	return h.handleData("telemetry", deviceID, js)
}

// HandleRegistration stores registration data in influxDB.
func (h *DataHandler) HandleRegistration(deviceID string, js devcore.JSON) error {
	return h.handleData("registration", deviceID, js)
}

func (h *DataHandler) handleData(dataID string, deviceID string, js devcore.JSON) error {
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

	if err := h.client.WritePoint(h.ctx, point); err != nil {
		return fmt.Errorf("influxdb-data-handler: failed to write to DB: %w", err)
	}

	return h.clock.SetTimestamp(unixTimestamp.Unix())
}
