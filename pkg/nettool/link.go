package nettool

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/vishvananda/netlink"
)

// CreateOrUpdateBridge creates or updates bridge and sets its as the gateway of container network
func CreateOrUpdateBridge(name, ip string, mtu int) (*netlink.Bridge, error) {
	br := &netlink.Bridge{
		LinkAttrs: netlink.LinkAttrs{
			Name:   name,
			MTU:    mtu,
			TxQLen: -1,
		},
	}

	// ip address for bridge
	ipaddr, ipnet, err := net.ParseCIDR(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ip address %q: %v", ip, err)
	}
	ipnet.IP = ipaddr
	addr := &netlink.Addr{IPNet: ipnet}

	l, err := netlink.LinkByName(name)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); ok {
			if err := netlink.LinkAdd(br); err != nil {
				return nil, fmt.Errorf("failed to create bridge %q with error: %v", name, err)
			}
			if err = netlink.AddrAdd(br, addr); err != nil {
				return nil, fmt.Errorf("failed to set address: %q for bridge %q: %v", addr, name, err)
			}
			if err = netlink.LinkSetUp(br); err != nil {
				return nil, fmt.Errorf("failed to set bridge %q up: %v", name, err)
			}
			return br, nil
		} else {
			return nil, fmt.Errorf("could not find link %s: %v", name, err)
		}
	}
	currentBr, ok := l.(*netlink.Bridge)
	if !ok {
		return nil, fmt.Errorf("link %s already exists but is not a bridge type", name)
	}
	addrs, err := netlink.AddrList(currentBr, netlink.FAMILY_ALL)
	if err != nil {
		return nil, fmt.Errorf("failed to list address for bridge %q: %v", name, err)
	}
	switch {
	case len(addrs) > 1:
		return nil, fmt.Errorf("unexpected addresses for bridge %q: %v", name, addrs)
	case len(addrs) == 1 && !addr.Equal(addrs[0]):
		if err = netlink.AddrReplace(currentBr, addr); err != nil {
			return nil, fmt.Errorf("failed to replace address: %q for bridge %q: %v", addr, name, err)
		}
	default:
		if err = netlink.AddrAdd(currentBr, addr); err != nil {
			return nil, fmt.Errorf("failed to set address: %q for bridge %q: %v", addr, name, err)
		}
	}
	return currentBr, nil
}

// SetupVeth sets up a pair of virtual ethernet devices in container netns
// and then move the host-side veth into the hostNS namespace.
func SetupVeth(netns ns.NetNS, br *netlink.Bridge, ifName, ip string, mtu int) error {
	err := netns.Do(func(hostNS ns.NetNS) error {
		hostVethName, veth, err := makeVethPair(ifName, mtu)
		if err != nil {
			return err
		}
		ipaddr, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			return fmt.Errorf("failed to parse ip address %q: %v", ip, err)
		}
		ipnet.IP = ipaddr
		if err = netlink.AddrAdd(veth, &netlink.Addr{IPNet: ipnet}); err != nil {
			return fmt.Errorf("failed to set address: %q for veth %q: %v", ipnet, ifName, err)
		}
		if err = netlink.LinkSetUp(veth); err != nil {
			return fmt.Errorf("failed to set veth %q up: %v", ifName, err)
		}

		hostVeth, err := netlink.LinkByName(hostVethName)
		if err != nil {
			return fmt.Errorf("failed to lookup hostveth %q: %v", hostVethName, err)
		}
		if err = netlink.LinkSetNsFd(hostVeth, int(hostNS.Fd())); err != nil {
			return fmt.Errorf("failed to set hostveth %q to host netns: %v", hostVethName, err)
		}
		err = hostNS.Do(func(_ ns.NetNS) error {
			hostVeth, err = netlink.LinkByName(hostVethName)
			if err != nil {
				return fmt.Errorf("failed to lookup hostveth %q in %q: %v", hostVethName, hostNS.Path(), err)
			}
			if err = netlink.LinkSetUp(hostVeth); err != nil {
				return fmt.Errorf("failed to set veth %q up: %v", hostVethName, err)
			}

			// connect the host veth to the bridge
			if err = netlink.LinkSetMaster(hostVeth, br); err != nil {
				return fmt.Errorf("failed to connect %q to bridge %v: %v", hostVethName, br.Name, err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to set hostveth %q: %v", hostVethName, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to set veth %q: %v", ifName, err)
	}
	return nil
}

// makeVethPair create veth pair and peer name with random string with "veth" prefix
func makeVethPair(name string, mtu int) (string, *netlink.Veth, error) {
	peerName, err := generateRandomVethName()
	if err != nil {
		return "", nil, err
	}
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: name,
			//Flags: net.FlagUp,
			MTU: mtu,
		},
		PeerName: peerName,
	}

	err = netlink.LinkAdd(veth)
	switch {
	case err == nil:
		return peerName, veth, nil
	case os.IsExist(err):
		return "", nil, fmt.Errorf("failed to create veth because veth name %q already exists", name)
	default:
		return "", nil, fmt.Errorf("failed to create veth %q with error: %v", name, err)
	}
}

// generateRandomVethName generate random string start with "veth"
func generateRandomVethName() (string, error) {
	rd := make([]byte, 4)
	if _, err := rand.Read(rd); err != nil {
		return "", fmt.Errorf("failed to gererate random veth name: %v", err)
	}

	return fmt.Sprintf("veth%x", rd), nil
}

// GetVethIPInNS return the IP address for the ifName in container Namespace
func GetVethIPInNS(netns ns.NetNS, ifName string) (string, error) {
	ip := ""
	err := netns.Do(func(_ ns.NetNS) error {
		l, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("failed to lookup veth %q in %q: %v", ifName, netns.Path(), err)
		}
		veth, ok := l.(*netlink.Veth)
		if !ok {
			return fmt.Errorf("link %s already exists but is not a veth type", ifName)
		}
		addrs, err := netlink.AddrList(veth, netlink.FAMILY_ALL)
		if err != nil {
			return fmt.Errorf("failed to list address for veth %q: %v", ifName, err)
		}
		switch {
		case len(addrs) > 1:
			return fmt.Errorf("unexpected addresses for veth %q: %v", ifName, addrs)
		case len(addrs) == 1:
			ip = addrs[0].IPNet.String()
		default:
			return fmt.Errorf("no address set for veth %q", ifName)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return ip, nil
}
