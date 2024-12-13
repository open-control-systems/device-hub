package device

// Device represents an IoT device.
type Device interface {
	// Update device data.
	Update() error
}
