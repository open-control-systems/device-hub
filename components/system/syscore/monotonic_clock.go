package syscore

import "time"

// MonotonicClock to read monotonic time.
type MonotonicClock interface {
	// Now returns a monotonic clock reading.
	Now() time.Time
}
