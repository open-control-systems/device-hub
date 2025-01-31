package stcore

import "github.com/open-control-systems/device-hub/components/status"

// NoopDB is a non-operational database.
type NoopDB struct{}

// Read is non-operational.
func (*NoopDB) Read(_ string) ([]byte, error) {
	return []byte{}, status.StatusNoData
}

// Write is non-operational.
func (*NoopDB) Write(_ string, _ []byte) error {
	return nil
}

// Remove is non-operational.
func (*NoopDB) Remove(_ string) error {
	return nil
}

// ForEach is non-operational.
func (*NoopDB) ForEach(_ func(key string, b []byte) error) error {
	return nil
}

// Close is non-operational.
func (*NoopDB) Close() error {
	return nil
}
