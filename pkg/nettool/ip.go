package nettool

import (
	"fmt"
	"net"
)

var ReservedIPs []string

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

func RecycleIP(ip string) error {
	for i, rip := range ReservedIPs {
		if rip == ip {
			ReservedIPs = append(ReservedIPs[:i], ReservedIPs[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("the IP not found in the reserved IP list")
}
