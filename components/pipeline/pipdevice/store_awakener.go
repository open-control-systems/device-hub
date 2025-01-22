package pipdevice

import "github.com/open-control-systems/device-hub/components/system/syssched"

// StoreAwakener to notify the awakener that the store operation has happened.
type StoreAwakener struct {
	awakener syssched.Awakener
	store    Store
}

// NewStoreAwakener is an initialization of StoreAwakener.
func NewStoreAwakener(a syssched.Awakener, s Store) *StoreAwakener {
	return &StoreAwakener{
		awakener: a,
		store:    s,
	}
}

// Add adds the device and notifies the awakener.
func (a *StoreAwakener) Add(uri string, desc string) error {
	err := a.store.Add(uri, desc)
	if err == nil {
		a.awakener.Awake()
	}

	return err
}

// Remove removes the device associated with the provided URI.
func (a *StoreAwakener) Remove(uri string) error {
	return a.store.Remove(uri)
}

// GetDesc returns descriptions for registered devices.
func (a *StoreAwakener) GetDesc() []StoreItem {
	return a.store.GetDesc()
}
