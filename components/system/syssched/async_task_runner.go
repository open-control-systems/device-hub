package syssched

import (
	"context"
	"time"
)

// AsyncTaskRunner periodically runs task in the standalone goroutine.
type AsyncTaskRunner struct {
	ctx            context.Context
	doneCh         chan struct{}
	task           Task
	reporter       ErrorReporter
	updateInterval time.Duration
}

// NewAsyncTaskRunner is an initialization of AsyncTaskRunner.
func NewAsyncTaskRunner(
	ctx context.Context,
	task Task,
	reporter ErrorReporter,
	updateInterval time.Duration,
) *AsyncTaskRunner {
	return &AsyncTaskRunner{
		ctx:            ctx,
		doneCh:         make(chan struct{}),
		task:           task,
		reporter:       reporter,
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

	for {
		select {
		case <-ticker.C:
			if err := r.task.Run(); err != nil {
				if r.reporter != nil {
					r.reporter.ReportError(err)
				}
			}

		case <-r.ctx.Done():
			return
		}
	}
}
