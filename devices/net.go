package devices

import (
	"log"
	"strings"

	"github.com/VictoriaMetrics/metrics"
	"github.com/shirou/gopsutil/v3/net"
)

// Interface represents a network interface
type Interface struct {
	// The OS-defined interface name
	IFace string
	// IsVPN will be set to true if the interface name begins with "tun", e.g. "tun0"
	IsVPN bool
	// BytesRecv is the OS-reported sum of bytes received over the interface
	BytesRecv uint64
	// BytesSent is the OS-reported sum of bytes sent over the interface
	BytesSent uint64
	// recvMetric is the metrics counter for received bytes (total)
	recvMetric *metrics.Counter
	// sentMetric is the metrics counter for sent bytes (total)
	sentMetric *metrics.Counter
}

type Network struct {
	// Data is the list of filtered interfaces, indexed by the interface name
	Data map[string]Interface
	// TotalBytesRecv is the last seen total number of bytes received, across all interfaces
	TotalBytesRecv uint64
	// TotalBytesSent is the last seen total number of bytes sent, across all interfaces
	TotalBytesSent uint64
	// RecentBytesRecv is the number of bytes received between the last two updates
	RecentBytesRecv uint64
	// RecentBytesSent is the number of bytes sent between the last two updates
	RecentBytesSent uint64
}

func NewNetwork() Network {
	return Network{Data: make(map[string]Interface)}
}

// LocalNetwork sets up tracking for a filtered list of interfaces. ifaces contains the list filter:
// 1. Included interfaces are simply the interface name, e.g. "eth0"
// 2. Excluded interfaces are prefixed by `!`, e.g. "!wlan0"
// 3. If the list contains *only* exclusions, then all interfaces not excluded are included
// 4. If the list contains any non-exclusions, then only those interfaces are included
// 5. Exclusion overrides inclusion
// 6. If the interface name begins with "tun", and `excludeVPNs` is true, then the interface is
//    excluded.
// filter is a comma-separated list of interface rules.
func LocalNetwork(filter []string, excludeVPNs bool) Network {
	interfaces, err := net.IOCounters(true)
	if err != nil {
		return Network{Data: make(map[string]Interface)}
	}
	nw := Network{Data: make(map[string]Interface)}
	// Build a map with wanted status for each interfaces.
	// 1. If no filter was provided, include everything.
	// 2. If no includes are provide, include everything that isn't excluded
	if len(filter) != 0 {
		excludes := make(map[string]bool)
		includes := make(map[string]bool)
		for _, iface := range filter {
			// "all" is synonymous with the empty includes set
			if iface == "all" {
				break
			}
			if strings.HasPrefix(iface, "!") {
				excludes[strings.TrimPrefix(iface, "!")] = true
			} else {
				includes[iface] = true
			}
		}
		for _, iface := range interfaces {
			// If is in exclude, exclude
			if _, exclude := excludes[iface.Name]; exclude {
				continue
			}
			// If is VPN and exclude VPNs, exclude
			if strings.HasPrefix(iface.Name, "tun") && excludeVPNs {
				continue
			}
			// If includes is not empty & not listed, exclude
			if _, ok := includes[iface.Name]; len(includes) != 0 && !ok {
				continue
			}
			// If we make it all the way to here, include it
			nw.Data[iface.Name] = Interface{
				IFace: iface.Name,
				IsVPN: strings.HasPrefix(iface.Name, "tun"),
			}
		}
	} else {
		// Include everything (except maybe VPNs)
		for _, iface := range interfaces {
			if strings.HasPrefix(iface.Name, "tun") && excludeVPNs {
				continue
			}
			nw.Data[iface.Name] = Interface{
				IFace: iface.Name,
				IsVPN: strings.HasPrefix(iface.Name, "tun"),
			}
		}
	}
	return nw
}

func (n *Network) Update() error {
	interfaces, err := net.IOCounters(true)
	if err != nil {
		return err
	}
	// Total sent & received across all devices this update
	var ttlRecv, ttlSent uint64
	for _, iface := range interfaces {
		intf, ok := n.Data[iface.Name]
		if ok { // Simple case
			if intf.sentMetric != nil {
				intf.sentMetric.Add(int(iface.BytesRecv - intf.BytesRecv))
				intf.recvMetric.Add(int(iface.BytesSent - intf.BytesSent))
			}
			intf.BytesRecv = iface.BytesRecv
			intf.BytesSent = iface.BytesSent
			ttlRecv += iface.BytesRecv
			ttlSent += iface.BytesSent
			n.Data[iface.Name] = intf
		}
	}

	n.RecentBytesRecv = ttlRecv - n.TotalBytesRecv
	n.RecentBytesSent = ttlSent - n.TotalBytesSent

	if n.RecentBytesRecv < 0 {
		log.Printf("illogical bytes received for network; previous %d > %d new", n.TotalBytesRecv, ttlRecv)
		// recover from error
		n.RecentBytesRecv = 0
	}
	if n.RecentBytesSent < 0 {
		log.Printf("illogical bytes sent for network; previous %d > %d new", n.TotalBytesSent, ttlSent)
		// recover from error
		n.RecentBytesSent = 0
	}

	// Set the TX memory ("previous" values)
	n.TotalBytesRecv = ttlRecv
	n.TotalBytesSent = ttlSent

	return nil
}

// EnableMetrics creates two counters -- recv and sent -- which tally the total
// bytes sent and received through the filtered interfaces.
func (net *Network) EnableMetrics(s *metrics.Set) {
	for k, v := range net.Data {
		v.recvMetric = s.NewCounter(makeName("net", k, "recv"))
		v.sentMetric = s.NewCounter(makeName("net", k, "sent"))
		net.Data[k] = v
	}
}

func interfaces() ([]string, error) {
	interfaces, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}
	rv := make([]string, len(interfaces))
	for i, intf := range interfaces {
		rv[i] = intf.Name
	}
	return rv, nil
}
