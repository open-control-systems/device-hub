package syssched

// Task represents an entity of the execution.
type Task interface {
	// Run executes a single operational loop.
	Run() error
}
