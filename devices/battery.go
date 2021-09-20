package devices

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
	"github.com/distatus/battery"
)

type BatteryInfo struct {
	Full       float64
	Current    float64
	Charging   bool
	ChargeRate float64
	Design     float64
}

type Batteries struct {
	Data []BatteryInfo
	// PercentFull is [0,1]
	PercentFull float64
}

func NewBatteries() Batteries {
	return Batteries{Data: make([]BatteryInfo, 0)}
}

// LOCAL(host) batteries
func LocalBatteries() Batteries {
	bats, err := battery.GetAll()
	if err != nil {
		return Batteries{Data: make([]BatteryInfo, 0)}
	}
	return Batteries{Data: make([]BatteryInfo, len(bats))}
}

func (b *Batteries) Update() error {
	bats, err := battery.GetAll()
	if err != nil {
		return fmt.Errorf("error setting up batteries: %v", err)
	}
	if len(bats) < 1 {
		return fmt.Errorf("no batteries")
	}
	// battery library does not provide uniquely identifying metadata for each
	// battery, so we try to use the design capacity as it should be immutable.
	// Since systems could still exist that have multiple duplicate batteries, this
	// code checks for dupes and prevents replacement.
	fullSum := 0.0
	currentSum := 0.0
	for i, bat := range bats {
		if i >= len(b.Data) {
			b.Data = append(b.Data, BatteryInfo{})
		}
		b.Data[i].Full = bat.Full
		b.Data[i].Current = bat.Current
		b.Data[i].Charging = bat.State == battery.Charging
		b.Data[i].ChargeRate = bat.ChargeRate
		// Calculate the % full, weighted across battery size
		fullSum += bat.Full / float64(len(bats))
		currentSum += bat.Current / float64(len(bats))
	}
	b.PercentFull = (currentSum / fullSum)
	return nil
}

func (b Batteries) EnableMetrics(s *metrics.Set) {
	batsSeen := make(map[float64]int)
	for _, bat := range b.Data {
		key := fmt.Sprintf("%d:%f", batsSeen[bat.Design], bat.Design)
		batsSeen[bat.Design]++
		s.NewGauge(makeName("batt", key, "total"), func() float64 {
			if bat.Current == 0 || bat.Full == 0 {
				return 0.0
			}
			return bat.Current / bat.Full
		})
	}
}
