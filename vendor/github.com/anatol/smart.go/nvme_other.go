// go:build !linux && !darwin
//go:build !linux && !darwin
// +build !linux,!darwin

package smart

type NVMeDevice struct{}

func OpenNVMe(name string) (*NVMeDevice, error) {
	return nil, ErrOSUnsupported
}

func (d *NVMeDevice) Close() error {
	return ErrOSUnsupported
}

func (d *NVMeDevice) ReadSMART() (*NvmeSMARTLog, error) {
	return nil, ErrOSUnsupported
}

func (d *NVMeDevice) readControllerIdentifyData() (*NvmeIdentController, error) {
	return nil, ErrOSUnsupported
}

func (d *NVMeDevice) readNamespaceIdentifyData(nsid int) (*NvmeIdentNamespace, error) {
	return nil, ErrOSUnsupported
}
