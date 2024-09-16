// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package memory

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/linuxpath"
	"github.com/jaypipes/ghw/pkg/unitutil"
)

func CachesForNode(ctx *context.Context, nodeID int) ([]*Cache, error) {
	// The /sys/devices/node/nodeX directory contains a subdirectory called
	// 'cpuX' for each logical processor assigned to the node. Each of those
	// subdirectories containers a 'cache' subdirectory which contains a number
	// of subdirectories beginning with 'index' and ending in the cache's
	// internal 0-based identifier. Those subdirectories contain a number of
	// files, including 'shared_cpu_list', 'size', and 'type' which we use to
	// determine cache characteristics.
	paths := linuxpath.New(ctx)
	path := filepath.Join(
		paths.SysDevicesSystemNode,
		fmt.Sprintf("node%d", nodeID),
	)
	caches := make(map[string]*Cache)

	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		filename := file.Name()
		if !strings.HasPrefix(filename, "cpu") {
			continue
		}
		if filename == "cpumap" || filename == "cpulist" {
			// There are two files in the node directory that start with 'cpu'
			// but are not subdirectories ('cpulist' and 'cpumap'). Ignore
			// these files.
			continue
		}
		// Grab the logical processor ID by cutting the integer from the
		// /sys/devices/system/node/nodeX/cpuX filename
		cpuPath := filepath.Join(path, filename)
		lpID, _ := strconv.Atoi(filename[3:])

		// Inspect the caches for each logical processor. There will be a
		// /sys/devices/system/node/nodeX/cpuX/cache directory containing a
		// number of directories beginning with the prefix "index" followed by
		// a number. The number indicates the level of the cache, which
		// indicates the "distance" from the processor. Each of these
		// directories contains information about the size of that level of
		// cache and the processors mapped to it.
		cachePath := filepath.Join(cpuPath, "cache")
		if _, err = os.Stat(cachePath); errors.Is(err, os.ErrNotExist) {
			continue
		}
		cacheDirFiles, err := ioutil.ReadDir(cachePath)
		if err != nil {
			return nil, err
		}
		for _, cacheDirFile := range cacheDirFiles {
			cacheDirFileName := cacheDirFile.Name()
			if !strings.HasPrefix(cacheDirFileName, "index") {
				continue
			}
			cacheIndex, _ := strconv.Atoi(cacheDirFileName[5:])

			// The cache information is repeated for each node, so here, we
			// just ensure that we only have a one Cache object for each
			// unique combination of level, type and processor map
			level := memoryCacheLevel(ctx, paths, nodeID, lpID, cacheIndex)
			cacheType := memoryCacheType(ctx, paths, nodeID, lpID, cacheIndex)
			sharedCpuMap := memoryCacheSharedCPUMap(ctx, paths, nodeID, lpID, cacheIndex)
			cacheKey := fmt.Sprintf("%d-%d-%s", level, cacheType, sharedCpuMap)

			cache, exists := caches[cacheKey]
			if !exists {
				size := memoryCacheSize(ctx, paths, nodeID, lpID, level)
				cache = &Cache{
					Level:             uint8(level),
					Type:              cacheType,
					SizeBytes:         uint64(size) * uint64(unitutil.KB),
					LogicalProcessors: make([]uint32, 0),
				}
				caches[cacheKey] = cache
			}
			cache.LogicalProcessors = append(
				cache.LogicalProcessors,
				uint32(lpID),
			)
		}
	}

	cacheVals := make([]*Cache, len(caches))
	x := 0
	for _, c := range caches {
		// ensure the cache's processor set is sorted by logical process ID
		sort.Sort(SortByLogicalProcessorId(c.LogicalProcessors))
		cacheVals[x] = c
		x++
	}

	return cacheVals, nil
}

func memoryCacheLevel(ctx *context.Context, paths *linuxpath.Paths, nodeID int, lpID int, cacheIndex int) int {
	levelPath := filepath.Join(
		paths.NodeCPUCacheIndex(nodeID, lpID, cacheIndex),
		"level",
	)
	levelContents, err := ioutil.ReadFile(levelPath)
	if err != nil {
		ctx.Warn("%s", err)
		return -1
	}
	// levelContents is now a []byte with the last byte being a newline
	// character. Trim that off and convert the contents to an integer.
	level, err := strconv.Atoi(string(levelContents[:len(levelContents)-1]))
	if err != nil {
		ctx.Warn("Unable to parse int from %s", levelContents)
		return -1
	}
	return level
}

func memoryCacheSize(ctx *context.Context, paths *linuxpath.Paths, nodeID int, lpID int, cacheIndex int) int {
	sizePath := filepath.Join(
		paths.NodeCPUCacheIndex(nodeID, lpID, cacheIndex),
		"size",
	)
	sizeContents, err := ioutil.ReadFile(sizePath)
	if err != nil {
		ctx.Warn("%s", err)
		return -1
	}
	// size comes as XK\n, so we trim off the K and the newline.
	size, err := strconv.Atoi(string(sizeContents[:len(sizeContents)-2]))
	if err != nil {
		ctx.Warn("Unable to parse int from %s", sizeContents)
		return -1
	}
	return size
}

func memoryCacheType(ctx *context.Context, paths *linuxpath.Paths, nodeID int, lpID int, cacheIndex int) CacheType {
	typePath := filepath.Join(
		paths.NodeCPUCacheIndex(nodeID, lpID, cacheIndex),
		"type",
	)
	cacheTypeContents, err := ioutil.ReadFile(typePath)
	if err != nil {
		ctx.Warn("%s", err)
		return CACHE_TYPE_UNIFIED
	}
	switch string(cacheTypeContents[:len(cacheTypeContents)-1]) {
	case "Data":
		return CACHE_TYPE_DATA
	case "Instruction":
		return CACHE_TYPE_INSTRUCTION
	default:
		return CACHE_TYPE_UNIFIED
	}
}

func memoryCacheSharedCPUMap(ctx *context.Context, paths *linuxpath.Paths, nodeID int, lpID int, cacheIndex int) string {
	scpuPath := filepath.Join(
		paths.NodeCPUCacheIndex(nodeID, lpID, cacheIndex),
		"shared_cpu_map",
	)
	sharedCpuMap, err := ioutil.ReadFile(scpuPath)
	if err != nil {
		ctx.Warn("%s", err)
		return ""
	}
	return string(sharedCpuMap[:len(sharedCpuMap)-1])
}
