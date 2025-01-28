package syssched

// FanoutStarter to start all at once.
type FanoutStarter struct {
	starters []Starter
}

// Start starts all the registered starters.
func (s *FanoutStarter) Start() error {
	for _, starter := range s.starters {
		if err := starter.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Add adds the starter to be started on Start() call.
func (s *FanoutStarter) Add(starter Starter) {
	s.starters = append(s.starters, starter)
}
