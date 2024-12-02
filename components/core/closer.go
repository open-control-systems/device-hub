package core

type Closer interface {
	// Close the resource.
	Close() error
}
