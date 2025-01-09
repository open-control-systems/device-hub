package syscore

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/open-control-systems/device-hub/components/status"
)

type testSystemClock struct {
	timestamp int64
	setErr    error
	getErr    error
}

func (c *testSystemClock) GetTimestamp() (int64, error) {
	if c.getErr != nil {
		return -1, c.getErr
	}

	return c.timestamp, nil
}

func (c *testSystemClock) SetTimestamp(timestamp int64) error {
	if c.setErr != nil {
		return c.setErr
	}

	c.timestamp = timestamp

	return nil
}

func TestSystemClockSynchronizerSynchronizeLocalError(t *testing.T) {
	local := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
		getErr:    status.StatusError,
	}

	remoteLast := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
	}

	remoteCurr := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
	}

	synchronizer := NewSystemClockSynchronizer(local, remoteLast, remoteCurr)
	require.Equal(t, status.StatusError, synchronizer.Synchronize())
}

func TestSystemClockSynchronizerSynchronizeRemoteLastError(t *testing.T) {
	local := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
	}

	remoteLast := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
		getErr:    status.StatusError,
	}

	remoteCurr := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
	}

	synchronizer := NewSystemClockSynchronizer(local, remoteLast, remoteCurr)
	require.Equal(t, status.StatusError, synchronizer.Synchronize())
}

func TestSystemClockSynchronizerSynchronizeRemoteLastAheadOfLocal(t *testing.T) {
	localTimestamp := int64(10)
	remoteLastTimestamp := localTimestamp * 2
	require.NotEqual(t, localTimestamp, remoteLastTimestamp)

	local := &testSystemClock{
		timestamp: localTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteLast := &testSystemClock{
		timestamp: remoteLastTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteCurr := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
	}

	synchronizer := NewSystemClockSynchronizer(local, remoteLast, remoteCurr)
	require.Equal(t, status.StatusError, synchronizer.Synchronize())
}

func TestSystemClockSynchronizerSynchronizeRemoteCurrError(t *testing.T) {
	localTimestamp := int64(10)
	remoteLastTimestamp := localTimestamp / 2
	require.NotEqual(t, localTimestamp, remoteLastTimestamp)

	local := &testSystemClock{
		timestamp: localTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteLast := &testSystemClock{
		timestamp: remoteLastTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteCurr := &testSystemClock{
		timestamp: -1,
		setErr:    status.StatusNotSupported,
		getErr:    status.StatusError,
	}

	synchronizer := NewSystemClockSynchronizer(local, remoteLast, remoteCurr)
	require.Equal(t, status.StatusError, synchronizer.Synchronize())
}

func TestSystemClockSynchronizerSynchronizeRemoteCurrAheadOfLocal(t *testing.T) {
	localTimestamp := int64(10)
	remoteLastTimestamp := localTimestamp / 2
	remoteCurrTimestamp := localTimestamp * 2
	require.NotEqual(t, localTimestamp, remoteLastTimestamp)
	require.NotEqual(t, remoteCurrTimestamp, remoteLastTimestamp)

	local := &testSystemClock{
		timestamp: localTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteLast := &testSystemClock{
		timestamp: remoteLastTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteCurr := &testSystemClock{
		timestamp: remoteCurrTimestamp,
		setErr:    status.StatusNotSupported,
	}

	synchronizer := NewSystemClockSynchronizer(local, remoteLast, remoteCurr)
	require.Equal(t, status.StatusError, synchronizer.Synchronize())
}

func TestSystemClockSynchronizerSynchronizeRemoteSetTimestampError(t *testing.T) {
	localTimestamp := int64(10)
	remoteLastTimestamp := localTimestamp / 2
	remoteCurrTimestamp := int64(-1)
	require.NotEqual(t, localTimestamp, remoteLastTimestamp)
	require.NotEqual(t, remoteCurrTimestamp, remoteLastTimestamp)

	local := &testSystemClock{
		timestamp: localTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteLast := &testSystemClock{
		timestamp: remoteLastTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteCurr := &testSystemClock{
		timestamp: remoteCurrTimestamp,
		setErr:    status.StatusNotSupported,
	}

	synchronizer := NewSystemClockSynchronizer(local, remoteLast, remoteCurr)
	require.Equal(t, status.StatusNotSupported, synchronizer.Synchronize())
}

func TestSystemClockSynchronizerSynchronize(t *testing.T) {
	localTimestamp := int64(10)
	remoteLastTimestamp := localTimestamp / 2
	remoteCurrTimestamp := int64(-1)
	require.NotEqual(t, localTimestamp, remoteLastTimestamp)
	require.NotEqual(t, remoteCurrTimestamp, remoteLastTimestamp)

	local := &testSystemClock{
		timestamp: localTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteLast := &testSystemClock{
		timestamp: remoteLastTimestamp,
		setErr:    status.StatusNotSupported,
	}

	remoteCurr := &testSystemClock{
		timestamp: remoteCurrTimestamp,
	}

	synchronizer := NewSystemClockSynchronizer(local, remoteLast, remoteCurr)
	require.Nil(t, synchronizer.Synchronize())
	require.Equal(t, remoteLastTimestamp, remoteLast.timestamp)
	require.Equal(t, localTimestamp, local.timestamp)
	require.Equal(t, localTimestamp, remoteCurr.timestamp)
}
