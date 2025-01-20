package stcore

// DB is a key-value database to store blobs of data.
//
// Remarks:
//   - Implementation should be thread-safe.
type DB interface {
	// Read reads a blob from the database.
	//
	// Remarks:
	//  - Implementation should return status.StatusNoData if blob doesn't exist.
	Read(key string) (Blob, error)

	// Write write a blob to the database.
	Write(key string, blob Blob) error

	// Remove removes a blob from the database.
	//
	// Remarks:
	//  - Implementation should return nil if blob doesn't exist.
	Remove(key string) error

	// ForEach iterates over all data in the database.
	ForEach(fn func(key string, b Blob) error) error

	// Close releases all resources for the database.
	Close() error
}
