package netif

import (
	"errors"
	"fmt"
	"net"
	"strconv"
)

type AddrSource int

const (
	DHCP AddrSource = 1 + iota
	STATIC
	LOOPBACK
	MANUAL
)

type AddrFamily int

const (
	INET AddrFamily = 1 + iota
	INET6
)

// A representation of a network adapter
type NetworkAdapter struct {
	Name    string
	Auto    bool
	Hotplug bool
	IPs     []*NetworkIP
}

type NetworkIP struct {
	AddrSource     AddrSource
	AddrFamily     AddrFamily
	Address        net.IP
	Netmask        net.IPMask
	Broadcast      net.IP
	Network        net.IP
	Metric         *int
	Gateway        net.IP
	DNSNameServers []net.IP
	WiFiName       string
	WiFiPassword   string
	Others         []string
}

type valueValidator struct {
	Type     string
	Required bool
	In       []string
}

var valueValidators = map[string]valueValidator{
	"name":      {Required: true},
	"addrFam":   {In: []string{"inet", "inet6"}},
	"source":    {In: []string{"dhcp", "static", "loopback", "manual"}},
	"auto":      {Type: "bool"},
	"hotplug":   {Type: "bool"},
	"address":   {Type: "IP"},
	"netmask":   {Type: "IP"},
	"broadcast": {Type: "IP"},
	"network":   {Type: "IP"},
	"gateway":   {Type: "IP"},
}

func (na *NetworkIP) validateAll() error {
	/*for k, v := range valueValidators {
		val := nil

	}*/
	return nil
}

func (na *NetworkAdapter) validateName() error {
	return nil
}

func (na *NetworkIP) validateAddress() error {
	return nil
}

func (na *NetworkIP) validateNetmask() error {
	return nil
}

func (na *NetworkIP) validateNetwork() error {
	return nil
}

func (na *NetworkIP) validateBroadcast() error {
	return nil
}

func (na *NetworkIP) validateGateway() error {
	return nil
}

func (na *NetworkIP) validateAddrFamily() error {
	return nil
}

func (na *NetworkIP) validateSource() error {
	return nil
}

func (na *NetworkIP) validateIP(strIP string) (net.IP, error) {
	var ip net.IP
	if ip = net.ParseIP(strIP); ip == nil {
		return nil, errors.New("invalid IP address")
	}
	return ip, nil
}

func (na *NetworkIP) SetAddress(address string) error {
	ip, ipNet, err := net.ParseCIDR(address)
	if err == nil {
		na.Address = ip
		na.Netmask = ipNet.Mask
	} else {
		ip, err = na.validateIP(address)
		if err == nil {
			na.Address = ip
		}
	}
	return err
}

func (na *NetworkIP) SetNetmask(address string) error {
	addr, err := na.validateIP(address)
	if err == nil {
		na.Netmask = net.IPMask(addr)
	} else {
		var prefix int
		prefix, err = strconv.Atoi(address)
		if err == nil {
			if na.AddrFamily == INET6 {
				na.Netmask = net.CIDRMask(prefix, 128)
			} else {
				na.Netmask = net.CIDRMask(prefix, 32)
			}
		}
	}
	return err
}

func (na *NetworkIP) SetBroadcast(address string) error {
	addr, err := na.validateIP(address)
	if err == nil {
		na.Broadcast = addr
	}
	return err
}

func (na *NetworkIP) SetNetwork(address string) error {
	addr, err := na.validateIP(address)
	if err == nil {
		na.Network = addr
	}
	return err
}

func (na *NetworkIP) SetMetric(address string) error {
	addr, err := strconv.Atoi(address)
	if err == nil {
		na.Metric = &addr
	}
	return err
}

func (na *NetworkIP) SetGateway(address string) error {
	addr, err := na.validateIP(address)
	if err == nil {
		na.Gateway = addr
	}
	return err
}

func (na *NetworkIP) SetDNSNameServers(address string) error {
	addr, err := na.validateIP(address)
	if err == nil {
		na.DNSNameServers = append(na.DNSNameServers, addr)
	}
	return err
}

func (na *NetworkIP) SetWifiName(name string) error {
	na.WiFiName = name
	return nil
}

func (na *NetworkIP) SetWifiPassword(password string) error {
	na.WiFiPassword = password
	return nil
}

func (na *NetworkIP) SetOthers(address string) error {
	na.Others = append(na.Others, address)
	return nil
}

func (na *NetworkIP) SetConfigType(configType string) error {
	switch configType {
	case "DHCP":
		na.AddrSource = DHCP
	case "STATIC":
		na.AddrSource = STATIC
	default:
		return fmt.Errorf("unexpected configType: %s", configType)
	}
	return nil
}

func (na *NetworkAdapter) ParseAddressSource(AddressSource string) (AddrSource, error) {
	// Parse the address source for an interface
	var src AddrSource
	switch AddressSource {
	case "static":
		src = STATIC
	case "dhcp":
		src = DHCP
	case "loopback":
		src = LOOPBACK
	case "manual":
		src = MANUAL
	default:
		return -1, errors.New("invalid address source")
	}
	return src, nil
}

func (na *NetworkAdapter) ParseAddressFamily(AddressFamily string) (AddrFamily, error) {
	// Parse the address family for an interface
	var fam AddrFamily
	switch AddressFamily {
	case "inet":
		fam = INET
	case "inet6":
		fam = INET6
	default:
		return -1, errors.New("invalid address family")

	}
	return fam, nil
}

func (na *NetworkIP) DNSConcatString() string {
	message := ""
	for i, dns := range na.DNSNameServers {
		if i == len(na.DNSNameServers)-1 {
			message += dns.String()
		} else {
			message += dns.String() + " "
		}
	}
	return message
}
