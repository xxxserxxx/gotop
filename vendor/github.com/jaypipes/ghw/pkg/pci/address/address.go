//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package address

import (
	"regexp"
	"strings"
)

var (
	regexAddress *regexp.Regexp = regexp.MustCompile(
		`^(([0-9a-f]{0,4}):)?([0-9a-f]{2}):([0-9a-f]{2})\.([0-9a-f]{1})$`,
	)
)

// Address contains the components of a PCI Address
type Address struct {
	Domain   string
	Bus      string
	Device   string
	Function string
}

// String() returns the canonical [D]BDF representation of this Address
func (addr *Address) String() string {
	return addr.Domain + ":" + addr.Bus + ":" + addr.Device + "." + addr.Function
}

// FromString returns an Address struct from an ddress string in either
// $BUS:$DEVICE.$FUNCTION (BDF) format or it can be a full PCI address that
// includes the 4-digit $DOMAIN information as well:
// $DOMAIN:$BUS:$DEVICE.$FUNCTION.
//
// Returns "" if the address string wasn't a valid PCI address.
func FromString(address string) *Address {
	addrLowered := strings.ToLower(address)
	matches := regexAddress.FindStringSubmatch(addrLowered)
	if len(matches) == 6 {
		dom := "0000"
		if matches[1] != "" {
			dom = matches[2]
		}
		return &Address{
			Domain:   dom,
			Bus:      matches[3],
			Device:   matches[4],
			Function: matches[5],
		}
	}
	return nil
}
