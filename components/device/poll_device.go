package device

import (
	"encoding/json"
	"fmt"
)

// PollDevice actively fetches telemetry and registration data.
type PollDevice struct {
	registrationFetcher Fetcher
	telemetryFetcher    Fetcher
	dataHandler         DataHandler
	deviceID            string
}

// NewPollDevice initializes polling device.
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

// Update fetches telemetry and registration data and pass them to the underlying handlers.
func (d *PollDevice) Update() error {
	registrationData, err := d.fetchRegistration()
	if err != nil {
		return err
	}

	telemetryData, err := d.fetchTelemetry()
	if err != nil {
		return err
	}

	if err := d.dataHandler.HandleRegistration(d.deviceID, registrationData); err != nil {
		return err
	}

	if err := d.dataHandler.HandleTelemetry(d.deviceID, telemetryData); err != nil {
		return err
	}

	return nil
}

func (d *PollDevice) fetchRegistration() (JSON, error) {
	buf, err := d.registrationFetcher.Fetch()
	if err != nil {
		return nil, err
	}

	var js JSON
	err = json.Unmarshal(buf, &js)
	if err != nil {
		return nil, err
	}

	if err := validateTimestamp(js); err != nil {
		return nil, err
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

	if d.deviceID != "" && d.deviceID != deviceID {
		return nil, fmt.Errorf(
			"poll-device: failed to fetch registration: device ID mismatch: want=%s got=%s",
			d.deviceID, deviceID,
		)
	}

	d.deviceID = deviceID

	return js, nil
}

func (d *PollDevice) fetchTelemetry() (JSON, error) {
	buf, err := d.telemetryFetcher.Fetch()
	if err != nil {
		return nil, err
	}

	var js JSON

	err = json.Unmarshal(buf, &js)
	if err != nil {
		return nil, err
	}

	if err := validateTimestamp(js); err != nil {
		return nil, err
	}

	return js, nil
}

func validateTimestamp(js JSON) error {
	ts, ok := js["timestamp"]
	if !ok {
		return fmt.Errorf("poll-device: failed to fetch data: missing timestamp field")
	}

	timestamp, ok := ts.(float64)
	if !ok {
		return fmt.Errorf("poll-device: failed to fetch data: invalid type for timestamp")
	}

	if timestamp == -1 {
		return fmt.Errorf("poll-device: failed to fetch data: invalid timestamp")
	}

	return nil
}
