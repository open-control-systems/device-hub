package syssched

import (
	"context"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
)

// AsyncTaskRunner periodically runs task in the standalone goroutine.
type AsyncTaskRunner struct {
	ctx            context.Context
	doneCh         chan struct{}
	awakeCh        chan struct{}
	task           Task
	handler        core.ErrorHandler
	updateInterval time.Duration
}

// NewAsyncTaskRunner is an initialization of AsyncTaskRunner.
func NewAsyncTaskRunner(
	ctx context.Context,
	task Task,
	handler core.ErrorHandler,
	updateInterval time.Duration,
) *AsyncTaskRunner {
	return &AsyncTaskRunner{
		ctx:            ctx,
		doneCh:         make(chan struct{}),
		awakeCh:        make(chan struct{}, 1),
		task:           task,
		handler:        handler,
		updateInterval: updateInterval,
	}
}

// Start begins asynchronous task processing.
func (r *AsyncTaskRunner) Start() {
	go r.run()
}

// Close ends asynchronous task processing.
func (r *AsyncTaskRunner) Close() error {
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

	ticker := time.NewTicker(r.updateInterval)
	defer ticker.Stop()

	r.runTask()

	for {
		select {
		case <-ticker.C:
			r.runTask()

		case <-r.awakeCh:
			r.runTask()

		case <-r.ctx.Done():
			return
		}
	}
}

func (r *AsyncTaskRunner) runTask() {
	if err := r.task.Run(); err != nil {
		if r.handler != nil {
			r.handler.HandleError(err)
		}
	}
}
