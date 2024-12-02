package device

// JSON device data.
type Json = map[string]interface{}

type DataHandler interface {
	// Handle telemetry data from the device.
	HandleTelemetry(deviceId string, js Json) error

	// Handle registration data from the device.
	HandleRegistration(deviceId string, js Json) error
}
