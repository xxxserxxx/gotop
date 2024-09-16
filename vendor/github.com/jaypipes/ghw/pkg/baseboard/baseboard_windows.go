// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package baseboard

import (
	"github.com/StackExchange/wmi"
)

const wqlBaseboard = "SELECT Manufacturer, SerialNumber, Tag, Version, Product FROM Win32_BaseBoard"

type win32Baseboard struct {
	Manufacturer *string
	SerialNumber *string
	Tag          *string
	Version      *string
	Product      *string
}

func (i *Info) load() error {
	// Getting data from WMI
	var win32BaseboardDescriptions []win32Baseboard
	if err := wmi.Query(wqlBaseboard, &win32BaseboardDescriptions); err != nil {
		return err
	}
	if len(win32BaseboardDescriptions) > 0 {
		i.AssetTag = *win32BaseboardDescriptions[0].Tag
		i.SerialNumber = *win32BaseboardDescriptions[0].SerialNumber
		i.Vendor = *win32BaseboardDescriptions[0].Manufacturer
		i.Version = *win32BaseboardDescriptions[0].Version
		i.Product = *win32BaseboardDescriptions[0].Product
	}

	return nil
}
