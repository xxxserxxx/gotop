//go:build !linux
// +build !linux

package devices

import "github.com/shirou/gopsutil/v3/cpu"

func CpuCount() (int, error) {
	return cpu.Counts(false)
}
