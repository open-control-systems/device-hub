package sysmdns

import "net"

// Service is mDNS service.
type Service struct {
	// Instance is the mDNS service instance name, e.g. "Bonsai GrowLab Firmware".
	Instance string

	// Name is the mDNS service name, e.g. "_http._tcp".
	Name string

	// Hostname is the machine DNS name, e.g. "bonsai-growlab.local".
	Hostname string

	// Port is the mDNS service port, e.g. 80.
	Port int

	// TxtRecords are the service txt records, e.g. ["api_base_path=/api/"]
	TxtRecords []string

	// AddrsIPv4 are the IPv4 addresses for the service.
	AddrsIPv4 []net.IP

	// AddrsIPv6 are the IPv6 addresses for the service.
	AddrsIPv6 []net.IP
}

// AddTxtRecord adds txt record to the service.
func (s *Service) AddTxtRecord(key string, value string) {
	s.TxtRecords = append(s.TxtRecords, key+"="+value)
}
