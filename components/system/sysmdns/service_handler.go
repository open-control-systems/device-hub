package sysmdns

// ServiceHandler is a mDNS service handler.
type ServiceHandler interface {
	// HandleService handles the mDNS service discovered over local network.
	HandleService(service Service) error
}
