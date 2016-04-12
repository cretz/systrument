package util

import "net"

func LocalIPAddresses() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	ret := []string{}
	for _, addr := range addrs {
		switch addr := addr.(type) {
		case *net.IPNet:
			ret = append(ret, addr.IP.String())
		case *net.IPAddr:
			ret = append(ret, addr.IP.String())
		}
	}
	return ret, nil
}
