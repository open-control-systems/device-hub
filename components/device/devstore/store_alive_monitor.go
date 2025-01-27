package devstore

import (
	"sync"
	"time"

	"github.com/open-control-systems/device-hub/components/system/syscore"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

// StoreAliveMonitor monitors the operational health of devices. If a device isn't
// active for a period of time, it is considered to be inactive and is removed.
type StoreAliveMonitor struct {
	maxInactiveInterval time.Duration
	clock               syscore.MonotonicClock
	store               Store

	mu      sync.Mutex
	devices map[string]time.Time
}

// NewStoreAliveMonitor is an initialization of StoreAliveMonitor.
//
// Parameters:
//   - store to automatically add/remove devices.
//   - clock to measure time for how long device is inactive.
//   - maxInactiveInterval is maximum allowed interval for a device to be inactive.
func NewStoreAliveMonitor(
	clock syscore.MonotonicClock,
	store Store,
	maxInactiveInterval time.Duration,
) *StoreAliveMonitor {
	monitor := &StoreAliveMonitor{
		maxInactiveInterval: maxInactiveInterval,
		store:               store,
		clock:               clock,
		devices:             make(map[string]time.Time),
	}

	monitor.restoreDevices()

	return monitor
}

// Monitor returns the alive notifier for the device associated with the provided URI.
func (m *StoreAliveMonitor) Monitor(uri string) syssched.AliveNotifier {
	return &storeAliveNotifier{
		uri:     uri,
		monitor: m,
	}
}

// Add adds the device to the underlying store and starts monitoring its well-being.
func (m *StoreAliveMonitor) Add(uri string, desc string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.store.Add(uri, desc); err != nil {
		return err
	}

	m.devices[uri] = m.clock.Now()

	return nil
}

// Remove removes the device associated with the provided URI.
func (m *StoreAliveMonitor) Remove(uri string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.store.Remove(uri); err != nil {
		return err
	}

	delete(m.devices, uri)

	return nil
}

// GetDesc returns descriptions for registered devices.
func (m *StoreAliveMonitor) GetDesc() []StoreItem {
	return m.store.GetDesc()
}

// HandleError handles Run() error.
func (*StoreAliveMonitor) HandleError(err error) {
	syscore.LogErr.Printf("store-alive-monitor: failed to verify inactive devices: %v\n", err)
}

// Run verifies if added devices are still alive.
func (m *StoreAliveMonitor) Run() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for uri, updateTime := range m.devices {
		now := m.clock.Now()

		diff := now.Sub(updateTime)
		if diff < m.maxInactiveInterval {
			continue
		}

		syscore.LogWrn.Printf("store-alive-monitor: removing inactive device:"+
			" uri=%s cur_inactive=%s max_inactive=%s\n", uri, diff, m.maxInactiveInterval)

		if err := m.store.Remove(uri); err != nil {
			return err
		}

		delete(m.devices, uri)
	}

	return nil
}

func (m *StoreAliveMonitor) restoreDevices() {
	for _, desc := range m.GetDesc() {
		m.devices[desc.URI] = m.clock.Now()
	}
}

func (m *StoreAliveMonitor) notifyAlive(uri string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.devices[uri] = m.clock.Now()
}

type storeAliveNotifier struct {
	uri     string
	monitor *StoreAliveMonitor
}

func (n *storeAliveNotifier) NotifyAlive() {
	n.monitor.notifyAlive(n.uri)
}
