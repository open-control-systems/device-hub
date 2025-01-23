package sysmdns

import "net"

// Service is a single mDNS service discovered on the local network.
type Service interface {
	// Instance returns mDNS service instance name, e.g. "Bonsai GrowLab Firmware".
	Instance() string

	// Service returns mDNS service name, e.g. "_http._tcp".
	Name() string

	// Hostname returns host machine DNS name, e.g. "bonsai-growlab.local".
	Hostname() string

	// Port returns service port, e.g. 80.
	Port() int

	// TxtRecords returns service txt records, e.g. ["api_base_path=/api/", "api_version=v1"]
	TxtRecords() []string

	// Host machine IP addresses. Service should contain at least one resolved IP address.
	Addrs() []net.IP
}
