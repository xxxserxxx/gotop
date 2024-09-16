package smart

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

func OpenScsi(name string) (*ScsiDevice, error) {
	fd, err := unix.Open(name, unix.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	scsi := ScsiDevice{
		fd: fd,
	}

	i, err := scsi.Inquiry()
	if err != nil {
		unix.Close(fd)
		return nil, err
	}

	// see this comment to understand why we check for deviceType
	// https://github.com/systemd/systemd/blob/58551e6ebc465227d0add8c714f9f38213b6878a/src/udev/ata_id/ata_id.c#L324-L344
	deviceType := i.Peripheral & 0x1f
	if deviceType != 0 {
		unix.Close(fd)
		return nil, fmt.Errorf("not a direct access block device")
	}

	if bytes.Equal(i.VendorIdent[:], []byte(_SATA_IDENT)) {
		unix.Close(fd)
		return nil, fmt.Errorf("it is SATA device")
	}

	return &scsi, nil
}

func (d *ScsiDevice) Close() error {
	return unix.Close(d.fd)
}

// SCSI CDB types
type (
	cdb6  [6]byte
	cdb10 [10]byte
	cdb16 [16]byte
)

type sgioError struct {
	hostStatus   uint32
	deviceStatus uint32
	driverStatus uint32
}

func (e sgioError) Error() string {
	return fmt.Sprintf("SCSI status: %#02x, transport status: %#02x, driver status: %#02x",
		e.deviceStatus, e.hostStatus, e.driverStatus)
}

func (d *ScsiDevice) Capacity() (uint64, error) {
	cdb := cdb10{_SCSI_READ_CAPACITY_10}

	respBuf := make([]byte, 8)

	if err := scsiSendCdb(d.fd, cdb[:], respBuf); err != nil {
		return 0, err
	}

	var r struct {
		LastLba uint32
		LbSize  uint32
	}
	if err := binary.Read(bytes.NewBuffer(respBuf[:]), binary.BigEndian, &r); err != nil {
		return 0, err
	}

	return uint64(r.LastLba+1) * uint64(r.LbSize), nil
}

func (d *ScsiDevice) Inquiry() (*ScsiInquiry, error) {
	return scsiInquiry(d.fd)
}

func scsiInquiry(fd int) (*ScsiInquiry, error) {
	var resp ScsiInquiry

	respBuf := make([]byte, 36) // 36 is the min response size

	cdb := cdb6{_SCSI_INQUIRY}
	binary.BigEndian.PutUint16(cdb[3:5], uint16(len(respBuf)))

	if err := scsiSendCdb(fd, cdb[:], respBuf); err != nil {
		return &resp, err
	}

	if err := binary.Read(bytes.NewBuffer(respBuf), binary.BigEndian, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func scsiInquiryVpd(fd int, page uint8, respBuf []byte) error {
	cdb := cdb6{_SCSI_INQUIRY, 1 /*enable VPD*/, page}
	binary.BigEndian.PutUint16(cdb[3:5], uint16(len(respBuf)))

	return scsiSendCdb(fd, cdb[:], respBuf)
}

func (d *ScsiDevice) SerialNumber() (string, error) {
	buf := make([]byte, 256)
	if err := scsiInquiryVpd(d.fd, 0x80, buf); err != nil {
		return "", nil
	}

	if buf[1] != 0x80 {
		return "", fmt.Errorf("invalid INQUIRY return page: %0x", buf[1])
	}
	length := buf[3]

	return string(buf[4 : 4+length]), nil
}

// SCSI ioctl v3 header
type sgIoHdr struct {
	interfaceId    int32   // 'S' for SCSI generic (required)
	dxferDirection int32   // data transfer direction
	cmdLen         uint8   // SCSI command length (<= 16 bytes)
	mxSbLen        uint8   // max length to write to sbp
	iovecCount     uint16  // 0 implies no scatter gather
	dxferLen       uint32  // byte count of data transfer
	dxferp         uintptr // points to data transfer memory or scatter gather list
	cmdp           uintptr // points to command to perform
	sbp            uintptr // points to sense_buffer memory
	timeout        uint32  // MAX_UINT -> no timeout (unit: millisec)
	flags          uint32  // 0 -> default, see SG_FLAG...
	packId         int32   // unused internally (normally)
	usrPtr         uintptr // unused internally
	status         uint8   // SCSI status
	maskedStatus   uint8   // shifted, masked scsi status
	msgStatus      uint8   // messaging level data (optional)
	sbLenWr        uint8   // byte count actually written to sbp
	hostStatus     uint16  // errors from host adapter
	driverStatus   uint16  // errors from software driver
	resid          int32   // dxfer_len - actual_transferred
	duration       uint32  // time taken by cmd (unit: millisec)
	info           uint32  // auxiliary information
}

// SCSI ioctl v4 header
type sgIoV4 struct {
	guard       int32  /* [i] 'Q' to differentiate from v3 */
	protocol    uint32 /* [i] 0 -> SCSI , .... */
	subprotocol uint32 /* [i] 0 -> SCSI command, 1 -> SCSI task management function, .... */

	requestLen      uint32 /* [i] in bytes */
	request         uint64 /* [i], [*i] {SCSI: cdb} */
	requestTag      uint64 /* [i] {SCSI: task tag (only if flagged)} */
	requestAttr     uint32 /* [i] {SCSI: task attribute} */
	requestPriority uint32 /* [i] {SCSI: task priority} */
	requestExtra    uint32 /* [i] {spare, for padding} */
	maxResponseLen  uint32 /* [i] in bytes */
	response        uint64 /* [i], [*o] {SCSI: (auto)sense data} */

	/* "dout_": data out (to device) "din_": data in (from device) */
	doutIovecCount uint32 /* [i] 0 -> "flat" dout transfer else dout_xfer points to array of iovec */
	doutXferLen    uint32 /* [i] bytes to be transferred to device */
	dinIovecCount  uint32 /* [i] 0 -> "flat" din transfer */
	dinXferLen     uint32 /* [i] bytes to be transferred from device */
	doutXferp      uint64 /* [i], [*i] */
	dinXferp       uint64 /* [i], [*o] */

	timeout uint32 /* [i] units: millisecond */
	flags   uint32 /* [i] bit mask */
	usrPtr  uint64 /* [i->o] unused internally */
	spareIn uint32 /* [i] */

	driverStatus    uint32 /* [o] 0 -> ok */
	transportStatus uint32 /* [o] 0 -> ok */
	deviceStatus    uint32 /* [o] {SCSI: command completion status} */
	retryDelay      uint32 /* [o] {SCSI: status auxiliary information} */
	info            uint32 /* [o] additional information */
	duration        uint32 /* [o] time to complete, in milliseconds */
	responseLen     uint32 /* [o] bytes of response actually written */
	dinResid        int32  /* [o] dinXferLen - actual_din_xfer_len */
	doutResid       int32  /* [o] doutXferLen - actual_dout_xfer_len */
	generatedTag    uint64 /* [o] {SCSI: transport generated task tag} */
	spareOut        uint32 /* [o] */

	_ uint32 // padding
}

func scsiSendCdb(fd int, cdb []byte, respBuf []byte) error {
	senseBuf := make([]byte, 32)

	/*
		// TODO: make it work with sg_io_v4 data structure
		hdr := sgIoV4{
			guard:          'Q',
			timeout:        _DEFAULT_TIMEOUT,
			requestLen:     uint32(len(cdb)),
			request:        uint64(uintptr(unsafe.Pointer(&cdb[0]))),
			maxResponseLen: uint32(len(senseBuf)),
			response:       uint64(uintptr(unsafe.Pointer(&senseBuf[0]))),
			dinXferLen:     uint32(len(respBuf)),
			dinXferp:       uint64(uintptr(unsafe.Pointer(&respBuf[0]))),
		}
	*/

	hdr := sgIoHdr{
		interfaceId:    'S',
		dxferDirection: _SG_DXFER_FROM_DEV,
		timeout:        _DEFAULT_TIMEOUT,
		cmdLen:         uint8(len(cdb)),
		mxSbLen:        uint8(len(senseBuf)),
		dxferLen:       uint32(len(respBuf)),
		dxferp:         uintptr(unsafe.Pointer(&respBuf[0])),
		cmdp:           uintptr(unsafe.Pointer(&cdb[0])),
		sbp:            uintptr(unsafe.Pointer(&senseBuf[0])),
	}

	if err := ioctl(uintptr(fd), _SG_IO, uintptr(unsafe.Pointer(&hdr))); err != nil {
		return err
	}

	if hdr.info&_SG_INFO_OK_MASK != _SG_INFO_OK {
		return sgioError{
			deviceStatus: uint32(hdr.status),
			hostStatus:   uint32(hdr.hostStatus),
			driverStatus: uint32(hdr.driverStatus),
		}
	}
	return nil
}
