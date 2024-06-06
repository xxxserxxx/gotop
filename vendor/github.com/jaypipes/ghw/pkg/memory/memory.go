//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package memory

import (
	"fmt"
	"math"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/marshal"
	"github.com/jaypipes/ghw/pkg/option"
	"github.com/jaypipes/ghw/pkg/unitutil"
	"github.com/jaypipes/ghw/pkg/util"
)

type Module struct {
	Label        string `json:"label"`
	Location     string `json:"location"`
	SerialNumber string `json:"serial_number"`
	SizeBytes    int64  `json:"size_bytes"`
	Vendor       string `json:"vendor"`
}

type Area struct {
	TotalPhysicalBytes int64 `json:"total_physical_bytes"`
	TotalUsableBytes   int64 `json:"total_usable_bytes"`
	// An array of sizes, in bytes, of memory pages supported in this area
	SupportedPageSizes []uint64  `json:"supported_page_sizes"`
	Modules            []*Module `json:"modules"`
}

func (a *Area) String() string {
	tpbs := util.UNKNOWN
	if a.TotalPhysicalBytes > 0 {
		tpb := a.TotalPhysicalBytes
		unit, unitStr := unitutil.AmountString(tpb)
		tpb = int64(math.Ceil(float64(a.TotalPhysicalBytes) / float64(unit)))
		tpbs = fmt.Sprintf("%d%s", tpb, unitStr)
	}
	tubs := util.UNKNOWN
	if a.TotalUsableBytes > 0 {
		tub := a.TotalUsableBytes
		unit, unitStr := unitutil.AmountString(tub)
		tub = int64(math.Ceil(float64(a.TotalUsableBytes) / float64(unit)))
		tubs = fmt.Sprintf("%d%s", tub, unitStr)
	}
	return fmt.Sprintf("memory (%s physical, %s usable)", tpbs, tubs)
}

type Info struct {
	ctx *context.Context
	Area
}

func New(opts ...*option.Option) (*Info, error) {
	ctx := context.New(opts...)
	info := &Info{ctx: ctx}
	if err := ctx.Do(info.load); err != nil {
		return nil, err
	}
	return info, nil
}

func (i *Info) String() string {
	return i.Area.String()
}

// simple private struct used to encapsulate memory information in a top-level
// "memory" YAML/JSON map/object key
type memoryPrinter struct {
	Info *Info `json:"memory"`
}

// YAMLString returns a string with the memory information formatted as YAML
// under a top-level "memory:" key
func (i *Info) YAMLString() string {
	return marshal.SafeYAML(i.ctx, memoryPrinter{i})
}

// JSONString returns a string with the memory information formatted as JSON
// under a top-level "memory:" key
func (i *Info) JSONString(indent bool) string {
	return marshal.SafeJSON(i.ctx, memoryPrinter{i}, indent)
}
