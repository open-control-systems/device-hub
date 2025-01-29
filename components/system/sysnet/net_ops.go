package sysnet

import "net"

// FilterInterfaces filters network interfaces.
//
// Remarks:
//   - Loopback and down interfaces are filtered by default.
func FilterInterfaces(fn func(iface net.Interface) bool) ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var ret []net.Interface

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		if !fn(iface) {
			continue
		}

		ret = append(ret, iface)
	}

	return ret, nil
}
