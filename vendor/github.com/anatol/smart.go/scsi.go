package smart

// https://www.seagate.com/files/staticfiles/support/docs/manual/Interface%20manuals/100293068j.pdf

type ScsiDevice struct {
	fd int
}

func (d *ScsiDevice) Type() string {
	return "scsi"
}

const (
	_SG_IO = 0x2285

	_SG_INFO_OK_MASK = 0x1
	_SG_INFO_OK      = 0x0 /* no sense, host nor driver "noise" */
	_SG_INFO_CHECK   = 0x1 /* something abnormal happened */

	// Timeout in milliseconds
	_DEFAULT_TIMEOUT = 20000

	// SCSI commands used by this package
	_SCSI_INQUIRY          = 0x12
	_SCSI_MODE_SENSE_6     = 0x1a
	_SCSI_READ_CAPACITY_10 = 0x25

	_SG_DXFER_NONE        = -1
	_SG_DXFER_TO_DEV      = -2
	_SG_DXFER_FROM_DEV    = -3
	_SG_DXFER_TO_FROM_DEV = -4
)

type ScsiInquiry struct {
	Peripheral   uint8 // peripheral qualifier + device type
	Rmb          uint8
	Version      uint8
	Flags        uint8
	RespLength   uint8
	Flags2       [3]uint8
	VendorIdent  [8]byte // if "ATA     " then it is ATA device
	ProductIdent [16]byte
	ProductRev   [4]byte
}

func (d *ScsiDevice) ReadGenericAttributes() (*GenericAttributes, error) {
	return nil, ErrOSUnsupported
}
