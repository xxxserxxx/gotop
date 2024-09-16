package smart

import (
	"bytes"
	"fmt"
	"regexp"
	"time"
)

// https://people.freebsd.org/~imp/asiabsdcon2015/works/d2161r5-ATAATAPI_Command_Set_-_3.pdf
// https://www.t10.org/ftp/t10/document.04/04-262r8.pdf

const (
	_SATA_IDENT           = "ATA     "
	_SCSI_ATA_PASSTHRU_16 = 0x85

	// ATA commands
	_ATA_SMART           = 0xb0
	_ATA_IDENTIFY_DEVICE = 0xec

	// ATA feature register values for SMART
	_SMART_READ_DATA       = 0xd0
	_SMART_READ_THRESHOLDS = 0xd1
	_SMART_READ_LOG        = 0xd5
	_SMART_RETURN_STATUS   = 0xda
)

// AtaIdentifyDevice ATA IDENTIFY DEVICE struct. ATA8-ACS defines this as a page of 16-bit words.
// Some fields span multiple words (e.g., model number). Some fields use less than a
// single word, and are bitmasked together with other fields. Since many of the fields are now
// retired / obsolete, we only define the fields that are currently used by this package.
type AtaIdentifyDevice struct {
	GeneralConfig       uint16    // Word 0, general configuration. If bit 15 is zero, device is ATA.
	_                   [9]uint16 // ...
	SerialNumberRaw     [20]byte  // Word 10..19, device serial number, padded with spaces (20h).
	_                   [3]uint16 // ...
	FirmwareRevisionRaw [8]byte   // Word 23..26, device firmware revision, padded with spaces (20h).
	ModelNumberRaw      [40]byte  // Word 27..46, device model number, padded with spaces (20h).
	_                   [2]uint16 // ...
	// Capabilities (see 7.12.7.17)
	// Word 49
	// 15:14 Reserved for the IDENTIFY PACKET DEVICE command.
	// 13: 1 = Standby timer values as specified in this standard are supported.
	//     0 = Standby timer values shall be vendor specific.
	// 12 Reserved for the IDENTIFY PACKET DEVICE command.
	// 11: 1 = IORDY (see ATA8-APT) supported
	//     0 = IORDY (see ATA8-APT) may be supported
	// 10 IORDY (see ATA8-APT) may be disabled
	// 9 Shall be set to one (i.e., LBA is supported).
	// 8 DMA supported
	// 7:2 Reserved
	// 1:0 Long Physical Sector Alignment Error reporting
	// Word 50 Capabilities (see 7.12.7.17)
	// 15 Shall be cleared to zero
	// 14 Shall be set to one
	// 13:2 Reserved
	// 1 Obsolete
	// 0: 1 = There is a minimum Standby time value and it is vendor specific.
	//    0 = There is no minimum Standby timer value.
	Capabilities  [2]uint16 // Word 49..50
	_             [9]uint16
	CapacityLba28 uint32 // Word 60..61
	_             [13]uint16
	QueueDepth    uint16 // Word 75, Maximum queue depth – 1
	// Serial ATA Capabilities (see 7.12.6.34)
	// bit 15 Supports READ LOG DMA EXT as equivalent to READ LOG EXT
	// bit 14 Supports Device Automatic Partial to Slumber transitions
	// bit 13 Supports Host Automatic Partial to Slumber transitions
	// bit 12 Supports NCQ priority information
	// bit 11 Supports Unload while NCQ commands are outstanding
	// bit 10 Supports the SATA Phy Event Counters log
	// bit 9 Supports receipt of host initiated power management requests
	// bit 8 Supports the NCQ feature set
	// bit 7:4 Reserved for Serial ATA
	// bit 3 Supports SATA Gen3 Signaling Speed (6.0Gb/s)
	// bit 2 Supports SATA Gen2 Signaling Speed (3.0Gb/s)
	// bit 1 Supports SATA Gen1 Signaling Speed (1.5Gb/s)
	// bit 0 Shall be cleared to zero
	SATACap uint16 // Word 76, SATA capabilities.
	// Serial ATA Additional Capabilities
	// bit 15:9 Reserved for Serial ATA
	// bit 8 Power Disable feature always enabled
	// bit 7 Supports DevSleep to ReducedPwrState
	// bit 6 Supports RECEIVE FPDMA QUEUED and SEND FPDMA QUEUED commands
	// bit 5 Supports NCQ NON-DATA Command
	// bit 4 Supports NCQ Streaming
	// bit 3:1 Coded value indicating current negotiated Serial ATA signal speed
	// bit 0 Shall be cleared to zero
	SATACapAddl uint16 // Word 77, SATA additional capabilities.
	// Serial ATA features supported (see 7.12.6.36)
	// bit 15:13 Reserved for Serial ATA
	// bit 12 Power Disable feature supported
	// bit 11 Device supports Rebuild Assist feature set
	// bit 10 Reserved for Serial ATA
	// bit 9 Device supports Hybrid Information
	// bit 8 Device Sleep feature supported
	// bit 7 Device supports NCQ Autosense
	// bit 6 Device supports Software Settings Preservation
	// bit 5 Device supports Hardware Feature Control
	// bit 4 Device supports in-order data delivery
	// bit 3 Device supports initiating power management
	// bit 2 Device supports DMA Setup auto-activation
	// bit 1 Device supports nonzero buffer offsets
	// bit 0 Shall be cleared to zero
	FeaturesSupported uint16 // Word 78, Serial ATA features supported.
	// Serial ATA features enabled (see 7.12.6.37)
	// bit 15:12 Reserved for Serial ATA
	// bit 11 Rebuild Assist feature set enabled
	// bit 10 Power Disable feature enabled
	// bit 9 Hybrid Information enabled
	// bit 8 Device Sleep feature enabled
	// bit 7 Automatic Partial to Slumber transitions enabled
	// bit 6 Software Settings Preservation enabled
	// bit 5 Hardware Feature Control is enabled
	// bit 4 In-order data delivery enabled
	// bit 3 Device initiated power management enabled
	// bit 2 DMA Setup auto-activation enabled
	// bit 1 Nonzero buffer offsets enabled
	// bit 0 Shall be cleared to zero
	FeaturesEnabled uint16 // Word 79, Serial ATA features enabled.
	MajorVersion    uint16 // Word 80, major version number.
	MinorVersion    uint16 // Word 81, minor version number.
	// Commands and feature sets supported (see 7.12.6.40)
	// bit 15 Obsolete
	// bit 14 The NOP command is supported.
	// bit 13 The READ BUFFER command is supported.
	// bit 12 The WRITE BUFFER command is supported.
	// bit 11:10 Obsolete
	// bit 9 Shall be cleared to zero (i.e., the DEVICE RESET command (see ACS-3) is not supported)
	// bit 8:7 Obsolete
	// bit 6 Read look-ahead is supported.
	// bit 5 The volatile write cache is supported.
	// bit 4 Shall be cleared to zero (i.e., the PACKET feature set (see ACS-3) is not supported).
	// bit 3 Shall be set to one (i.e., the Power Management feature set is supported)
	// bit 2 Obsolete
	// bit 1 The Security feature set is supported.
	// bit 0 The SMART feature set is supported.
	CommandsSupported1 uint16 // Word 82, commands and feature sets supported (see 7.12.6.40)
	// Commands and feature sets supported (see 7.12.6.40)
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one
	// bit 13 The FLUSH CACHE EXT command is supported.
	// bit 12 Shall be set to one (i.e., the FLUSH CACHE command is supported).
	// bit 11 Obsolete
	// bit 10 The 48-bit Address feature set is supported.
	// bit 9:7 Obsolete
	// bit 6 SET FEATURES subcommand is required to spin-up after power-up.
	// bit 5 The PUIS feature set is supported.
	// bit 4 Obsolete
	// bit 3 The APM feature set is supported.
	// bit 2 Reserved for CFA
	// bit 1 Obsolete
	// bit 0 The DOWNLOAD MICROCODE command is supported.
	CommandsSupported2 uint16 // Word 83, commands and feature sets supported (see 7.12.6.40)
	// Commands and feature sets supported (see 7.12.6.40)
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one
	// bit 13 The IDLE IMMEDIATE command with UNLOAD feature is supported.
	// bit 12:9 Obsolete
	// bit 8 Shall be set to one (i.e., the World Wide Name is supported)
	// bit 7 Obsolete
	// bit 6 The WRITE DMA FUA EXT command is supported.
	// bit 5 The GPL feature set is supported
	// bit 4 The Streaming feature set is supported
	// bit 3 Obsolete
	// bit 2 Reserved
	// bit 1 The SMART self-test is supported.
	// bit 0 SMART error logging is supported.
	CommandsSupported3 uint16 // Word 84, commands and feature sets supported (see 7.12.6.40)
	// Commands and feature sets supported or enabled (see 7.12.6.41)
	// bit 15 Obsolete
	// bit 14 The NOP command is supported.
	// bit 13 The READ BUFFER command is supported.
	// bit 12 The WRITE BUFFER command is supported.
	// bit 11:10 Obsolete
	// bit 9 Shall be cleared to zero (i.e., the DEVICE RESET command (see ACS-3) is not supported)
	// bit 8:7 Obsolete
	// bit 6 Read look-ahead is enabled.
	// bit 5 The volatile write cache is enabled.
	// bit 4 Shall be cleared to zero (i.e., the PACKET feature set (see ACS-3) is not supported)
	// bit 3 Shall be set to one (i.e., the Power Management feature set is supported)
	// bit 2 Obsolete
	// bit 1 The Security feature set is enabled.
	// bit 0 The SMART feature set is enabled.
	CommandsEnabled1 uint16 // Word 85, supported commands and feature sets.
	// Commands and feature sets supported or enabled (see 7.12.6.41)
	// bit 15 Words 119..120 are valid.
	// bit 14 Reserved
	// bit 13 FLUSH CACHE EXT command supported.
	// bit 12 FLUSH CACHE command supported.
	// bit 11 Obsolete
	// bit 10 The 48-bit Address features set is supported.
	// bit 9:7 Obsolete
	// bit 6 SET FEATURES subcommand is required to spin-up after power-up.
	// bit 5 The PUIS feature set is enabled.
	// bit 4 Obsolete
	// bit 3 The APM feature set is enabled.
	// bit 2 Reserved for CFA
	// bit 1 Obsolete
	// bit 0 The DOWNLOAD MICROCODE command is supported
	CommandsEnabled2 uint16 // Word 86, Commands and feature sets supported or enabled
	// Commands and feature sets supported or enabled (see 7.12.6.41)
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one
	// bit 13 The IDLE IMMEDIATE command with UNLOAD FEATURE is supported.
	// bit 12:9 Obsolete
	// bit 8 Shall be set to one (i.e., the World Wide Name is supported)
	// bit 7 Obsolete
	// bit 6 The WRITE DMA FUA EXT command is supported.
	// bit 5 The GPL feature set is supported.
	// bit 4:3 Obsolete
	// bit 2 Media serial number is valid.
	// bit 1 SMART self-test is supported.
	// bit 0 SMART error logging is supported.
	CommandsEnabled3 uint16 // Word 87
	// Ultra DMA modes (see 7.12.6.42)
	// bit 15 Reserved
	// bit 14 Ultra DMA mode 6 is selected.
	// bit 13 Ultra DMA mode 5 is selected.
	// bit 12 Ultra DMA mode 4 is selected.
	// bit 11 Ultra DMA mode 3 is selected.
	// bit 10 Ultra DMA mode 2 is selected.
	// bit 9 Ultra DMA mode 1 is selected.
	// bit 8 Ultra DMA mode 0 is selected.
	// bit 7 Reserved
	// bit 6 Ultra DMA mode 6 and below are supported.
	// bit 5 Ultra DMA mode 5 and below are supported.
	// bit 4 Ultra DMA mode 4 and below are supported.
	// bit 3 Ultra DMA mode 3 and below are supported.
	// bit 2 Ultra DMA mode 2 and below are supported.
	// bit 1 Ultra DMA mode 1 and below are supported.
	// bit 0 Ultra DMA mode 0 is supported.
	DMAModes uint16    // Word 88, Ultra DMA modes (see 7.12.6.42)
	_        [4]uint16 // ...
	// Hardware reset results (see 7.12.6.47)
	// For SATA devices, word 93 shall be set to the value 0000h.
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one for PATA devices
	// bit 13
	//    1 = device detected the CBLID- above ViHB (see ATA8-APT).
	//    0 = device detected the CBLID- below ViL (see ATA8-APT).
	// bit 12:8 Device 1 hardware reset result.
	//    Device 0 shall clear these bits to zero.
	//    Device 1 shall set these bits as follows:
	//      bit 12 Reserved
	//      bit 11 Device 1 asserted PDIAG-.
	//      bit 10:9 These bits indicate how Device 1 determined the device number:
	//        00 = Reserved
	//        01 = a jumper was used.
	//        10 = the CSEL signal was used.
	//        11 = some other method was used or the method is unknown.
	// bit 8 Shall be set to one
	// bit 7:0 Device 0 hardware reset result.
	//   Device 1 shall clear these bits to zero.
	//   Device 0 shall set these bits as follows:
	// bit 7 Reserved
	// bit 6 Device 0 responds when Device 1 is selected.
	// bit 5 Device 0 detected the assertion of DASP-.
	// bit 4 Device 0 detected the assertion of PDIAG-.
	// bit 3 Device 0 passed diagnostics.
	// bit 2:1 These bits indicate how Device 0 determined the device number:
	//    00 = Reserved
	//    01 = a jumper was used.
	//    10 = the CSEL signal was used.
	//    11 = some other method was used or the method is unknown.
	// bit 0 Shall be set to one for PATA devices
	ResetResults  uint16    // Word 93, Hardware reset results (see 7.12.6.47)
	_             [6]uint16 // ...
	CapacityLba48 uint64    // Word 100..103
	_             [2]uint16 // ...
	// Physical sector size / logical sector size (see 7.12.6.56)
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one
	// bit 13 Device has multiple logical sectors per physical sector.
	// bit 12 Device Logical Sector longer than 256 words
	// bit 11:4 Reserved
	// bit 3:0 2^X logical sectors per physical sector
	LogicalPerPhisicalSectors uint16 // Word 106, Physical sector size / logical sector size (see 7.12.6.56)
	InterSeekDelay            uint16 // Word 107, Inter-seek delay for ISO/IEC 7779 standard acoustic testing (see 7.12.6.57)
	// In the IDENTIFY DEVICE data (see 7.12.7) and the IDENTIFY PACKET DEVICE data (see 7.13.6):
	// bits 15:12 shall contain the NAA field (see A.11.5.8.2);
	// bits 11:0 and word 109 bits 15:4 shall contain the IEEE OUI field (see A.11.5.8.2); and
	// bits 3:0, word 110, and word 111 shall contain the UNIQUE ID field (see A.11.5.8.2).
	WWNRaw [4]uint16 // Word 108..111, WWN (World Wide Name).
	_      [5]uint16
	// Logical sector size
	LogicalSectorSize [2]uint16 // Word 117..118
	// Commands and feature sets supported (Continued from words 82..84) (see 7.12.6.40)
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one
	// bit 13:10 Reserved
	// bit 9 DSN feature set is supported.
	// bit 8 Accessible Max Address Configuration feature set is supported.
	// bit 7 EPC feature set is supported.
	// bit 6 Sense Data Reporting feature set is supported.
	// bit 5 The Free-fall Control feature set is supported.
	// bit 4 Download Microcode mode 3 is supported.
	// bit 3 The READ LOG DMA EXT command and WRITE LOG DMA EXT command are supported.
	// bit 2 The WRITE UNCORRECTABLE EXT command is supported.
	// bit 1 The Write-Read-Verify feature set is supported.
	// bit 0 Obsolete
	CommandsSupported4 uint16 // Word 119, Commands and feature sets supported (Continued from words 82..84) (see 7.12.6.40)
	// Commands and feature sets supported or enabled (Continued from words 85..87) (see 7.12.6.41)
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one
	// bit 13:10 Reserved
	// bit 9 DSN feature set is enabled.
	// bit 8 Reserved
	// bit 7 EPC feature set is enabled.
	// bit 6 Sense Data Reporting feature set is enabled.
	// bit 5 The Free-fall Control feature set is enabled.
	// bit 4 Download Microcode mode 3 is supported.
	// bit 3 The READ LOG DMA EXT command and WRITE LOG DMA EXT command are supported.
	// bit 2 The WRITE UNCORRECTABLE EXT command is supported.
	// bit 1 The Write-Read-Verify feature set is enabled.
	// bit 0 Obsolete
	CommandsEnabled4 uint16     //  Commands and feature sets supported or enabled (Continued from words 85..87) (see 7.12.6.41)
	_                [88]uint16 // ...
	// Alignment of logical sectors within a physical sector (see 7.12.7.75)
	// bit 15 Shall be cleared to zero
	// bit 14 Shall be set to one
	// bits 13:0 Logical sector offset within the first physical sector where the first logical sector is placed
	LogicalSectorOffset uint16 // Word 209
	_                   [7]uint16
	RotationRate        uint16     // Word 217, nominal media rotation rate.
	_                   [4]uint16  // ...
	TransportMajor      uint16     // Word 222, transport major version number.
	_                   [33]uint16 // ...
} // 512 bytes

func (a *AtaIdentifyDevice) IsGeneralPurposeLoggingCapable() bool {
	enabled := uint16(1) << 14
	enabledMask := uint16(0b11) << 14

	glLoggigAttr := uint16(1) << 5

	if a.CommandsSupported3&enabledMask == enabled {
		return a.CommandsSupported3&enabledMask&glLoggigAttr == glLoggigAttr
	}
	if a.CommandsEnabled3&enabledMask == enabled {
		return a.CommandsEnabled3&enabledMask&glLoggigAttr == glLoggigAttr
	}

	return false
}

func (i *AtaIdentifyDevice) ModelNumber() string {
	return fromAtaString(i.ModelNumberRaw[:])
}

func (i *AtaIdentifyDevice) SerialNumber() string {
	return fromAtaString(i.SerialNumberRaw[:])
}

func (i *AtaIdentifyDevice) FirmwareRevision() string {
	return fromAtaString(i.FirmwareRevisionRaw[:])
}

// WWN converts raw wwn format to uint64 number
func (i *AtaIdentifyDevice) WWN() uint64 {
	raw := i.WWNRaw

	return uint64(raw[0])<<48 + uint64(raw[1])<<32 + uint64(raw[2])<<16 + uint64(raw[3])
}

// ATA strings have each pair of bytes swapped. See 3.4.9 paragraph of the spec.
// convert ATA strings to regularly ordered string
func fromAtaString(in []byte) string {
	swapped := make([]byte, len(in))
	for i := 0; i < len(in); i += 2 {
		// swap paired bytes
		swapped[i], swapped[i+1] = in[i+1], in[i]
	}

	swapped = bytes.TrimSpace(swapped)
	return string(swapped)
}

// Capacity returns size of the device sectors and capacity of the device itself
func (i *AtaIdentifyDevice) Capacity() (sectors, capacity, logicalSectorSize, physicalSectorSize, logicalSectorOffset uint64) {
	var lba28, lba48 uint64

	if i.Capabilities[0]&0x0200 == 0 {
		// LBA is not supported
		return
	}

	lba28 = uint64(i.CapacityLba28)

	if i.CommandsSupported2&0x400 != 0 {
		// The 48-bit Address feature set is supported
		lba48 = i.CapacityLba48
	}

	if lba28 == 0 && lba48 == 0 {
		// capacity unknown (ATAPI CD/DVD)
		return
	}

	logicalSectorSize = 512
	physicalSectorSize = 512

	// Long Logical/Physical Sectors (LLS/LPS) ?
	if i.LogicalPerPhisicalSectors&0x1000 != 0 {
		// Logical sector size is specified in 16-bit words
		logicalSectorSize = (uint64(i.LogicalSectorSize[1])<<16 | uint64(i.LogicalSectorSize[0])) << 1
		physicalSectorSize = logicalSectorSize
	}
	if i.LogicalPerPhisicalSectors&0x2000 != 0 {
		// Physical sector size is multiple of logical sector size
		physicalSectorSize <<= i.LogicalPerPhisicalSectors & 0x0f
	}
	logicalSectorOffset = uint64(i.LogicalSectorOffset&0x3fff) * logicalSectorSize

	// Some early 4KiB LLS disks (Samsung N3U-3) return bogus lba28 value
	if lba48 >= lba28 || (lba48 != 0 && logicalSectorSize > 512) {
		sectors = lba48
	} else {
		sectors = lba28
	}

	capacity = sectors * logicalSectorSize

	return
}

//go:generate go run generator/ata_attributes.go

const (
	ataDeviceAttributeRestrictionNone = iota
	ataDeviceAttributeRestrictionHDDOnly
	ataDeviceAttributeRestrictionSSDOnly
)

const (
	AtaDeviceAttributeTypeRaw8          = iota + 1 // formatting is raw[40:48], raw[32:40], raw[24:32], raw[16:24], raw[8:16], raw[0:8]
	AtaDeviceAttributeTypeRaw16                    // formatting is raw[32:48], raw[16:32], raw[0:16]
	AtaDeviceAttributeTypeRaw48                    // formatting is raw[0:48]
	AtaDeviceAttributeTypeHex48                    // formatting is raw[0:48] in hex
	AtaDeviceAttributeTypeRaw56                    // formatting is raw[0:56]
	AtaDeviceAttributeTypeHex56                    // formatting is raw[0:56] in hex
	AtaDeviceAttributeTypeRaw64                    // formatting is raw[0:64]
	AtaDeviceAttributeTypeHex64                    // formatting is raw[0:64] in hex
	AtaDeviceAttributeTypeRaw16OptRaw16            // formatting is raw[0:16] and optional raw[32:48], raw[16:32]
	AtaDeviceAttributeTypeRaw16OptAvg16            // formatting is raw[0:6] and optional raw[16:32] as average
	AtaDeviceAttributeTypeRaw24OptRaw8             // formatting is raw[0:24] and optional raw[40:48], raw[32:40], raw[24:32]
	AtaDeviceAttributeTypeRaw24DivRaw24            // formatting is raw[24:48]/raw[0:24]
	AtaDeviceAttributeTypeRaw24DivRaw32            // formatting is raw[32:56]/raw[0:32]
	AtaDeviceAttributeTypeSec2Hour                 // formatting is hours = raw/3600; minutes = (raw-3600*hours)/60; seconds = raw%60
	AtaDeviceAttributeTypeMin2Hour                 // formatting is hours = raw[0:32]/60; minutes = raw[0:32]%60; optional raw[32:48] as ???
	AtaDeviceAttributeTypeHalfMin2Hour             // formatting is hours = raw/120; minutes = (raw-120*hours)/2
	AtaDeviceAttributeTypeMsec24Hour32             // formatting is hours = raw[32:64]; milliseconds = raw[0:32]
	AtaDeviceAttributeTypeTempMinMax               // formatting is too complicated
	AtaDeviceAttributeTypeTemp10X                  // formatting is temp = raw[0:32]/10 in Celsius
)

type ataDeviceAttr struct {
	name        string
	typ         int
	byteOrder   string
	restriction int
}

const (
	ataFirmwareBugNoLogDir = 1 << iota
	ataFirmwareBugSamsung
	ataFirmwareBugSamsung2
	ataFirmwareBugSamsung3
	ataFirmwareBugXErrorLBA
)

type ataDeviceInfo struct {
	model      string
	modelRe    string
	firmwareRe string
	presets    map[int]ataDeviceAttr
	bug        int
}

// AtaSmartAttrRaw individual SMART attribute (12 bytes)
type AtaSmartAttrRaw struct {
	Id          uint8
	Flags       uint16
	Current     uint8   // current value from the device
	Worst       uint8   // worst value
	VendorBytes [6]byte // vendor-specific (and sometimes device-specific) data
	Reserved    uint8
}

// AtaSmartPageRaw is page of 30 SMART attributes as per ATA spec
type AtaSmartPageRaw struct {
	Version uint16
	Attrs   [30]AtaSmartAttrRaw
}

// SMART log address 00h
type AtaSmartLogDirectory struct {
	Version uint16
	Address [255]struct {
		NumPages byte
		_        byte // Reserved
	}
}

// SMART log address 01h
type AtaSmartErrorLogSummary struct {
	Version    byte
	LogIndex   byte
	LogData    [5][90]byte // TODO: Expand out to error log structure
	ErrorCount uint16      // Device error count
	_          [57]byte    // Reserved
	Checksum   byte        // Two's complement checksum of first 511 bytes
}

// SMART log address 06h
type AtaSmartSelfTestLog struct {
	Version uint16
	Entry   [21]struct {
		LBA_7          byte   // Content of the LBA field (7:0) when subcommand was issued
		Status         byte   // Self-test execution status
		LifeTimestamp  uint16 // Power-on lifetime of the device in hours when subcommand was completed
		Checkpoint     byte
		LBA            uint32 // LBA of first error (28-bit addressing)
		VendorSpecific [15]byte
	}
	VendorSpecific uint16
	Index          byte
	_              uint16 // Reserved
	Checksum       byte   // Two's complement checksum of first 511 bytes
}

type SataDevice struct {
	fd               int
	attributeMapping map[int]ataDeviceAttr
	firmwareBug      int
}

// AtaSmartThresholdEntry individual SMART attribute threshold (12 bytes)
type AtaSmartThresholdEntry struct {
	Id        uint8    // SMART attribute ID
	Threshold uint8    // Threshold value
	_         [10]byte // Reserved
}

// AtaSmartThresholdsPageRaw is page of 30 SMART attributes thresholds as per ATA spec
type AtaSmartThresholdsPageRaw struct {
	Revnumber        uint16
	ThresholdEntries [30]AtaSmartThresholdEntry
	_                [149]byte // Reserved
	Chksum           byte      // Two's complement checksum of first 511 bytes
}

func (d *SataDevice) Type() string {
	return "sata"
}

func (d *SataDevice) ReadGenericAttributes() (*GenericAttributes, error) {
	page, err := d.ReadSMARTData()
	if err != nil {
		return nil, err
	}

	a := GenericAttributes{}
	for _, attr := range page.Attrs {
		switch attr.Name {
		case "Airflow_Temperature_Cel":
			fallthrough
		case "Temperature_Celsius":
			fallthrough
		case "Temperature_Celsius_X10":
			current, _, _, _, err := attr.ParseAsTemperature()
			if err != nil {
				return nil, err
			}
			a.Temperature = uint64(current)
		case "Total_LBAs_Read":
			a.Read = attr.ValueRaw
		case "Total_LBAs_Written":
			a.Written = attr.ValueRaw
		case "Power_On_Hours":
			a.PowerOnHours = attr.ValueRaw
		case "Power_Cycle_Count":
			a.PowerCycles = attr.ValueRaw
		}
	}

	return &a, nil
}

func findAttributesMapping(model, firmware string) (map[int]ataDeviceAttr, int, error) {
	db, err := findMatchingDbRecord(model, firmware)
	if err != nil {
		return nil, 0, err
	}

	if db == nil {
		return ataDefaultAttributes, 0, nil
	}

	if db.presets == nil {
		return ataDefaultAttributes, db.bug, nil
	}

	// combine default and db attributes mapping
	attributeMapping := make(map[int]ataDeviceAttr)
	for k, v := range ataDefaultAttributes {
		attributeMapping[k] = v
	}
	for k, v := range db.presets {
		attributeMapping[k] = v
	}
	return attributeMapping, db.bug, nil
}

func findMatchingDbRecord(model, firmware string) (*ataDeviceInfo, error) {
	for _, db := range ataDevicesDatabase {
		matched, err := regexp.MatchString(db.modelRe, model)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}

		if db.firmwareRe != "" {
			matched, err := regexp.MatchString(db.firmwareRe, firmware)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}

		return &db, nil
	}

	return nil, nil
}

const (
	// Prefailure bit. If the flag is 0 it corresponds to "Old_age" value, "Pre-fail" otherwise.
	AtaAttributeFlagPrefailure = 1 << 0
	// Online bit. If the flag is 0 it corresponds to "Offline" value, "Always" otherwise.
	AtaAttributeFlagOnline = 1 << 1
)

type AtaSmartAttr struct {
	Id          uint8
	Flags       uint16
	Current     uint8   // raw value from the device
	Worst       uint8   // worst value
	VendorBytes [6]byte // vendor-specific (and sometimes device-specific) data

	// computed values
	Name     string
	Type     int
	ValueRaw uint64
}

func (a AtaSmartAttr) ParseAsDuration() (time.Duration, error) {
	switch a.Type {
	case AtaDeviceAttributeTypeMin2Hour:
		return time.Duration(a.ValueRaw&0xffffffff) * time.Minute, nil
	case AtaDeviceAttributeTypeSec2Hour:
		return time.Duration(a.ValueRaw) * time.Second, nil
	case AtaDeviceAttributeTypeHalfMin2Hour:
		// 30-second counter
		return time.Duration(a.ValueRaw) * 30 * time.Second, nil
	case AtaDeviceAttributeTypeMsec24Hour32:
		// hours + milliseconds
		hours := a.ValueRaw & 0xffffffff
		milliseconds := a.ValueRaw >> 32
		return time.Duration(hours)*time.Hour + time.Duration(milliseconds)*time.Millisecond, nil
	default:
		return 0, fmt.Errorf("unknown time type: %d", a.Type)
	}
}

// ParseAsTemperature returns temperature, range and optional over temperature counter
// if a.Type is equal AtaDeviceAttributeTypeTemp10X then the returned temperature is round down to the closest integer value;
// and the rest of the temperature fields are zeros.
//
// if a.Type is equal AtaDeviceAttributeTypeTempMinMax then the first returned value is temperature and then 3 optional fields follow -
// Min, Max and OverTemperature counter. The optional values are not supported by all harddrives.
func (a AtaSmartAttr) ParseAsTemperature() (int /* val */, int /* low */, int /* hi */, int /* ?? */, error) {
	switch a.Type {
	case AtaDeviceAttributeTypeTemp10X:
		return int(a.ValueRaw&0xffff) / 10, 0, 0, 0, nil
	case AtaDeviceAttributeTypeTempMinMax:
		// consult smartmontools ata_format_attr_raw_value() function that has more information about this algorithm
		word0 := uint16(a.ValueRaw)
		word1 := uint16(a.ValueRaw >> 16)
		word2 := uint16(a.ValueRaw >> 32)

		raw0 := int8(a.ValueRaw)
		raw1 := int8(a.ValueRaw >> 8)
		raw2 := int8(a.ValueRaw >> 16)
		raw3 := int8(a.ValueRaw >> 24)
		raw4 := int8(a.ValueRaw >> 32)
		// raw5 := int8(a.ValueRaw >> 40)

		var hi, lo int8

		ctw0 := checkTempWord(word0)
		if word2 == 0 {
			if word1 == 0 && ctw0 != 0 {
				// 00 00 00 00 xx TT
				return int(raw0), 0, 0, 0, nil
			} else if ctw0 != 0 && checkTempRange(raw0, raw2, raw3, &lo, &hi) {
				// 00 00 HL LH xx TT
				return int(raw0), int(lo), int(hi), 0, nil
			} else if raw3 == 0 && checkTempRange(raw0, raw1, raw2, &lo, &hi) {
				// 00 00 00 HL LH TT
				return int(raw0), int(lo), int(hi), 0, nil
			}
		} else if ctw0 != 0 {
			if (ctw0&checkTempWord(word1)&checkTempWord(word2)) != 0x00 && checkTempRange(raw0, raw2, raw4, &lo, &hi) {
				// xx HL xx LH xx TT
				return int(raw0), int(lo), int(hi), 0, nil
			} else if word2 < 0x7fff && checkTempRange(raw0, raw2, raw3, &lo, &hi) && hi >= 40 {
				// CC CC HL LH xx TT
				return int(raw0), int(lo), int(hi), int(word2), nil
			}
		}
		return int(raw0), 0, 0, 0, nil
	default:
		return 0, 0, 0, 0, fmt.Errorf("unknown temperature min/max type: %d", a.Type)
	}
}

func checkTempWord(word uint16) int {
	if word <= 0x7f {
		return 0x11 // >= 0, signed byte or word
	}
	if word <= 0xff {
		return 0x01 // < 0, signed byte
	}
	if 0xff80 <= word {
		return 0x10 // < 0, signed word
	}
	return 0x00
}

func checkTempRange(t int8, t1 int8, t2 int8, lo *int8, hi *int8) bool {
	if t1 > t2 {
		tx := t1
		t1 = t2
		t2 = tx
	}

	if -60 <= t1 && t1 <= t && t <= t2 && t2 <= 120 && !(t1 == -1 && t2 <= 0) {
		*lo = t1
		*hi = t2
		return true
	}
	return false
}

type AtaSmartPage struct {
	Version uint16
	Attrs   map[uint8]AtaSmartAttr
}

func (d *SataDevice) ReadSMARTData() (*AtaSmartPage, error) {
	pageRaw, err := d.readSMARTData()
	if err != nil {
		return nil, err
	}

	page := AtaSmartPage{}
	page.Version = pageRaw.Version
	page.Attrs = make(map[uint8]AtaSmartAttr)

	for _, a := range pageRaw.Attrs {
		if a.Id == 0 {
			break
		}

		n := AtaSmartAttr{}
		n.Id = a.Id
		n.Flags = a.Flags
		n.Current = a.Current
		n.Worst = a.Worst
		n.VendorBytes = a.VendorBytes

		// extract raw value
		if mapping, ok := d.attributeMapping[int(a.Id)]; ok {
			n.ValueRaw = computeAttributeRawValue(mapping, a.VendorBytes, a.Reserved, a.Current, a.Worst)
			n.Type = mapping.typ
			n.Name = mapping.name
		}

		page.Attrs[n.Id] = n
	}

	return &page, nil
}

func computeAttributeRawValue(mapping ataDeviceAttr, vendorBytes [6]byte, reserved uint8, current uint8, worst uint8) uint64 {
	byteOrder := mapping.byteOrder
	switch mapping.typ {
	case AtaDeviceAttributeTypeRaw64:
		fallthrough
	case AtaDeviceAttributeTypeHex64:
		byteOrder = "543210wv"

	case AtaDeviceAttributeTypeRaw56:
		fallthrough
	case AtaDeviceAttributeTypeHex56:
		fallthrough
	case AtaDeviceAttributeTypeRaw24DivRaw32:
		fallthrough
	case AtaDeviceAttributeTypeMsec24Hour32:
		byteOrder = "r543210"

	default:
		byteOrder = "543210"
	}

	// Build 64-bit value from selected bytes
	var rawValue uint64

	for _, c := range byteOrder {
		var b uint8

		switch c {
		case '0':
			b = vendorBytes[0]
		case '1':
			b = vendorBytes[1]
		case '2':
			b = vendorBytes[2]
		case '3':
			b = vendorBytes[3]
		case '4':
			b = vendorBytes[4]
		case '5':
			b = vendorBytes[5]
		case 'r':
			b = reserved
		case 'v':
			b = current
		case 'w':
			b = worst
		}

		rawValue <<= 8
		rawValue |= uint64(b)
	}
	return rawValue
}

type AtaSmartThresholdsPage struct {
	Revnumber  uint16
	Thresholds map[uint8]uint8
}

func (d *SataDevice) ReadSMARTThresholds() (*AtaSmartThresholdsPage, error) {
	pageRaw, err := d.readSMARTThresholds()
	if err != nil {
		return nil, err
	}

	page := AtaSmartThresholdsPage{}
	page.Revnumber = pageRaw.Revnumber
	page.Thresholds = make(map[uint8]uint8)

	for _, a := range pageRaw.ThresholdEntries {
		if a.Id == 0 {
			break
		}

		page.Thresholds[a.Id] = a.Threshold
	}

	return &page, nil
}
