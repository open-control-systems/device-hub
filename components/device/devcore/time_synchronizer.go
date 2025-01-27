package devcore

// TimeSynchronizer synchronizes time between local and remote resources.
type TimeSynchronizer interface {
	// TimeSynchronizer synchronizes the UNIX time for a device.
	Synchronize() error
}
