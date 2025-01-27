package syssched

// AliveNotifier to notify when an operation is running normally.
type AliveNotifier interface {
	// Notify indicates that an operation is running normally.
	NotifyAlive()
}
