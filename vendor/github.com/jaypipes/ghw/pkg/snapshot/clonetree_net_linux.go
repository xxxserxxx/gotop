//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

import (
	"strings"
)

// ExpectedCloneNetContent returns a slice of strings pertaning to the network interfaces ghw
// cares about. We cannot use a static list because we want to filter away the virtual devices,
// which  ghw doesn't concern itself about. So we need to do some runtime discovery.
// Additionally, we want to make sure to clone the backing device data.
func ExpectedCloneNetContent() []string {
	ifaceEntries := []string{
		"addr_assign_type",
		// intentionally avoid to clone "address" to avoid to leak any host-idenfifiable data.
	}

	filterLink := func(linkDest string) bool {
		return !strings.Contains(linkDest, "devices/virtual/net")
	}

	return cloneContentByClass("net", ifaceEntries, filterNone, filterLink)
}
