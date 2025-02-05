package devstore

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/open-control-systems/device-hub/components/device/devcore"
	"github.com/open-control-systems/device-hub/components/http/htcore"
	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/open-control-systems/device-hub/components/system/syscore"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

// CacheStoreParams represents various configuration options for a cache store.
type CacheStoreParams struct {
	HTTP struct {
		// FetchInterval - how often to fetch data from the device.
		FetchInterval time.Duration

		// FetchTimeout - how long to wait for the response from the device.
		FetchTimeout time.Duration
	}

	TimeSync struct {
		// MaxDriftInterval is a maximum allowed time difference between local
		// and device UNIX time.
		MaxDriftInterval time.Duration
	}
}

// CacheStore allows to cache information about the added devices in the persistent storage.
type CacheStore struct {
	ctx             context.Context
	localClock      syscore.SystemClock
	remoteLastClock syscore.SystemClock
	dataHandler     devcore.DataHandler
	resolveStore    *sysnet.ResolveStore
	aliveMonitor    AliveMonitor
	params          CacheStoreParams

	mu    sync.Mutex
	db    stcore.DB
	nodes map[string]*storeNode
}

// NewCacheStore is an initialization of CacheStore.
//
// Parameters:
//   - ctx - parent context.
//   - localClock to handle local UNIX time.
//   - remoteLastClock to get the last persisted UNIX time.
//   - dataHandler to handle device data.
//   - db to persist device registration life-cycle.
//   - resolveStore to manage device host resolving.
//   - params - various configuration options for a cache store.
func NewCacheStore(
	ctx context.Context,
	localClock syscore.SystemClock,
	remoteLastClock syscore.SystemClock,
	dataHandler devcore.DataHandler,
	db stcore.DB,
	resolveStore *sysnet.ResolveStore,
	params CacheStoreParams,
) *CacheStore {
	s := &CacheStore{
		ctx:             ctx,
		localClock:      localClock,
		remoteLastClock: remoteLastClock,
		dataHandler:     dataHandler,
		params:          params,
		db:              db,
		resolveStore:    resolveStore,
		nodes:           make(map[string]*storeNode),
	}

	s.restoreNodes()

	return s
}

// SetAliveMonitor sets the device inactivity monitor.
func (s *CacheStore) SetAliveMonitor(monitor AliveMonitor) {
	s.aliveMonitor = monitor
}

// Start starts data processing for cached devices.
func (s *CacheStore) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, node := range s.nodes {
		if err := node.start(); err != nil {
			return err
		}
	}

	return nil
}

// Stop stops data processing for added devices.
func (s *CacheStore) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, node := range s.nodes {
		if err := node.stop(); err != nil {
			syscore.LogErr.Printf("failed to stop device: uri=%s err=%v\n",
				node.uri, err)
		}
	}

	s.nodes = nil

	return nil
}

// Add caches the device information in the persistent storage.
func (s *CacheStore) Add(uri string, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.nodes[uri]; ok {
		return ErrDeviceExist
	}

	now := time.Now()

	node, err := s.makeNode(uri, desc, now)
	if err != nil {
		return err
	}

	item := StorageItem{
		Desc:      desc,
		Timestamp: now.Unix(),
	}

	buf, err := item.MarshalBinary()
	if err != nil {
		return err
	}

	if err := s.db.Write(uri, buf); err != nil {
		return fmt.Errorf("failed to persist device information: uri=%s err=%v", uri, err)
	}

	if err := node.start(); err != nil {
		return err
	}

	s.nodes[uri] = node

	syscore.LogInf.Printf("device added: uri=%s desc=%s\n", uri, desc)

	return nil
}

// Remove removes the device if it exists.
func (s *CacheStore) Remove(uri string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[uri]
	if !ok {
		return status.StatusNoData
	}

	if err := s.db.Remove(uri); err != nil {
		return err
	}

	if err := node.stop(); err != nil {
		return fmt.Errorf("failed to stop device: uri=%s err=%v", uri, err)
	}

	delete(s.nodes, uri)

	syscore.LogInf.Printf("device removed: uri=%s\n", uri)

	return nil
}

// GetDesc returns descriptions for registered devices.
func (s *CacheStore) GetDesc() []StoreItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	var items []StoreItem

	for _, node := range s.nodes {
		items = append(items, StoreItem{
			URI:       node.uri,
			Desc:      node.desc,
			ID:        node.holder.Get(),
			CreatedAt: node.createdAt,
		})
	}

	return items
}

func (s *CacheStore) restoreNodes() {
	var unrestoredURIs []string

	err := s.db.ForEach(func(uri string, buf []byte) error {
		if err := s.restoreNode(uri, buf); err != nil {
			syscore.LogErr.Printf("failed to restore device: uri=%s err=%v\n",
				uri, err)

			unrestoredURIs = append(unrestoredURIs, uri)
		}

		return nil
	})
	if err != nil {
		panic("failed to restore nodes: invalid state: " + err.Error())
	}

	if len(unrestoredURIs) == 0 {
		return
	}

	for _, uri := range unrestoredURIs {
		if err := s.db.Remove(uri); err != nil {
			syscore.LogErr.Printf("failed to remove unrestored device:"+
				" uri=%s err=%v\n", uri, err)
		} else {
			syscore.LogErr.Printf("unrestored device removed: uri=%s\n", uri)
		}
	}
}

func (s *CacheStore) restoreNode(uri string, buf []byte) error {
	var item StorageItem
	if _, err := item.Unmarshal(buf); err != nil {
		return err
	}

	node, err := s.makeNode(uri, item.Desc, time.Unix(item.Timestamp, 0))
	if err != nil {
		return err
	}

	s.nodes[uri] = node

	syscore.LogInf.Printf("device restored: uri=%s desc=%s\n",
		uri, item.Desc)

	return nil
}

func (s *CacheStore) makeNode(uri string, desc string, now time.Time) (*storeNode, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	deviceType := parseDeviceType(u.Scheme)

	switch deviceType {
	case deviceTypeHTTP:
		return s.makeNodeHTTP(u, uri, desc, now)
	default:
		return nil, status.StatusNotSupported
	}
}

func (s *CacheStore) makeNodeHTTP(
	u *url.URL,
	uri string,
	desc string,
	now time.Time,
) (*storeNode, error) {
	if u.Port() == "" {
		return nil, fmt.Errorf("HTTP port is missed")
	}

	ctx, cancelFunc := context.WithCancel(s.ctx)
	stopper := &syssched.FanoutStopper{}

	holder := devcore.NewIDHolder(s.dataHandler)

	runner := syssched.NewAsyncTaskRunner(
		ctx,
		s.newHTTPDevice(
			ctx,
			stopper,
			holder,
			s.localClock,
			s.remoteLastClock,
			uri,
			desc,
			u.Hostname(),
		),
		&logErrorHandler{uri: uri, desc: desc},
		syssched.AsyncTaskRunnerParams{
			UpdateInterval: s.params.HTTP.FetchInterval,
		},
	)

	stopper.Add(desc, runner)

	return &storeNode{
		uri:        uri,
		desc:       desc,
		createdAt:  now.Format(time.RFC1123),
		cancelFunc: cancelFunc,
		stopper:    stopper,
		holder:     holder,
		runner:     runner,
	}, nil
}

func (s *CacheStore) newHTTPDevice(
	ctx context.Context,
	stopper *syssched.FanoutStopper,
	dataHandler devcore.DataHandler,
	localClock syscore.SystemClock,
	remoteLastClock syscore.SystemClock,
	uri string,
	desc string,
	hostname string,
) syssched.Task {
	remoteCurrClock := htcore.NewSystemClock(
		ctx,
		s.makeHTTPClient(stopper, uri, desc, hostname),
		uri+"/system/time",
		s.params.HTTP.FetchTimeout,
	)

	clockSynchronizer := syscore.NewSystemClockSynchronizer(
		localClock, remoteLastClock, remoteCurrClock)

	var clockVerifier devcore.TimeVerifier
	if maxDriftInterval := s.params.TimeSync.MaxDriftInterval; maxDriftInterval == 0 {
		clockVerifier = &devcore.BasicTimeVerifier{}
	} else {
		clockVerifier = devcore.NewDriftTimeVerifier(localClock, maxDriftInterval)
	}

	task := devcore.NewPollDevice(
		htcore.NewURLFetcher(
			ctx,
			s.makeHTTPClient(stopper, uri, desc, hostname),
			uri+"/registration",
			s.params.HTTP.FetchTimeout,
		),
		htcore.NewURLFetcher(
			ctx,
			s.makeHTTPClient(stopper, uri, desc, hostname),
			uri+"/telemetry",
			s.params.HTTP.FetchTimeout,
		),
		dataHandler,
		clockSynchronizer,
		clockVerifier,
	)

	if s.aliveMonitor != nil {
		notifier := s.aliveMonitor.Monitor(uri)

		return syssched.NewTaskAliveNotifier(task, notifier)
	}

	return task
}

func (s *CacheStore) makeHTTPClient(
	stopper *syssched.FanoutStopper,
	uri string,
	desc string,
	hostname string,
) *htcore.HTTPClient {
	if !strings.Contains(uri, ".local") {
		return htcore.NewDefaultClient()
	}

	s.resolveStore.Add(hostname)

	stopper.Add("resolve-store-"+desc, syssched.FuncStopper(func() error {
		s.resolveStore.Remove(hostname)

		return nil
	}))

	return htcore.NewResolveClient(s.resolveStore)
}

type deviceType int

const (
	deviceTypeUnsupported deviceType = iota
	deviceTypeHTTP
)

func parseDeviceType(scheme string) deviceType {
	if scheme == "http" || scheme == "https" {
		return deviceTypeHTTP
	}

	return deviceTypeUnsupported
}

type storeNode struct {
	uri        string
	desc       string
	createdAt  string
	cancelFunc context.CancelFunc
	stopper    *syssched.FanoutStopper
	holder     *devcore.IDHolder
	runner     *syssched.AsyncTaskRunner
}

func (s *storeNode) start() error {
	return s.runner.Start()
}

func (s *storeNode) stop() error {
	s.cancelFunc()

	return s.stopper.Stop()
}
