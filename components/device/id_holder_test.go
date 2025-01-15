package device

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/require"
)

type testIDHolderDataHandler struct {
	deviceID     string
	telemetry    JSON
	registration JSON
}

func (h *testIDHolderDataHandler) HandleRegistration(deviceID string, js JSON) error {
	h.deviceID = deviceID
	h.registration = maps.Clone(js)

	return nil
}

func (h *testIDHolderDataHandler) HandleTelemetry(deviceID string, js JSON) error {
	h.deviceID = deviceID
	h.telemetry = maps.Clone(js)

	return nil
}

func TestIDHolderGet(t *testing.T) {
	handler := &testIDHolderDataHandler{}

	holder := NewIDHolder(handler)
	require.Empty(t, holder.Get())

	testTelemetry := make(JSON)
	testTelemetry["foo"] = "bar"

	testRegistration := make(JSON)
	testRegistration["baz"] = "bug"

	deviceID := "0xABCD"

	require.Nil(t, holder.HandleTelemetry(deviceID, testTelemetry))
	require.Empty(t, holder.Get())

	require.True(t, maps.Equal(testTelemetry, handler.telemetry))

	require.Nil(t, holder.HandleRegistration(deviceID, testRegistration))
	require.True(t, maps.Equal(testRegistration, handler.registration))
	require.True(t, maps.Equal(testTelemetry, handler.telemetry))

	require.Equal(t, deviceID, holder.Get())
}

func TestIDHolderGetIDChanged(t *testing.T) {
	handler := &testIDHolderDataHandler{}

	holder := NewIDHolder(handler)

	testRegistration := make(JSON)
	testRegistration["baz"] = "bug"

	deviceID := "0xABCD"

	require.Nil(t, holder.HandleRegistration(deviceID, testRegistration))
	require.Equal(t, deviceID, holder.Get())

	newDeviceID := "0XDCBA"
	require.NotEqual(t, deviceID, newDeviceID)

	require.Nil(t, holder.HandleRegistration(newDeviceID, testRegistration))
	require.Equal(t, newDeviceID, holder.Get())
}
