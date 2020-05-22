package widgets

import (
	"fmt"
	"log"

	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/distatus/battery"

	"github.com/xxxserxxx/gotop/v4/termui"
)

type BatteryGauge struct {
	*termui.Gauge
}

func NewBatteryGauge() *BatteryGauge {
	self := &BatteryGauge{Gauge: termui.NewGauge()}
	self.Title = " Power Level "

	self.update()

	go func() {
		for range time.NewTicker(time.Second).C {
			self.Lock()
			self.update()
			self.Unlock()
		}
	}()

	return self
}

func (b *BatteryGauge) EnableMetric() {
	metrics.NewGauge(makeName("battery", "total"), func() float64 {
		return float64(b.Percent)
	})
}

func (b *BatteryGauge) update() {
	bats, err := battery.GetAll()
	if err != nil {
		log.Printf("error setting up batteries: %v", err)
		return
	}
	mx := 0.0
	cu := 0.0
	charging := "%d%% ⚡%s"
	rate := 0.0
	for _, bat := range bats {
		mx += bat.Full
		cu += bat.Current
		if rate < bat.ChargeRate {
			rate = bat.ChargeRate
		}
		if bat.State == battery.Charging {
			charging = "%d%% 🔌%s"
		}
	}
	tn := (mx - cu) / rate
	d, _ := time.ParseDuration(fmt.Sprintf("%fh", tn))
	b.Percent = int((cu / mx) * 100.0)
	b.Label = fmt.Sprintf(charging, b.Percent, d.Truncate(time.Minute))
}
