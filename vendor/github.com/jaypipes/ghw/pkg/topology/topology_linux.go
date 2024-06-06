// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package topology

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw/pkg/context"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/linuxpath"
	"github.com/jaypipes/ghw/pkg/memory"
)

func (i *Info) load() error {
	i.Nodes = topologyNodes(i.ctx)
	if len(i.Nodes) == 1 {
		i.Architecture = ARCHITECTURE_SMP
	} else {
		i.Architecture = ARCHITECTURE_NUMA
	}
	return nil
}

func topologyNodes(ctx *context.Context) []*Node {
	paths := linuxpath.New(ctx)
	nodes := make([]*Node, 0)

	files, err := ioutil.ReadDir(paths.SysDevicesSystemNode)
	if err != nil {
		ctx.Warn("failed to determine nodes: %s\n", err)
		return nodes
	}
	for _, file := range files {
		filename := file.Name()
		if !strings.HasPrefix(filename, "node") {
			continue
		}
		node := &Node{}
		nodeID, err := strconv.Atoi(filename[4:])
		if err != nil {
			ctx.Warn("failed to determine node ID: %s\n", err)
			return nodes
		}
		node.ID = nodeID
		cores, err := cpu.CoresForNode(ctx, nodeID)
		if err != nil {
			ctx.Warn("failed to determine cores for node: %s\n", err)
			return nodes
		}
		node.Cores = cores
		caches, err := memory.CachesForNode(ctx, nodeID)
		if err != nil {
			ctx.Warn("failed to determine caches for node: %s\n", err)
			return nodes
		}
		node.Caches = caches

		distances, err := distancesForNode(ctx, nodeID)
		if err != nil {
			ctx.Warn("failed to determine node distances for node: %s\n", err)
			return nodes
		}
		node.Distances = distances

		area, err := memory.AreaForNode(ctx, nodeID)
		if err != nil {
			ctx.Warn("failed to determine memory area for node: %s\n", err)
			return nodes
		}
		node.Memory = area

		nodes = append(nodes, node)
	}
	return nodes
}

func distancesForNode(ctx *context.Context, nodeID int) ([]int, error) {
	paths := linuxpath.New(ctx)
	path := filepath.Join(
		paths.SysDevicesSystemNode,
		fmt.Sprintf("node%d", nodeID),
		"distance",
	)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	items := strings.Fields(strings.TrimSpace(string(data)))
	dists := make([]int, len(items)) // TODO: can a NUMA cell be offlined?
	for idx, item := range items {
		dist, err := strconv.Atoi(item)
		if err != nil {
			return dists, err
		}
		dists[idx] = dist
	}
	return dists, nil
}
