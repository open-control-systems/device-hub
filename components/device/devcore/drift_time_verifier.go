package devcore

import (
	"time"

	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// DriftTimeVerifier checks the timestamp difference between local and device UNIX time.
type DriftTimeVerifier struct {
	clock            syscore.SystemClock
	maxDriftInterval time.Duration
}

// NewDriftTimeVerifier is an initialization of DriftTimeVerifier.
func NewDriftTimeVerifier(
	clock syscore.SystemClock,
	maxDriftInterval time.Duration,
) *DriftTimeVerifier {
	return &DriftTimeVerifier{
		clock:            clock,
		maxDriftInterval: maxDriftInterval,
	}
}

// VerifyTime returns true if the time difference between local and device UNIX time
// is within the allowed range.
func (v *DriftTimeVerifier) VerifyTime(deviceTs int64) bool {
	if deviceTs < 0 {
		return false
	}

	localTs, err := v.clock.GetTimestamp()
	if err != nil {
		syscore.LogErr.Printf("failed to get local time: %v", err)

		return false
	}

	if deviceTs < localTs {
		return localTs-deviceTs < int64(v.maxDriftInterval.Seconds())
	}

	// If the device time is from the future, we can't make any assumptions about its validity.
	// It's the valid case, when the local clock and the device clock are drifting a bit, no
	// need to perform the synchronization in this case. As the device data can be persisted
	// based on its timestamp, downsyncing the device time can lead to undefined results.
	// On the other hand, up-syncing the device time can only lead to the gaps in the
	// persistent storage. If it's a problem, some kind of metric should be introduced here
	// to check how often such case can happen.
	return true
}
