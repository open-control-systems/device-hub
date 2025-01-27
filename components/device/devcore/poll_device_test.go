package devcore

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-control-systems/device-hub/components/status"
)

type testFetcher[T any] struct {
	data T
	err  error
}

func (f *testFetcher[T]) Fetch() ([]byte, error) {
	if f.err != nil {
		return nil, f.err
	}

	buf, err := json.Marshal(f.data)
	if err != nil {
		return nil, err
	}

	return buf, nil
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

type testTimeSynchronizer struct {
	err       error
	callCount int
}

func (s *testTimeSynchronizer) Synchronize() error {
	s.callCount++

	if s.err != nil {
		return s.err
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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
		err:  errors.New("failed to fetch"),
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.Equal(t, "", dataHandler.registration.DeviceID)

	err := device.Run()
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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
		err:  nil,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.Equal(t, "", dataHandler.registration.DeviceID)

	err := device.Run()
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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.Equal(t, "", dataHandler.registration.DeviceID)

	err := device.Run()
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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.Equal(t, "", dataHandler.registration.DeviceID)

	err := device.Run()
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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(-1),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
		err:  errors.New("failed to fetch"),
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{
		err: errors.New("failed to handle"),
	}

	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.Equal(t, "", dataHandler.registration.DeviceID)

	err := device.Run()
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

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.Equal(t, "", dataHandler.registration.DeviceID)

	require.Nil(t, device.Run())
	require.Equal(t, deviceID, dataHandler.registration.DeviceID)

	changedDeviceID := "0xCBDE"
	require.NotEqual(t, deviceID, changedDeviceID)

	registrationData.DeviceID = changedDeviceID
	registrationFetcher.data = registrationData

	err := device.Run()
	require.Nil(t, err)
	require.Equal(t, changedDeviceID, dataHandler.registration.DeviceID)
}

func TestPollDeviceSynchronizeTimeTelemetryAndRegistration(t *testing.T) {
	deviceID := "0xABCD"
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: -1,
	}

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   -1,
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.NotNil(t, device.Run())
	require.Equal(t, 1, timeSynchronizer.callCount)

	testTimestamp := float64(13)

	telemetryData.Timestamp = testTimestamp
	telemetryFetcher.data = telemetryData

	registrationData.Timestamp = testTimestamp
	registrationFetcher.data = registrationData

	require.Nil(t, device.Run())
	require.Equal(t, 1, timeSynchronizer.callCount)
	require.Equal(t, deviceID, dataHandler.registration.DeviceID)
	require.Equal(t, testTimestamp, dataHandler.registration.Timestamp)
	require.Equal(t, testTimestamp, dataHandler.telemetry.Timestamp)
}

func TestPollDeviceSynchronizeTimeRegistration(t *testing.T) {
	deviceID := "0xABCD"
	testTemperature := 42.135
	testStatus := "foo"

	testTimestamp := float64(13)

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: -1,
	}

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(testTimestamp),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.NotNil(t, device.Run())
	require.Equal(t, 1, timeSynchronizer.callCount)

	registrationData.Timestamp = testTimestamp
	registrationFetcher.data = registrationData

	require.Nil(t, device.Run())
	require.Equal(t, 1, timeSynchronizer.callCount)
	require.Equal(t, deviceID, dataHandler.registration.DeviceID)
	require.Equal(t, testTimestamp, dataHandler.registration.Timestamp)
	require.Equal(t, testTimestamp, dataHandler.telemetry.Timestamp)
}

func TestPollDeviceSynchronizeTimeTelemetry(t *testing.T) {
	deviceID := "0xABCD"
	testTemperature := 42.135
	testStatus := "foo"

	testTimestamp := float64(13)

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: testTimestamp,
	}

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   float64(-1),
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{}
	timeSynchronizer := testTimeSynchronizer{}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.NotNil(t, device.Run())
	require.Equal(t, 1, timeSynchronizer.callCount)

	telemetryData.Timestamp = testTimestamp
	telemetryFetcher.data = telemetryData

	require.Nil(t, device.Run())
	require.Equal(t, 1, timeSynchronizer.callCount)
	require.Equal(t, deviceID, dataHandler.registration.DeviceID)
	require.Equal(t, testTimestamp, dataHandler.registration.Timestamp)
	require.Equal(t, testTimestamp, dataHandler.telemetry.Timestamp)
}

func TestPollDeviceSynchronizeTimeError(t *testing.T) {
	deviceID := "0xABCD"
	testTemperature := 42.135
	testStatus := "foo"

	registrationData := testRegistrationData{
		DeviceID:  deviceID,
		Timestamp: -1,
	}

	registrationFetcher := testFetcher[testRegistrationData]{
		data: registrationData,
	}

	telemetryData := testTelemetryData{
		Timestamp:   -1,
		Temperature: float64(testTemperature),
		Status:      testStatus,
	}

	telemetryFetcher := testFetcher[testTelemetryData]{
		data: telemetryData,
	}

	dataHandler := testDataHandler{}

	timeSynchronizer := testTimeSynchronizer{
		err: errors.New("failed to sync time"),
	}

	device := NewPollDevice(
		&registrationFetcher,
		&telemetryFetcher,
		&dataHandler,
		&timeSynchronizer,
	)

	require.NotNil(t, device.Run())
	require.Equal(t, 1, timeSynchronizer.callCount)

	require.NotNil(t, device.Run())
	require.Equal(t, 2, timeSynchronizer.callCount)
	require.Equal(t, "", dataHandler.registration.DeviceID)
	require.Equal(t, float64(0), dataHandler.registration.Timestamp)
	require.Equal(t, float64(0), dataHandler.telemetry.Timestamp)
}
