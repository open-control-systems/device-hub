package device

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type testFetcher struct {
	data []byte
	err  error
}

func (f *testFetcher) Fetch() ([]byte, error) {
	return f.data, f.err
}

type testRegistrationData struct {
	DeviceId  string  `json:"device_id"`
	Timestamp float64 `json:"timestamp"`
}

type testTelemetryData struct {
	Timestamp   float64 `json:"timestamp"`
	Temperature float32 `json:"temperature"`
	Status      string  `json:"status"`
}

type testDataHandler struct {
	telemetry    testTelemetryData
	registration testRegistrationData
	err          error
}

func (d *testDataHandler) HandleTelemetry(deviceId string, js Json) error {
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

func (d *testDataHandler) HandleRegistration(deviceId string, js Json) error {
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
	deviceId := "0xABCD"

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
		Timestamp: 123,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   123,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.Nil(t, device.Update())
	require.Equal(t, deviceId, dataHandler.registration.DeviceId)
	require.Equal(t, telemetryData, dataHandler.telemetry)
}

func TestPollDeviceRunFetchRegistrationError(t *testing.T) {
	deviceId := "0xABCD"

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
		Timestamp: 123,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  errors.New("failed to fetch"),
	}

	telemetryData := testTelemetryData{
		Timestamp:   123,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.NotNil(t, device.Update())
	require.Empty(t, dataHandler.registration.DeviceId)
}

func TestPollDeviceRunFetchTelemetryError(t *testing.T) {
	deviceId := "0xABCD"

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
		Timestamp: 123,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   123,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.NotNil(t, device.Update())
	require.Empty(t, dataHandler.registration.DeviceId)
}

func TestPollDeviceRunEmptyDeviceId(t *testing.T) {
	deviceId := ""

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
		Timestamp: 123,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   123,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.NotNil(t, device.Update())
	require.Empty(t, dataHandler.registration.DeviceId)
}

func TestPollDeviceRunInvalidTimestampRegistration(t *testing.T) {
	deviceId := "0xABCD"

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
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
		Timestamp:   123,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.NotNil(t, device.Update())
	require.Empty(t, dataHandler.registration.DeviceId)
}

func TestPollDeviceRunInvalidTimestampTelemetry(t *testing.T) {
	deviceId := "0xABCD"

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
		Timestamp: 123,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   -1,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.NotNil(t, device.Update())
	require.Empty(t, dataHandler.registration.DeviceId)
}

func TestPollDeviceRunDataHandlerFailed(t *testing.T) {
	deviceId := "0xABCD"

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
		Timestamp: 123,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   123,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.NotNil(t, device.Update())
	require.Empty(t, dataHandler.registration.DeviceId)
}

func TestPollDeviceRunDeviceIdChanged(t *testing.T) {
	deviceId := "0xABCD"

	registrationData := testRegistrationData{
		DeviceId:  deviceId,
		Timestamp: 123,
	}

	buf, err := json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher := testFetcher{
		data: buf,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   123,
		Temperature: 123.3,
		Status:      "foo",
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
	require.Equal(t, "", dataHandler.registration.DeviceId)

	require.Nil(t, device.Update())
	require.Equal(t, deviceId, dataHandler.registration.DeviceId)

	changedDeviceId := "0xCBDE"
	require.NotEqual(t, deviceId, changedDeviceId)

	registrationData.DeviceId = changedDeviceId

	buf, err = json.Marshal(registrationData)
	require.Nil(t, err)
	require.NotEmpty(t, buf)

	registrationFetcher.data = buf

	require.NotNil(t, device.Update())
	require.Equal(t, deviceId, dataHandler.registration.DeviceId)
}
