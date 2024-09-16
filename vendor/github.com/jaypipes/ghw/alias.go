//
// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package ghw

import (
	"github.com/jaypipes/ghw/pkg/baseboard"
	"github.com/jaypipes/ghw/pkg/bios"
	"github.com/jaypipes/ghw/pkg/block"
	"github.com/jaypipes/ghw/pkg/chassis"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/gpu"
	"github.com/jaypipes/ghw/pkg/memory"
	"github.com/jaypipes/ghw/pkg/net"
	"github.com/jaypipes/ghw/pkg/option"
	"github.com/jaypipes/ghw/pkg/pci"
	pciaddress "github.com/jaypipes/ghw/pkg/pci/address"
	"github.com/jaypipes/ghw/pkg/product"
	"github.com/jaypipes/ghw/pkg/topology"
)

type WithOption = option.Option

var (
	WithChroot      = option.WithChroot
	WithSnapshot    = option.WithSnapshot
	WithAlerter     = option.WithAlerter
	WithNullAlerter = option.WithNullAlerter
	// match the existing environ variable to minimize surprises
	WithDisableWarnings = option.WithNullAlerter
	WithDisableTools    = option.WithDisableTools
	WithPathOverrides   = option.WithPathOverrides
)

type SnapshotOptions = option.SnapshotOptions

type PathOverrides = option.PathOverrides

type CPUInfo = cpu.Info

var (
	CPU = cpu.New
)

type MemoryArea = memory.Area
type MemoryInfo = memory.Info
type MemoryCacheType = memory.CacheType
type MemoryModule = memory.Module

const (
	MEMORY_CACHE_TYPE_UNIFIED     = memory.CACHE_TYPE_UNIFIED
	MEMORY_CACHE_TYPE_INSTRUCTION = memory.CACHE_TYPE_INSTRUCTION
	MEMORY_CACHE_TYPE_DATA        = memory.CACHE_TYPE_DATA
)

var (
	Memory = memory.New
)

type BlockInfo = block.Info
type Disk = block.Disk
type Partition = block.Partition

var (
	Block = block.New
)

type DriveType = block.DriveType

const (
	DRIVE_TYPE_UNKNOWN = block.DRIVE_TYPE_UNKNOWN
	DRIVE_TYPE_HDD     = block.DRIVE_TYPE_HDD
	DRIVE_TYPE_FDD     = block.DRIVE_TYPE_FDD
	DRIVE_TYPE_ODD     = block.DRIVE_TYPE_ODD
	DRIVE_TYPE_SSD     = block.DRIVE_TYPE_SSD
)

type StorageController = block.StorageController

const (
	STORAGE_CONTROLLER_UNKNOWN = block.STORAGE_CONTROLLER_UNKNOWN
	STORAGE_CONTROLLER_IDE     = block.STORAGE_CONTROLLER_IDE
	STORAGE_CONTROLLER_SCSI    = block.STORAGE_CONTROLLER_SCSI
	STORAGE_CONTROLLER_NVME    = block.STORAGE_CONTROLLER_NVME
	STORAGE_CONTROLLER_VIRTIO  = block.STORAGE_CONTROLLER_VIRTIO
	STORAGE_CONTROLLER_MMC     = block.STORAGE_CONTROLLER_MMC
)

type NetworkInfo = net.Info
type NIC = net.NIC
type NICCapability = net.NICCapability

var (
	Network = net.New
)

type BIOSInfo = bios.Info

var (
	BIOS = bios.New
)

type ChassisInfo = chassis.Info

var (
	Chassis = chassis.New
)

type BaseboardInfo = baseboard.Info

var (
	Baseboard = baseboard.New
)

type TopologyInfo = topology.Info
type TopologyNode = topology.Node

var (
	Topology = topology.New
)

type Architecture = topology.Architecture

const (
	ARCHITECTURE_SMP  = topology.ARCHITECTURE_SMP
	ARCHITECTURE_NUMA = topology.ARCHITECTURE_NUMA
)

type PCIInfo = pci.Info
type PCIAddress = pciaddress.Address
type PCIDevice = pci.Device

var (
	PCI                  = pci.New
	PCIAddressFromString = pciaddress.FromString
)

type ProductInfo = product.Info

var (
	Product = product.New
)

type GPUInfo = gpu.Info
type GraphicsCard = gpu.GraphicsCard

var (
	GPU = gpu.New
)
