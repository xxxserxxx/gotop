package devices

import (
	psMem "github.com/shirou/gopsutil/mem"
)

func init() {
	mf := func(mems map[string]MemoryInfo) map[string]error {
		mainMemory, err := psMem.VirtualMemory()
		if err != nil {
			return map[string]error{"Main": err}
		}
		pressure, err := psMem.Pressure()
		if err != nil {
			return map[string]error{"Pressure": err}
		}
		mems["Main"] = MemoryInfo{
			Total:       mainMemory.Total,
			Used:        mainMemory.Used,
			UsedPercent: mainMemory.UsedPercent,
			Pressure:    pressure.SomeAvg10,
		}
		return nil
	}
	RegisterMem(mf)
}
