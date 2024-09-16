// go:build !linux
//go:build !linux
// +build !linux

package smart

func OpenSata(name string) (*SataDevice, error) {
	return nil, ErrOSUnsupported
}

func (d *SataDevice) Close() error {
	return ErrOSUnsupported
}

func (d *SataDevice) Identify() (*AtaIdentifyDevice, error) {
	return nil, ErrOSUnsupported
}

func (d *SataDevice) readSMARTLog(logPage uint8) ([]byte, error) {
	return nil, ErrOSUnsupported
}

func (d *SataDevice) readSMARTData() (*AtaSmartPageRaw, error) {
	return nil, ErrOSUnsupported
}

func (d *SataDevice) ReadSMARTLogDirectory() (*AtaSmartLogDirectory, error) {
	return nil, ErrOSUnsupported
}

func (d *SataDevice) ReadSMARTErrorLogSummary() (*AtaSmartErrorLogSummary, error) {
	return nil, ErrOSUnsupported
}

func (d *SataDevice) ReadSMARTSelfTestLog() (*AtaSmartSelfTestLog, error) {
	return nil, ErrOSUnsupported
}

func (d *SataDevice) readSMARTThresholds() (*AtaSmartThresholdsPageRaw, error) {
	return nil, ErrOSUnsupported
}
