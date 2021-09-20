package devices

import (
	"github.com/VictoriaMetrics/metrics"
	"github.com/shirou/gopsutil/v3/mem"
)

// TODO Swap memory values for remote devices is bogus
type MemoryInfo struct {
	Total       uint64
	Used        uint64
	UsedPercent float64
}

type Memory map[string]*MemoryInfo

func NewMemory() Memory {
	return make(Memory)
}

func LocalMemory() Memory {
	m := NewMemory()
	m["Main"] = &MemoryInfo{}
	m["Swap"] = &MemoryInfo{}
	m.Update()
	return m
}

func (m Memory) Update() error {
	mainMemory, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	me := m["Main"]
	me.Total = mainMemory.Total
	me.Used = mainMemory.Used
	me.UsedPercent = mainMemory.UsedPercent
	return m.UpdateSwap()
}

func (m Memory) EnableMetrics(s *metrics.Set) {
	m.Update()
	for k, v := range m {
		vp := v
		s.NewGauge(makeName("memory", "total", k), func() float64 {
			return float64(vp.Total)
		})
		s.NewGauge(makeName("memory", "used", k), func() float64 {
			return float64(vp.Used)
		})
	}
}
