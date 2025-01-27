package devstore

import "github.com/open-control-systems/device-hub/components/system/syssched"

// AliveMonitor to monitor the device well-being.
type AliveMonitor interface {
	// Monitor monitors the device well-being.
	//
	// Remarks:
	//	- If the device isn't marked as alive with the returned alive notifier,
	//	  it can be considered as inactive.
	Monitor(uri string) syssched.AliveNotifier
}
