package device

// Fetcher fetches the device data from the arbitrary source.
type Fetcher interface {
	// Fetch the device data.
	Fetch() ([]byte, error)
}
