package devcore

import (
	"encoding/json"
	"fmt"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// PollDevice actively fetches telemetry and registration data.
type PollDevice struct {
	registrationFetcher Fetcher
	telemetryFetcher    Fetcher
	dataHandler         DataHandler
	timeSynchronizer    TimeSynchronizer
	timeVerifier        TimeVerifier
	deviceID            string
}

// NewPollDevice initializes polling device.
//
// Parameters:
//   - registrationFetcher to fetch device registration data.
//   - telemetryFetcher to fetch device telemetry data.
//   - dataHandler to handle fetched telemetry and registration data.
//   - timeSynchronizer to synchronize the UNIX time for a device.
func NewPollDevice(
	registrationFetcher Fetcher,
	telemetryFetcher Fetcher,
	dataHandler DataHandler,
	timeSynchronizer TimeSynchronizer,
	timeVerifier TimeVerifier,
) *PollDevice {
	return &PollDevice{
		registrationFetcher: registrationFetcher,
		telemetryFetcher:    telemetryFetcher,
		dataHandler:         dataHandler,
		timeSynchronizer:    timeSynchronizer,
		timeVerifier:        timeVerifier,
	}
}

// Run fetches telemetry and registration data and pass them to the underlying handlers.
func (d *PollDevice) Run() error {
	registrationData, err := d.fetchRegistration()
	if err != nil {
		syscore.LogErr.Printf("fetch registration failed: %v", err)

		return status.StatusError
	}

	telemetryData, err := d.fetchTelemetry()
	if err != nil {
		syscore.LogErr.Printf("fetch telemetry failed: %v", err)

		return status.StatusError
	}

	if err := d.dataHandler.HandleRegistration(d.deviceID, registrationData); err != nil {
		syscore.LogErr.Printf("handle registration failed: %v", err)

		return status.StatusError
	}

	if err := d.dataHandler.HandleTelemetry(d.deviceID, telemetryData); err != nil {
		syscore.LogErr.Printf("handle telemetry failed: %v", err)

		return status.StatusError
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

	err = d.parseDeviceID(js)
	if err != nil {
		return nil, err
	}

	if err := d.validateTimestamp(js); err != nil {
		return nil, err
	}

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

	if err := d.validateTimestamp(js); err != nil {
		return nil, err
	}

	return js, nil
}

func (d *PollDevice) validateTimestamp(js JSON) error {
	ts, ok := js["timestamp"]
	if !ok {
		return fmt.Errorf("poll-device: failed to fetch data: missing timestamp field")
	}

	timestamp, ok := ts.(float64)
	if !ok {
		return fmt.Errorf("poll-device: failed to fetch data: invalid type for timestamp")
	}

	if !d.timeVerifier.VerifyTime(int64(timestamp)) {
		syscore.LogInf.Printf("start syncing time for device: ID=%v", d.deviceID)

		if err := d.timeSynchronizer.SyncTime(); err != nil {
			return fmt.Errorf("failed to sync device time: %v", err)
		}

		return fmt.Errorf("failed to fetch data: invalid timestamp")
	}

	return nil
}

func (d *PollDevice) parseDeviceID(js JSON) error {
	id, ok := js["device_id"]
	if !ok {
		return fmt.Errorf(
			"poll-device: failed to fetch registration: missing device_id field")
	}

	deviceID, ok := id.(string)
	if !ok {
		return fmt.Errorf(
			"poll-device: failed to fetch registration: invalid type for device_id")
	}

	if d.deviceID != "" && d.deviceID != deviceID {
		return fmt.Errorf(
			"poll-device: failed to fetch registration: device ID mismatch: want=%s got=%s",
			d.deviceID, deviceID,
		)
	}

	if d.deviceID == "" {
		syscore.LogInf.Printf("device ID received: %s", deviceID)

		d.deviceID = deviceID
	}

	return nil
}
