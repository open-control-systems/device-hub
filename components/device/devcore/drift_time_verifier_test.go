package devcore

import (
	"testing"
	"time"

	"github.com/open-control-systems/device-hub/components/status"
	"github.com/stretchr/testify/require"
)

type testDriftTimeVerifierTestClock struct {
	timestamp int64
	err       error
}

func (c *testDriftTimeVerifierTestClock) GetTimestamp() (int64, error) {
	if c.err != nil {
		return -1, c.err
	}

	return c.timestamp, c.err
}

func (c *testDriftTimeVerifierTestClock) SetTimestamp(timestamp int64) error {
	if c.err != nil {
		return c.err
	}

	c.timestamp = timestamp

	return nil
}

func TestDriftTimeVerifierVerifyTimeDeviceTimeIsInvalid(t *testing.T) {
	clock := &testDriftTimeVerifierTestClock{}

	verifier := NewDriftTimeVerifier(clock, time.Second)
	require.False(t, verifier.VerifyTime(-1))
}

func TestDriftTimeVerifierVerifyTimeFailedToGetTimestamp(t *testing.T) {
	clock := &testDriftTimeVerifierTestClock{
		err: status.StatusTimeout,
	}

	verifier := NewDriftTimeVerifier(clock, time.Second)
	require.False(t, verifier.VerifyTime(1))
}

func TestDriftTimeVerifierVerifyTimeDeviceFromFuture(t *testing.T) {
	localTs := int64(100)
	deviceTs := int64(200)
	require.True(t, localTs < deviceTs)

	clock := &testDriftTimeVerifierTestClock{
		timestamp: localTs,
	}

	verifier := NewDriftTimeVerifier(clock, time.Second)
	require.True(t, verifier.VerifyTime(deviceTs))
}

func TestDriftTimeVerifierVerifyTime(t *testing.T) {
	localTs := int64(200)

	clock := &testDriftTimeVerifierTestClock{
		timestamp: localTs,
	}

	verifier := NewDriftTimeVerifier(clock, time.Minute)

	for n := int64(0); n < int64(time.Minute.Seconds()); n++ {
		require.True(t, verifier.VerifyTime(localTs-n))
	}
	require.False(t, verifier.VerifyTime(localTs-int64(time.Minute.Seconds())))
}
