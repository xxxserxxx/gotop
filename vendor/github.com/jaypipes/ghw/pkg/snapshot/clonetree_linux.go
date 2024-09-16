//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func setupScratchDir(scratchDir string) error {
	var createPaths = []string{
		"sys/block",
	}

	for _, path := range createPaths {
		if err := os.MkdirAll(filepath.Join(scratchDir, path), os.ModePerm); err != nil {
			return err
		}
	}

	return createBlockDevices(scratchDir)
}

// ExpectedCloneStaticContent return a slice of glob patterns which represent the pseudofiles
// ghw cares about, and which are independent from host specific topology or configuration,
// thus are safely represented by a static slice - e.g. they don't need to be discovered at runtime.
func ExpectedCloneStaticContent() []string {
	return []string{
		"/proc/cpuinfo",
		"/proc/meminfo",
		"/proc/self/mounts",
		"/sys/devices/system/cpu/cpu*/cache/index*/*",
		"/sys/devices/system/cpu/cpu*/topology/*",
		"/sys/devices/system/memory/block_size_bytes",
		"/sys/devices/system/memory/memory*/online",
		"/sys/devices/system/memory/memory*/state",
		"/sys/devices/system/node/has_*",
		"/sys/devices/system/node/online",
		"/sys/devices/system/node/possible",
		"/sys/devices/system/node/node*/cpu*",
		"/sys/devices/system/node/node*/distance",
		"/sys/devices/system/node/node*/meminfo",
		"/sys/devices/system/node/node*/memory*",
		"/sys/devices/system/node/node*/hugepages/hugepages-*/*",
	}
}

type filterFunc func(string) bool

// cloneContentByClass copies all the content related to a given device class
// (devClass), possibly filtering out devices whose name does NOT pass a
// filter (filterName). Each entry in `/sys/class/$CLASS` is actually a
// symbolic link. We can filter out entries depending on the link target.
// Each filter is a simple function which takes the entry name or the link
// target and must return true if the entry should be collected, false
// otherwise. Last, explicitly collect a list of attributes for each entry,
// given as list of glob patterns as `subEntries`.
// Return the final list of glob patterns to be collected.
func cloneContentByClass(devClass string, subEntries []string, filterName filterFunc, filterLink filterFunc) []string {
	var fileSpecs []string

	// warning: don't use the context package here, this means not even the linuxpath package.
	// TODO(fromani) remove the path duplication
	sysClass := filepath.Join("sys", "class", devClass)
	entries, err := ioutil.ReadDir(sysClass)
	if err != nil {
		// we should not import context, hence we can't Warn()
		return fileSpecs
	}
	for _, entry := range entries {
		devName := entry.Name()

		if !filterName(devName) {
			continue
		}

		devPath := filepath.Join(sysClass, devName)
		dest, err := os.Readlink(devPath)
		if err != nil {
			continue
		}

		if !filterLink(dest) {
			continue
		}

		// so, first copy the symlink itself
		fileSpecs = append(fileSpecs, devPath)
		// now we have to clone the content of the actual entry
		// related (and found into a subdir of) the backing hardware
		// device
		devData := filepath.Clean(filepath.Join(sysClass, dest))
		for _, subEntry := range subEntries {
			fileSpecs = append(fileSpecs, filepath.Join(devData, subEntry))
		}
	}

	return fileSpecs
}

// filterNone allows all content, filtering out none of it
func filterNone(_ string) bool {
	return true
}
