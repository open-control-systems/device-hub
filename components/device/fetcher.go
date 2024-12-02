package device

type Fetcher interface {
	// Fetch the device data.
	Fetch() ([]byte, error)
}
