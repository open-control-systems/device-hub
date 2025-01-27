package devcore

// JSON device data.
type JSON = map[string]any

// DataHandler handles varios data types from a device.
type DataHandler interface {
	// HandleTelemetry handles the telemetry data from the device.
	HandleTelemetry(deviceID string, js JSON) error

	// HandleRegistration handles the registration data from the device.
	HandleRegistration(deviceID string, js JSON) error
}
