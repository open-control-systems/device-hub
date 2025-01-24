package syssched

// FanoutStarter to start all at once.
type FanoutStarter struct {
	starters []Starter
}

// Start starts all the registered starters.
func (s *FanoutStarter) Start() {
	for _, starter := range s.starters {
		starter.Start()
	}
}

// Add adds the starter to be started on Start() call.
func (s *FanoutStarter) Add(starter Starter) {
	s.starters = append(s.starters, starter)
}
