package core

// Closer implementation should free all allocated resources.
type Closer interface {
	// Close the resource.
	Close() error
}
