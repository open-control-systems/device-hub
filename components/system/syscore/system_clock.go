package syscore

// SystemClock represents a UNIX time of the resource.
type SystemClock interface {
	// SetTimestamp sets the UNIX time for the resource.
	//
	// Requirements:
	//  - Implementation should be thread safe.
	SetTimestamp(timestamp int64) error

	// GetTimestamp returns the UNIX time for the resource.
	//
	// Notes:
	//  - -1 should be returned on error.
	//
	// Requirements:
	//  - Implementation should be thread safe.
	GetTimestamp() (int64, error)
}
