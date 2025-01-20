package pipdevice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/device"
	"github.com/open-control-systems/device-hub/components/status"
	"github.com/open-control-systems/device-hub/components/storage/stcore"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// PipelineStoreParams represents various configuration options for device pipelines.
type PipelineStoreParams struct {
	HTTP struct {
		// FetchInterval - how often to fetch data from the device.
		FetchInterval time.Duration

		// FetchTimeout - how long to wait for the response from the device.
		FetchTimeout time.Duration
	}
}

// PipelineStore allows to add/remove device pipelines.
type PipelineStore struct {
	ctx             context.Context
	localClock      syscore.SystemClock
	remoteLastClock syscore.SystemClock
	dataHandler     device.DataHandler
	errorReporter   device.ErrorReporter
	params          PipelineStoreParams

	mu    sync.Mutex
	db    stcore.DB
	nodes map[string]*storeNode
}

// PipelineStoreItem is a description of a single device.
type PipelineStoreItem struct {
	URI       string `json:"uri"`
	Desc      string `json:"desc"`
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

// NewPipelineStore is a PipelineStore initialization.
//
// Parameters:
//   - ctx - parent context.
//   - closer to register all resources that should be closed.
//   - localClock to handle local UNIX time.
//   - remoteLastClock to get the last persisted UNIX time.
//   - dataHandler to handle device data.
//   - errorReporter to report errors from the device.
//   - db to persist device registration life-cycle.
//   - params - various configuration options for device pipelines.
func NewPipelineStore(
	ctx context.Context,
	localClock syscore.SystemClock,
	remoteLastClock syscore.SystemClock,
	dataHandler device.DataHandler,
	errorReporter device.ErrorReporter,
	db stcore.DB,
	params PipelineStoreParams,
) *PipelineStore {
	s := &PipelineStore{
		ctx:             ctx,
		localClock:      localClock,
		remoteLastClock: remoteLastClock,
		dataHandler:     dataHandler,
		errorReporter:   errorReporter,
		params:          params,
		db:              db,
		nodes:           make(map[string]*storeNode),
	}

	if err := s.restoreNodes(); err != nil {
		core.LogErr.Printf("device-pipeline-store: failed to restore nodes: %v\n", err)
	}

	return s
}

// Start starts data processing for cached devices.
func (s *PipelineStore) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, node := range s.nodes {
		node.start()
	}
}

// Close stops data processing for added devices.
func (s *PipelineStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, node := range s.nodes {
		if err := node.close(); err != nil {
			core.LogErr.Printf("pipeline-store: failed to close device: uri=%s err=%v\n",
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
func (s *PipelineStore) Add(uri string, desc string) error {
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

		panic(fmt.Sprintf("device-pipeline-store: failed to add device: invalid state:"+
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

	node.start()

	return nil
}

// Remove removes the device associated with the provided URI.
//
// Parameters:
//   - uri - unique device identifier.
func (s *PipelineStore) Remove(uri string) error {
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
		return fmt.Errorf("failed to stop HTTP pipeline: uri=%s err=%v", uri, err)
	}

	delete(s.nodes, key)

	return nil
}

// GetDesc returns descriptions for registered devices.
func (s *PipelineStore) GetDesc() []PipelineStoreItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	var items []PipelineStoreItem

	for _, node := range s.nodes {
		items = append(items, PipelineStoreItem{
			URI:       node.uri,
			Desc:      node.desc,
			ID:        node.holder.Get(),
			CreatedAt: node.createdAt,
		})
	}

	return items
}

func (s *PipelineStore) restoreNodes() error {
	return s.db.ForEach(func(key string, blob stcore.Blob) error {
		var item storageItem
		if err := json.Unmarshal(blob.Data, &item); err != nil {
			return err
		}

		node, err := s.makeNode(item.URI, item.Desc, time.Unix(item.Timestamp, 0))
		if err != nil {
			panic(fmt.Sprintf("device-pipeline-store: failed to restore device:"+
				"uri=%s desc=%s err=%v", item.URI, item.Desc, err))
		}

		s.nodes[key] = node

		core.LogInf.Printf("device-pipeline-store: device restored: uri=%s desc=%s\n",
			item.URI, item.Desc)

		return nil
	})
}

func (s *PipelineStore) makeNode(uri string, desc string, now time.Time) (*storeNode, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	deviceType := parseDeviceType(u.Scheme)
	if deviceType == unsupportedDevice {
		return nil, status.StatusNotSupported
	}

	ctx, cancelFunc := context.WithCancel(s.ctx)
	closer := &core.FanoutCloser{}

	holder := device.NewIDHolder(s.dataHandler)

	pipeline := NewHTTPPipeline(
		ctx,
		closer,
		holder,
		s.errorReporter,
		s.localClock,
		s.remoteLastClock,
		HTTPPipelineParams{
			BaseURL:       uri,
			Desc:          desc,
			FetchInterval: s.params.HTTP.FetchInterval,
			FetchTimeout:  s.params.HTTP.FetchTimeout,
		})
	closer.Add(desc, pipeline)

	return &storeNode{
		uri:        uri,
		desc:       desc,
		createdAt:  now.Format(time.RFC1123),
		cancelFunc: cancelFunc,
		closer:     closer,
		holder:     holder,
		pipeline:   pipeline,
	}, nil
}

type deviceType int

const (
	unsupportedDevice = iota
	httpDevice
)

func parseDeviceType(scheme string) deviceType {
	if scheme == "http" || scheme == "https" {
		return httpDevice
	}

	return unsupportedDevice
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
	pipeline   *HTTPPipeline
}

func (s *storeNode) start() {
	s.pipeline.Start()
}

func (s *storeNode) close() error {
	s.cancelFunc()

	return s.closer.Close()
}

func hashURI(uri string) string {
	hash := sha256.Sum256([]byte(uri))
	return hex.EncodeToString(hash[:])
}
