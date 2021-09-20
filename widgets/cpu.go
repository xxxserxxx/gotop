package widgets

import (
	"fmt"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/xxxserxxx/gotop/v4/devices"

	"github.com/gizak/termui/v3"
)

const AVRG = "AVRG"

// TODO Maybe group CPUs in columns if space permits
// TODO add CPU freq
type CPUWidget struct {
	*LineGraph
	cpuCount        int
	ShowAverageLoad bool
	ShowPerCPULoad  bool
	updateInterval  time.Duration
	cpuLoads        map[string]float64
	average         ewma.MovingAverage
	cpus            []*devices.CPUs
	keys            [][]string
}

func NewCPUWidget(updateInterval time.Duration, horizontalScale int, showAverageLoad bool, showPerCPULoad bool) *CPUWidget {
	self := &CPUWidget{
		LineGraph:       NewLineGraph(),
		cpuCount:        len(cpuLabels),
		updateInterval:  updateInterval,
		ShowAverageLoad: showAverageLoad,
		ShowPerCPULoad:  showPerCPULoad,
		cpuLoads:        make(map[string]float64),
		average:         ewma.NewMovingAverage(),
	}
	self.LabelStyles[AVRG] = termui.ModifierBold
	self.Title = tr.Value("widget.label.cpu")
	self.HorizontalScale = horizontalScale

	if !(self.ShowAverageLoad || self.ShowPerCPULoad) {
		if self.cpuCount <= 8 {
			self.ShowPerCPULoad = true
		} else {
			self.ShowAverageLoad = true
		}
	}

	if self.ShowAverageLoad {
		self.Data[AVRG] = []float64{0}
	}

	return self
}

var cpuLabels []string

func (cpu *CPUWidget) Attach(cs *devices.CPUs) {
	cpu.cpus = append(cpu.cpus, cs)

	cpu.cpuCount = 0
	cpu.keys = make([][]string, len(cpu.cpus))
	if cpu.ShowPerCPULoad {
		cpu.Data = make(map[string][]float64)
		for _, c := range cpu.cpus {
			cpu.cpuCount += len(c.Data)
		}
		formatString := "%s%1d"
		if cpu.cpuCount > 10 {
			formatString = "%s%02d"
		}
		for i, ci := range cpu.cpus {
			cpu.keys[i] = make([]string, len(ci.Data))
			for j, _ := range ci.Data {
				key := fmt.Sprintf(formatString, ci.Name, j)
				cpu.Data[key] = make([]float64, 0)
				cpu.keys[i][j] = key
			}
		}
	}
}

func (cpu *CPUWidget) Scale(i int) {
	cpu.LineGraph.HorizontalScale = i
}

func (cpu *CPUWidget) Update() {
	cpus := make(map[string]int)
	// AVG = ((AVG*i)+n)/(i+1)
	var sum float64
	if cpu.ShowPerCPULoad {
		for i, c := range cpu.cpus {
			sum += c.Average
			for j, percent := range c.Data {
				key := cpu.keys[i][j]
				cpu.Data[key] = append(cpu.Data[key], float64(percent))
				cpu.Labels[key] = fmt.Sprintf("%3d%%", int(percent))
				cpu.cpuLoads[key] = float64(percent)
			}
		}
	} else {
		for _, c := range cpu.cpus {
			sum += c.Average
		}
	}
	if cpu.ShowAverageLoad {
		cpu.average.Add(sum / float64(len(cpus)))
		avg := cpu.average.Value()
		cpu.Data[AVRG] = append(cpu.Data[AVRG], avg)
		cpu.Labels[AVRG] = fmt.Sprintf("%3.0f%%", avg)
		cpu.cpuLoads[AVRG] = avg
	}
}
