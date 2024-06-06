package smart

// #cgo darwin LDFLAGS: -framework IOKit -framework CoreFoundation
// #include "nvme_darwin.h"
import "C"

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"
)

type NVMeDevice struct {
	ptr unsafe.Pointer // opaque pointer to underlying macosx nvme struct
}

func OpenNVMe(name string) (*NVMeDevice, error) {
	dev := NVMeDevice{}

	if res := C.smart_nvme_open_darwin(C.CString(name), &dev.ptr); res != 0 {
		return nil, fmt.Errorf("open darwin device error: 0x%x", res)
	}

	return &dev, nil
}

func (d *NVMeDevice) Close() error {
	C.smart_nvme_close_darwin(d.ptr)
	d.ptr = unsafe.Pointer(nil)
	return nil
}

func (d *NVMeDevice) ReadSMART() (*NvmeSMARTLog, error) {
	buf := make([]byte, 512)
	if err := C.smart_nvme_readsmart_darwin(d.ptr, unsafe.Pointer(&buf[0])); err != 0 {
		return nil, fmt.Errorf("smart_nvme_readsmart_darwin: %v", err)
	}
	var sl NvmeSMARTLog
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &sl); err != nil {
		return nil, err
	}

	return &sl, nil
}

func (d *NVMeDevice) readControllerIdentifyData() (*NvmeIdentController, error) {
	buf := make([]byte, 4096)
	if err := C.smart_nvme_identify_darwin(d.ptr, unsafe.Pointer(&buf[0]), 0); err != 0 {
		return nil, fmt.Errorf("smart_nvme_identify_darwin: %v", err)
	}
	var controller NvmeIdentController
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &controller); err != nil {
		return nil, err
	}

	return &controller, nil
}

func (d *NVMeDevice) readNamespaceIdentifyData(nsid int) (*NvmeIdentNamespace, error) {
	var n NvmeIdentNamespace
	if err := C.smart_nvme_identify_darwin(d.ptr, unsafe.Pointer(&n), C.uint(nsid)); err != 0 {
		return nil, fmt.Errorf("smart_nvme_identify_darwin: %X", err)
	}

	return &n, nil
}
