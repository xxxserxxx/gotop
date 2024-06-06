// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package baseboard

import (
	"github.com/jaypipes/ghw/pkg/linuxdmi"
)

func (i *Info) load() error {
	i.AssetTag = linuxdmi.Item(i.ctx, "board_asset_tag")
	i.SerialNumber = linuxdmi.Item(i.ctx, "board_serial")
	i.Vendor = linuxdmi.Item(i.ctx, "board_vendor")
	i.Version = linuxdmi.Item(i.ctx, "board_version")
	i.Product = linuxdmi.Item(i.ctx, "board_name")

	return nil
}
