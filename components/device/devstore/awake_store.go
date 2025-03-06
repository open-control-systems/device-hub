package devstore

import "github.com/open-control-systems/device-hub/components/system/syssched"

// AwakeStore to notify the awakener that the store operation has happened.
type AwakeStore struct {
	awakener syssched.Awakener
	store    Store
}

// NewAwakeStore is an initialization of AwakeStore.
func NewAwakeStore(a syssched.Awakener, s Store) *AwakeStore {
	return &AwakeStore{
		awakener: a,
		store:    s,
	}
}

// Add adds the device and notifies the awakener.
func (s *AwakeStore) Add(uri string, typ string, desc string) error {
	err := s.store.Add(uri, typ, desc)
	if err == nil {
		s.awakener.Awake()
	}

	return err
}

// Remove removes the device associated with the provided URI.
func (s *AwakeStore) Remove(uri string) error {
	return s.store.Remove(uri)
}

// GetDesc returns descriptions for registered devices.
func (s *AwakeStore) GetDesc() []StoreItem {
	return s.store.GetDesc()
}
