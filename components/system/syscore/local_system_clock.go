package syscore

import (
	"time"

	"golang.org/x/sys/unix"
)

// LocalSystemClock is used to set/get the current UNIX time.
type LocalSystemClock struct{}

// SetTimestamp sets the UNIX time via settimeofday(2) system call.
//
// References:
//   - https://linux.die.net/man/2/settimeofday
func (*LocalSystemClock) SetTimestamp(timestamp int64) error {
	tv := unix.Timeval{
		Sec:  timestamp,
		Usec: 0,
	}

	return unix.Settimeofday(&tv)
}

// GetTimestamp returns the current UNIX time.
func (*LocalSystemClock) GetTimestamp() (int64, error) {
	return time.Now().Unix(), nil
}
