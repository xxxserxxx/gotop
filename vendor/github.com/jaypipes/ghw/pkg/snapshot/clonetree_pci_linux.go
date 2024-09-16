//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package snapshot

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	pciaddr "github.com/jaypipes/ghw/pkg/pci/address"
)

const (
	// root directory: entry point to start scanning the PCI forest
	// warning: don't use the context package here, this means not even the linuxpath package.
	// TODO(fromani) remove the path duplication
	sysBusPCIDir = "/sys/bus/pci/devices"
)

// ExpectedClonePCIContent return a slice of glob patterns which represent the pseudofiles
// ghw cares about, pertaining to PCI devices only.
// Beware: the content is host-specific, because the PCI topology is host-dependent and unpredictable.
func ExpectedClonePCIContent() []string {
	fileSpecs := []string{
		"/sys/bus/pci/drivers/*",
	}
	pciRoots := []string{
		sysBusPCIDir,
	}
	for {
		if len(pciRoots) == 0 {
			break
		}
		pciRoot := pciRoots[0]
		pciRoots = pciRoots[1:]
		specs, roots := scanPCIDeviceRoot(pciRoot)
		pciRoots = append(pciRoots, roots...)
		fileSpecs = append(fileSpecs, specs...)
	}
	return fileSpecs
}

// scanPCIDeviceRoot reports a slice of glob patterns which represent the pseudofiles
// ghw cares about pertaining to all the PCI devices connected to the bus connected from the
// given root; usually (but not always) a CPU packages has 1+ PCI(e) roots, forming the first
// level; more PCI bridges are (usually) attached to this level, creating deep nested trees.
// hence we need to scan all possible roots, to make sure not to miss important devices.
//
// note about notifying errors. This function and its helper functions do use trace() everywhere
// to report recoverable errors, even though it would have been appropriate to use Warn().
// This is unfortunate, and again a byproduct of the fact we cannot use context.Context to avoid
// circular dependencies.
// TODO(fromani): switch to Warn() as soon as we figure out how to break this circular dep.
func scanPCIDeviceRoot(root string) (fileSpecs []string, pciRoots []string) {
	trace("scanning PCI device root %q\n", root)

	perDevEntries := []string{
		"class",
		"device",
		"driver",
		"irq",
		"local_cpulist",
		"modalias",
		"numa_node",
		"revision",
		"vendor",
	}
	entries, err := ioutil.ReadDir(root)
	if err != nil {
		return []string{}, []string{}
	}
	for _, entry := range entries {
		entryName := entry.Name()
		if addr := pciaddr.FromString(entryName); addr == nil {
			// doesn't look like a entry we care about
			// This is by far and large the most likely path
			// hence we should NOT trace/warn here.
			continue
		}

		entryPath := filepath.Join(root, entryName)
		pciEntry, err := findPCIEntryFromPath(root, entryName)
		if err != nil {
			trace("error scanning %q: %v", entryName, err)
			continue
		}

		trace("PCI entry is %q\n", pciEntry)
		fileSpecs = append(fileSpecs, entryPath)
		for _, perNetEntry := range perDevEntries {
			fileSpecs = append(fileSpecs, filepath.Join(pciEntry, perNetEntry))
		}

		if isPCIBridge(entryPath) {
			trace("adding new PCI root %q\n", entryName)
			pciRoots = append(pciRoots, pciEntry)
		}
	}
	return fileSpecs, pciRoots
}

func findPCIEntryFromPath(root, entryName string) (string, error) {
	entryPath := filepath.Join(root, entryName)
	fi, err := os.Lstat(entryPath)
	if err != nil {
		return "", fmt.Errorf("stat(%s) failed: %v\n", entryPath, err)
	}
	if fi.Mode()&os.ModeSymlink == 0 {
		// regular file, nothing to resolve
		return entryPath, nil
	}
	// resolve symlink
	target, err := os.Readlink(entryPath)
	trace("entry %q is symlink resolved to %q\n", entryPath, target)
	if err != nil {
		return "", fmt.Errorf("readlink(%s) failed: %v - skipped\n", entryPath, err)
	}
	return filepath.Clean(filepath.Join(root, target)), nil
}

func isPCIBridge(entryPath string) bool {
	subNodes, err := ioutil.ReadDir(entryPath)
	if err != nil {
		// this is so unlikely we don't even return error. But we trace just in case.
		trace("error scanning device entry path %q: %v", entryPath, err)
		return false
	}
	for _, subNode := range subNodes {
		if !subNode.IsDir() {
			continue
		}
		if addr := pciaddr.FromString(subNode.Name()); addr != nil {
			// we got an entry in the directory pertaining to this device
			// which is a directory itself and it is named like a PCI address.
			// Hence we infer the device we are considering is a PCI bridge of sorts.
			// This is is indeed a bit brutal, but the only possible alternative
			// (besides blindly copying everything in /sys/bus/pci/devices) is
			// to detect the type of the device and pick only the bridges.
			// This approach duplicates the logic within the `pci` subkpg
			// - or forces us into awkward dep cycles, and has poorer forward
			// compatibility.
			return true
		}
	}
	return false
}
