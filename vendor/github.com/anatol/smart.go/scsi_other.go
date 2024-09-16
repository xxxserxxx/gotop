// go:build !linux
//go:build !linux
// +build !linux

package smart

func OpenScsi(name string) (*ScsiDevice, error) {
	return nil, ErrOSUnsupported
}

func (d *ScsiDevice) Close() error {
	return ErrOSUnsupported
}

func (d *ScsiDevice) Capacity() (uint64, error) {
	return 0, ErrOSUnsupported
}

func (d *ScsiDevice) Inquiry() (*ScsiInquiry, error) {
	return nil, ErrOSUnsupported
}

func (d *ScsiDevice) SerialNumber() (string, error) {
	return "", ErrOSUnsupported
}
