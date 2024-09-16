// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package linuxpath

import (
	"fmt"
	"path/filepath"

	"github.com/jaypipes/ghw/pkg/context"
)

// PathRoots holds the roots of all the filesystem subtrees
// ghw wants to access.
type PathRoots struct {
	Etc  string
	Proc string
	Run  string
	Sys  string
	Var  string
}

// DefaultPathRoots return the canonical default value for PathRoots
func DefaultPathRoots() PathRoots {
	return PathRoots{
		Etc:  "/etc",
		Proc: "/proc",
		Run:  "/run",
		Sys:  "/sys",
		Var:  "/var",
	}
}

// PathRootsFromContext initialize PathRoots from the given Context,
// allowing overrides of the canonical default paths.
func PathRootsFromContext(ctx *context.Context) PathRoots {
	roots := DefaultPathRoots()
	if pathEtc, ok := ctx.PathOverrides["/etc"]; ok {
		roots.Etc = pathEtc
	}
	if pathProc, ok := ctx.PathOverrides["/proc"]; ok {
		roots.Proc = pathProc
	}
	if pathRun, ok := ctx.PathOverrides["/run"]; ok {
		roots.Run = pathRun
	}
	if pathSys, ok := ctx.PathOverrides["/sys"]; ok {
		roots.Sys = pathSys
	}
	if pathVar, ok := ctx.PathOverrides["/var"]; ok {
		roots.Var = pathVar
	}
	return roots
}

type Paths struct {
	VarLog                 string
	ProcMeminfo            string
	ProcCpuinfo            string
	ProcMounts             string
	SysKernelMMHugepages   string
	SysBlock               string
	SysDevicesSystemNode   string
	SysDevicesSystemMemory string
	SysBusPciDevices       string
	SysClassDRM            string
	SysClassDMI            string
	SysClassNet            string
	RunUdevData            string
}

// New returns a new Paths struct containing filepath fields relative to the
// supplied Context
func New(ctx *context.Context) *Paths {
	roots := PathRootsFromContext(ctx)
	return &Paths{
		VarLog:                 filepath.Join(ctx.Chroot, roots.Var, "log"),
		ProcMeminfo:            filepath.Join(ctx.Chroot, roots.Proc, "meminfo"),
		ProcCpuinfo:            filepath.Join(ctx.Chroot, roots.Proc, "cpuinfo"),
		ProcMounts:             filepath.Join(ctx.Chroot, roots.Proc, "self", "mounts"),
		SysKernelMMHugepages:   filepath.Join(ctx.Chroot, roots.Sys, "kernel", "mm", "hugepages"),
		SysBlock:               filepath.Join(ctx.Chroot, roots.Sys, "block"),
		SysDevicesSystemNode:   filepath.Join(ctx.Chroot, roots.Sys, "devices", "system", "node"),
		SysDevicesSystemMemory: filepath.Join(ctx.Chroot, roots.Sys, "devices", "system", "memory"),
		SysBusPciDevices:       filepath.Join(ctx.Chroot, roots.Sys, "bus", "pci", "devices"),
		SysClassDRM:            filepath.Join(ctx.Chroot, roots.Sys, "class", "drm"),
		SysClassDMI:            filepath.Join(ctx.Chroot, roots.Sys, "class", "dmi"),
		SysClassNet:            filepath.Join(ctx.Chroot, roots.Sys, "class", "net"),
		RunUdevData:            filepath.Join(ctx.Chroot, roots.Run, "udev", "data"),
	}
}

func (p *Paths) NodeCPU(nodeID int, lpID int) string {
	return filepath.Join(
		p.SysDevicesSystemNode,
		fmt.Sprintf("node%d", nodeID),
		fmt.Sprintf("cpu%d", lpID),
	)
}

func (p *Paths) NodeCPUCache(nodeID int, lpID int) string {
	return filepath.Join(
		p.NodeCPU(nodeID, lpID),
		"cache",
	)
}

func (p *Paths) NodeCPUCacheIndex(nodeID int, lpID int, cacheIndex int) string {
	return filepath.Join(
		p.NodeCPUCache(nodeID, lpID),
		fmt.Sprintf("index%d", cacheIndex),
	)
}
