//go:build linux || darwin || windows

package devices

import (
	"github.com/shirou/gopsutil/v3/host"
)

func (t *Temperature) Update() error {
	sensors, err := host.SensorsTemperatures()
	if err != nil {
		return err
	}
	tmps := make(map[string]float64)
	for _, sensor := range sensors {
		if _, ok := t.temps[sensor.SensorKey]; ok {
			tmps[sensor.SensorKey] = sensor.Temperature
		}
	}
	t.temps = tmps
	return nil
}
