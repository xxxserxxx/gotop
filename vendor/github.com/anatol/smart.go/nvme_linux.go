package smart

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// https:// nvmexpress.org/wp-content/uploads/NVM-Express-Base-Specification-2.0b-2021.12.18-Ratified.pdf
// https:// nvmexpress.org/wp-content/uploads/NVM-Express-NVM-Command-Set-Specification-1.0b-2021.12.18-Ratified.pdf

// include/uapi/linux/nvme_ioctl.h

var nvmeIoctlAdmin64Cmd = iowr('N', 0x47, unsafe.Sizeof(nvmePassthruCmd64{}))

type nvmePassthruCmd64 struct {
	opcode      uint8
	flags       uint8
	_           uint16
	nsid        uint32
	cdw2        uint32
	cdw3        uint32
	metadata    uint64
	addr        uint64
	metadataLen uint32
	dataLen     uint32
	cdw10       uint32
	cdw11       uint32
	cdw12       uint32
	cdw13       uint32
	cdw14       uint32
	cdw15       uint32
	timeoutMs   uint32
	_           uint32
	result      uint64
}

type NVMeDevice struct {
	fd int
}

func OpenNVMe(name string) (*NVMeDevice, error) {
	fd, err := unix.Open(name, unix.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	dev := NVMeDevice{
		fd: fd,
	}
	return &dev, nil
}

func (d *NVMeDevice) Close() error {
	return unix.Close(d.fd)
}

func (d *NVMeDevice) ReadSMART() (*NvmeSMARTLog, error) {
	buf := make([]byte, 512)
	if err := nvmeReadLogPage(d.fd, nvmeLogSmartInformation, buf); err != nil {
		return nil, err
	}
	var sl NvmeSMARTLog
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &sl); err != nil {
		return nil, err
	}

	return &sl, nil
}

func nvmeReadLogPage(fd int, logID uint8, buf []byte) error {
	bufLen := len(buf)

	if (bufLen < 4) || (bufLen > 0x4000) || (bufLen%4 != 0) {
		return fmt.Errorf("invalid buffer size")
	}

	cmd := nvmePassthruCmd64{
		opcode:  nvmeAdminGetLogPage,
		nsid:    0xffffffff, // controller-level SMART info
		addr:    uint64(uintptr(unsafe.Pointer(&buf[0]))),
		dataLen: uint32(bufLen),
		cdw10:   uint32(logID) | (((uint32(bufLen) / 4) - 1) << 16),
	}

	return ioctl(uintptr(fd), nvmeIoctlAdmin64Cmd, uintptr(unsafe.Pointer(&cmd)))
}

func (d *NVMeDevice) readIdentifyData(nsid, cns int, data []byte) error {
	cmd := nvmePassthruCmd64{
		opcode:  nvmeAdminIdentify,
		nsid:    uint32(nsid),
		addr:    uint64(uintptr(unsafe.Pointer(&data[0]))),
		dataLen: uint32(len(data)),
		cdw10:   uint32(cns),
	}

	return ioctl(uintptr(d.fd), nvmeIoctlAdmin64Cmd, uintptr(unsafe.Pointer(&cmd)))
}

func (d *NVMeDevice) readControllerIdentifyData() (*NvmeIdentController, error) {
	buf := make([]byte, 4096)
	if err := d.readIdentifyData(0, 1, buf); err != nil {
		return nil, err
	}
	var controller NvmeIdentController
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &controller); err != nil {
		return nil, err
	}

	return &controller, nil
}

func (d *NVMeDevice) readNamespaceIdentifyData(nsid int) (*NvmeIdentNamespace, error) {
	buf := make([]byte, 4096)
	if err := d.readIdentifyData(nsid, 0, buf); err != nil {
		return nil, err
	}
	var namespace NvmeIdentNamespace
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &namespace); err != nil {
		return nil, err
	}

	return &namespace, nil
}
