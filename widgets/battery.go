package widgets

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/xxxserxxx/gotop/v4/devices"
)

type BatteryWidget struct {
	*LineGraph
	updateInterval time.Duration
	d              []*devices.Batteries
}

func NewBatteryWidget(horizontalScale int) *BatteryWidget {
	bw := &BatteryWidget{
		LineGraph:      NewLineGraph(),
		updateInterval: time.Minute,
		d:              make([]*devices.Batteries, 0),
	}
	bw.Title = tr.Value("widget.label.battery")
	bw.HorizontalScale = horizontalScale

	return bw
}

func (b *BatteryWidget) Attach(n *devices.Batteries) {
	b.d = append(b.d, n)
	// intentional duplicate
	// adds 2 datapoints to the graph, otherwise the dot is difficult to see
	b.Update()
	b.Update()

}

func makeID(i int) string {
	return tr.Value("widget.label.batt") + strconv.Itoa(i)
}

func (b *BatteryWidget) Scale(i int) {
	b.LineGraph.HorizontalScale = i
}

func (b *BatteryWidget) Update() {
	for _, bats := range b.d {
		for i, battery := range bats.Data {
			if battery.Full == 0.0 {
				continue
			}
			id := makeID(i)
			perc := battery.Current / battery.Full
			percentFull := math.Abs(perc) * 100.0
			b.Data[id] = append(b.Data[id], percentFull)
			b.Labels[id] = fmt.Sprintf("%3.0f%% %.0f/%.0f", percentFull, math.Abs(battery.Current), math.Abs(battery.Full))
		}
	}
}
