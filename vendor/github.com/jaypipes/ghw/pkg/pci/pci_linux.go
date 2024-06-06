// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package pci

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/jaypipes/pcidb"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/linuxpath"
	"github.com/jaypipes/ghw/pkg/option"
	pciaddr "github.com/jaypipes/ghw/pkg/pci/address"
	"github.com/jaypipes/ghw/pkg/topology"
	"github.com/jaypipes/ghw/pkg/util"
)

const (
	// found running `wc` against real linux systems
	modAliasExpectedLength = 54
)

func (i *Info) load() error {
	// when consuming snapshots - most notably, but not only, in tests,
	// the context pkg forces the chroot value to the unpacked snapshot root.
	// This is intentional, intentionally transparent and ghw is prepared to handle this case.
	// However, `pcidb` is not. It doesn't know about ghw snaphots, nor it should.
	// so we need to complicate things a bit. If the user explicitely supplied
	// a chroot option, then we should honor it all across the stack, and passing down
	// the chroot to pcidb is the right thing to do. If, however, the chroot was
	// implcitely set by snapshot support, then this must be consumed by ghw only.
	// In this case we should NOT pass it down to pcidb.
	chroot := i.ctx.Chroot
	if i.ctx.SnapshotPath != "" {
		chroot = option.DefaultChroot
	}
	db, err := pcidb.New(pcidb.WithChroot(chroot))
	if err != nil {
		return err
	}
	i.Classes = db.Classes
	i.Vendors = db.Vendors
	i.Products = db.Products
	i.Devices = i.ListDevices()
	return nil
}

func getDeviceModaliasPath(ctx *context.Context, pciAddr *pciaddr.Address) string {
	paths := linuxpath.New(ctx)
	return filepath.Join(
		paths.SysBusPciDevices,
		pciAddr.String(),
		"modalias",
	)
}

func getDeviceRevision(ctx *context.Context, pciAddr *pciaddr.Address) string {
	paths := linuxpath.New(ctx)
	revisionPath := filepath.Join(
		paths.SysBusPciDevices,
		pciAddr.String(),
		"revision",
	)

	if _, err := os.Stat(revisionPath); err != nil {
		return ""
	}
	revision, err := ioutil.ReadFile(revisionPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(revision))
}

func getDeviceNUMANode(ctx *context.Context, pciAddr *pciaddr.Address) *topology.Node {
	paths := linuxpath.New(ctx)
	numaNodePath := filepath.Join(paths.SysBusPciDevices, pciAddr.String(), "numa_node")

	if _, err := os.Stat(numaNodePath); err != nil {
		return nil
	}

	nodeIdx := util.SafeIntFromFile(ctx, numaNodePath)
	if nodeIdx == -1 {
		return nil
	}

	return &topology.Node{
		ID: nodeIdx,
	}
}

func getDeviceDriver(ctx *context.Context, pciAddr *pciaddr.Address) string {
	paths := linuxpath.New(ctx)
	driverPath := filepath.Join(paths.SysBusPciDevices, pciAddr.String(), "driver")

	if _, err := os.Stat(driverPath); err != nil {
		return ""
	}

	dest, err := os.Readlink(driverPath)
	if err != nil {
		return ""
	}
	return filepath.Base(dest)
}

type deviceModaliasInfo struct {
	vendorID     string
	productID    string
	subproductID string
	subvendorID  string
	classID      string
	subclassID   string
	progIfaceID  string
}

func parseModaliasFile(fp string) *deviceModaliasInfo {
	if _, err := os.Stat(fp); err != nil {
		return nil
	}
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil
	}

	return parseModaliasData(string(data))
}

func parseModaliasData(data string) *deviceModaliasInfo {
	// extra sanity check to avoid segfaults. We actually expect
	// the data to be exactly long `modAliasExpectedlength`, but
	// we will happily ignore any extra data we don't know how to
	// handle.
	if len(data) < modAliasExpectedLength {
		return nil
	}
	// The modalias file is an encoded file that looks like this:
	//
	// $ cat /sys/devices/pci0000\:00/0000\:00\:03.0/0000\:03\:00.0/modalias
	// pci:v000010DEd00001C82sv00001043sd00008613bc03sc00i00
	//
	// It is interpreted like so:
	//
	// pci: -- ignore
	// v000010DE -- PCI vendor ID
	// d00001C82 -- PCI device ID (the product/model ID)
	// sv00001043 -- PCI subsystem vendor ID
	// sd00008613 -- PCI subsystem device ID (subdevice product/model ID)
	// bc03 -- PCI base class
	// sc00 -- PCI subclass
	// i00 -- programming interface
	vendorID := strings.ToLower(data[9:13])
	productID := strings.ToLower(data[18:22])
	subvendorID := strings.ToLower(data[28:32])
	subproductID := strings.ToLower(data[38:42])
	classID := data[44:46]
	subclassID := data[48:50]
	progIfaceID := data[51:53]
	return &deviceModaliasInfo{
		vendorID:     vendorID,
		productID:    productID,
		subproductID: subproductID,
		subvendorID:  subvendorID,
		classID:      classID,
		subclassID:   subclassID,
		progIfaceID:  progIfaceID,
	}
}

// Returns a pointer to a pcidb.Vendor struct matching the supplied vendor
// ID string. If no such vendor ID string could be found, returns the
// pcidb.Vendor struct populated with "unknown" vendor Name attribute and
// empty Products attribute.
func findPCIVendor(info *Info, vendorID string) *pcidb.Vendor {
	vendor := info.Vendors[vendorID]
	if vendor == nil {
		return &pcidb.Vendor{
			ID:       vendorID,
			Name:     util.UNKNOWN,
			Products: []*pcidb.Product{},
		}
	}
	return vendor
}

// Returns a pointer to a pcidb.Product struct matching the supplied vendor
// and product ID strings. If no such product could be found, returns the
// pcidb.Product struct populated with "unknown" product Name attribute and
// empty Subsystems attribute.
func findPCIProduct(
	info *Info,
	vendorID string,
	productID string,
) *pcidb.Product {
	product := info.Products[vendorID+productID]
	if product == nil {
		return &pcidb.Product{
			ID:         productID,
			Name:       util.UNKNOWN,
			Subsystems: []*pcidb.Product{},
		}
	}
	return product
}

// Returns a pointer to a pcidb.Product struct matching the supplied vendor,
// product, subvendor and subproduct ID strings. If no such product could be
// found, returns the pcidb.Product struct populated with "unknown" product
// Name attribute and empty Subsystems attribute.
func findPCISubsystem(
	info *Info,
	vendorID string,
	productID string,
	subvendorID string,
	subproductID string,
) *pcidb.Product {
	product := info.Products[vendorID+productID]
	subvendor := info.Vendors[subvendorID]
	if subvendor != nil && product != nil {
		for _, p := range product.Subsystems {
			if p.ID == subproductID {
				return p
			}
		}
	}
	return &pcidb.Product{
		VendorID: subvendorID,
		ID:       subproductID,
		Name:     util.UNKNOWN,
	}
}

// Returns a pointer to a pcidb.Class struct matching the supplied class ID
// string. If no such class ID string could be found, returns the
// pcidb.Class struct populated with "unknown" class Name attribute and
// empty Subclasses attribute.
func findPCIClass(info *Info, classID string) *pcidb.Class {
	class := info.Classes[classID]
	if class == nil {
		return &pcidb.Class{
			ID:         classID,
			Name:       util.UNKNOWN,
			Subclasses: []*pcidb.Subclass{},
		}
	}
	return class
}

// Returns a pointer to a pcidb.Subclass struct matching the supplied class
// and subclass ID strings.  If no such subclass could be found, returns the
// pcidb.Subclass struct populated with "unknown" subclass Name attribute
// and empty ProgrammingInterfaces attribute.
func findPCISubclass(
	info *Info,
	classID string,
	subclassID string,
) *pcidb.Subclass {
	class := info.Classes[classID]
	if class != nil {
		for _, sc := range class.Subclasses {
			if sc.ID == subclassID {
				return sc
			}
		}
	}
	return &pcidb.Subclass{
		ID:                    subclassID,
		Name:                  util.UNKNOWN,
		ProgrammingInterfaces: []*pcidb.ProgrammingInterface{},
	}
}

// Returns a pointer to a pcidb.ProgrammingInterface struct matching the
// supplied class, subclass and programming interface ID strings.  If no such
// programming interface could be found, returns the
// pcidb.ProgrammingInterface struct populated with "unknown" Name attribute
func findPCIProgrammingInterface(
	info *Info,
	classID string,
	subclassID string,
	progIfaceID string,
) *pcidb.ProgrammingInterface {
	subclass := findPCISubclass(info, classID, subclassID)
	for _, pi := range subclass.ProgrammingInterfaces {
		if pi.ID == progIfaceID {
			return pi
		}
	}
	return &pcidb.ProgrammingInterface{
		ID:   progIfaceID,
		Name: util.UNKNOWN,
	}
}

// GetDevice returns a pointer to a Device struct that describes the PCI
// device at the requested address. If no such device could be found, returns nil.
func (info *Info) GetDevice(address string) *Device {
	// check cached data first
	if dev := info.lookupDevice(address); dev != nil {
		return dev
	}

	pciAddr := pciaddr.FromString(address)
	if pciAddr == nil {
		info.ctx.Warn("error parsing the pci address %q", address)
		return nil
	}

	// no cached data, let's get the information from system.
	fp := getDeviceModaliasPath(info.ctx, pciAddr)
	if fp == "" {
		info.ctx.Warn("error finding modalias info for device %q", address)
		return nil
	}

	modaliasInfo := parseModaliasFile(fp)
	if modaliasInfo == nil {
		info.ctx.Warn("error parsing modalias info for device %q", address)
		return nil
	}

	device := info.getDeviceFromModaliasInfo(address, modaliasInfo)
	device.Revision = getDeviceRevision(info.ctx, pciAddr)
	if info.arch == topology.ARCHITECTURE_NUMA {
		device.Node = getDeviceNUMANode(info.ctx, pciAddr)
	}
	device.Driver = getDeviceDriver(info.ctx, pciAddr)
	return device
}

// ParseDevice returns a pointer to a Device given its describing data.
// The PCI device obtained this way may not exist in the system;
// use GetDevice to get a *Device which is found in the system
func (info *Info) ParseDevice(address, modalias string) *Device {
	modaliasInfo := parseModaliasData(modalias)
	if modaliasInfo == nil {
		return nil
	}
	return info.getDeviceFromModaliasInfo(address, modaliasInfo)
}

func (info *Info) getDeviceFromModaliasInfo(address string, modaliasInfo *deviceModaliasInfo) *Device {
	vendor := findPCIVendor(info, modaliasInfo.vendorID)
	product := findPCIProduct(
		info,
		modaliasInfo.vendorID,
		modaliasInfo.productID,
	)
	subsystem := findPCISubsystem(
		info,
		modaliasInfo.vendorID,
		modaliasInfo.productID,
		modaliasInfo.subvendorID,
		modaliasInfo.subproductID,
	)
	class := findPCIClass(info, modaliasInfo.classID)
	subclass := findPCISubclass(
		info,
		modaliasInfo.classID,
		modaliasInfo.subclassID,
	)
	progIface := findPCIProgrammingInterface(
		info,
		modaliasInfo.classID,
		modaliasInfo.subclassID,
		modaliasInfo.progIfaceID,
	)

	return &Device{
		Address:              address,
		Vendor:               vendor,
		Subsystem:            subsystem,
		Product:              product,
		Class:                class,
		Subclass:             subclass,
		ProgrammingInterface: progIface,
	}
}

// ListDevices returns a list of pointers to Device structs present on the
// host system
// DEPRECATED. Will be removed in v1.0. Please use
// github.com/jaypipes/pcidb to explore PCIDB information
func (info *Info) ListDevices() []*Device {
	paths := linuxpath.New(info.ctx)
	devs := make([]*Device, 0)
	// We scan the /sys/bus/pci/devices directory which contains a collection
	// of symlinks. The names of the symlinks are all the known PCI addresses
	// for the host. For each address, we grab a *Device matching the
	// address and append to the returned array.
	links, err := ioutil.ReadDir(paths.SysBusPciDevices)
	if err != nil {
		info.ctx.Warn("failed to read /sys/bus/pci/devices")
		return nil
	}
	var dev *Device
	for _, link := range links {
		addr := link.Name()
		dev = info.GetDevice(addr)
		if dev == nil {
			info.ctx.Warn("failed to get device information for PCI address %s", addr)
		} else {
			devs = append(devs, dev)
		}
	}
	return devs
}
