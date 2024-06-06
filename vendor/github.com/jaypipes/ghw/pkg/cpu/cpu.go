//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package cpu

import (
	"fmt"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/marshal"
	"github.com/jaypipes/ghw/pkg/option"
)

// ProcessorCore describes a physical host processor core. A processor core is
// a separate processing unit within some types of central processing units
// (CPU).
type ProcessorCore struct {
	// ID is the `uint32` identifier that the host gave this core. Note that
	// this does *not* necessarily equate to a zero-based index of the core
	// within a physical package. For example, the core IDs for an Intel Core
	// i7 are 0, 1, 2, 8, 9, and 10
	ID int `json:"id"`
	// Index is the zero-based index of the core on the physical processor
	// package
	Index int `json:"index"`
	// NumThreads is the number of hardware threads associated with the core
	NumThreads uint32 `json:"total_threads"`
	// LogicalProcessors is a slice of ints representing the logical processor
	// IDs assigned to any processing unit for the core
	LogicalProcessors []int `json:"logical_processors"`
}

// String returns a short string indicating important information about the
// processor core
func (c *ProcessorCore) String() string {
	return fmt.Sprintf(
		"processor core #%d (%d threads), logical processors %v",
		c.Index,
		c.NumThreads,
		c.LogicalProcessors,
	)
}

// Processor describes a physical host central processing unit (CPU).
type Processor struct {
	// ID is the physical processor `uint32` ID according to the system
	ID int `json:"id"`
	// NumCores is the number of physical cores in the processor package
	NumCores uint32 `json:"total_cores"`
	// NumThreads is the number of hardware threads in the processor package
	NumThreads uint32 `json:"total_threads"`
	// Vendor is a string containing the vendor name
	Vendor string `json:"vendor"`
	// Model` is a string containing the vendor's model name
	Model string `json:"model"`
	// Capabilities is a slice of strings indicating the features the processor
	// has enabled
	Capabilities []string `json:"capabilities"`
	// Cores is a slice of ProcessorCore` struct pointers that are packed onto
	// this physical processor
	Cores []*ProcessorCore `json:"cores"`
}

// HasCapability returns true if the Processor has the supplied cpuid
// capability, false otherwise. Example of cpuid capabilities would be 'vmx' or
// 'sse4_2'. To see a list of potential cpuid capabilitiies, see the section on
// CPUID feature bits in the following article:
//
// https://en.wikipedia.org/wiki/CPUID
func (p *Processor) HasCapability(find string) bool {
	for _, c := range p.Capabilities {
		if c == find {
			return true
		}
	}
	return false
}

// String returns a short string describing the Processor
func (p *Processor) String() string {
	ncs := "cores"
	if p.NumCores == 1 {
		ncs = "core"
	}
	nts := "threads"
	if p.NumThreads == 1 {
		nts = "thread"
	}
	return fmt.Sprintf(
		"physical package #%d (%d %s, %d hardware %s)",
		p.ID,
		p.NumCores,
		ncs,
		p.NumThreads,
		nts,
	)
}

// Info describes all central processing unit (CPU) functionality on a host.
// Returned by the `ghw.CPU()` function.
type Info struct {
	ctx *context.Context
	// TotalCores is the total number of physical cores the host system
	// contains
	TotalCores uint32 `json:"total_cores"`
	// TotalThreads is the total number of hardware threads the host system
	// contains
	TotalThreads uint32 `json:"total_threads"`
	// Processors is a slice of Processor struct pointers, one for each
	// physical processor package contained in the host
	Processors []*Processor `json:"processors"`
}

// New returns a pointer to an Info struct that contains information about the
// CPUs on the host system
func New(opts ...*option.Option) (*Info, error) {
	ctx := context.New(opts...)
	info := &Info{ctx: ctx}
	if err := ctx.Do(info.load); err != nil {
		return nil, err
	}
	return info, nil
}

// String returns a short string indicating a summary of CPU information
func (i *Info) String() string {
	nps := "packages"
	if len(i.Processors) == 1 {
		nps = "package"
	}
	ncs := "cores"
	if i.TotalCores == 1 {
		ncs = "core"
	}
	nts := "threads"
	if i.TotalThreads == 1 {
		nts = "thread"
	}
	return fmt.Sprintf(
		"cpu (%d physical %s, %d %s, %d hardware %s)",
		len(i.Processors),
		nps,
		i.TotalCores,
		ncs,
		i.TotalThreads,
		nts,
	)
}

// simple private struct used to encapsulate cpu information in a top-level
// "cpu" YAML/JSON map/object key
type cpuPrinter struct {
	Info *Info `json:"cpu"`
}

// YAMLString returns a string with the cpu information formatted as YAML
// under a top-level "cpu:" key
func (i *Info) YAMLString() string {
	return marshal.SafeYAML(i.ctx, cpuPrinter{i})
}

// JSONString returns a string with the cpu information formatted as JSON
// under a top-level "cpu:" key
func (i *Info) JSONString(indent bool) string {
	return marshal.SafeJSON(i.ctx, cpuPrinter{i}, indent)
}
