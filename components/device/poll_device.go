package device

import (
	"encoding/json"
	"fmt"
)

// Actively fetch telemetry and registration data.
type PollDevice struct {
	registrationFetcher Fetcher
	telemetryFetcher    Fetcher
	dataHandler         DataHandler
	deviceId            string
}

// Initialize polling device.
//
// Parameters:
//   - registrationFetcher to fetch device registration data.
//   - telemetryFetcher to fetch device telemetry data.
//   - dataHandler to handle fetched telemetry and registration data.
func NewPollDevice(
	registrationFetcher Fetcher,
	telemetryFetcher Fetcher,
	dataHandler DataHandler,
) *PollDevice {
	return &PollDevice{
		registrationFetcher: registrationFetcher,
		telemetryFetcher:    telemetryFetcher,
		dataHandler:         dataHandler,
	}
}

// Fetch telemetry and registration data and pass them to the underlying handler.
func (d *PollDevice) Update() error {
	registrationData, err := d.fetchRegistration()
	if err != nil {
		return err
	}

	telemetryData, err := d.fetchTelemetry()
	if err != nil {
		return err
	}

	if err := d.dataHandler.HandleRegistration(d.deviceId, registrationData); err != nil {
		return err
	}

	if err := d.dataHandler.HandleTelemetry(d.deviceId, telemetryData); err != nil {
		return err
	}

	return nil
}

func (d *PollDevice) fetchRegistration() (Json, error) {
	buf, err := d.registrationFetcher.Fetch()
	if err != nil {
		return nil, err
	}

	var js Json
	err = json.Unmarshal(buf, &js)
	if err != nil {
		return nil, err
	}

	ts, ok := js["timestamp"]
	if !ok {
		return nil, fmt.Errorf(
			"poll-device: failed to fetch registration: missing timestamp field")
	}

	timestamp, ok := ts.(float64)
	if !ok {
		return nil, fmt.Errorf(
			"poll-device: failed to fetch registration: invalid type for timestamp")
	}

	if timestamp == -1 {
		return nil, fmt.Errorf("poll-device: failed to fetch registration: invalid timestamp")
	}

	id, ok := js["device_id"]
	if !ok {
		return nil, fmt.Errorf(
			"poll-device: failed to fetch registration: missing device_id field")
	}

	deviceID, ok := id.(string)
	if !ok {
		return nil, fmt.Errorf(
			"poll-device: failed to fetch registration: invalid type for device_id")
	}

	if d.deviceId != "" && d.deviceId != deviceID {
		return nil, fmt.Errorf(
			"poll-device: failed to fetch registration: device ID mismatch: want=%s got=%s",
			d.deviceId, deviceID,
		)
	}

	d.deviceId = deviceID

	return js, nil
}

func (d *PollDevice) fetchTelemetry() (Json, error) {
	buf, err := d.telemetryFetcher.Fetch()
	if err != nil {
		return nil, err
	}

	var js Json

	err = json.Unmarshal(buf, &js)
	if err != nil {
		return nil, err
	}

	ts, ok := js["timestamp"]
	if !ok {
		return nil, fmt.Errorf("poll-device: failed to fetch telemetry: missing timestamp field")
	}

	timestamp, ok := ts.(float64)
	if !ok {
		return nil, fmt.Errorf("poll-device: failed to fetch telemetry: invalid type for timestamp")
	}

	if timestamp == -1 {
		return nil, fmt.Errorf("poll-device: failed to fetch telemetry: invalid timestamp")
	}

	return js, nil
}
