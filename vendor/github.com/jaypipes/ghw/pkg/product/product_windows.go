// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package product

import (
	"github.com/StackExchange/wmi"

	"github.com/jaypipes/ghw/pkg/util"
)

const wqlProduct = "SELECT Caption, Description, IdentifyingNumber, Name, SKUNumber, Vendor, Version, UUID FROM Win32_ComputerSystemProduct"

type win32Product struct {
	Caption           *string
	Description       *string
	IdentifyingNumber *string
	Name              *string
	SKUNumber         *string
	Vendor            *string
	Version           *string
	UUID              *string
}

func (i *Info) load() error {
	// Getting data from WMI
	var win32ProductDescriptions []win32Product
	// Assuming the first product is the host...
	if err := wmi.Query(wqlProduct, &win32ProductDescriptions); err != nil {
		return err
	}
	if len(win32ProductDescriptions) > 0 {
		i.Family = util.UNKNOWN
		i.Name = *win32ProductDescriptions[0].Name
		i.Vendor = *win32ProductDescriptions[0].Vendor
		i.SerialNumber = *win32ProductDescriptions[0].IdentifyingNumber
		i.UUID = *win32ProductDescriptions[0].UUID
		i.SKU = *win32ProductDescriptions[0].SKUNumber
		i.Version = *win32ProductDescriptions[0].Version
	}

	return nil
}
