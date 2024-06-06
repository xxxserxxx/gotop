// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package net

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/linuxpath"
)

const (
	_WARN_ETHTOOL_NOT_INSTALLED = `ethtool not installed. Cannot grab NIC capabilities`
)

func (i *Info) load() error {
	i.NICs = nics(i.ctx)
	return nil
}

func nics(ctx *context.Context) []*NIC {
	nics := make([]*NIC, 0)

	paths := linuxpath.New(ctx)
	files, err := ioutil.ReadDir(paths.SysClassNet)
	if err != nil {
		return nics
	}

	etAvailable := ctx.EnableTools
	if etAvailable {
		if etInstalled := ethtoolInstalled(); !etInstalled {
			ctx.Warn(_WARN_ETHTOOL_NOT_INSTALLED)
			etAvailable = false
		}
	}

	for _, file := range files {
		filename := file.Name()
		// Ignore loopback...
		if filename == "lo" {
			continue
		}

		netPath := filepath.Join(paths.SysClassNet, filename)
		dest, _ := os.Readlink(netPath)
		isVirtual := false
		if strings.Contains(dest, "devices/virtual/net") {
			isVirtual = true
		}

		nic := &NIC{
			Name:      filename,
			IsVirtual: isVirtual,
		}

		mac := netDeviceMacAddress(paths, filename)
		nic.MacAddress = mac
		if etAvailable {
			nic.Capabilities = netDeviceCapabilities(ctx, filename)
		} else {
			nic.Capabilities = []*NICCapability{}
		}

		nic.PCIAddress = netDevicePCIAddress(paths.SysClassNet, filename)

		nics = append(nics, nic)
	}
	return nics
}

func netDeviceMacAddress(paths *linuxpath.Paths, dev string) string {
	// Instead of use udevadm, we can get the device's MAC address by examing
	// the /sys/class/net/$DEVICE/address file in sysfs. However, for devices
	// that have addr_assign_type != 0, return None since the MAC address is
	// random.
	aatPath := filepath.Join(paths.SysClassNet, dev, "addr_assign_type")
	contents, err := ioutil.ReadFile(aatPath)
	if err != nil {
		return ""
	}
	if strings.TrimSpace(string(contents)) != "0" {
		return ""
	}
	addrPath := filepath.Join(paths.SysClassNet, dev, "address")
	contents, err = ioutil.ReadFile(addrPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(contents))
}

func ethtoolInstalled() bool {
	_, err := exec.LookPath("ethtool")
	return err == nil
}

func netDeviceCapabilities(ctx *context.Context, dev string) []*NICCapability {
	caps := make([]*NICCapability, 0)
	path, _ := exec.LookPath("ethtool")
	cmd := exec.Command(path, "-k", dev)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		msg := fmt.Sprintf("could not grab NIC capabilities for %s: %s", dev, err)
		ctx.Warn(msg)
		return caps
	}

	// The out variable will now contain something that looks like the
	// following.
	//
	// Features for enp58s0f1:
	// rx-checksumming: on
	// tx-checksumming: off
	//     tx-checksum-ipv4: off
	//     tx-checksum-ip-generic: off [fixed]
	//     tx-checksum-ipv6: off
	//     tx-checksum-fcoe-crc: off [fixed]
	//     tx-checksum-sctp: off [fixed]
	// scatter-gather: off
	//     tx-scatter-gather: off
	//     tx-scatter-gather-fraglist: off [fixed]
	// tcp-segmentation-offload: off
	//     tx-tcp-segmentation: off
	//     tx-tcp-ecn-segmentation: off [fixed]
	//     tx-tcp-mangleid-segmentation: off
	//     tx-tcp6-segmentation: off
	// < snipped >
	scanner := bufio.NewScanner(&out)
	// Skip the first line...
	scanner.Scan()
	for scanner.Scan() {
		line := strings.TrimPrefix(scanner.Text(), "\t")
		caps = append(caps, netParseEthtoolFeature(line))
	}
	return caps
}

// netParseEthtoolFeature parses a line from the ethtool -k output and returns
// a NICCapability.
//
// The supplied line will look like the following:
//
// tx-checksum-ip-generic: off [fixed]
//
// [fixed] indicates that the feature may not be turned on/off. Note: it makes
// no difference whether a privileged user runs `ethtool -k` when determining
// whether [fixed] appears for a feature.
func netParseEthtoolFeature(line string) *NICCapability {
	parts := strings.Fields(line)
	cap := strings.TrimSuffix(parts[0], ":")
	enabled := parts[1] == "on"
	fixed := len(parts) == 3 && parts[2] == "[fixed]"
	return &NICCapability{
		Name:      cap,
		IsEnabled: enabled,
		CanEnable: !fixed,
	}
}

func netDevicePCIAddress(netDevDir, netDevName string) *string {
	// what we do here is not that hard in the end: we need to navigate the sysfs
	// up to the directory belonging to the device backing the network interface.
	// we can make few relatively safe assumptions, but the safest way is follow
	// the right links. And so we go.
	// First of all, knowing the network device name we need to resolve the backing
	// device path to its full sysfs path.
	// say we start with netDevDir="/sys/class/net" and netDevName="enp0s31f6"
	netPath := filepath.Join(netDevDir, netDevName)
	dest, err := os.Readlink(netPath)
	if err != nil {
		// bail out with empty value
		return nil
	}
	// now we have something like dest="../../devices/pci0000:00/0000:00:1f.6/net/enp0s31f6"
	// remember the path is relative to netDevDir="/sys/class/net"

	netDev := filepath.Clean(filepath.Join(netDevDir, dest))
	// so we clean "/sys/class/net/../../devices/pci0000:00/0000:00:1f.6/net/enp0s31f6"
	// leading to "/sys/devices/pci0000:00/0000:00:1f.6/net/enp0s31f6"
	// still not there. We need to access the data of the pci device. So we jump into the path
	// linked to the "device" pseudofile
	dest, err = os.Readlink(filepath.Join(netDev, "device"))
	if err != nil {
		// bail out with empty value
		return nil
	}
	// we expect something like="../../../0000:00:1f.6"

	devPath := filepath.Clean(filepath.Join(netDev, dest))
	// so we clean "/sys/devices/pci0000:00/0000:00:1f.6/net/enp0s31f6/../../../0000:00:1f.6"
	// leading to "/sys/devices/pci0000:00/0000:00:1f.6/"
	// finally here!

	// to which bus is this device connected to?
	dest, err = os.Readlink(filepath.Join(devPath, "subsystem"))
	if err != nil {
		// bail out with empty value
		return nil
	}
	// ok, this is hacky, but since we need the last *two* path components and we know we
	// are running on linux...
	if !strings.HasSuffix(dest, "/bus/pci") {
		// unsupported and unexpected bus!
		return nil
	}

	pciAddr := filepath.Base(devPath)
	return &pciAddr
}
