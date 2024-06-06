//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package net

import (
	"fmt"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/marshal"
	"github.com/jaypipes/ghw/pkg/option"
)

type NICCapability struct {
	Name      string `json:"name"`
	IsEnabled bool   `json:"is_enabled"`
	CanEnable bool   `json:"can_enable"`
}

type NIC struct {
	Name         string           `json:"name"`
	MacAddress   string           `json:"mac_address"`
	IsVirtual    bool             `json:"is_virtual"`
	Capabilities []*NICCapability `json:"capabilities"`
	PCIAddress   *string          `json:"pci_address,omitempty"`
	// TODO(fromani): add other hw addresses (USB) when we support them
}

func (n *NIC) String() string {
	isVirtualStr := ""
	if n.IsVirtual {
		isVirtualStr = " (virtual)"
	}
	return fmt.Sprintf(
		"%s%s",
		n.Name,
		isVirtualStr,
	)
}

type Info struct {
	ctx  *context.Context
	NICs []*NIC `json:"nics"`
}

// New returns a pointer to an Info struct that contains information about the
// network interface controllers (NICs) on the host system
func New(opts ...*option.Option) (*Info, error) {
	ctx := context.New(opts...)
	info := &Info{ctx: ctx}
	if err := ctx.Do(info.load); err != nil {
		return nil, err
	}
	return info, nil
}

func (i *Info) String() string {
	return fmt.Sprintf(
		"net (%d NICs)",
		len(i.NICs),
	)
}

// simple private struct used to encapsulate net information in a
// top-level "net" YAML/JSON map/object key
type netPrinter struct {
	Info *Info `json:"network"`
}

// YAMLString returns a string with the net information formatted as YAML
// under a top-level "net:" key
func (i *Info) YAMLString() string {
	return marshal.SafeYAML(i.ctx, netPrinter{i})
}

// JSONString returns a string with the net information formatted as JSON
// under a top-level "net:" key
func (i *Info) JSONString(indent bool) string {
	return marshal.SafeJSON(i.ctx, netPrinter{i}, indent)
}
