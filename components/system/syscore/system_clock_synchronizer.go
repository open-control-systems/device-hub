package syscore

import (
	"github.com/open-control-systems/device-hub/components/status"
)

// SystemClockSynchronizer synchronizes the UNIX time between local and remote resources.
type SystemClockSynchronizer struct {
	local      SystemClock
	remoteLast SystemClock
	remoteCurr SystemClock
}

// NewSystemClockSynchronizer initializes the component for the UNIX time synchronization.
//
// Parameters:
//   - local - UNIX time of the local resource.
//   - remoteLast - last know UNIX time for the remote resource.
//   - remoteCurr - current UNIX time for the remote resource.
func NewSystemClockSynchronizer(
	local SystemClock,
	remoteLast SystemClock,
	remoteCurr SystemClock,
) *SystemClockSynchronizer {
	return &SystemClockSynchronizer{
		local:      local,
		remoteLast: remoteLast,
		remoteCurr: remoteCurr,
	}
}

// SyncTime synchronizes the UNIX time between local and remote resources.
func (s *SystemClockSynchronizer) SyncTime() error {
	localTs, err := s.local.GetTimestamp()
	if err != nil {
		return err
	}

	remoteLastTs, err := s.remoteLast.GetTimestamp()
	if err != nil {
		return err
	}

	if localTs < remoteLastTs {
		LogWrn.Printf(
			"system-clock-synchronizer: unable to sync: last remote is ahead of local: "+
				"local=%v remote=%v", localTs, remoteLastTs)

		return status.StatusError
	}

	remoteCurrTs, err := s.remoteCurr.GetTimestamp()
	if err != nil {
		return err
	}

	if localTs < remoteCurrTs {
		LogWrn.Printf(
			"unable to sync: current remote is ahead of local local=%v remote=%v",
			localTs, remoteCurrTs)

		return status.StatusError
	}

	if err := s.remoteCurr.SetTimestamp(localTs); err != nil {
		return err
	}

	LogInf.Printf(
		"system-clock-synchronizer: time synced: local=%v remote_last=%v remote_curr=%v",
		localTs, remoteLastTs, remoteCurrTs)

	return nil
}
