package devcore

// FuncSynchronizer is a function type that implements the TimeSynchronizer interface.
type FuncSynchronizer func() error

// SyncTime calls the function to synchronize local and device UNIX time.
func (s FuncSynchronizer) SyncTime() error {
	return s()
}
