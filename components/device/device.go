package device

type Device interface {
	// Update device data.
	Update() error
}
