package devcore

// TimeVerifier to verify the UNIX timestamp of the device.
type TimeVerifier interface {
	// VerifyTime returns true if the provided UNIX timestamp is valid.
	VerifyTime(timestamp int64) bool
}
