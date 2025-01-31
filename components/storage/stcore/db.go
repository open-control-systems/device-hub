package stcore

// DB is a key-value database to store arbitrary data.
//
// Remarks:
//   - Implementation should be thread-safe.
type DB interface {
	// Read reads data for the given key.
	//
	// Remarks:
	//  - Implementation should return status.StatusNoData if data doesn't exist.
	Read(key string) ([]byte, error)

	// Write writes data to the database.
	Write(key string, value []byte) error

	// Remove removes data from the database.
	//
	// Remarks:
	//  - Implementation should return nil if value doesn't exist.
	Remove(key string) error

	// ForEach iterates over all data in the database.
	ForEach(fn func(key string, value []byte) error) error

	// Close releases all resources for the database.
	Close() error
}
