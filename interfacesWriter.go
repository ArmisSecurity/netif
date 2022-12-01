package netif

import (
	"fmt"
	"net"
	"os"

	"strings"

	"github.com/n-marshall/fn"
)

func (is *InterfaceSet) Write(opts ...fn.Option) error {
	fnConfig := fn.MakeConfig(
		fn.Defaults{"path": "output"},
		opts,
	)
	path := fnConfig.GetString("path")

	// try to open the interface file for writing
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// write interface file
	return is.WriteToFile(f)
}

func (is *InterfaceSet) WriteToFile(f *os.File) error {
	fmt.Fprintf(f, "# interfaces(5) file used by ifup(8) and ifdown(8)\n"+
		"# Include files from /etc/network/interfaces.d:\n")

	for _, other := range is.others {
		fmt.Fprintln(f, other)
	}
	fmt.Fprintln(f)

	for _, adapter := range is.Adapters {
		adapterString, err := adapter.writeString()
		if err != nil {
			return err
		}
		fmt.Fprintf(f, "%s\n\n", adapterString)
	}
	return nil
}

func (a *NetworkAdapter) writeString() (string, error) {
	var lines []string
	if a.Auto {
		lines = append(lines, fmt.Sprintf("auto %s", a.Name))
	}
	if a.Hotplug {
		lines = append(lines, fmt.Sprintf("allow-hotplug %s", a.Name))
	}

	for _, ip := range a.IPs {
		lines = append(lines, ip.writeAddressFamily(a.Name))
		lines = append(lines, ip.writeIPLines()...)
	}

	return strings.Join(lines, "\n"), nil
}

func (a *NetworkIP) GetAddrFamilyString() string {
	switch a.AddrFamily {
	case INET:
		return "inet"
	case INET6:
		return "inet6"
	}
	return "inet"
}

func (a *NetworkIP) GetSourceFamilyString() string {
	switch a.AddrSource {
	case DHCP:
		return "dhcp"
	case STATIC:
		return "static"
	case LOOPBACK:
		return "loopback"
	case MANUAL:
		return "manual"
	}
	return "dhcp"
}

func (a *NetworkIP) writeAddressFamily(name string) string {
	var familyStr = a.GetAddrFamilyString()
	var sourceStr = a.GetSourceFamilyString()
	return fmt.Sprintf("iface %s %s %s", name, familyStr, sourceStr)
}

func (a *NetworkIP) writeIPLines() (lines []string) {
	if a.Address != nil {
		lines = append(lines, fmt.Sprintf("  address %s", a.Address))
	}
	if a.Netmask != nil {
		if a.AddrFamily == INET6 {
			prefix, _ := a.Netmask.Size()
			lines = append(lines, fmt.Sprintf("  netmask %d", prefix))
		} else {
			lines = append(lines, fmt.Sprintf("  netmask %s", net.IP(a.Netmask)))
		}
	}
	if a.Broadcast != nil {
		lines = append(lines, fmt.Sprintf("  broadcast %s", a.Broadcast))
	}
	if a.Network != nil {
		lines = append(lines, fmt.Sprintf("  network %s", a.Network))
	}
	if a.Metric != nil {
		lines = append(lines, fmt.Sprintf("  metric %d", *a.Metric))
	}
	if a.Gateway != nil {
		lines = append(lines, fmt.Sprintf("  gateway %s", a.Gateway))
	}
	if len(a.DNSNameServers) > 0 {
		lines = append(lines, fmt.Sprintf("  dns-nameservers %s", a.DNSConcatString()))
	}
	if a.WiFiName != "" {
		lines = append(lines, fmt.Sprintf("  wpa-ssid %s", a.WiFiName))
	}
	if a.WiFiPassword != "" {
		lines = append(lines, fmt.Sprintf("  wpa-psk %s", a.WiFiPassword))
	}
	for _, other := range a.Others {
		lines = append(lines, fmt.Sprintf("  %s", other))
	}
	return
}
