package smart

import (
	"errors"
	"fmt"
)

// ErrOSUnsupported is returned on unsupported operating systems.
var ErrOSUnsupported = errors.New("os not supported")

type GenericAttributes struct {
	// Temperature represents the device temperature in Celsius
	Temperature uint64
	// Read represents a number of data units (LBA) read
	Read uint64
	// Written represents a number of data units (LBA) written
	Written uint64
	// PowerOnHours represents a power on time in hours
	PowerOnHours uint64
	// PowerCycles represents a power cycles
	PowerCycles uint64
}

type Device interface {
	Type() string
	Close() error

	// ReadGenericAttributes is an *experimental* API for quick access to the most common device SMART properties
	// This API as well as content of GenericAttributes is subject for a change.
	ReadGenericAttributes() (*GenericAttributes, error)
}

func Open(path string) (Device, error) {
	n, err := OpenNVMe(path)
	if err == nil {
		_, _, err := n.Identify()
		if err == nil {
			return n, nil
		}
		n.Close()
	}

	a, err := OpenSata(path)
	if err == nil {
		return a, nil
	}

	s, err := OpenScsi(path)
	if err == nil {
		return s, nil
	}

	return nil, fmt.Errorf("unknown drive type")
}
