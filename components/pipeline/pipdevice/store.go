package pipdevice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/device"
	"github.com/open-control-systems/device-hub/components/http/htcore"
	"github.com/open-control-systems/device-hub/components/pipeline/piphttp"
	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/open-control-systems/device-hub/components/system/syscore"
	"github.com/open-control-systems/device-hub/components/system/sysnet"
	"github.com/open-control-systems/device-hub/components/system/syssched"
)

// StoreParams represents various configuration options for a device.
type StoreParams struct {
	HTTP struct {
		// FetchInterval - how often to fetch data from the device.
		FetchInterval time.Duration

		// FetchTimeout - how long to wait for the response from the device.
		FetchTimeout time.Duration
	}
}

// Store allows to add/remove device.
type Store struct {
	ctx             context.Context
	localClock      syscore.SystemClock
	remoteLastClock syscore.SystemClock
	dataHandler     device.DataHandler
	resolveStore    *sysnet.ResolveStore
	params          StoreParams

	mu    sync.Mutex
	db    stcore.DB
	nodes map[string]*storeNode
}

// StoreItem is a description of a single device.
type StoreItem struct {
	URI       string `json:"uri"`
	Desc      string `json:"desc"`
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

// NewStore is an initialization of Store.
//
// Parameters:
//   - ctx - parent context.
//   - closer to register all resources that should be closed.
//   - localClock to handle local UNIX time.
//   - remoteLastClock to get the last persisted UNIX time.
//   - dataHandler to handle device data.
//   - db to persist device registration life-cycle.
//   - params - various configuration options for a device.
func NewStore(
	ctx context.Context,
	localClock syscore.SystemClock,
	remoteLastClock syscore.SystemClock,
	dataHandler device.DataHandler,
	db stcore.DB,
	resolveStore *sysnet.ResolveStore,
	params StoreParams,
) *Store {
	s := &Store{
		ctx:             ctx,
		localClock:      localClock,
		remoteLastClock: remoteLastClock,
		dataHandler:     dataHandler,
		params:          params,
		db:              db,
		resolveStore:    resolveStore,
		nodes:           make(map[string]*storeNode),
	}

	if err := s.restoreNodes(); err != nil {
		core.LogErr.Printf("device-store: failed to restore nodes: %v\n", err)
	}

	return s
}

// Start starts data processing for cached devices.
func (s *Store) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, node := range s.nodes {
		node.start()
	}
}

// Close stops data processing for added devices.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, node := range s.nodes {
		if err := node.close(); err != nil {
			core.LogErr.Printf("device-store: failed to close device: uri=%s err=%v\n",
				node.uri, err)
		}
	}

	s.nodes = nil

	return nil
}

// Add adds the device.
//
// Parameters:
//   - uri - device URI, how device can be reached.
//   - desc - human readable device description.
//
// Remarks:
//   - uri should be unique
//
// URI examples:
//   - http://bonsai-growlab.local/api/v1. mDNS HTTP API
//   - http://192.168.4.1:17321. Static IP address.
//
// Desc examples:
//   - room-plant-zamioculcas
//   - living-room-light-bulb
func (s *Store) Add(uri string, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := hashURI(uri)

	if _, ok := s.nodes[key]; ok {
		return fmt.Errorf("device with uri=%s already exists", uri)
	}

	now := time.Now()

	node, err := s.makeNode(uri, desc, now)
	if err != nil {
		return err
	}

	if blob, err := s.db.Read(key); err == nil {
		var item storageItem

		if err := json.Unmarshal(blob.Data, &item); err != nil {
			return err
		}

		if item.URI != uri {
			return fmt.Errorf("failed to save device info: collision")
		}

		panic(fmt.Sprintf("device-store: failed to add device: invalid state:"+
			" uri=%s desc=%s", uri, desc))
	}

	item := storageItem{
		URI:       uri,
		Desc:      desc,
		Timestamp: now.Unix(),
	}

	buf, err := json.Marshal(item)
	if err != nil {
		return err
	}

	blob := stcore.Blob{
		Data: buf,
	}

	if err := s.db.Write(key, blob); err != nil {
		return fmt.Errorf("failed to persist device information: uri=%s err=%v", uri, err)
	}

	s.nodes[key] = node

	core.LogInf.Printf("device-store: device added: uri=%s desc=%s\n", uri, desc)

	node.start()

	return nil
}

// Remove removes the device associated with the provided URI.
//
// Parameters:
//   - uri - unique device identifier.
func (s *Store) Remove(uri string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := hashURI(uri)

	node, ok := s.nodes[key]
	if !ok {
		return status.StatusNoData
	}

	if err := s.db.Remove(key); err != nil {
		return err
	}

	if err := node.close(); err != nil {
		return fmt.Errorf("failed to stop device: uri=%s err=%v", uri, err)
	}

	delete(s.nodes, key)

	core.LogInf.Printf("device-store: device removed: uri=%s\n", uri)

	return nil
}

// GetDesc returns descriptions for registered devices.
func (s *Store) GetDesc() []StoreItem {
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

func (s *Store) restoreNodes() error {
	return s.db.ForEach(func(key string, blob stcore.Blob) error {
		var item storageItem
		if err := json.Unmarshal(blob.Data, &item); err != nil {
			return err
		}

		node, err := s.makeNode(item.URI, item.Desc, time.Unix(item.Timestamp, 0))
		if err != nil {
			panic(fmt.Sprintf("device-store: failed to restore device:"+
				"uri=%s desc=%s err=%v", item.URI, item.Desc, err))
		}

		s.nodes[key] = node

		core.LogInf.Printf("device-store: device restored: uri=%s desc=%s\n",
			item.URI, item.Desc)

		return nil
	})
}

func (s *Store) makeNode(uri string, desc string, now time.Time) (*storeNode, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	deviceType := parseDeviceType(u.Scheme)
	if deviceType == deviceTypeUnsupported {
		return nil, status.StatusNotSupported
	}

	ctx, cancelFunc := context.WithCancel(s.ctx)
	closer := &core.FanoutCloser{}

	holder := device.NewIDHolder(s.dataHandler)

	runner := syssched.NewAsyncTaskRunner(
		ctx,
		s.newHTTPDevice(
			ctx,
			closer,
			holder,
			s.localClock,
			s.remoteLastClock,
			uri,
			u.Host,
			desc,
		),
		&logErrorHandler{uri: uri, desc: desc},
		s.params.HTTP.FetchInterval,
	)

	closer.Add(desc, runner)

	return &storeNode{
		uri:        uri,
		desc:       desc,
		createdAt:  now.Format(time.RFC1123),
		cancelFunc: cancelFunc,
		closer:     closer,
		holder:     holder,
		runner:     runner,
	}, nil
}

func (s *Store) newHTTPDevice(
	ctx context.Context,
	closer *core.FanoutCloser,
	dataHandler device.DataHandler,
	localClock syscore.SystemClock,
	remoteLastClock syscore.SystemClock,
	baseURL string,
	host string,
	desc string,
) syssched.Task {
	remoteCurrClock := piphttp.NewSystemClock(
		ctx,
		s.makeHTTPClient(closer, baseURL, host, desc),
		baseURL+"/system/time",
		s.params.HTTP.FetchTimeout,
	)

	clockSynchronizer := syscore.NewSystemClockSynchronizer(
		localClock, remoteLastClock, remoteCurrClock)

	return device.NewPollDevice(
		htcore.NewURLFetcher(
			ctx,
			s.makeHTTPClient(closer, baseURL, host, desc),
			baseURL+"/registration",
			s.params.HTTP.FetchTimeout,
		),
		htcore.NewURLFetcher(
			ctx,
			s.makeHTTPClient(closer, baseURL, host, desc),
			baseURL+"/telemetry",
			s.params.HTTP.FetchTimeout,
		),
		dataHandler,
		clockSynchronizer,
	)
}

func (s *Store) makeHTTPClient(
	closer *core.FanoutCloser,
	uri string,
	host string,
	desc string,
) *htcore.HTTPClient {
	if !strings.Contains(uri, ".local") {
		return htcore.NewDefaultClient()
	}

	s.resolveStore.Add(host)

	closer.Add("resolve-store-"+desc, core.FuncCloser(func() error {
		s.resolveStore.Remove(host)

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

type storageItem struct {
	URI       string `json:"uri"`
	Desc      string `json:"desc"`
	Timestamp int64  `json:"ts"`
}

type storeNode struct {
	uri        string
	desc       string
	createdAt  string
	cancelFunc context.CancelFunc
	closer     *core.FanoutCloser
	holder     *device.IDHolder
	runner     *syssched.AsyncTaskRunner
}

func (s *storeNode) start() {
	s.runner.Start()
}

func (s *storeNode) close() error {
	s.cancelFunc()

	return s.closer.Close()
}

func hashURI(uri string) string {
	hash := sha256.Sum256([]byte(uri))
	return hex.EncodeToString(hash[:])
}
