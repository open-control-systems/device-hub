package syssched

import (
	"context"
	"time"
)

// AsyncTaskRunnerParams represents various configuration options for AsyncTaskRunner.
type AsyncTaskRunnerParams struct {
	// UpdateInterval is how often a task should be run.
	UpdateInterval time.Duration

	// ExitOnSuccess is to stop asynchronous task processing after first successful execution.
	ExitOnSuccess bool
}

// AsyncTaskRunner periodically runs task in the standalone goroutine.
type AsyncTaskRunner struct {
	ctx     context.Context
	doneCh  chan struct{}
	awakeCh chan struct{}
	task    Task
	handler ErrorHandler
	params  AsyncTaskRunnerParams
}

// NewAsyncTaskRunner is an initialization of AsyncTaskRunner.
func NewAsyncTaskRunner(
	ctx context.Context,
	task Task,
	handler ErrorHandler,
	params AsyncTaskRunnerParams,
) *AsyncTaskRunner {
	return &AsyncTaskRunner{
		ctx:     ctx,
		doneCh:  make(chan struct{}),
		awakeCh: make(chan struct{}, 1),
		task:    task,
		handler: handler,
		params:  params,
	}
}

// Start begins asynchronous task processing.
func (r *AsyncTaskRunner) Start() error {
	go r.run()

	return nil
}

// Stop ends asynchronous task processing.
func (r *AsyncTaskRunner) Stop() error {
	<-r.doneCh

	return nil
}

// Awake wakes up the underlying goroutine.
func (r *AsyncTaskRunner) Awake() {
	select {
	case r.awakeCh <- struct{}{}:
	default:
	}
}

func (r *AsyncTaskRunner) run() {
	defer close(r.doneCh)

	ticker := time.NewTicker(r.params.UpdateInterval)
	defer ticker.Stop()

	if r.runTask() {
		return
	}

	for {
		select {
		case <-ticker.C:
			if r.runTask() {
				return
			}

		case <-r.awakeCh:
			if r.runTask() {
				return
			}

		case <-r.ctx.Done():
			return
		}
	}
}

func (r *AsyncTaskRunner) runTask() bool {
	if err := r.task.Run(); err != nil {
		if r.handler != nil {
			r.handler.HandleError(err)
		}

		return false
	}

	return r.params.ExitOnSuccess
}
