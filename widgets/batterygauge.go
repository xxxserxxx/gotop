package widgets

import (
	"fmt"
	"time"

	"github.com/xxxserxxx/gotop/v4/devices"
)

// FIXME gauge isn't updating.

type BatteryGauge struct {
	*Gauge
	d []*devices.Batteries
}

// NewBatteryGauge creates a gauge widget representing the percentage of how
// full power a battery is.
func NewBatteryGauge() *BatteryGauge {
	self := &BatteryGauge{
		Gauge: NewGauge(),
		d:     make([]*devices.Batteries, 0),
	}
	self.Title = tr.Value("widget.label.gauge")

	return self
}

func (g *BatteryGauge) Attach(b *devices.Batteries) {
	g.d = append(g.d, b)
}

// Only report battery errors once.
var errLogged = false

func (b *BatteryGauge) Update() {
	mx := 0.0
	cu := 0.0
	// default: discharging
	formatString := "%d%% ⚡%s"
	var rate time.Duration
	for _, bats := range b.d {
		for _, bat := range bats.Data {
			if bat.Full == 0.0 {
				continue
			}
			mx += bat.Full
			cu += bat.Current
			if bat.Charging {
				fullTime := (mx - cu) / bat.ChargeRate
				rate, _ = time.ParseDuration(fmt.Sprintf("%fh", fullTime))
				formatString = "%d%% 🔌%s"
			} else {
				runTime := cu / bat.ChargeRate
				rate, _ = time.ParseDuration(fmt.Sprintf("%fh", runTime))
			}
		}
	}
	b.Percent = int((cu / mx) * 100.0)
	b.Label = fmt.Sprintf(formatString, b.Percent, rate.Truncate(time.Minute))
}
