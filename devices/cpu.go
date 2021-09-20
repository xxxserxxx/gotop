package devices

import (
	"log"

	"github.com/VictoriaMetrics/metrics"
	"github.com/shirou/gopsutil/v3/cpu"
)

type CPUs struct {
	Name    string
	Data    []float64
	Average float64
	Logical bool
}

func NewCPUs(name string, logical bool) CPUs {
	return CPUs{
		Name:    name,
		Data:    make([]float64, 0),
		Logical: logical,
	}
}

func LocalCPUs(logical bool) CPUs {
	l := NewCPUs("CPU", logical)
	vals, err := cpu.Percent(0, logical)
	if err != nil {
		log.Printf("couldn't get local CPU information: %s", err)
		return l
	}
	l.Data = vals
	return l
}

// CPUPercent calculates the percentage of cpu used either per CPU or combined.
// Returns one value per cpu, or a single value if percpu is set to false.
func (c *CPUs) Update() error {
	vals, err := cpu.Percent(0, c.Logical)
	if err != nil {
		return err
	}
	c.Data = vals
	c.Average = 0
	for _, v := range vals {
		c.Average += v
	}
	c.Average = c.Average / float64(len(vals))
	return nil
}

func (c *CPUs) EnableMetrics(s *metrics.Set) {
	s.NewGauge(makeName("cpu", "avg"), func() float64 {
		return c.Average
	})
	for i := range c.Data {
		idx := i
		s.NewGauge(makeName("cpu", i), func() float64 {
			return c.Data[idx]
		})
	}
}
