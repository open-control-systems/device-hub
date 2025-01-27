package syscore

import "time"

// LocalMonotonicClock is wrapper around standard time.Time package.
type LocalMonotonicClock struct{}

// Now returns the current local time.
func (*LocalMonotonicClock) Now() time.Time {
	return time.Now()
}
