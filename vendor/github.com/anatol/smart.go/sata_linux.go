package smart

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"golang.org/x/sys/unix"
)

func OpenSata(name string) (*SataDevice, error) {
	fd, err := unix.Open(name, unix.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	i, err := scsiInquiry(fd)
	if err != nil {
		unix.Close(fd)
		return nil, err
	}

	if !bytes.Equal(i.VendorIdent[:], []byte(_SATA_IDENT)) {
		unix.Close(fd)
		return nil, fmt.Errorf("it is not a SATA device")
	}

	dev := SataDevice{
		fd,
		nil,
		0,
	}

	id, err := dev.Identify()
	if err != nil {
		return nil, err
	}
	mapping, bug, err := findAttributesMapping(id.ModelNumber(), id.FirmwareRevision())
	if err != nil {
		return nil, err
	}
	dev.attributeMapping = mapping
	dev.firmwareBug = bug

	return &dev, nil
}

func (d *SataDevice) Close() error {
	return unix.Close(d.fd)
}

func (d *SataDevice) Identify() (*AtaIdentifyDevice, error) {
	var resp AtaIdentifyDevice

	respBuf := make([]byte, 512)

	cdb := cdb16{_SCSI_ATA_PASSTHRU_16}
	cdb[1] = 0x08                  // ATA protocol (4 << 1, PIO data-in)
	cdb[2] = 0x0e                  // BYT_BLOK = 1, T_LENGTH = 2, T_DIR = 1
	cdb[14] = _ATA_IDENTIFY_DEVICE // command

	if err := scsiSendCdb(d.fd, cdb[:], respBuf); err != nil {
		return &resp, fmt.Errorf("sendCDB ATA IDENTIFY: %v", err)
	}

	if err := binary.Read(bytes.NewBuffer(respBuf), binary.LittleEndian, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (d *SataDevice) readSMARTLog(logPage uint8) ([]byte, error) {
	respBuf := make([]byte, 512)

	cdb := cdb16{_SCSI_ATA_PASSTHRU_16}
	cdb[1] = 0x08            // ATA protocol (4 << 1, PIO data-in)
	cdb[2] = 0x0e            // BYT_BLOK = 1, T_LENGTH = 2, T_DIR = 1
	cdb[4] = _SMART_READ_LOG // feature LSB
	cdb[6] = 0x01            // sector count
	cdb[8] = logPage         // SMART log page number
	cdb[10] = 0x4f           // low lba_mid
	cdb[12] = 0xc2           // low lba_high
	cdb[14] = _ATA_SMART     // command

	if err := scsiSendCdb(d.fd, cdb[:], respBuf); err != nil {
		return nil, fmt.Errorf("scsiSendCdb SMART READ LOG: %v", err)
	}

	return respBuf, nil
}

func (d *SataDevice) readSMARTData() (*AtaSmartPageRaw, error) {
	cdb := cdb16{_SCSI_ATA_PASSTHRU_16}
	cdb[1] = 0x08             // ATA protocol (4 << 1, PIO data-in)
	cdb[2] = 0x0e             // BYT_BLOK = 1, T_LENGTH = 2, T_DIR = 1
	cdb[4] = _SMART_READ_DATA // feature LSB
	cdb[10] = 0x4f            // low lba_mid
	cdb[12] = 0xc2            // low lba_high
	cdb[14] = _ATA_SMART      // command

	respBuf := make([]byte, 512)

	if err := scsiSendCdb(d.fd, cdb[:], respBuf); err != nil {
		return nil, fmt.Errorf("scsiSendCdb SMART READ DATA: %v", err)
	}

	page := AtaSmartPageRaw{}
	if err := binary.Read(bytes.NewBuffer(respBuf[:362]), binary.LittleEndian, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

func (d *SataDevice) ReadSMARTLogDirectory() (*AtaSmartLogDirectory, error) {
	buf, err := d.readSMARTLog(0x00)
	if err != nil {
		return nil, err
	}

	dir := AtaSmartLogDirectory{}
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &dir); err != nil {
		return nil, err
	}

	return &dir, nil
}

func (d *SataDevice) ReadSMARTErrorLogSummary() (*AtaSmartErrorLogSummary, error) {
	buf, err := d.readSMARTLog(0x01)
	if err != nil {
		return nil, err
	}

	summary := AtaSmartErrorLogSummary{}
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

func (d *SataDevice) ReadSMARTSelfTestLog() (*AtaSmartSelfTestLog, error) {
	buf, err := d.readSMARTLog(0x06)
	if err != nil {
		return nil, err
	}

	log := AtaSmartSelfTestLog{}
	if err := binary.Read(bytes.NewBuffer(buf), binary.LittleEndian, &log); err != nil {
		return nil, err
	}

	return &log, nil
}

func (d *SataDevice) readSMARTThresholds() (*AtaSmartThresholdsPageRaw, error) {
	cdb := cdb16{_SCSI_ATA_PASSTHRU_16}
	cdb[1] = 0x08                   // ATA protocol (4 << 1, PIO data-in)
	cdb[2] = 0x0e                   // BYT_BLOK = 1, T_LENGTH = 2, T_DIR = 1
	cdb[4] = _SMART_READ_THRESHOLDS // feature LSB
	cdb[8] = 0x1                    // low lba_low
	cdb[10] = 0x4f                  // low lba_mid
	cdb[12] = 0xc2                  // low lba_high
	cdb[14] = _ATA_SMART            // command

	respBuf := make([]byte, 512)

	if err := scsiSendCdb(d.fd, cdb[:], respBuf); err != nil {
		return nil, fmt.Errorf("scsiSendCdb SMART READ THRESHOLD: %v", err)
	}

	if !checksum(respBuf) {
		return nil, fmt.Errorf("invalid checksum for SMART THRESHOLD data")
	}

	page := AtaSmartThresholdsPageRaw{}
	if err := binary.Read(bytes.NewBuffer(respBuf[:]), binary.LittleEndian, &page); err != nil {
		return nil, err
	}

	return &page, nil
}

func checksum(data []byte) bool {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return sum == 0
}
