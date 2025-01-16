package pipdevice

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/open-control-systems/device-hub/components/core"
	"github.com/open-control-systems/device-hub/components/device"
	"github.com/open-control-systems/device-hub/components/system/syscore"
)

// PipelineStore allows to add/remove device pipelines.
type PipelineStore struct {
	ctx             context.Context
	localClock      syscore.SystemClock
	remoteLastClock syscore.SystemClock
	dataHandler     device.DataHandler

	mu    sync.Mutex
	nodes map[string]*storeNode
}

// PipelineStoreItem is a description of a single device.
type PipelineStoreItem struct {
	URI      string `json:"uri"`
	ID       string `json:"id"`
	DeviceID string `json:"device_id"`
}

// NewPipelineStore is a PipelineStore initialization.
//
// Parameters:
//   - ctx - parent context.
//   - closer to register all resources that should be closed.
//   - localClock to handle local UNIX time.
//   - remoteLastClock to get the last persisted UNIX time.
//   - dataHandler to handle device data.
func NewPipelineStore(
	ctx context.Context,
	localClock syscore.SystemClock,
	remoteLastClock syscore.SystemClock,
	dataHandler device.DataHandler,
) *PipelineStore {
	return &PipelineStore{
		ctx:             ctx,
		localClock:      localClock,
		remoteLastClock: remoteLastClock,
		dataHandler:     dataHandler,
		nodes:           make(map[string]*storeNode),
	}
}

// Close stops data processing for added devices.
func (s *PipelineStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for uri, node := range s.nodes {
		if err := node.close(); err != nil {
			core.LogErr.Printf("pipeline-store: failed to close device: URI=%v err=%v\n",
				uri, err)
		}
	}

	s.nodes = nil

	return nil
}

// Add adds the device.
//
// Parameters:
//   - uri - device URI, how device can be reached.
//   - id - human readable device identifier.
//
// Remarks:
//   - uri should be unique
//
// URI examples:
//   - http://bonsai-growlab.local/api/v1. mDNS HTTP API
//   - http://192.168.4.1:17321. Static IP address.
//
// ID examples:
//   - room-plant-zamioculcas
//   - living-room-light-bulb
func (s *PipelineStore) Add(uri string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.nodes[uri]; ok {
		return fmt.Errorf("error: device with uri=%s already exists", uri)
	}

	ctx, cancelFunc := context.WithCancel(s.ctx)
	closer := &core.FanoutCloser{}

	node := &storeNode{
		cancelFunc: cancelFunc,
		closer:     closer,
		pipeline: NewHTTPPipeline(
			ctx,
			closer,
			s.dataHandler,
			s.localClock,
			s.remoteLastClock,
			HTTPPipelineParams{
				ID:            id,
				BaseURL:       uri,
				FetchInterval: time.Second * 5,
				FetchTimeout:  time.Second * 10,
			}),
	}

	s.nodes[uri] = node

	node.pipeline.Start()

	return nil
}

// Remove removes the device associated with the provided URI.
//
// Parameters:
//   - uri - unique device identifier.
func (s *PipelineStore) Remove(uri string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[uri]
	if !ok {
		return fmt.Errorf("device with uri=%s doesn't exist", uri)
	}

	if err := node.close(); err != nil {
		return fmt.Errorf("failed to stop HTTP pipeline: uri=%v err=%v", uri, err)
	}

	delete(s.nodes, uri)

	return nil
}

// Get returns descriptions for registered devices.
func (s *PipelineStore) GetDesc() []PipelineStoreItem {
	s.mu.Lock()
	defer s.mu.Unlock()

	var items []PipelineStoreItem

	for uri, node := range s.nodes {
		items = append(items, PipelineStoreItem{
			URI:      uri,
			ID:       node.pipeline.GetID(),
			DeviceID: node.pipeline.GetDeviceID(),
		})
	}

	return items
}

type storeNode struct {
	cancelFunc context.CancelFunc
	closer     *core.FanoutCloser
	pipeline   *HTTPPipeline
}

func (s *storeNode) close() error {
	s.cancelFunc()

	return s.closer.Close()
}
