package syssched

// Stopper implementation should free all allocated resources.
type Stopper interface {
	// Stop stops the resource.
	Stop() error
}
