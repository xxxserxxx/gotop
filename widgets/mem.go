package widgets

import (
	"fmt"
	"time"

	"github.com/xxxserxxx/gotop/v4/devices"
	"github.com/xxxserxxx/gotop/v4/utils"
)

// TODO Colors are wrong for #mem > 2
type MemWidget struct {
	*LineGraph
	updateInterval time.Duration
	mems           devices.Memory
}

func NewMemWidget(updateInterval time.Duration, horizontalScale int) *MemWidget {
	widg := &MemWidget{
		LineGraph:      NewLineGraph(),
		updateInterval: updateInterval,
		mems:           devices.NewMemory(),
	}
	widg.Title = tr.Value("widget.label.mem")
	widg.HorizontalScale = horizontalScale

	return widg
}

func (mw *MemWidget) Attach(m devices.Memory) {
	for name, me := range m {
		mw.mems[name] = me
		mw.Data[name] = []float64{0}
	}
}

func (widg *MemWidget) Update() {
	for name, mem := range widg.mems {
		if mem.Total > 0 {
			widg.Data[name] = append(widg.Data[name], mem.UsedPercent)
			memoryTotalBytes, memoryTotalMagnitude := utils.ConvertBytes(mem.Total)
			memoryUsedBytes, memoryUsedMagnitude := utils.ConvertBytes(mem.Used)
			widg.Labels[name] = fmt.Sprintf("%3.0f%% %5.1f%s/%.0f%s",
				mem.UsedPercent,
				memoryUsedBytes+0.5,
				memoryUsedMagnitude,
				memoryTotalBytes+0.5,
				memoryTotalMagnitude,
			)
		}
	}
}

func (mem *MemWidget) Scale(i int) {
	mem.LineGraph.HorizontalScale = i
}
