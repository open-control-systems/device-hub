package stcore

import "context"

// SystemClockReader to read the UNIX timestamp from the persistent storage.
type SystemClockReader interface {
	// ReadTimestamp reads the UNIX timestamp from the persistent storage.
	ReadTimestamp(context.Context) (int64, error)
}
