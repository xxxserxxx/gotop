//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

import (
	"strings"
)

// ExpectedCloneGPUContent returns a slice of strings pertaining to the GPU devices ghw
// cares about. We cannot use a static list because we want to grab only the first cardX data
// (see comment in pkg/gpu/gpu_linux.go)
// Additionally, we want to make sure to clone the backing device data.
func ExpectedCloneGPUContent() []string {
	cardEntries := []string{
		"device",
	}

	filterName := func(cardName string) bool {
		if !strings.HasPrefix(cardName, "card") {
			return false
		}
		if strings.ContainsRune(cardName, '-') {
			return false
		}
		return true
	}

	return cloneContentByClass("drm", cardEntries, filterName, filterNone)
}
