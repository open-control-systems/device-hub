package sysnet

import "strings"

// MdnsServiceType represents known mDNS service types.
//
// References:
//   - See common services: http://www.dns-sd.org/serviceTypes.html
//   - https://datatracker.ietf.org/doc/html/rfc2782
//   - https://datatracker.ietf.org/doc/html/rfc6335
//   - https://www.ietf.org/rfc/rfc6763.txt
type MdnsServiceType int

const (
	// MdnsServiceTypeHTTP is used for a HTTP mDNS service type.
	MdnsServiceTypeHTTP MdnsServiceType = iota
)

// String returns string representation of the mDNS service type.
func (s MdnsServiceType) String() string {
	switch s {
	case MdnsServiceTypeHTTP:
		return "_http"
	default:
		return "<none>"
	}
}

// MdnsProto represents known transport protocols.
type MdnsProto int

const (
	// MdnsProtoTCP is used for application protocols that run over TCP.
	MdnsProtoTCP MdnsProto = iota
)

// String returns string representation of the mDNS protocol.
func (p MdnsProto) String() string {
	switch p {
	case MdnsProtoTCP:
		return "_tcp"
	default:
		return "<none>"
	}
}

// MdnsServiceName makes mDNS service name from the provided mDNS service type and protocol.
//
// Examples:
//   - _http._tcp - HTTP service over TCP protocol.
func MdnsServiceName(serviceType MdnsServiceType, proto MdnsProto) string {
	return strings.Join([]string{serviceType.String(), proto.String()}, ".")
}
