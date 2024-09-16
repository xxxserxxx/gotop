// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package product

import (
	"github.com/jaypipes/ghw/pkg/linuxdmi"
)

func (i *Info) load() error {

	i.Family = linuxdmi.Item(i.ctx, "product_family")
	i.Name = linuxdmi.Item(i.ctx, "product_name")
	i.Vendor = linuxdmi.Item(i.ctx, "sys_vendor")
	i.SerialNumber = linuxdmi.Item(i.ctx, "product_serial")
	i.UUID = linuxdmi.Item(i.ctx, "product_uuid")
	i.SKU = linuxdmi.Item(i.ctx, "product_sku")
	i.Version = linuxdmi.Item(i.ctx, "product_version")

	return nil
}
