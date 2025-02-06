package syssched

import "github.com/open-control-systems/device-hub/components/system/syscore"

// CrashTask reports the crash from the underlying task to the log.
type CrashTask struct {
	task Task
}

// NewCrashTask is an initialization of CrashTask.
func NewCrashTask(task Task) *CrashTask {
	return &CrashTask{
		task: task,
	}
}

// Run wraps the underlying task with panic-recovery mechanism.
func (t *CrashTask) Run() error {
	defer func() {
		if e := recover(); e != nil {
			syscore.LogCrash(e)
		}
	}()

	return t.task.Run()
}
