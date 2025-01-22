package syssched

// Awakener to wake up an execution.
type Awakener interface {
	// Awake wakes up an execution.
	Awake()
}
