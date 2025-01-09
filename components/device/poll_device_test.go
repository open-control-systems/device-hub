package device

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-control-systems/device-hub/components/status"
)

type testFetcher struct {
	data []byte
	err  error
}

func (f *testFetcher) Fetch() ([]byte, error) {
	return f.data, f.err
}

type testRegistrationData struct {
	DeviceID  string  `json:"device_id"`
	Timestamp float64 `json:"timestamp"`
}

type testTelemetryData struct {
	Timestamp   float64 `json:"timestamp"`
	Temperature float64 `json:"temperature"`
	Status      string  `json:"status"`
}

type testDataHandler struct {
	telemetry    testTelemetryData
	registration testRegistrationData
	err          error
}

func (d *testDataHandler) HandleTelemetry(_ string, js JSON) error {
	if d.err != nil {
		return d.err
	}

	buf, err := json.Marshal(js)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(buf, &d.telemetry); err != nil {
		return err
	}

	return nil
}

func (d *testDataHandler) HandleRegistration(_ string, js JSON) error {
	if d.err != nil {
		return d.err
	}

	buf, err := json.Marshal(js)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(buf, &d.registration); err != nil {
		return err
	}

	return nil
}

func TestPollDeviceRun(t *testing.T) {
	deviceID := "0xABCD"
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: float64(testTimestamp),
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	dataHandler := testDataHandler{}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	require.Nil(t, device.Run())
	require.Equal(t, deviceID, dataHandler.registration.DeviceID)
	require.Equal(t, telemetryData, dataHandler.telemetry)
}

func TestPollDeviceRunFetchRegistrationError(t *testing.T) {
	deviceID := "0xABCD"
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: float64(testTimestamp),
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  errors.New("failed to fetch"),
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	dataHandler := testDataHandler{}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	err = device.Run()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, status.StatusError))
	require.Empty(t, dataHandler.registration.DeviceID)
}

func TestPollDeviceRunFetchTelemetryError(t *testing.T) {
	deviceID := "0xABCD"
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: float64(testTimestamp),
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	err = device.Run()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, status.StatusError))
	require.Empty(t, dataHandler.registration.DeviceID)
}

func TestPollDeviceRunEmptyDeviceId(t *testing.T) {
	deviceID := ""
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: float64(testTimestamp),
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	err = device.Run()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, status.StatusError))
	require.Empty(t, dataHandler.registration.DeviceID)
}

func TestPollDeviceRunInvalidTimestampRegistration(t *testing.T) {
	deviceID := "0xABCD"
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: -1,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	err = device.Run()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, status.StatusError))
	require.Empty(t, dataHandler.registration.DeviceID)
}

func TestPollDeviceRunInvalidTimestampTelemetry(t *testing.T) {
	deviceID := "0xABCD"
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: float64(testTimestamp),
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(-1),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	err = device.Run()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, status.StatusError))
	require.Empty(t, dataHandler.registration.DeviceID)
}

func TestPollDeviceRunDataHandlerFailed(t *testing.T) {
	deviceID := "0xABCD"
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: float64(testTimestamp),
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	dataHandler := testDataHandler{
		err: errors.New("failed to handle"),
	}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	err = device.Run()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, status.StatusError))
	require.Empty(t, dataHandler.registration.DeviceID)
}

func TestPollDeviceRunDeviceIdChanged(t *testing.T) {
	deviceID := "0xABCD"
	testTimestamp := 13
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: float64(testTimestamp),
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	buf, err = json.Marshal(telemetryData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	telemetryFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	dataHandler := testDataHandler{}

	device := NewPollDevice(&registrationFetcher, &telemetryFetcher, &dataHandler)
	require.Equal(t, "", dataHandler.registration.DeviceID)

	require.Nil(t, device.Run())
	require.Equal(t, deviceID, dataHandler.registration.DeviceID)

	changedDeviceID := "0xCBDE"
	require.NotEqual(t, deviceID, changedDeviceID)

	registrationData.DeviceID = changedDeviceID

	buf, err = json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher.data = buf

	err = device.Run()
	require.NotNil(t, err)
	require.True(t, errors.Is(err, status.StatusError))
	require.Equal(t, deviceID, dataHandler.registration.DeviceID)
}
