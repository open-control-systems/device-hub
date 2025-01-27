package syssched

// TaskAliveNotifier to notify if the task is running without errors.
type TaskAliveNotifier struct {
	task     Task
	notifier AliveNotifier
}

// NewTaskAliveNotifier is an initialization of TaskAliveNotifier.
func NewTaskAliveNotifier(task Task, notifier AliveNotifier) *TaskAliveNotifier {
	return &TaskAliveNotifier{
		task:     task,
		notifier: notifier,
	}
}

// Run runs the task and notify if it is running without errors.
func (n *TaskAliveNotifier) Run() error {
	if err := n.task.Run(); err != nil {
		return err
	}

	n.notifier.NotifyAlive()

	return nil
}
