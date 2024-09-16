//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package bios

import (
	"fmt"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/marshal"
	"github.com/jaypipes/ghw/pkg/option"
	"github.com/jaypipes/ghw/pkg/util"
)

// Info defines BIOS release information
type Info struct {
	ctx     *context.Context
	Vendor  string `json:"vendor"`
	Version string `json:"version"`
	Date    string `json:"date"`
}

func (i *Info) String() string {

	vendorStr := ""
	if i.Vendor != "" {
		vendorStr = " vendor=" + i.Vendor
	}
	versionStr := ""
	if i.Version != "" {
		versionStr = " version=" + i.Version
	}
	dateStr := ""
	if i.Date != "" && i.Date != util.UNKNOWN {
		dateStr = " date=" + i.Date
	}

	res := fmt.Sprintf(
		"bios%s%s%s",
		vendorStr,
		versionStr,
		dateStr,
	)
	return res
}

// New returns a pointer to a Info struct containing information
// about the host's BIOS
func New(opts ...*option.Option) (*Info, error) {
	ctx := context.New(opts...)
	info := &Info{ctx: ctx}
	if err := ctx.Do(info.load); err != nil {
		return nil, err
	}
	return info, nil
}

// simple private struct used to encapsulate BIOS information in a top-level
// "bios" YAML/JSON map/object key
type biosPrinter struct {
	Info *Info `json:"bios"`
}

// YAMLString returns a string with the BIOS information formatted as YAML
// under a top-level "dmi:" key
func (info *Info) YAMLString() string {
	return marshal.SafeYAML(info.ctx, biosPrinter{info})
}

// JSONString returns a string with the BIOS information formatted as JSON
// under a top-level "bios:" key
func (info *Info) JSONString(indent bool) string {
	return marshal.SafeJSON(info.ctx, biosPrinter{info}, indent)
}
