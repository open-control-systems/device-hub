package syssched

import "github.com/open-control-systems/device-hub/components/system/syscore"

// FanoutStopper propagates stop call to the underlying stoppers.
type FanoutStopper struct {
	nodes []node
}

// Add addes stopper with id to be notified when the stop event is happened.
func (s *FanoutStopper) Add(id string, stopper Stopper) {
	s.nodes = append(s.nodes, node{id: id, s: stopper})
}

// Stop stops all registered stoppers.
func (s *FanoutStopper) Stop() error {
	for _, node := range s.nodes {
		if err := node.s.Stop(); err != nil {
			syscore.LogErr.Printf(
				"fanout-stopper: failed to stop: id=%s err=%v\n", node.id, err)
		}
	}

	return nil
}

type node struct {
	id string
	s  Stopper
}
