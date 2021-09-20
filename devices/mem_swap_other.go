//go:build !freebsd
// +build !freebsd

package devices

import "github.com/shirou/gopsutil/v3/mem"

// UpdateSwap adds a "Swap" memory entry.
func (m Memory) UpdateSwap() error {
	memory, err := mem.SwapMemory()
	if err != nil {
		return err
	}
	me := m["Swap"]
	me.Total = memory.Total
	me.Used = memory.Used
	me.UsedPercent = memory.UsedPercent
	return nil
}
