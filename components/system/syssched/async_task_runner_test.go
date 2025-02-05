package syssched

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/stretchr/testify/require"
)

type testAsyncTaskRunnerTestTask struct {
	mu        sync.Mutex
	err       error
	callCount int
}

func (t *testAsyncTaskRunnerTestTask) Run() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.callCount++

	if t.err != nil {
		return t.err
	}

	return nil
}

func (t *testAsyncTaskRunnerTestTask) getCallCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.callCount
}

func (t *testAsyncTaskRunnerTestTask) setError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.err = err
}

func TestAsyncTaskRunnerExitOnSuccess(t *testing.T) {
	task := &testAsyncTaskRunnerTestTask{
		err: status.StatusNotSupported,
	}
	ctx := context.Background()

	runner := NewAsyncTaskRunner(ctx, task, nil, AsyncTaskRunnerParams{
		UpdateInterval: time.Millisecond * 100,
		ExitOnSuccess:  true,
	})
	require.Nil(t, runner.Start())

	for task.getCallCount() < 2 {
		time.Sleep(time.Millisecond * 50)
	}

	task.setError(nil)
	require.Nil(t, runner.Stop())
}
