//go:build !linux
// +build !linux

// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package pci

import (
	"runtime"

	"github.com/pkg/errors"
)

func (i *Info) load() error {
	return errors.New("pciFillInfo not implemented on " + runtime.GOOS)
}

// GetDevice returns a pointer to a Device struct that describes the PCI
// device at the requested address. If no such device could be found, returns
// nil
func (info *Info) GetDevice(address string) *Device {
	return nil
}

// ListDevices returns a list of pointers to Device structs present on the
// host system
func (info *Info) ListDevices() []*Device {
	return nil
}
