//go:build linux || windows

package devices

import (
	"log"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
)

// All possible thermometers
func thermalSensorNames() []string {
	sensors, err := host.SensorsTemperatures()
	if err != nil {
		log.Printf("gopsutil reports %s", err)
		if len(sensors) == 0 {
			log.Printf("no temperature sensors returned")
			return []string{}
		}
	}
	rv := make([]string, len(sensors))
	for i, sensor := range sensors {
		label := sensor.SensorKey
		label = strings.TrimSuffix(sensor.SensorKey, "_input")
		label = strings.TrimSuffix(label, "_thermal")
		rv[i] = label
	}
	return rv
}

func thermalSensorLabels() []string {
	sensors, err := host.SensorsTemperatures()
	if err != nil {
		log.Printf("gopsutil reports %s", err)
		if len(sensors) == 0 {
			log.Printf("no temperature sensors returned")
			return []string{}
		}
	}
	rv := make([]string, len(sensors))
	for i, sensor := range sensors {
		rv[i] = sensor.SensorKey
	}
	return rv
}
