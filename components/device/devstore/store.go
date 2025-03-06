package devstore

import "errors"

// StoreItem is a description of a single device.
type StoreItem struct {
	URI       string `json:"uri"`
	Type      string `json:"type"`
	Desc      string `json:"desc"`
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

// ErrDeviceExist is returned if the device already exists in the store.
var ErrDeviceExist = errors.New("device already exists")

// Store to manage device registration life-cycle.
type Store interface {
	// Add adds the device.
	//
	// Parameters:
	//   - uri - device URI, how device can be reached.
	//   - typ - device type, to distinguish one device from another.
	//   - desc - human readable device description.
	//
	// Remarks:
	//   - uri should be unique.
	//   - ErrDeviceExist is returned if the device already exists.
	//
	// URI examples:
	//   - http://bonsai-growlab.local:12345/api/v1. mDNS HTTP API
	//   - http://192.168.4.1:17321. Static IP address.
	//
	// Typ examples:
	//  - bonsai-growlab
	//  - bonsai-zero-ar-1
	//  - bonsai-zero-a-4
	//  - bonsai-lite-a-4
	//
	// Desc examples:
	//   - room-plant-zamioculcas
	//   - living-room-light-bulb
	Add(uri string, typ string, desc string) error

	// Remove removes the device associated with the provided URI.
	//
	// Parameters:
	//   - uri - unique device identifier.
	Remove(uri string) error

	// GetDesc returns descriptions for registered devices.
	GetDesc() []StoreItem
}
