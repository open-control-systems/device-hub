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

func (r *AsyncTaskRunner) run() {
	defer close(r.doneCh)

	ticker := time.NewTicker(r.updateInterval)
	defer ticker.Stop()

	r.runTask()

	for {
		select {
		case <-ticker.C:
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
