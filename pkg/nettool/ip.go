package nettool

import (
	"net"
)

type AllocatedIP struct {
	Version string `json:"version"`
	Address string `json:"address"`
	Gateway string `json:"gateway"`
}

func GetAllIPs(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		tempIPNet := &net.IPNet{IP: ip, Mask: ipnet.Mask}
		ips = append(ips, tempIPNet.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	ip = ip.To4()
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] > 0 {
			break
		}
	}
}
