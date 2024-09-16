// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package chassis

import (
	"github.com/StackExchange/wmi"

	"github.com/jaypipes/ghw/pkg/util"
)

const wqlChassis = "SELECT Caption, Description, Name, Manufacturer, Model, SerialNumber, Tag, TypeDescriptions, Version FROM CIM_Chassis"

type win32Chassis struct {
	Caption          *string
	Description      *string
	Name             *string
	Manufacturer     *string
	Model            *string
	SerialNumber     *string
	Tag              *string
	TypeDescriptions []string
	Version          *string
}

func (i *Info) load() error {
	// Getting data from WMI
	var win32ChassisDescriptions []win32Chassis
	if err := wmi.Query(wqlChassis, &win32ChassisDescriptions); err != nil {
		return err
	}
	if len(win32ChassisDescriptions) > 0 {
		i.AssetTag = *win32ChassisDescriptions[0].Tag
		i.SerialNumber = *win32ChassisDescriptions[0].SerialNumber
		i.Type = util.UNKNOWN // TODO:
		i.TypeDescription = *win32ChassisDescriptions[0].Model
		i.Vendor = *win32ChassisDescriptions[0].Manufacturer
		i.Version = *win32ChassisDescriptions[0].Version
	}
	return nil
}
