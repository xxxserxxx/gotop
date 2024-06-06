//go:build !linux && !windows
// +build !linux,!windows

// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package bios

import (
	"runtime"

	"github.com/pkg/errors"
)

func (i *Info) load() error {
	return errors.New("biosFillInfo not implemented on " + runtime.GOOS)
}
