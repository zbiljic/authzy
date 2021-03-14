package netutil

import (
	"errors"
	"net"
)

var (
	// ErrMissingPrivateIP error is returned when it is not possible to
	// determine local IP address.
	ErrMissingPrivateIP = errors.New("failed to find local IP address")
)

// PrivateIP returns private IP address of the machine.
func PrivateIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", nil
	}

	for _, address := range addrs {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip := ipNet.IP
				_, cidr24BitBlock, _ := net.ParseCIDR("10.0.0.0/8")
				_, cidr20BitBlock, _ := net.ParseCIDR("172.16.0.0/12")
				_, cidr16BitBlock, _ := net.ParseCIDR("192.168.0.0/16")
				private := cidr24BitBlock.Contains(ip) || cidr20BitBlock.Contains(ip) || cidr16BitBlock.Contains(ip)
				if private {
					return ip.String(), nil
				}
			}
		}
	}

	return "", ErrMissingPrivateIP
}
