// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.
//

package topology

import (
	"encoding/binary"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	rcFailure                                = 0
	sizeofLogicalProcessorInfo               = 32
	errInsufficientBuffer      syscall.Errno = 122

	relationProcessorCore    = 0
	relationNUMANode         = 1
	relationCache            = 2
	relationProcessorPackage = 3
	relationGroup            = 4
)

func (i *Info) load() error {
	nodes, err := topologyNodes()
	if err != nil {
		return err
	}
	i.Nodes = nodes
	if len(nodes) == 1 {
		i.Architecture = ARCHITECTURE_SMP
	} else {
		i.Architecture = ARCHITECTURE_NUMA
	}
	return nil
}

func topologyNodes() ([]*Node, error) {
	nodes := make([]*Node, 0)
	lpis, err := getWin32LogicalProcessorInfos()
	if err != nil {
		return nil, err
	}
	for _, lpi := range lpis {
		switch lpi.relationship {
		case relationNUMANode:
			nodes = append(nodes, &Node{
				ID: lpi.numaNodeID(),
			})
		case relationProcessorCore:
			// TODO(jaypipes): associated LP to processor core
		case relationProcessorPackage:
			// ignore
		case relationCache:
			// TODO(jaypipes) handle cache layers
		default:
			return nil, fmt.Errorf("Unknown LOGICAL_PROCESSOR_RELATIONSHIP value: %d", lpi.relationship)

		}
	}
	return nodes, nil
}

// This is the CACHE_DESCRIPTOR struct in the Win32 API
type cacheDescriptor struct {
	level         uint8
	associativity uint8
	lineSize      uint16
	size          uint32
	cacheType     uint32
}

// This is the SYSTEM_LOGICAL_PROCESSOR_INFORMATION struct in the Win32 API
type logicalProcessorInfo struct {
	processorMask uint64
	relationship  uint64
	// The following dummyunion member is a representation of this part of
	// the SYSTEM_LOGICAL_PROCESSOR_INFORMATION struct:
	//
	//  union {
	//    struct {
	//      BYTE Flags;
	//    } ProcessorCore;
	//    struct {
	//      DWORD NodeNumber;
	//    } NumaNode;
	//    CACHE_DESCRIPTOR Cache;
	//    ULONGLONG        Reserved[2];
	//  } DUMMYUNIONNAME;
	dummyunion [16]byte
}

// numaNodeID returns the NUMA node's identifier from the logical processor
// information struct by grabbing the integer representation of the struct's
// NumaNode unioned data element
func (lpi *logicalProcessorInfo) numaNodeID() int {
	if lpi.relationship != relationNUMANode {
		return -1
	}
	return int(binary.LittleEndian.Uint16(lpi.dummyunion[0:]))
}

// ref: https://docs.microsoft.com/en-us/windows/win32/api/sysinfoapi/nf-sysinfoapi-getlogicalprocessorinformation
func getWin32LogicalProcessorInfos() (
	[]*logicalProcessorInfo,
	error,
) {
	lpis := make([]*logicalProcessorInfo, 0)
	win32api := syscall.NewLazyDLL("kernel32.dll")
	glpi := win32api.NewProc("GetLogicalProcessorInformation")

	// The way the GetLogicalProcessorInformation (GLPI) Win32 API call
	// works is wonky, but consistent with the Win32 API calling structure.
	// Basically, you need to first call the GLPI API with a NUL pointerr
	// and a pointer to an integer. That first call to the API should
	// return ERROR_INSUFFICIENT_BUFFER, which is the indication that the
	// supplied buffer pointer is NUL and needs to have memory allocated to
	// it of an amount equal to the value of the integer pointer argument.
	// Once the buffer is allocated this amount of space, the GLPI API call
	// is again called. This time, the return value should be 0 and the
	// buffer will have been set to an array of
	// SYSTEM_LOGICAL_PROCESSOR_INFORMATION structs.
	toAllocate := uint32(0)
	// first, figure out how much we need
	rc, _, win32err := glpi.Call(uintptr(0), uintptr(unsafe.Pointer(&toAllocate)))
	if rc == rcFailure {
		if win32err != errInsufficientBuffer {
			return nil, fmt.Errorf("GetLogicalProcessorInformation Win32 API initial call failed to return ERROR_INSUFFICIENT_BUFFER")
		}
	} else {
		// This shouldn't happen because buffer hasn't yet been allocated...
		return nil, fmt.Errorf("GetLogicalProcessorInformation Win32 API initial call returned success instead of failure with ERROR_INSUFFICIENT_BUFFER")
	}

	// OK, now we actually allocate a raw buffer to fill with some number
	// of SYSTEM_LOGICAL_PROCESSOR_INFORMATION structs
	b := make([]byte, toAllocate)
	rc, _, win32err = glpi.Call(uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&toAllocate)))
	if rc == rcFailure {
		return nil, fmt.Errorf("GetLogicalProcessorInformation Win32 API call failed to set supplied buffer. Win32 system error: %s", win32err)
	}

	for x := uint32(0); x < toAllocate; x += sizeofLogicalProcessorInfo {
		lpiraw := b[x : x+sizeofLogicalProcessorInfo]
		lpi := &logicalProcessorInfo{
			processorMask: binary.LittleEndian.Uint64(lpiraw[0:]),
			relationship:  binary.LittleEndian.Uint64(lpiraw[8:]),
		}
		copy(lpi.dummyunion[0:16], lpiraw[16:32])
		lpis = append(lpis, lpi)
	}
	return lpis, nil
}
