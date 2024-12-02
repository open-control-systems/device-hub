package core

// Propagate close call to the underlying closers.
type FanoutCloser struct {
	closers []node
}

// Add closer with id to be notified when the close event is happened.
func (c *FanoutCloser) Add(id string, closer Closer) {
	c.closers = append(c.closers, node{id: id, c: closer})
}

// Close all.
func (c *FanoutCloser) Close() error {
	for _, node := range c.closers {
		if err := node.c.Close(); err != nil {
			LogErr.Printf("fanout-closer: failed to close: id=%s err=%v\n", node.id, err)
		}
	}

	return nil
}

type node struct {
	id string
	c  Closer
}
