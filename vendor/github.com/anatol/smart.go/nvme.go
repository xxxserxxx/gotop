package smart

import "bytes"

// Uint128 is a stopgap until https:// github.com/golang/go/issues/9455 is implemented
type Uint128 struct {
	// uint128 is represented as pair of uint64. Val[0] represents lower part of the uint128 value.
	Val [2]uint64
}

// Identify Controller Data Structure (CNS 01h)
// The Identify Controller data structure is returned to the host for the controller processing
// the command.
type NvmeIdentController struct {
	// PCI Vendor ID (VID): Contains the company vendor identifier that is assigned by the
	// PCI SIG. This is the same value as reported in the ID register in the PCI Header section
	// of the NVMe over PCIe Transport Specification.
	VendorID uint16
	// PCI Subsystem Vendor ID (SSVID): Contains the company vendor identifier that is
	// assigned by the PCI SIG for the subsystem. This is the same value as reported in the
	// SS register in the PCI Header section of the NVMe over PCIe Transport Specification.
	Ssvid uint16
	// Serial Number (SN): Contains the serial number for the NVM subsystem that is
	// assigned by the vendor as an ASCII string.
	SerialNumberRaw [20]byte
	// Model Number (MN): Contains the model number for the NVM subsystem that is
	// assigned by the vendor as an ASCII string.
	ModelNumberRaw [40]byte
	// Firmware Revision (FR): Contains the currently active firmware revision, as an ASCII
	// string, for the domain of which this controller is a part. This is the same revision
	// information that may be retrieved with the Get Log Page command.
	FirmwareRevRaw [8]byte
	// Recommended Arbitration Burst (RAB): This is the recommended Arbitration Burst
	// size. The value is in commands and is reported as a power of two (2^n). This is the
	// same units as the Arbitration Burst size.
	Rab uint8
	// IEEE OUI Identifier (IEEE): Contains the Organization Unique Identifier (OUI) for the
	// controller vendor. The OUI shall be a valid IEEE/RAC assigned identifier that may be
	// registered at http:// standards.ieee.org/develop/regauth/oui/public.html
	IEEE [3]byte
	// Controller Multi-Path I/O and Namespace Sharing Capabilities (CMIC): This field
	// specifies multi-path I/O and namespace sharing capabilities of the controller and NVM
	// subsystem.
	// Bits 7:4 are reserved.
	// Bit 3 if set to ‘1’, then the NVM subsystem supports Asymmetric Namespace Access
	// Reporting (refer to section 8.1). If cleared to ‘0’, then the NVM subsystem does not
	// support Asymmetric Namespace Access Reporting.
	// Bit 2 if set to ‘1’, then the controller is associated with an SR-IOV Virtual Function. If
	// cleared to ‘0’, then the controller is associated with a PCI Function or a Fabrics
	// connection.
	// Bit 1 if set to ‘1’, then the NVM subsystem may contain two or more controllers. If cleared
	// to ‘0’, then the NVM subsystem contains only a single controller. As described in section
	// 2.4.1, an NVM subsystem that contains multiple controllers may be used by multiple
	// hosts, or may provide multiple paths for a single host.
	// Bit 0 if set to ‘1’, then the NVM subsystem may contain more than one NVM subsystem
	// port. If cleared to ‘0’, then the NVM subsystem contains only a single NVM subsystem
	// port.
	Cmic uint8
	// Maximum Data Transfer Size (MDTS): This field indicates the maximum data transfer
	// size for a command that transfers data between host-accessible memory (refer to
	// section 1.5.26) and the controller. The host should not submit a command that exceeds
	// this maximum data transfer size. If a command is submitted that exceeds this transfer
	// size, then the command is aborted with a status code of Invalid Field in Command. The
	// value is in units of the minimum memory page size (CAP.MPSMIN) and is reported as
	// a power of two (2^n). A value of 0h indicates that there is no maximum data transfer
	// size. This field includes the length of metadata, if metadata is interleaved with the user
	// data. This field does not apply to commands that do not transfer data between host-accessible
	// memory and the controller (e.g., the Verify command, the Write
	// Uncorrectable command, and the Write Zeroes command); refer to the ONCS field for
	// restrictions on these commands and other commands that transfer data.
	// If SGL Bit Bucket descriptors are supported, their lengths shall be included in
	// determining if a command exceeds the Maximum Data Transfer Size for destination
	// data buffers. Their length in a source data buffer is not included for a Maximum Data
	// Transfer Size calculation.
	Mdts uint8
	// Controller ID (CNTLID): Contains the NVM subsystem unique controller identifier
	// associated with the controller.
	Cntlid uint16
	// Version (VER): This field contains the value reported in the Version property defined in
	// section 3.1.3.2. Implementations compliant to NVM Express Base Specification revision
	// 1.2 or later shall report a non-zero value in this field.
	Ver uint32
	// RTD3 Resume Latency (RTD3R): This field indicates the expected latency in
	// microseconds to resume from Runtime D3 (RTD3). Refer to section 8.15.4. A value of
	// 0h indicates RTD3 Resume Latency is not reported.
	Rtd3r uint32
	// RTD3 Entry Latency (RTD3E): This field indicates the typical latency in microseconds
	// to enter Runtime D3 (RTD3). Refer to section 8.15.4. A value of 0h indicates RTD3
	// Entry Latency is not reported.
	Rtd3e uint32 // RTD3 Entry Latency
	// Optional Asynchronous Events Supported (OAES): This field indicates the optional
	// asynchronous events supported by the controller. A controller shall not send optional
	// asynchronous events before they are enabled by host software.
	// Bit 31 is set to ‘1’ if the controller supports sending Discovery Log Page Change
	// Notifications. If cleared to ‘0’, then the controller does not support the Discovery Log
	// Page Change Notification events.
	// Bits 30:28 are reserved.
	// Bit 27 is set to ‘1’ if the controller supports the Zone Descriptor Changed Notices event
	// and the associated Changed Zone List log page (refer to the Zoned Namespace
	// Command Set specification). If cleared to ‘0’, then the controller does not support the
	// Zone Descriptor Changed Notices event nor the associated Changed Zone List log
	// page.
	// Bits 26:16 are reserved.
	// Bit 15 is set to ‘1’ if the controller supports the Normal NVM Subsystem Shutdown event.
	// If cleared to ‘0’, then the controller does not support the Normal NVM Subsystem
	// Shutdown event.
	// Bit 14 is set to ‘1’ if the controller supports the Endurance Group Event Aggregate Log
	// Page Change Notices event. If cleared to ‘0’, then the controller does not support the
	// Endurance Group Event Aggregate Log Page Change Notices event.
	// Bit 13 is set to ‘1’ if the controller supports the LBA Status Information Alert Notices
	// event (refer to the NVM Command Set specification). If cleared to ‘0’, then the controller
	// does not support the LBA Status Information Alert Notices event.
	// Bit 12 is set to ‘1’ if the controller supports the Predictable Latency Event Aggregate Log
	// Change Notices event. If cleared to ‘0’, then the controller does not support the
	// Predictable Latency Event Aggregate Log Change Notices event.
	// Bit 11 is set to ‘1’ if the controller supports sending Asymmetric Namespace Access
	// Change Notices. If cleared to ‘0’, then the controller does not support the Asymmetric
	// Namespace Access Change Notices event.
	// Bit 10 is reserved.
	// Bit 9 is set to ‘1’ if the controller supports the Firmware Activation Notices event. If
	// cleared to ‘0’, then the controller does not support the Firmware Activation Notices
	// event.
	// Bit 8 is set to ‘1’ if the controller supports the Namespace Attribute Notices event and
	// the associated Changed Namespace List log page. If cleared to ‘0’, then the controller
	// does not support the Namespace Attribute Notices event nor the associated Changed
	// Namespace List log page.
	// Bits 7:0 are reserved.
	Oaes uint32 // Optional Asynchronous Events Supported
	// Controller Attributes (CTRATT): This field indicates attributes of the controller.
	// Bits 31:16 Reserved
	// Bit 15 - Extended LBA Formats Supported (ELBAS): If set to ‘1’ indicates that
	// the controller supports the I/O command set specific extended protection
	// information formats (refer to the Protection Information Formats section of
	// the applicable I/O command set specification).
	// If cleared to ‘0’ indicates that the controller does not support the I/O
	// command set specific extended protection information formats (refer to the
	// Protection Information Formats section of the NVM Command Set
	// Specification).
	// Refer to the LBA Format Extension Enable (LBAFEE) field in the Host
	// Behavior Support feature (refer to section 5.27.1.18) for details for host
	// software to enable the controller to operate on namespaces using the
	// protection information formats.
	// NOTE: This bit field applies to all I/O Command Sets. The original name
	// has been retained for historical continuity.
	// Bit 14:
	// Delete NVM Set: If set to ‘1’, then the controller supports the Delete NVM
	// Set operation (refer to section 8.3.3). If cleared to ‘0’, then the controller
	// does not support the Delete NVM Set operation.
	// Bit 13:
	// Delete Endurance Group: If set to ‘1’, then the controller supports the
	// Delete Endurance Group operation (refer to section 8.3.3). If cleared to ‘0’,
	// then the controller does not support the Delete Endurance Group
	// operation.
	// 12
	// Variable Capacity Management: If set to ‘1’, then the controller supports
	// Variable Capacity Management (refer to section 8.3.3). If cleared to ‘0’,
	// then the controller does not support Variable Capacity Management.
	// 11
	// Fixed Capacity Management: If set to ‘1’, then the controller supports
	// Fixed Capacity Management (refer to section 8.3.2). If cleared to ‘0’, then
	// the controller does not support Fixed Capacity Management.
	// 10
	// Multi-Domain Subsystem (MDS): If set to ‘1’, then the NVM subsystem
	// supports the multiple domains (refer to section 3.2.4). If cleared to ‘0’, then
	// the NVM subsystem does not support the reporting of multiple domains
	// and the NVM subsystem consists of a single domain.
	// 9
	// UUID List: If set to ‘1’, then the controller supports reporting of a UUID List
	// (refer to Figure 284). If cleared to ‘0’, then the controller does not support
	// reporting of a UUID List (refer to section 8.25).
	// 8
	// SQ Associations: If set to ‘1’, then the controller supports SQ Associations
	// (refer to section 8.22). If cleared to ‘0’, then the controller does not support
	// SQ Associations.
	// 7
	// Namespace Granularity: If set to ‘1’, then the controller supports reporting
	// of Namespace Granularity (refer to section 5.17.2.15). If cleared to ‘0’, the
	// controller does not support reporting of Namespace Granularity. If the
	// Namespace Management capability (refer to section 8.11) is not
	// supported, then this bit shall be cleared to ‘0’.
	// 6
	// Traffic Based Keep Alive Support (TBKAS): If set to '1‘, then the
	// controller supports restarting the Keep Alive Timer if an Admin command
	// or an I/O command is processed during the Keep Alive Timeout Interval
	// (refer to section 3.9.2). If cleared to '0‘, then the controller supports
	// restarting the Keep Alive Timer only if a Keep Alive command is processed
	// during the Keep Alive Timeout Interval (refer to section 3.9.1).
	// Predictable Latency Mode: If set to ‘1’, then the controller supports
	// Predictable Latency Mode (refer to section 8.16). If cleared to ‘0’, then the
	// controller does not support Predictable Latency Mode.
	// 4
	// Endurance Groups: If set to ‘1’, then the controller supports Endurance
	// Groups (refer to section 3.2.3). If cleared to ‘0’, then the controller does not
	// support Endurance Groups.
	// NVM Express® Base Specification, revision 2.0b
	// 245
	// Figure 275: Identify – Identify Controller Data Structure, I/O Command Set Independent
	// Bytes I/O1
	// Admin1
	// Disc1 Description
	// 3
	// Read Recovery Levels: If set to ‘1’, then the controller supports Read
	// Recovery Levels (refer to section 8.17). If cleared to ‘0’, then the controller
	// does not support Read Recovery Levels.
	// 2
	// NVM Sets: If set to ‘1’, then the controller supports NVM Sets (refer to
	// section 3.2.2). If cleared to ‘0’, then the controller does not support NVM
	// Sets.
	// 1
	// Non-Operational Power State Permissive Mode: If set to ‘1’, then the
	// controller supports host control of whether the controller may temporarily
	// exceed the power of a non-operational power state for the purpose of
	// executing controller initiated background operations in a non-operational
	// power state (i.e., Non-Operational Power State Permissive Mode
	// supported). If cleared to ‘0’, then the controller does not support host
	// control of whether the controller may exceed the power of a nonoperational state for the purpose of executing controller initiated
	// background operations in a non-operational state (i.e., Non-Operational
	// Power State Permissive Mode not supported). Refer to section 5.27.1.14.
	// 0
	// Host Identifier Support: If set to ‘1’, then the controller supports a 128-bit
	// Host Identifier. Bit 0 if cleared to ‘0’, then the controller does not support a
	// 128-bit Host Identifier.
	Ctratt uint32
	// Read Recovery Levels Supported (RRLS): If Read Recovery Levels (RRL) are
	// supported, then this field shall be supported. If a bit is set to ‘1’, then the corresponding
	// Read Recovery Level is supported. If a bit is cleared to ‘0’, then the corresponding Read
	// Recovery Level is not supported.
	// Bit Definition
	// 0 Read Recovery Level 0
	// 1 Read Recovery Level 1
	// 2 Read Recovery Level 2
	// 3 Read Recovery Level 3
	// 4 Read Recovery Level 4 – Default1
	// 5 Read Recovery Level 5
	// 6 Read Recovery Level 6
	// 7 Read Recovery Level 7
	// 8 Read Recovery Level 8
	// 9 Read Recovery Level 9
	// 10 Read Recovery Level 10
	// 11 Read Recovery Level 11
	// 12 Read Recovery Level 12
	// 13 Read Recovery Level 13
	// 14 Read Recovery Level 14
	// 15 Read Recovery Level 15 – Fast Fail1
	// NOTE:
	// 1. If Read Recovery Levels are supported, then this bit shall be set to ‘1’.
	Rrls uint16
	_    [9]byte
	// Controller Type (CNTRLTYPE): This field specifies the controller type. A value of 0h
	// indicates that the controller type is not reported.
	// Implementations compliant to NVM Express Base Specification revision 1.4 or later shall
	// report a controller type (i.e., the value 0h is reserved and shall not be used).
	// Implementations compliant to an earlier specification version may report a value of 0h
	// to indicate that a controller type is not reported.
	// Value Controller Type
	// 0h Reserved (controller type not reported)
	// 1h I/O controller
	// 2h Discovery controller
	// 3h Administrative controller
	// 4h to FFh Reserved
	CntrlType uint8
	// FRU Globally Unique Identifier (FGUID): This field contains a 128-bit value that is
	// globally unique for a given Field Replaceable Unit (FRU). Refer to the NVM Express®
	// Management Interface Specification for the definition of a FRU. This field remains fixed
	// throughout the life of the FRU. This field shall contain the same value for each controller
	// associated with a given FRU.
	// This field uses the EUI-64 based 16-byte designator format. Bytes 122:120 contain the
	// 24-bit Organizationally Unique Identifier (OUI) value assigned by the IEEE Registration
	// Authority. Bytes 127:123 contain an extension identifier assigned by the corresponding
	// organization. Bytes 119:112 contain the vendor specific extension identifier assigned
	// by the corresponding organization. Refer to the IEEE EUI-64 guidelines for more
	// information. This field is big endian (refer to section 4.3.4).
	// When not implemented, this field contains a value of 0h.
	Fguid [16]byte
	// Command Retry Delay Time 1 (CRDT1): If the Do Not Retry (DNR) bit is cleared to ‘0’
	// in the CQE and the Command Retry Delay (CRD) field is set to 01b in the CQE, then
	// this value indicates the command retry delay time in units of 100 milliseconds.
	// Command Retry Delay Time 2 (CRDT2): If the DNR bit is cleared to ‘0’ in the CQE
	Crdt1 uint16
	// Command Retry Delay Time 2 (CRDT2): If the DNR bit is cleared to ‘0’ in the CQE
	// and the CRD field is set to 10b in the CQE, then this value indicates the command retry
	// delay time in units of 100 milliseconds.
	// Command Retry Delay Time 3 (CRDT3): If the DNR bit is cleared to ‘0’ in the CQE
	Crdt2 uint16
	// Command Retry Delay Time 3 (CRDT3): If the DNR bit is cleared to ‘0’ in the CQE
	// and CRD field is set to 11b in the CQE, then this value indicates the command retry
	// delay time in units of 100 milliseconds.
	Crdt3 uint16
	_     [119]byte // ...
	// NVM Subsystem Report (NVMSR): This field reports information associated with the
	// NVM subsystem. If the controller is compliant to the NVMe Management Interface
	// Specification, then at least one bit in this field is set to ‘1’. If the NVM subsystem does
	// not support the NVMe Management Interface Specification, then this field shall be
	// cleared to 0h. Refer to the NVMe Management Interface Specification.
	// Bits Description
	// 7:2 Reserved
	// 1
	// NVMe Enclosure (NVMEE): If set to ‘1’, then the NVM subsystem is part of
	// an NVMe Enclosure. If cleared to ‘0’, then the NVM subsystem is not part of
	// an NVMe Enclosure.
	// 0
	// NVMe Storage Device (NVMESD): If set to ‘1’, then the NVM subsystem is
	// part of an NVMe Storage Device. If cleared to ‘0’, then the NVM subsystem
	// is not part of an NVMe Storage Device.
	Nvmsr uint8
	// VPD Write Cycle Information (VWCI): This field indicates information about the
	// remaining number of times that VPD contents are able to be updated using the VPD
	// Write command. Refer to the NVMe Management Interface Specification for details on
	// VPD contents and the VPD Write command.
	// Bits Description
	// 7
	// VPD Write Cycles Remaining Valid (VWCRV): If this bit is set to ‘1’, then
	// the VPD Write Cycles Remaining field is valid. If this bit is cleared to ‘0’, then
	// the VPD Write Cycles Remaining field is invalid and cleared to ‘0’.
	// 6:0
	// VPD Write Cycles Remaining (VWCR): If the VPD Write Cycle Remaining
	// Valid bit is set to ‘1’, then this field contains a value indicating the remaining
	// number of times that VPD contents are able to be updated in units of 256
	// bytes using the VPD Write command. For example, a 1 KiB FRU Information
	// Device that can be updated 8 times would indicate a value of 32 in this field.
	// If this field is set to 7Fh, then the remaining number of times that VPD
	// contents are able to be updated using the VPD Write command is greater
	// than or equal to 7Fh.
	// If the VPD Write Cycle Remaining Valid bit is cleared to ‘0’, then this field is
	// not valid and shall be cleared to a value of 0h.
	Vwci uint8
	// Management Endpoint Capabilities (MEC): This field indicates the capabilities of the
	// Management Endpoint in the NVM subsystem. Refer to the NVMe Management
	// Interface Specification for details.
	// Bits Description
	// 7:2 Reserved
	// 1
	// PCIe Port Management Endpoint (PCIEME): If set to ‘1’, then the NVM
	// subsystem contains a Management Endpoint on a PCIe port.
	// 0
	// SMBus/I2C Port Management Endpoint (SMBUSME): If set to ‘1’, then the
	// NVM subsystem contains a Management Endpoint on an SMBus/I2C port.
	Mec uint8
	// Optional Admin Command Support (OACS): This field indicates the optional Admin
	// commands and features supported by the controller. Refer to section 3.1.2.
	// Bits 15:11 are reserved.
	// Bit 10 if set to ‘1’, then the controller supports the Command and Feature Lockdown
	// capability (refer to section 8.4). If cleared to ‘0’, then the controller does not support the
	// Command and Feature Lockdown capability. This value shall be the same for all
	// controllers in the NVM subsystem.
	// Bit 9 if set to ‘1’, then the controller supports the Get LBA Status capability (refer to the
	// NVM Command Set Specification). If cleared to ‘0’, then the controller does not support
	// the Get LBA Status capability.
	// Bit 8 if set to '1', then the controller supports the Doorbell Buffer Config command. If
	// cleared to '0', then the controller does not support the Doorbell Buffer Config command.
	// Bit 7 if set to ‘1’, then the controller supports the Virtualization Management command.
	// If cleared to ‘0’, then the controller does not support the Virtualization Management
	// command.
	// Bit 6 if set to ‘1’, then the controller supports the NVMe-MI Send and NVMe-MI Receive
	// commands. If cleared to ‘0’, then the controller does not support the NVMe-MI Send
	// and NVMe-MI Receive commands.
	// Bit 5 if set to ‘1’, then the controller supports Directives. If cleared to ‘0’, then the
	// controller does not support Directives. A controller that supports Directives shall support
	// the Directive Send and Directive Receive commands. Refer to section 8.7.
	// Bit 4 if set to ‘1’, then the controller supports the Device Self-test command. If cleared
	// to ‘0’, then the controller does not support the Device Self-test command.
	// Bit 3 if set to ‘1’, then the controller supports the Namespace Management capability
	// (refer to section 8.11). If cleared to ‘0’, then the controller does not support the
	// Namespace Management capability.
	// Bit 2 if set to ‘1’, then the controller supports the Firmware Commit and Firmware Image
	// Download commands. If cleared to ‘0’, then the controller does not support the Firmware
	// Commit and Firmware Image Download commands.
	// Bit 1 if set to ‘1’, then the controller supports the Format NVM command. If cleared to
	// ‘0’, then the controller does not support the Format NVM command.
	// Bit 0 if set to ‘1’, then the controller supports the Security Send and Security Receive
	// commands. If cleared to ‘0’, then the controller does not support the Security Send and
	// Security Receive commands.
	Oacs uint16 // Optional Admin Command Support
	// Abort Command Limit (ACL): This field is used to convey the maximum number of
	// concurrently executing Abort commands supported by the controller (refer to section
	// 5.1). This is a 0’s based value. It is recommended that implementations support
	// concurrent execution of a minimum of four Abort commands.
	Acl uint8 // Abort Command Limit
	// Asynchronous Event Request Limit (AERL): This field is used to convey the
	// maximum number of concurrently outstanding Asynchronous Event Request
	// commands supported by the controller (refer to section 5.2). This is a 0’s based value.
	// It is recommended that implementations support a minimum of four Asynchronous
	// Event Request Limit commands outstanding simultaneously.
	Aerl uint8 // Asynchronous Event Request Limit
	// Firmware Updates (FRMW): This field indicates capabilities regarding firmware
	// updates. Refer to section 3.11 for more information on the firmware update process.
	// Bits Description
	// 7:6 Reserved
	// 5
	// Support Multiple Update Detection (SMUD): If set to ‘1’ indicates that the
	// controller is able to detect overlapping firmware/boot partition image update
	// command sequences (refer to section 3.11 and section 8.2.2). If cleared to ‘0’,
	// then the controller is not able to detect overlapping firmware/boot partition
	// image update command sequences.
	// 4
	// Firmware Activation Without Reset (FAWR): If set to ‘1’ indicates that the
	// controller supports firmware activation without a reset. If cleared to ‘0’, then the
	// controller requires a reset for firmware to be activated.
	// 3:1
	// Number Of Firmware Slots (NOFS): This field indicates the number of
	// firmware slots supported by the domain that contains this controller. This field
	// shall specify a value from one to seven, indicating that at least one firmware
	// slot is supported and up to seven maximum. This corresponds to firmware slots
	// 1 through 7
	// 0
	// First Firmware Slot Read Only (FFSRO): If set to ‘1’ indicates that the first
	// firmware slot (i.e., slot 1) is read only. If cleared to ‘0’, then the first firmware
	// slot (i.e., slot 1) is read/write. Implementations may choose to have a baseline
	// read only firmware image.
	Frmw uint8 // Firmware Updates
	// Log Page Attributes (LPA): This field indicates optional attributes for log pages that
	// are accessed via the Get Log Page command.
	// Bits Description
	// 7 Reserved
	// 6
	// If set to ‘1’, then the controller supports Data Area 4 for the Telemetry HostInitiated and Telemetry Controller-Initiated log. If cleared to ’0’, then the
	// controller does not support Data Area 4 for the Telemetry Host-Initiated and
	// Telemetry Controller-Initiated log pages.
	// 5
	// If set to ‘1’, then the controller supports:
	// • the Supported Log Pages log page (Log Identifier 0h);
	// • returning the scope of each command in the Commands Supported
	// and Effects log page (Log Identifier 05h);
	// • the Feature Identifiers Supported and Effects log page (Log
	// Identifier 12h); and
	// • the NVMe-MI Commands Supported and Effects log page (Log
	// Identifier 13h).
	// If cleared to ‘0’, then the controller:
	// • does not support returning the scope of each command in the
	// Commands Supported and Effects log page;
	// • may support the Supported Log Pages log page;
	// • may support the Feature Identifiers Supported and Effects log
	// page; and
	// • may support the NVMe-MI Commands Supported and Effects log
	// page.
	// 4
	// If set to ‘1’, then the controller supports the Persistent Event log. If cleared
	// to ‘0’, then the controller does not support the Persistent Event log.
	// 3
	// If set to ‘1’, then the controller supports the Telemetry Host-Initiated and
	// Telemetry Controller-Initiated log pages and sending Telemetry Log Notices.
	// If cleared to ’0’, then the controller does not support the Telemetry HostInitiated and Telemetry Controller-Initiated log pages and Telemetry Log
	// Notice events.
	// 2
	// If set to ‘1’, then the controller supports extended data for the Get Log Page
	// command (including extended Number of Dwords and Log Page Offset
	// fields). If cleared to ‘0’, then the controller does not support extended data
	// for the Get Log Page command.
	// 1
	// If set to ‘1’, then the controller supports the Commands Supported and
	// Effects log page. Bit 1 if cleared to ‘0’, then the controller does not support
	// the Commands Supported and Effects log page.
	// 0
	// If set to ‘1’, then the controller supports the SMART / Health Information log
	// page on a per namespace basis. If cleared to ‘0’, then the controller does not
	// support the SMART / Health Information log page on a per namespace basis.
	Lpa uint8 // Log Page Attributes
	// Error Log Page Entries (ELPE): This field indicates the maximum number of Error
	// Information log entries that are stored by the controller. This field is a 0’s based value.
	Elpe uint8 // Error Log Page Entries
	// Number of Power States Support (NPSS): This field indicates the number of NVM
	// Express power states supported by the controller. This is a 0’s based value. Refer to
	// section 8.15.
	// Power states are numbered sequentially starting at power state 0. A controller shall
	// support at least one power state (i.e., power state 0) and may support up to 31 additional
	// power states (i.e., up to 32 total).
	Npss uint8 // Number of Power States Support
	// Admin Vendor Specific Command Configuration (AVSCC): This field indicates the
	// configuration settings for Admin Vendor Specific command handling. Refer to section
	// 8.23.
	// Bits 7:1 are reserved.
	// Bit 0 if set to ‘1’ indicates that all Admin Vendor Specific Commands use the format
	// defined in Figure 88. If cleared to ‘0’ indicates that the format of all Admin Vendor
	// Specific Commands are vendor specific.
	Avscc uint8 // Admin Vendor Specific Command Configuration
	// Autonomous Power State Transition Attributes (APSTA): This field indicates the
	// attributes of the autonomous power state transition feature. Refer to section 8.15.2.
	// Bits 7:1 are reserved.
	// Bit 0 if set to ‘1’, then the controller supports autonomous power state transitions. If
	// cleared to ‘0’, then the controller does not support autonomous power state transitions.
	Apsta uint8 // Autonomous Power State Transition Attributes
	// Warning Composite Temperature Threshold (WCTEMP): This field indicates the
	// minimum Composite Temperature field value (reported in the SMART / Health
	// Information log in Figure 207) that indicates an overheating condition during which
	// controller operation continues. Immediate remediation is recommended (e.g., additional
	// cooling or workload reduction). The platform should strive to maintain a composite
	// temperature less than this value.
	// A value of 0h in this field indicates that no warning temperature threshold value is
	// reported by the controller. Implementations compliant to NVM Express Base
	// Specification revision 1.2 or later shall report a non-zero value in this field.
	// It is recommended that implementations report a value of 0157h in this field.
	Wctemp uint16 // Warning Composite Temperature Threshold
	// Critical Composite Temperature Threshold (CCTEMP): This field indicates the
	// minimum Composite Temperature field value (reported in the SMART / Health
	// Information log in Figure 207) that indicates a critical overheating condition (e.g., may
	// prevent continued normal operation, possibility of data loss, automatic device shutdown,
	// extreme performance throttling, or permanent damage).
	// A value of 0h in this field indicates that no critical temperature threshold value is reported
	// by the controller. Implementations compliant to NVM Express Base Specification
	// revision 1.2 or later shall report a non-zero value in this field.
	Cctemp uint16 // Critical Composite Temperature Threshold
	// Maximum Time for Firmware Activation (MTFA): Indicates the maximum time the
	// controller temporarily stops processing commands to activate the firmware image. This
	// field shall be valid if the controller supports firmware activation without a reset. This field
	// is specified in 100 millisecond units. A value of 0h indicates that the maximum time is
	// undefined.
	Mtfa uint16 // Maximum Time for Firmware Activation
	// Host Memory Buffer Preferred Size (HMPRE): This field indicates the preferred size
	// that the host is requested to allocate for the Host Memory Buffer feature in 4 KiB units.
	// This value shall be greater than or equal to the Host Memory Buffer Minimum Size. If
	// this field is non-zero, then the Host Memory Buffer feature is supported. If this field is
	// cleared to 0h, then the Host Memory Buffer feature is not supported.
	Hmpre uint32 // Host Memory Buffer Preferred Size
	// Host Memory Buffer Minimum Size (HMMIN): This field indicates the minimum size
	// that the host is requested to allocate for the Host Memory Buffer feature in 4 KiB units.
	// If this field is cleared to 0h, then the host is requested to allocate any amount of host
	// memory possible up to the HMPRE value
	Hmmin uint32 // Host Memory Buffer Minimum Size
	// Total NVM Capacity (TNVMCAP): This field indicates the total NVM capacity that is
	// accessible by the controller. The value is in bytes. This field shall be supported if the
	// Namespace Management capability (refer to section 8.11) is supported or if the
	// Capacity Management capability (refer to section 8.3) is supported.
	// Refer to section 3.8.
	Tnvmcap Uint128 // Total NVM Capacity
	// Unallocated NVM Capacity (UNVMCAP): This field indicates the unallocated NVM
	// capacity that is accessible by the controller. The value is in bytes. This field shall be
	// supported if the Namespace Management capability (refer to section 8.11) is supported
	// or if the Capacity Management capability (refer to section 8.3) is supported.
	// Refer to section 3.8.
	Unvmcap Uint128 // Unallocated NVM Capacity
	// Replay Protected Memory Block Support (RPMBS): This field indicates if the
	// controller supports one or more Replay Protected Memory Blocks (RPMBs) and the
	// capabilities. Refer to section 8.18.
	// Bits Description
	// 31:24
	// Access Size: If the Number of RPMB Units field is non-zero, then this field
	// indicates the maximum number of 512B units of data that may be read or
	// written per RPMB access by Security Send or Security Receive commands
	// for the controller. This is a 0’s based value. A value of 0h indicates support
	// for one unit of 512B of data.
	// If the Number of RPMB Units field is 0h, then this field shall be ignored.
	// 23:16
	// Total Size: If the Number of RPMB Units field is non-zero, then this field
	// indicates the number of 128 KiB units of data in each RPMB supported in
	// the controller. This is a 0’s based value. A value of 0h indicates support for
	// one unit of 128 KiB of data.
	// If the Number of RPMB Units field is 0h, this field shall be ignored.
	// 15:06 Reserved
	// 05:03
	// Authentication Method: This field indicates the authentication method
	// used to access all RPMBs in the controller. The values for this field are:
	// Value Definition
	// 000b HMAC SHA-256 (refer to RFC 6234)
	// 001b to 111b Reserved
	// 02:00
	// Number of RPMB Units: This field indicates the number of RPMB targets
	// the controller supports. All RPMB targets supported shall have the same
	// capabilities as defined in the RPMBS field. A value of 0h indicates the
	// controller does not support Replay Protected Memory Blocks. If this value
	// is non-zero, then the controller shall support the Security Send and Security
	// Receive commands.
	Rpmbs uint32 // Replay Protected Memory Block Support
	// Extended Device Self-test Time (EDSTT): If the Device Self-test command is
	// supported, then this field indicates the nominal amount of time in one minute units that
	// the controller takes to complete an extended device self-test operation when in power
	// state 0. If the Device Self-test command is not supported, then this field is reserved.
	Edstt uint16
	// Device Self-test Options (DSTO): This field indicates the optional Device Self-test
	// command or operation behaviors supported by the controller or NVM subsystem.
	// Bits 7:1 are reserved.
	// Bit 0 if set to ‘1’, then the NVM subsystem supports only one device self-test operation
	// in progress at a time. If cleared to ‘0’, then the NVM subsystem supports one device
	// self-test operation per controller at a time.
	Dsto uint8
	// Firmware Update Granularity (FWUG): This field indicates the granularity and
	// alignment requirement of the firmware image being updated by the Firmware Image
	// Download command (refer to section 5.13). If the values specified in the NUMD field or
	// the OFST field in the Firmware Image Download command do not conform to this
	// granularity and alignment requirement, then the firmware update may abort with a status
	// code of Invalid Field in Command. For the broadest interoperability with host software,
	// it is recommended that the controller set this value to the lowest value possible.
	// The value is reported in 4 KiB units (e.g., 1h corresponds to 4 KiB, 2h corresponds to
	// 8 KiB). A value of 0h indicates that no information on granularity is provided. A value of
	// FFh indicates there is no restriction (i.e., any granularity and alignment in dwords is
	// allowed).
	Fwug uint8
	// Keep Alive Support (KAS): This field indicates the granularity of the Keep Alive Timer
	// in 100 millisecond units (refer to section 3.9). If this field is cleared to 0h, then the Keep
	// Alive feature is not supported. The Keep Alive feature shall be supported for NVMe over
	// Fabrics implementations as described in section 3.9
	Kas uint16
	// Host Controlled Thermal Management Attributes (HCTMA): This field indicates the
	// attributes of the host controlled thermal management feature. Refer to section 8.15.5.
	// Bits 15:1 are reserved.
	// Bit 0 if set to ‘1’, then the controller supports host controlled thermal management. If
	// cleared to ‘0’, then the controller does not support host controlled thermal management.
	// If this bit is set to ‘1’, then the controller shall support the Set Features command and
	// Get Features command with the Feature Identifier field set to 10h.
	Hctma uint16
	// Minimum Thermal Management Temperature (MNTMT): This field indicates the
	// minimum temperature, in Kelvins, that the host may request in the Thermal
	// Management Temperature 1 field and Thermal Management Temperature 2 field of a
	// Set Features command with the Feature Identifier field set to 10h. A value of 0h
	// indicates that the controller does not report this field or the host controlled thermal
	// management feature (refer to section 8.15.5) is not supported.
	Mntmt uint16
	// Maximum Thermal Management Temperature (MXTMT): This field indicates the
	// maximum temperature, in Kelvins, that the host may request in the Thermal
	// Management Temperature 1 field and Thermal Management Temperature 2 field of the
	// Set Features command with the Feature Identifier set to 10h. A value of 0h indicates
	// that the controller does not report this field or the host controlled thermal management
	// feature is not supported.
	Mxtmt uint16
	// Sanitize Capabilities (SANICAP): This field indicates attributes for sanitize operations.
	// If the Sanitize command is supported, then this field shall be non-zero. If the Sanitize
	// command is not supported, then this field shall be cleared to 0h. Refer to section 8.21.
	// Bits Description
	// 31:30
	// No-Deallocate Modifies Media After Sanitize (NODMMAS):
	// This field indicates if media is additionally modified by the
	// controller after a sanitize operation successfully completes that
	// had been started by a Sanitize command with the No-Deallocate
	// After Sanitize bit set to ‘1’.
	// The work required for the associated additional media modification
	// is included both in the estimated time for each sanitize operation
	// and in the Sanitize Progress field (refer to Figure 267).
	// Value Definition
	// 00b
	// Additional media modification after sanitize
	// operation completes successfully is not
	// defined. Only controllers compliant with NVM
	// Express Base Specification revision 1.3 and
	// earlier or that have bits 2:0 of the SANICAP
	// field cleared to 0h shall be allowed to return
	// this value.
	// 01b
	// Media is not additionally modified by the
	// NVMe controller after sanitize operation
	// completes successfully.
	// 10b
	// Media is additionally modified by the NVMe
	// controller after sanitize operation completes
	// successfully. The Sanitize Operation
	// Completed event does not occur until the
	// additional media modification associated with
	// this field has completed.
	// 11b Reserved
	// If bits 2:0 of the SANICAP field are cleared to 000b, then the
	// controller shall clear this field to 00b.
	// 29
	// No-Deallocate Inhibited (NDI): If set to ‘1’ and the No-Deallocate
	// Response Mode bit is set to ‘1’, then the controller deallocates
	// after the sanitize operation even if the No-Deallocate After Sanitize
	// bit is set to ‘1’ in a Sanitize command.
	// If:
	// a) this bit is set to ‘1’;
	// b) the No-Deallocate After Sanitize bit is set to ‘1’ in a
	// Sanitize command, and:
	// 1) the No-Deallocate Response Mode bit (refer to
	// Figure 352) is cleared to ‘0’; or
	// 2) the Sanitize Config Feature (refer to section
	// 5.27.1.19) is not supported,
	// then the controller aborts the Sanitize command with a status code
	// of Invalid Field in Command.
	// If the No-Deallocate After Sanitize bit is cleared to ‘0’ in a Sanitize
	// command, then the value of this bit has no effect on the processing
	// that Sanitize command.
	// If the No-Deallocate After Sanitize bit is cleared to ‘0’ in a Sanitize
	// command, then the value of this bit has no effect on the processing
	// that Sanitize command.
	// If this bit is cleared to ‘0’, then the controller supports the NoDeallocate After Sanitize bit in a Sanitize command.
	// If bits 2:0 of the SANICAP field are cleared to 0h, then the
	// controller shall clear this bit to ‘0’
	// 28:03 Reserved
	// 02
	// Overwrite Support (OWS): If set to ‘1’, then the controller
	// supports the Overwrite sanitize operation. If cleared to ‘0’, then the
	// controller does not support the Overwrite sanitize operation.
	// 01
	// Block Erase Support (BES): If set to ‘1’, then the controller
	// supports the Block Erase sanitize operation. If cleared to ‘0’, then
	// the controller does not support the Block Erase sanitize operation.
	// 00
	// Crypto Erase Support (CES): If set to ‘1’, then the controller
	// supports the Crypto Erase sanitize operation. If cleared to ‘0’, then
	// the controller does not support the Crypto Erase sanitize
	// operation.
	Sanicap uint32
	_       [180]byte // ...
	Sqes    uint8     // Submission Queue Entry Size
	Cqes    uint8     // Completion Queue Entry Size
	_       [2]byte   // (defined in NVMe 1.3 spec)
	// Number of Namespaces (NN): This field indicates the maximum value of a valid NSID
	// for the NVM subsystem. Refer to the MNAN field for the number of supported
	// namespaces in the NVM subsystem.
	Nn    uint32                  // Number of Namespaces
	Oncs  uint16                  // Optional NVM Command Support
	Fuses uint16                  // Fused Operation Support
	Fna   uint8                   // Format NVM Attributes
	Vwc   uint8                   // Volatile Write Cache
	Awun  uint16                  // Atomic Write Unit Normal
	Awupf uint16                  // Atomic Write Unit Power Fail
	Nvscc uint8                   // NVM Vendor Specific Command Configuration
	_     uint8                   // ...
	Acwu  uint16                  // Atomic Compare & Write Unit
	_     [2]byte                 // ...
	Sgls  uint32                  // SGL Support
	_     [1508]byte              // ...
	Psd   [32]NvmeIdentPowerState // Power State Descriptors
	Vs    [1024]byte              // Vendor Specific
} // 4096 bytes

func (c *NvmeIdentController) ModelNumber() string {
	return string(bytes.TrimSpace(c.ModelNumberRaw[:]))
}

func (c *NvmeIdentController) SerialNumber() string {
	return string(bytes.TrimSpace(c.SerialNumberRaw[:]))
}

func (c *NvmeIdentController) FirmwareRev() string {
	return string(bytes.TrimSpace(c.FirmwareRevRaw[:]))
}

type NvmeIdentPowerState struct {
	MaxPower        uint16 // Maximum Power (specified in MaxPowerScale units)
	_               uint8
	Flags           uint8  // bit 0 - MaxPowerScale, bit 1 - Non-Operational State
	EntryLat        uint32 // Entry Latency
	ExitLat         uint32 // Exit Latency
	ReadThroughput  uint8
	ReadLatency     uint8
	WriteThroughput uint8
	WriteLatency    uint8
	IdlePower       uint16
	IdleScale       uint8
	_               uint8
	ActivePower     uint16
	ActiveWorkScale uint8 // Active Power Workload + Active Power Scale
	_               [9]byte
}

type NvmeLBAF struct {
	// Metadata Size (MS): This field indicates the number of metadata bytes provided per LBA based
	// on the LBA Data Size indicated. If there is no metadata supported, then this field shall be cleared
	// to 0h.
	// If metadata is supported, then the namespace may support the metadata being transferred as
	// part of an extended data LBA or as part of a separate contiguous buffer. If end-to-end data
	// protection is enabled, then the first eight bytes or last eight bytes of the metadata is the protection
	// information (refer to the DPS field in the Identify Namespace data structure).
	Ms uint16 // Metadata Size
	// LBA Data Size (LBADS): This field indicates the LBA data size supported. The value is reported
	// in terms of a power of two (2^n). A value smaller than 9 (i.e., 512 bytes) is not supported. If the
	// value reported is 0h, then the LBA format is not supported / used or is not currently available
	Ds uint8 // LBA Data Size
	// Relative Performance (RP): This field indicates the relative performance of the LBA format
	// indicated relative to other LBA formats supported by the controller. Depending on the size of the
	// LBA and associated metadata, there may be performance implications. The performance
	// analysis is based on better performance on a queue depth 32 with 4 KiB read workload. The
	// meanings of the values indicated are included in the following table.
	// Value Definition
	// 00b Best performance
	// 01b Better performance
	// 10b Good performance
	// 11b Degraded performance
	Rp uint8 // Relative Performance
}

type NvmeIdentNamespace struct {
	// Namespace Size (NSZE): This field indicates the total size of the namespace in logical
	// blocks. A namespace of size n consists of LBA 0 through (n - 1). The number of logical
	// blocks is based on the formatted LBA size.
	Nsze uint64 // Namespace Size
	// Namespace Capacity (NCAP): This field indicates the maximum number of logical
	// blocks that may be allocated in the namespace at any point in time. The number of logical
	// blocks is based on the formatted LBA size. Spare LBAs are not reported as part of this
	// field.
	// Refer to section 2.1.1 for details on the usage of this field.
	Ncap uint64 // Namespace Capacity
	// Namespace Utilization (NUSE): This field indicates the current number of logical blocks
	// allocated in the namespace. This field is less than or equal to the Namespace Capacity.
	// The number of logical blocks is based on the formatted LBA size.
	// Refer to section 2.1.1 for details on the usage of this field.
	Nuse uint64 // Namespace Utilization
	// Namespace Features (NSFEAT): This field defines features of the namespace.
	// Bits 7:5 are reserved.
	// Bit 4 (OPTPERF) if set to ‘1’ indicates that the fields NPWG, NPWA, NPDG, NPDA, and
	// NOWS are defined for this namespace and should be used by the host for I/O optimization
	// (refer to section 5.8.2). If cleared to ‘0’, then the controller does not support the fields
	// NPWG, NPWA, NPDG, NPDA, and NOWS for this namespace.
	// Bit 3 (UIDREUSE) This bit is as defined in the UIDREUSE bit in the I/O Command Set
	// Independent Identify Namespace data structure (refer to the I/O Command Set
	// Independent Identify Namespace data structure section in the NVMe Base Specification).
	// Bit 2 (DAE) if set to ‘1’ indicates that the controller supports the Deallocated or Unwritten
	// Logical Block error for this namespace. If cleared to ‘0’, then the controller does not
	// support the Deallocated or Unwritten Logical Block error for this namespace. Refer to
	// section 3.2.3.2.1.
	// Bit 1 (NSABP) if set to ‘1’ indicates that the fields NAWUN, NAWUPF, and NACWU are
	// defined for this namespace and should be used by the host for this namespace instead
	// of the AWUN, AWUPF, and ACWU fields in the Identify Controller data structure. If
	// cleared to ‘0’, then the controller does not support the fields NAWUN, NAWUPF, and
	// NACWU for this namespace. In this case, the host should use the AWUN, AWUPF, and
	// ACWU fields defined in the Identify Controller data structure in the NVMe Base
	// Specification. Refer to section 2.1.4.
	// Bit 0 (THINP) if set to ‘1’ indicates that the namespace supports thin provisioning. If
	// cleared to ‘0’ indicates that thin provisioning is not supported Refer to section 2.1.1 for
	// details on the usage of this bit.
	Nsfeat uint8 // Namespace Features
	// Number of LBA Formats (NLBAF): This field defines the number of supported LBA data
	// size and metadata size combinations supported by the namespace. LBA formats shall be
	// allocated in order (starting with 0) and packed sequentially. This is a 0’s based value. The
	// maximum number of LBA formats that may be indicated as supported is:
	// a) 16 if the LBA Format Extension Enable (LBAFEE) field is cleared to 0h in the
	// Host Behavior Support feature (refer to the Host Behavior Support section in the
	// NVMe Base Specification); or
	// b) 64 if the LBAFEE field is set to 1h in the Host Behavior Support feature (refer to
	// the Host Behavior Support section in the NVMe Base Specification).
	// The supported LBA formats are indicated in bytes 128 to 383 in this data structure. The
	// LBA Format fields with an index beyond the value set in this field are invalid and not
	// supported. LBA Formats that are valid, but not currently available may be indicated by
	// setting the LBA Data Size for that LBA Format to 0h.
	// The metadata may be either transferred as part of the LBA (creating an extended LBA
	// which is a larger LBA size that is exposed to the application) or may be transferred as a
	// separate contiguous buffer of data. The metadata shall not be split between the LBA and
	// a separate metadata buffer.
	// It is recommended that software and controllers transition to an LBA size that is 4 KiB or
	// larger for ECC efficiency at the controller. If providing metadata, it is recommended that
	// at least 8 bytes are provided per logical block to enable use with end-to-end data
	// protection, refer to section 5.8.3.
	Nlbaf uint8 // Number of LBA Formats
	// Formatted LBA Size (FLBAS): This field indicates the LBA data size & metadata size
	// combination that the namespace has been formatted with (refer to section 4.1.2).
	// Bits 7 is reserved.
	// Bits 6:5 indicate the most significant 2 bits of the Format Index of the supported LBA
	// Format indicated in this data structure that was used to format the namespace. If the
	// NLBAF field is less than or equal to 16, then the host should ignore these bits.
	// Bit 4 if set to ‘1’ indicates that the metadata is transferred at the end of the data LBA,
	// creating an extended data LBA. Bit 4 if cleared to ‘0’ indicates that all of the metadata for
	// a command is transferred as a separate contiguous buffer of data. Bit 4 is not applicable
	// when there is no metadata.
	// Bits 3:0 indicate the least significant 4 bits of the Format Index of the supported LBA
	// Format indicated in this data structure that was used to format the namespace.
	Flbas uint8 // Formatted LBA Size
	// Metadata Capabilities (MC): This field indicates the capabilities for metadata.
	// Bits 7:2 are reserved.
	// Bit 1 if set to ‘1’ indicates the namespace supports the metadata being transferred as part
	// of a separate buffer that is specified in the Metadata Pointer. Bit 1 if cleared to ‘0’ indicates
	// that the namespace does not support the metadata being transferred as part of a separate
	// buffer.
	// Bit 0 if set to ‘1’ indicates that the namespace supports the metadata being transferred as
	// part of an extended data LBA. Bit 0 if cleared to ‘0’ indicates that the namespace does
	// not support the metadata being transferred as part of an extended data LBA.
	Mc uint8 // Metadata Capabilities
	// End-to-end Data Protection Capabilities (DPC): This field indicates the capabilities for
	// the end-to-end data protection feature. Multiple bits may be set in this field. Refer to
	// section 5.2.
	// Bits Description
	// 7:5 Reserved
	// 4
	// Protection Information In Last Bytes (PIILB): If set to ‘1’ indicates
	// that the namespace supports protection information transferred as the
	// last bytes of metadata. If cleared to ‘0’ indicates that the namespace
	// does not support protection information transferred as the last bytes
	// of metadata.
	// 3
	// Protection Information In First Bytes (PIIFB): If set to ‘1’ indicates
	// that the namespace supports protection information transferred as the
	// first bytes of metadata. If cleared to ‘0’ indicates that the namespace
	// does not support protection information transferred as the first bytes
	// of metadata. For implementations compliant to revision 1.0 or later of
	// the NVM Command Set Specification, this bit shall be cleared to ‘0’.
	// 2
	// Protection Information Type 3 Supported (PIT3S): If set to ‘1’
	// indicates that the namespace supports Protection Information Type 3.
	// If cleared to ‘0’ indicates that the namespace does not support
	// Protection Information Type 3.
	// 1
	// Protection Information Type 2 Supported (PIT2S): If set to ‘1’
	// indicates that the namespace supports Protection Information Type 2.
	// If cleared to ‘0’ indicates that the namespace does not support
	// Protection Information Type 2.
	// 0
	// Protection Information Type 1 Supported (PIT1S): If set to ‘1’
	// indicates that the namespace supports Protection Information Type 1.
	// If cleared to ‘0’ indicates that the namespace does not support
	// Protection Information Type 1.
	Dpc uint8 // End-to-end Data Protection Capabilities
	// End-to-end Data Protection Type Settings (DPS): This field indicates the protection
	// information Type settings for the end-to-end data protection feature. Refer to section 5.2.
	// Bits Description
	// 7:4 Reserved
	// 3
	// Protection Information Position (PIP): This bit indicates that the
	// protection information, if enabled, is transferred as the first bytes of
	// metadata. Bit 3 if cleared to ‘0’ indicates that the protection information, if
	// enabled, is transferred as the last bytes of metadata. For implementations
	// compliant to version 1.0 or later of the NVM Command Set Specification,
	// this bit shall be cleared to ‘0’.
	// 2:0
	// Protection Information Type (PIT): This field indicates whether protection
	// information is enabled and the type of protection information enabled. The
	// values for this field have the following meanings:
	// Value Definition
	// 000b Protection information is not enabled
	// 001b Protection information is enabled, Type 1
	// 010b Protection information is enabled, Type 2
	// 011b Protection information is enabled, Type 3
	// 100b to 111b Reserved
	Dps uint8 // End-to-end Data Protection Type Settings
	// Namespace Multi-path I/O and Namespace Sharing Capabilities (NMIC): This field is
	// as defined in the I/O Command Set Independent Identify Namespace data structure (refer
	// to the I/O Command Set Independent Identify Namespace data structure section in the
	// NVMe Base Specification).
	Nmic uint8 // Namespace Multi-path I/O and Namespace Sharing Capabilities
	// Reservation Capabilities (RESCAP): This field is as defined in the I/O Command Set
	// Independent Identify Namespace data structure (refer to the I/O Command Set
	// Independent Identify Namespace data structure section in the NVMe Base Specification).
	Rescap uint8 // Reservation Capabilities
	// Format Progress Indicator (FPI): This field is as defined in the I/O Command Set
	// Independent Identify Namespace data structure (refer to the I/O Command Set
	// Independent Identify Namespace data structure section in the NVMe Base Specification)
	Fpi uint8 // Format Progress Indicator
	// Deallocate Logical Block Features (DLFEAT): This field indicates information about
	// features that affect deallocating logical blocks for this namespace.
	// Bits 7:5 are reserved.
	// Bit 4 if set to ‘1’ indicates that the Guard field for deallocated logical blocks that contain
	// protection information is set to the CRC for the value read from the deallocated logical
	// block and its metadata (excluding protection information). If cleared to ‘0’ indicates that
	// each byte in the Guard field for the deallocated logical blocks that contain protection
	// information is set to FFh.
	// Bit 3 if set to ‘1’ indicates that the controller supports the Deallocate bit in the Write Zeroes
	// command for this namespace. If cleared to ‘0’ indicates that the controller does not
	// support the Deallocate bit in the Write Zeroes command for this namespace. This bit shall
	// be set to the same value for all namespaces in the NVM subsystem.
	// Bits 2:0 indicate deallocated logical block read behavior. For a logical block that is
	// deallocated, this field indicates the values read from that deallocated logical block and its
	// metadata (excluding protection information). The values for this field have the following
	// meanings:
	// Value Definition
	// 000b The read behavior is not reported
	// 001b A deallocated logical block returns all bytes cleared to 0h
	// 010b A deallocated logical block returns all bytes set to FFh
	// 011b to 111b Reserved
	Dlfeat uint8 // Deallocate Logical Block Features
	// Namespace Atomic Write Unit Normal (NAWUN): This field indicates the namespace
	// specific size of the write operation guaranteed to be written atomically to the NVM during
	// normal operation. If the NSABP bit is cleared to ‘0’, then this field is reserved.
	// A value of 0h indicates that the size for this namespace is the same size as that reported
	// in the AWUN field of the Identify Controller data structure. All other values specify a size
	// in terms of logical blocks using the same encoding as the AWUN field. Refer to section
	// 2.1.4.
	Nawun uint16 // Namespace Atomic Write Unit Normal
	// Namespace Atomic Write Unit Power Fail (NAWUPF): This field indicates the
	// namespace specific size of the write operation guaranteed to be written atomically to the
	// NVM during a power fail or error condition. If the NSABP bit is cleared to ‘0’, then this field
	// is reserved.
	// A value of 0h indicates that the size for this namespace is the same size as that reported
	// in the AWUPF field of the Identify Controller data structure. All other values specify a size
	// in terms of logical blocks using the same encoding as the AWUPF field. Refer to section
	// 2.1.4
	Nawupf uint16 // Namespace Atomic Write Unit Power Fail
	// Namespace Atomic Compare & Write Unit (NACWU): This field indicates the
	// namespace specific size of the write operation guaranteed to be written atomically to the
	// NVM for a Compare and Write fused command. If the NSABP bit is cleared to ‘0’, then
	// this field is reserved.
	// A value of 0h indicates that the size for this namespace is the same size as that reported
	// in the ACWU field of the Identify Controller data structure. All other values specify a size
	// in terms of logical blocks using the same encoding as the ACWU field. Refer to section
	// 2.1.4
	Nacwu uint16 // Namespace Atomic Compare & Write Unit
	// Namespace Atomic Boundary Size Normal (NABSN): This field indicates the atomic
	// boundary size for this namespace for the NAWUN value. This field is specified in logical
	// blocks. Writes to this namespace that cross atomic boundaries are not guaranteed to be
	// atomic to the NVM with respect to other read or write commands.
	// A value of 0h indicates that there are no atomic boundaries for normal write operations.
	// All other values specify a size in terms of logical blocks using the same encoding as the
	// AWUN field. Refer to section 2.1.4.
	// Refer to section 5.8.2 for how this field is utilized.
	Nabsn uint16 // Namespace Atomic Boundary Size Normal
	// Namespace Atomic Boundary Offset (NABO): This field indicates the LBA on this
	// namespace where the first atomic boundary starts.
	// If the NABSN and NABSPF fields are cleared to 0h, then the NABO field shall be cleared
	// to 0h. NABO shall be less than or equal to NABSN and NABSPF. Refer to section 2.1.4.
	// Refer to section 5.8.2 for how this field is utilized
	Nabo uint16 // Namespace Atomic Boundary Offset
	// Namespace Atomic Boundary Size Power Fail (NABSPF): This field indicates the
	// atomic boundary size for this namespace specific to the Namespace Atomic Write Unit
	// Power Fail value. This field is specified in logical blocks. Writes to this namespace that
	// cross atomic boundaries are not guaranteed to be atomic with respect to other read or
	// write commands and there is no guarantee of data returned on subsequent reads of the
	// associated logical blocks.
	// A value of 0h indicates that there are no atomic boundaries for power fail or error
	// conditions. All other values specify a size in terms of logical blocks using the same
	// encoding as the AWUPF field. Refer to section 2.1.4
	Nabspf uint16 // Namespace Atomic Boundary Size Power Fail
	// Namespace Optimal I/O Boundary (NOIOB): This field indicates the optimal I/O
	// boundary for this namespace. This field is specified in logical blocks. The host should
	// construct Read and Write commands that do not cross the I/O boundary to achieve
	// optimal performance. A value of 0h indicates that no optimal I/O boundary is reported.
	// Refer to section 5.8.2 for how this field is utilized to improve performance and endurance.
	Noiob uint16 // Namespace Optimal I/O Boundary
	// NVM Capacity (NVMCAP): This field indicates the total size of the NVM allocated to this
	// namespace. The value is in bytes. This field shall be supported if the Namespace
	// Management capability (refer to section 5.3) is supported.
	// Note: This field may not correspond to the logical block size multiplied by the Namespace
	// Size field. Due to thin provisioning or other settings (e.g., endurance), this field may be
	// larger or smaller than the product of the logical block size and the Namespace Size
	// reported.
	// If the controller supports Asymmetric Namespace Access Reporting (refer to the CMIC
	// field), and the relationship between the controller and the namespace is in the ANA
	// Inaccessible state (refer to the ANA Inaccessible state section in the NVMe Base
	// Specification) or the ANA Persistent Loss state (refer to the ANA Persistent Loss state
	// section in the NVMe Base Specification), then this field shall be cleared to 0h.
	Nvmcap Uint128 // NVM Capacity
	// Namespace Preferred Write Granularity (NPWG): This field indicates the smallest
	// recommended write granularity in logical blocks for this namespace. This is a 0’s based
	// value. If the OPTPERF bit is cleared to ‘0’, then this field is reserved.
	// The size indicated should be less than or equal to Maximum Data Transfer Size (MDTS)
	// that is specified in units of minimum memory page size. The value of this field may change
	// if the namespace is reformatted. The size should be a multiple of Namespace Preferred
	// Write Alignment (NPWA).
	// Refer to section 5.8.2 for how this field is utilized to improve performance and endurance.
	Npwg uint16 // Namespace Preferred Write Granularity
	// Namespace Preferred Write Alignment (NPWA): This field indicates the recommended
	// write alignment in logical blocks for this namespace. This is a 0’s based value. If the
	// OPTPERF bit is cleared to ‘0’, then this field is reserved.
	// The value of this field may change if the namespace is reformatted.
	// Refer to section 5.8.2 for how this field is utilized to improve performance and endurance
	Npwa uint16 // Namespace Preferred Write Alignment
	// Namespace Preferred Deallocate Granularity (NPDG): This field indicates the
	// recommended granularity in logical blocks for the Dataset Management command with
	// the Attribute – Deallocate bit set to ‘1’ in Dword 11. This is a 0’s based value. If the
	// OPTPERF bit is cleared to ‘0’, then this field is reserved.
	// The value of this field may change if the namespace is reformatted. The size should be a
	// multiple of Namespace Preferred Deallocate Alignment (NPDA).
	// Refer to section 5.8.2 for how this field is utilized to improve performance and endurance.
	Npdg uint16 // Namespace Preferred Deallocate Granularity
	// Namespace Preferred Deallocate Alignment (NPDA): This field indicates the
	// recommended alignment in logical blocks for the Dataset Management command with the
	// Attribute – Deallocate bit set to ‘1’ in Dword 11. This is a 0’s based value. If the OPTPERF
	// bit is cleared to ‘0’, then this field is reserved.
	// The value of this field may change if the namespace is reformatted.
	// Refer to section 5.8.2 for how this field is utilized to improve performance and endurance.
	Npda uint16 // Namespace Preferred Deallocate Alignment
	// Namespace Optimal Write Size (NOWS): This field indicates the size in logical blocks
	// for optimal write performance for this namespace. This is a 0’s based value. If the
	// OPTPERF bit is cleared to ‘0’, then this field is reserved.
	// The size indicated should be less than or equal to Maximum Data Transfer Size (MDTS)
	// that is specified in units of minimum memory page size. The value of this field may change
	// if the namespace is reformatted. The value of this field should be a multiple of Namespace
	// Preferred Write Granularity (NPWG).
	// If the namespace is associated with an NVM set, NOWS defined for this namespace shall
	// be set to the Optimal Write Size field setting defined in NVM Set Attributes Entry (refer to
	// the Namespace Identification Descriptor in the NVMe Base Specification) for the NVM
	// Set with which this namespace is associated. If NOWS is not supported, the Optimal Write
	// Size field in NVM Sets Attributes Entry (refer to the Namespace Identification Descriptor
	// in the NVMe Base Specification) for the NVM Set with which this namespace is associated
	// should be used by the host for I/O optimization.
	// Refer to section 5.8.2 for how this field is utilized to improve performance and endurance
	Nows uint16 // Namespace Optimal Write Size
	// Maximum Single Source Range Length (MSSRL): This field indicates the maximum
	// number of logical blocks that may be specified in the Number of Logical Block field in each
	// valid Source Range Entries Descriptor of a Copy command (refer to section 3.2.2).
	// If the controller supports the Copy command, then this field shall be set to a non-zero
	// value
	Mssrl uint16 // Maximum Single Source Range Length
	// Maximum Copy Length (MCL): This field indicates the maximum number of logical
	// blocks that may be specified in a Copy command (i.e., the sum of the number of logical
	// blocks specified in all Source Range entries).
	// If the controller supports the Copy command, then this field shall be set to a non-zero
	// value.
	Mcl uint32 // Maximum Copy Length
	// Maximum Source Range Count (MSRC): This field indicates the maximum number of
	// Source Range entries that may be used to specify source data in a Copy command. This
	// is a 0’s based value.
	Msrc uint8 // Maximum Source Range Count
	_    [11]byte
	// ANA Group Identifier (ANAGRPID): This field is as defined in the I/O Command Set
	// Independent Identify Namespace data structure (refer to the I/O Command Set
	// Independent Identify Namespace data structure section in the NVMe Base Specification)
	Anagrpid uint32 // ANA Group Identifier
	_        [3]byte
	// Namespace Attributes (NSATTR): This field is as defined in the I/O Command Set
	// Independent Identify Namespace data structure (refer to the I/O Command Set
	// Independent Identify Namespace data structure section in the NVMe Base Specification).
	Nsattr uint8 // Namespace Attributes
	// NVM Set Identifier (NVMSETID): This field is as defined in the I/O Command Set
	// Independent Identify Namespace data structure (refer to the I/O Command Set
	// Independent Identify Namespace data structure section in the NVMe Base Specification).
	Nvmseid uint16 // NVM Set Identifier
	// Endurance Group Identifier (ENDGID): This field is as defined in the I/O Command Set
	// Independent Identify Namespace data structure (refer to the I/O Command Set
	// Independent Identify Namespace data structure section in the NVMe Base Specification).
	Endgid uint16 // Endurance Group Identifier
	// Namespace Globally Unique Identifier (NGUID): This field contains a 128-bit value that
	// is globally unique and assigned to the namespace when the namespace is created. This
	// field remains fixed throughout the life of the namespace and is preserved across
	// namespace and controller operations (e.g., Controller Level Reset, namespace format,
	// etc.).
	// This field uses the EUI-64 based 16-byte designator format. Bytes 114:112 contain the
	// 24-bit Organizationally Unique Identifier (OUI) value assigned by the IEEE Registration
	// Authority. Bytes 119:115 contain an extension identifier assigned by the corresponding
	// organization. Bytes 111:104 contain the vendor specific extension identifier assigned by
	// the corresponding organization. Refer to the IEEE EUI-64 guidelines for more information.
	// This field is big endian (refer to the Namespace Globally Unique Identifier section in the
	// NVMe Base Specification).
	// The controller shall specify a globally unique namespace identifier in this field, the EUI64
	// field, or a Namespace UUID in the Namespace Identification Descriptor (refer to the
	// Namespace Identification Descriptor figure in the NVMe Base Specification) when the
	// namespace is created. If the controller is not able to provide a globally unique identifier in
	// this field, then this field shall be cleared to 0h. Refer to the Unique Identifier section in the
	// NVMe Base Specification.
	Nguid [16]byte // Namespace Globally Unique Identifier
	// IEEE Extended Unique Identifier (EUI64): This field contains a 64-bit IEEE Extended
	// Unique Identifier (EUI-64) that is globally unique and assigned to the namespace when
	// the namespace is created. This field remains fixed throughout the life of the namespace
	// and is preserved across namespace and controller operations (e.g., Controller Level
	// Reset, namespace format, etc.).
	// The EUI-64 is a concatenation of a 24-bit or 36-bit Organizationally Unique Identifier (OUI
	// or OUI-36) value assigned by the IEEE Registration Authority and an extension identifier
	// assigned by the corresponding organization. Refer to the IEEE EUI-64 guidelines for more
	// information. This field is big endian (refer to the IEEE Extended Unique Identifier section
	// in the NVMe Base Specification).
	// The controller shall specify a globally unique namespace identifier in this field, the NGUID
	// field, or a Namespace UUID in the Namespace Identification Descriptor (refer to the
	// Namespace Identification Descriptor figure in the NVMe Base Specification) when the
	// namespace is created. If the controller is not able to provide a globally unique 64-bit
	// identifier in this field, then this field shall be cleared to 0h. Refer to the Unique Identifier
	// section in the NVMe Base Specification.
	Eui64 [8]byte // IEEE Extended Unique Identifier
	// LBA Format Support (LBAF): This field indicates the LBA format N that is supported
	// by the controller.
	Lbaf [64]NvmeLBAF // LBA Format Support
	// Vendor Specific
	Vs [3712]byte
} // 4096 bytes

func (ns *NvmeIdentNamespace) LbaSize() uint64 {
	return uint64(1) << ns.Lbaf[ns.Flbas&0xf].Ds
}

type NvmeSMARTLog struct {
	CritWarning            uint8  // Critical Warning
	Temperature            uint16 // Composite Temperature
	AvailSpare             uint8  // Available Spare
	SpareThresh            uint8  // Available Spare Threshold
	PercentUsed            uint8  // Percentage Used
	EnduranceCritWarning   uint8  // Endurance Group Critical Warning Summary
	_                      [25]byte
	DataUnitsRead          Uint128   // Data Units Read
	DataUnitsWritten       Uint128   // Data Units Written
	HostReads              Uint128   // Host Read Commands
	HostWrites             Uint128   // Host Write Commands
	CtrlBusyTime           Uint128   // Controller Busy Time
	PowerCycles            Uint128   // Power Cycles
	PowerOnHours           Uint128   // Power On Hours
	UnsafeShutdowns        Uint128   // Unsafe Shutdowns
	MediaErrors            Uint128   // Media and Data Integrity Errors
	NumErrLogEntries       Uint128   // Number of Error Information Log Entries
	WarningTempTime        uint32    // Warning Composite Temperature Time
	CritCompTime           uint32    // Critical Composite Temperature Time
	TempSensor             [8]uint16 // Temperature Sensors
	ThermalTransitionCount [2]uint32 // Thermal Management Transition Count
	ThermalManagementTime  [2]uint32 // Total Time For Thermal Management
	_                      [280]byte
} // 512 bytes

const (
	nvmeAdminDeleteSq      = 0x00
	nvmeAdminCreateSq      = 0x01
	nvmeAdminGetLogPage    = 0x02
	nvmeAdminDeleteCq      = 0x04
	nvmeAdminCreateCq      = 0x05
	nvmeAdminIdentify      = 0x06
	nvmeAdminAbortCmd      = 0x08
	nvmeAdminSetFeatures   = 0x09
	nvmeAdminGetFeatures   = 0x0a
	nvmeAdminAsyncEvent    = 0x0c
	nvmeAdminNsMgmt        = 0x0d
	nvmeAdminActivateFw    = 0x10
	nvmeAdminDownloadFw    = 0x11
	nvmeAdminDevSelfTest   = 0x14
	nvmeAdminNsAttach      = 0x15
	nvmeAdminKeepAlive     = 0x18
	nvmeAdminDirectiveSend = 0x19
	nvmeAdminDirectiveRecv = 0x1a
	nvmeAdminVirtualMgmt   = 0x1c
	nvmeAdminNvmeMiSend    = 0x1d
	nvmeAdminNvmeMiRecv    = 0x1e
	nvmeAdminDbbuf         = 0x7C
	nvmeAdminFormatNvm     = 0x80
	nvmeAdminSecuritySend  = 0x81
	nvmeAdminSecurityRecv  = 0x82
	nvmeAdminSanitizeNvm   = 0x84
	nvmeAdminGetLbaStatus  = 0x86
	nvmeAdminVendorStart   = 0xC0
)

const (
	nvmeLogSupportedPages    = 0x0
	nvmeLogErrorInformation  = 0x1
	nvmeLogSmartInformation  = 0x2
	nvmeLogFirmwareInfo      = 0x3
	nvmeLogChangedNamespace  = 0x4
	nvmeLogCommandsSupported = 0x5
	nvmeLogDeviceSelftest    = 0x6
)

func (d *NVMeDevice) Type() string {
	return "nvme"
}

func (d *NVMeDevice) ReadGenericAttributes() (*GenericAttributes, error) {
	log, err := d.ReadSMART()
	if err != nil {
		return nil, err
	}

	a := GenericAttributes{}
	a.Temperature = uint64(log.Temperature - 273) // NVMe reports the temperature in Kelvins, normalize it to Celsius
	a.Read = log.DataUnitsRead.Val[0]
	a.Written = log.DataUnitsWritten.Val[0]
	a.PowerOnHours = log.PowerOnHours.Val[0]
	a.PowerCycles = log.PowerCycles.Val[0]
	return &a, nil
}

func (d *NVMeDevice) Identify() (*NvmeIdentController, []NvmeIdentNamespace, error) {
	controller, err := d.readControllerIdentifyData()
	if err != nil {
		return nil, nil, err
	}

	var ns []NvmeIdentNamespace
	// QEMU has 256 namespaces for some reason, TODO: clarify
	for i := 0; i < int(controller.Nn); i++ {
		namespace, err := d.readNamespaceIdentifyData(i + 1)
		if err != nil {
			return nil, nil, err
		}
		if namespace.Nsze == 0 {
			continue
		}

		ns = append(ns, *namespace)
	}

	return controller, ns, nil
}
