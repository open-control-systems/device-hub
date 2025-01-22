package sysmdns

import "strings"

// ServiceType represents known mDNS service types.
//
// References:
//   - See common services: http://www.dns-sd.org/serviceTypes.html
//   - https://datatracker.ietf.org/doc/html/rfc2782
//   - https://datatracker.ietf.org/doc/html/rfc6335
//   - https://www.ietf.org/rfc/rfc6763.txt
type ServiceType int

const (
	// ServiceTypeHTTP is used for a HTTP mDNS service type.
	ServiceTypeHTTP ServiceType = iota
)

// String returns string representation of the mDNS service type.
func (t ServiceType) String() string {
	switch t {
	case ServiceTypeHTTP:
		return "_http"
	default:
		return "<none>"
	}
}

// Proto represents known transport protocols.
type Proto int

const (
	// ProtoTCP is used for application protocols that run over TCP.
	ProtoTCP Proto = iota
)

// String returns string representation of the mDNS protocol.
func (p Proto) String() string {
	switch p {
	case ProtoTCP:
		return "_tcp"
	default:
		return "<none>"
	}
}

// ServiceName makes mDNS service name from the provided mDNS service type and protocol.
//
// Examples:
//   - _http._tcp - HTTP service over TCP protocol.
func ServiceName(serviceType ServiceType, proto Proto) string {
	return strings.Join([]string{serviceType.String(), proto.String()}, ".")
}
