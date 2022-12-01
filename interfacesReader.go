package netif

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/n-marshall/fn"
)

// TODO get rid of interfaceReader
type InterfacesReader struct {
	filePath   string
	others     []string
	adapters   []*NetworkAdapter
	adapterMap map[string]*NetworkAdapter
}

func Parse(opts ...fn.Option) (*InterfaceSet, error) {
	var err error

	fnConfig := fn.MakeConfig(
		fn.Defaults{"path": "/etc/network/interfaces"},
		opts,
	)
	path := fnConfig.GetString("path")

	is := &InterfaceSet{
		InterfacesPath: path,
	}
	if is.others, is.Adapters, err = NewInterfacesReader(is.InterfacesPath).ParseInterfaces(); err != nil {
		return nil, err
	}

	return is, nil
}

func NewInterfacesReader(filePath string) *InterfacesReader {
	ir := InterfacesReader{filePath: filePath}
	ir.reset()

	return &ir
}

func (ir *InterfacesReader) ParseInterfaces() ([]string, []*NetworkAdapter, error) {
	// Reset this object in case is not new
	ir.reset()

	// Try to open the file
	f, err := os.Open(ir.filePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	// Treat each line from the file
	ir.readLinesFromFile(f)

	return ir.others, ir.adapters, nil
}

func (ir *InterfacesReader) readLinesFromFile(file *os.File) (err error) {
	s := bufio.NewScanner(file)

	//var a Adapter

	iface := ""
	for s.Scan() {
		// fmt.Printf("%s\n", s.Text())
		line := strings.TrimSpace(s.Text())

		// Identify the clauses by analyzing the first word of each line.
		// Go to the next line if the current line is a comment.
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Continue if line is empty
		if len(line) == 0 {
			continue
		}

		// Parse the line
		if err = ir.parseOthers(line, &iface); err != nil {
			return err
		}
		if err = ir.readAuto(line, &iface); err != nil {
			return err
		}
		if err = ir.readHotplug(line, &iface); err != nil {
			return err
		}
		if err = ir.parseIface(line, &iface); err != nil {
			return err
		}
		if err = ir.parseDetails(line, &iface); err != nil {
			return err
		}
	}
	return nil
}

func (ir *InterfacesReader) parseOthers(line string, iface *string) error {
	sline := strings.Fields(line)
	if sline[0] == "mapping" ||
		sline[0] == "rename" ||
		sline[0] == "source" ||
		sline[0] == "source-directory" {
		*iface = ""
		ir.others = append(ir.others, line)
	}
	return nil
}

func (ir *InterfacesReader) parseIface(line string, iface *string) error {
	sline := strings.Fields(line)

	if sline[0] == "iface" {
		*iface = sline[1]

		na, exists := ir.adapterMap[*iface]
		if !exists {
			na = &NetworkAdapter{Name: *iface}
			ir.adapters = append(ir.adapters, na)
			ir.adapterMap[*iface] = na
		}

		// Parse and set the address source
		src, err := na.ParseAddressSource(sline[3])
		if err != nil {
			return err
		}

		// Parse and set the address fadapterMapily
		fam, err := na.ParseAddressFamily(sline[2])
		if err != nil {
			return err
		}

		ip := &NetworkIP{AddrFamily: fam, AddrSource: src}
		na.IPs = append(na.IPs, ip)
	}
	return nil
}

func (ir *InterfacesReader) parseDetails(line string, iface *string) error {
	sline := strings.Fields(line)

	if sline[0] != "address" && // static
		sline[0] != "netmask" &&
		sline[0] != "broadcast" &&
		sline[0] != "network" &&
		sline[0] != "metric" &&
		sline[0] != "gateway" &&
		sline[0] != "pointopoint" &&
		sline[0] != "media" &&
		sline[0] != "hwaddress" &&
		sline[0] != "mtu" &&
		sline[0] != "hostname" && // dhcp
		sline[0] != "leasehours" &&
		sline[0] != "leasetime" &&
		sline[0] != "vendor" &&
		sline[0] != "client" &&
		sline[0] != "dns-nameservers" && // options
		sline[0] != "dns-domain" &&
		sline[0] != "dns-search" &&
		sline[0] != "up" &&
		sline[0] != "pre-up" &&
		sline[0] != "post-up" &&
		sline[0] != "down" &&
		sline[0] != "pre-down" &&
		sline[0] != "post-down" {
		return nil
	}
	if *iface == "" {
		return fmt.Errorf("invalid line: %s", line)
	}

	na := ir.adapterMap[*iface]
	ip := na.IPs[len(na.IPs)-1]

	switch sline[0] {
	case "address":
		if err := ip.SetAddress(sline[1]); err != nil {
			return err
		}
	case "netmask":
		if err := ip.SetNetmask(sline[1]); err != nil {
			return err
		}
	case "broadcast":
		if err := ip.SetBroadcast(sline[1]); err != nil {
			return err
		}
	case "network":
		if err := ip.SetNetwork(sline[1]); err != nil {
			return err
		}
	case "metric":
		if err := ip.SetMetric(sline[1]); err != nil {
			return err
		}
	case "gateway":
		if err := ip.SetGateway(sline[1]); err != nil {
			return err
		}
	case "dns-nameservers":
		for i := 1; i < len(sline); i++ {
			if err := ip.SetDNSNameServers(sline[i]); err != nil {
				return err
			}
		}
	case "wpa-ssid":
		if err := ip.SetWifiName(sline[1]); err != nil {
			return err
		}
	case "wpa-psk":
		if err := ip.SetWifiPassword(sline[1]); err != nil {
			return err
		}
	default:
		if err := ip.SetOthers(line); err != nil {
			return err
		}
	}

	return nil
}

func (ir *InterfacesReader) readAuto(line string, iface *string) error {
	sline := strings.Fields(line)

	if sline[0] == "auto" || sline[0] == "allow-auto" {
		*iface = ""
		na, exists := ir.adapterMap[sline[1]]
		if !exists {
			na = &NetworkAdapter{Name: sline[1]}
			ir.adapters = append(ir.adapters, na)
			ir.adapterMap[sline[1]] = na
		}
		na.Auto = true
	}
	return nil
}

func (ir *InterfacesReader) readHotplug(line string, iface *string) error {
	sline := strings.Fields(line)

	if sline[0] == "allow-hotplug" {
		*iface = ""
		na, exists := ir.adapterMap[sline[1]]
		if !exists {
			na = &NetworkAdapter{Name: sline[1]}
			ir.adapters = append(ir.adapters, na)
			ir.adapterMap[sline[1]] = na
		}
		na.Hotplug = true
	}
	return nil
}

func (ir *InterfacesReader) reset() {
	// Initialize a place to store create NetworkAdapter objects
	ir.others = nil
	ir.adapters = nil
	ir.adapterMap = make(map[string]*NetworkAdapter)
}
