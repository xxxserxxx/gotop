package widgets

import (
	"fmt"
	"time"

	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/utils"
)

const (
	// NetInterfaceAll enables all network interfaces
	NetInterfaceAll = "all"
	// NetInterfaceVpn is the VPN interface
	NetInterfaceVpn = "tun0"
)

type NetWidget struct {
	*SparklineGroup

	// used to calculate recent network activity
	totalBytesRecv uint64
	totalBytesSent uint64
	Mbps           bool
	Network        *devices.Network
	rx, tx, format string
	factor         float64
}

// TODO: state:merge #169 % option for network use (jrswab/networkPercentage)
func NewNetWidget(refresh time.Duration) *NetWidget {
	recvSparkline := NewSparkline()
	recvSparkline.Data = []int{}

	sentSparkline := NewSparkline()
	sentSparkline.Data = []int{}

	spark := NewSparklineGroup(recvSparkline, sentSparkline)
	self := &NetWidget{
		SparklineGroup: spark,
		factor:         float64(refresh / time.Second),
	}
	self.Title = tr.Value("widget.label.net")

	return self
}

func (n *NetWidget) Attach(ni *devices.Network) {
	n.Network = ni
	n.rx, n.tx = "RX/s", "TX/s"
	if n.Mbps {
		n.rx, n.tx = "mbps", "mbps"
	}
	n.format = " %s: %9.1f %2s/s"
}

func (net *NetWidget) Update() {
	net.Lines[0].Data = append(net.Lines[0].Data, int(net.Network.RecentBytesRecv))
	net.Lines[1].Data = append(net.Lines[1].Data, int(net.Network.RecentBytesSent))

	var total, recent uint64
	var label, unitRecent, rate string
	var recentConverted float64
	// render widget titles
	for i := 0; i < 2; i++ {
		if i == 0 {
			total, label, rate, recent = net.Network.TotalBytesRecv, "RX", net.rx, net.Network.RecentBytesRecv
		} else {
			total, label, rate, recent = net.Network.TotalBytesSent, "TX", net.tx, net.Network.RecentBytesSent
		}

		totalConverted, unitTotal := utils.ConvertBytes(total)
		if net.Mbps {
			recentConverted, unitRecent, net.format = float64(recent)*0.000008, "", " %s: %11.3f %2s"
		} else {
			recentConverted, unitRecent = utils.ConvertBytes(recent)
		}

		net.Lines[i].Title1 = fmt.Sprintf(" %s %s: %5.1f %s", tr.Value("total"), label, totalConverted, unitTotal)
		net.Lines[i].Title2 = fmt.Sprintf(net.format, rate, recentConverted/net.factor, unitRecent)
	}
}
