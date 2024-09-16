//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package pci

import (
	"encoding/json"
	"fmt"

	"github.com/jaypipes/pcidb"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/marshal"
	"github.com/jaypipes/ghw/pkg/option"
	pciaddr "github.com/jaypipes/ghw/pkg/pci/address"
	"github.com/jaypipes/ghw/pkg/topology"
	"github.com/jaypipes/ghw/pkg/util"
)

// backward compatibility, to be removed in 1.0.0
type Address pciaddr.Address

// backward compatibility, to be removed in 1.0.0
var AddressFromString = pciaddr.FromString

type Device struct {
	// The PCI address of the device
	Address   string         `json:"address"`
	Vendor    *pcidb.Vendor  `json:"vendor"`
	Product   *pcidb.Product `json:"product"`
	Revision  string         `json:"revision"`
	Subsystem *pcidb.Product `json:"subsystem"`
	// optional subvendor/sub-device information
	Class *pcidb.Class `json:"class"`
	// optional sub-class for the device
	Subclass *pcidb.Subclass `json:"subclass"`
	// optional programming interface
	ProgrammingInterface *pcidb.ProgrammingInterface `json:"programming_interface"`
	// Topology node that the PCI device is affined to. Will be nil if the
	// architecture is not NUMA.
	Node   *topology.Node `json:"node,omitempty"`
	Driver string         `json:"driver"`
}

type devIdent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type devMarshallable struct {
	Driver    string   `json:"driver"`
	Address   string   `json:"address"`
	Vendor    devIdent `json:"vendor"`
	Product   devIdent `json:"product"`
	Revision  string   `json:"revision"`
	Subsystem devIdent `json:"subsystem"`
	Class     devIdent `json:"class"`
	Subclass  devIdent `json:"subclass"`
	Interface devIdent `json:"programming_interface"`
}

// NOTE(jaypipes) Device has a custom JSON marshaller because we don't want
// to serialize the entire PCIDB information for the Vendor (which includes all
// of the vendor's products, etc). Instead, we simply serialize the ID and
// human-readable name of the vendor, product, class, etc.
func (d *Device) MarshalJSON() ([]byte, error) {
	dm := devMarshallable{
		Driver:  d.Driver,
		Address: d.Address,
		Vendor: devIdent{
			ID:   d.Vendor.ID,
			Name: d.Vendor.Name,
		},
		Product: devIdent{
			ID:   d.Product.ID,
			Name: d.Product.Name,
		},
		Revision: d.Revision,
		Subsystem: devIdent{
			ID:   d.Subsystem.ID,
			Name: d.Subsystem.Name,
		},
		Class: devIdent{
			ID:   d.Class.ID,
			Name: d.Class.Name,
		},
		Subclass: devIdent{
			ID:   d.Subclass.ID,
			Name: d.Subclass.Name,
		},
		Interface: devIdent{
			ID:   d.ProgrammingInterface.ID,
			Name: d.ProgrammingInterface.Name,
		},
	}
	return json.Marshal(dm)
}

func (d *Device) String() string {
	vendorName := util.UNKNOWN
	if d.Vendor != nil {
		vendorName = d.Vendor.Name
	}
	productName := util.UNKNOWN
	if d.Product != nil {
		productName = d.Product.Name
	}
	className := util.UNKNOWN
	if d.Class != nil {
		className = d.Class.Name
	}
	return fmt.Sprintf(
		"%s -> driver: '%s' class: '%s' vendor: '%s' product: '%s'",
		d.Address,
		d.Driver,
		className,
		vendorName,
		productName,
	)
}

type Info struct {
	arch topology.Architecture
	ctx  *context.Context
	// All PCI devices on the host system
	Devices []*Device
	// hash of class ID -> class information
	// DEPRECATED. Will be removed in v1.0. Please use
	// github.com/jaypipes/pcidb to explore PCIDB information
	Classes map[string]*pcidb.Class `json:"-"`
	// hash of vendor ID -> vendor information
	// DEPRECATED. Will be removed in v1.0. Please use
	// github.com/jaypipes/pcidb to explore PCIDB information
	Vendors map[string]*pcidb.Vendor `json:"-"`
	// hash of vendor ID + product/device ID -> product information
	// DEPRECATED. Will be removed in v1.0. Please use
	// github.com/jaypipes/pcidb to explore PCIDB information
	Products map[string]*pcidb.Product `json:"-"`
}

func (i *Info) String() string {
	return fmt.Sprintf("PCI (%d devices)", len(i.Devices))
}

// New returns a pointer to an Info struct that contains information about the
// PCI devices on the host system
func New(opts ...*option.Option) (*Info, error) {
	merged := option.Merge(opts...)
	ctx := context.New(merged)
	// by default we don't report NUMA information;
	// we will only if are sure we are running on NUMA architecture
	info := &Info{
		arch: topology.ARCHITECTURE_SMP,
		ctx:  ctx,
	}

	// we do this trick because we need to make sure ctx.Setup() gets
	// a chance to run before any subordinate package is created reusing
	// our context.
	loadDetectingTopology := func() error {
		topo, err := topology.New(context.WithContext(ctx))
		if err == nil {
			info.arch = topo.Architecture
		} else {
			ctx.Warn("error detecting system topology: %v", err)
		}
		return info.load()
	}

	var err error
	if context.Exists(merged) {
		err = loadDetectingTopology()
	} else {
		err = ctx.Do(loadDetectingTopology)
	}
	if err != nil {
		return nil, err
	}
	return info, nil
}

// lookupDevice gets a device from cached data
func (info *Info) lookupDevice(address string) *Device {
	for _, dev := range info.Devices {
		if dev.Address == address {
			return dev
		}
	}
	return nil
}

// simple private struct used to encapsulate PCI information in a top-level
// "pci" YAML/JSON map/object key
type pciPrinter struct {
	Info *Info `json:"pci"`
}

// YAMLString returns a string with the PCI information formatted as YAML
// under a top-level "pci:" key
func (i *Info) YAMLString() string {
	return marshal.SafeYAML(i.ctx, pciPrinter{i})
}

// JSONString returns a string with the PCI information formatted as JSON
// under a top-level "pci:" key
func (i *Info) JSONString(indent bool) string {
	return marshal.SafeJSON(i.ctx, pciPrinter{i}, indent)
}
