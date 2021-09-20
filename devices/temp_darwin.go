//go:build darwin
// +build darwin

package devices

import (
	"bytes"
	_ "embed"
	"encoding/csv"
	"github.com/shirou/gopsutil/v3/host"
	"io"
	"log"
)

// All possible thermometers
func thermalSensorNames() []string {
	// Otherwise, get the sensor data from the system & filter it
	ids := loadIDs()
	sensors, err := host.SensorsTemperatures()
	if err != nil {
		log.Printf("error getting sensor list for temps: %s", err)
		return []string{}
	}
	rv := make([]string, 0, len(sensors))
	for _, sensor := range sensors {
		// 0-value sensors are not implemented
		if sensor.Temperature == 0 {
			continue
		}
		if label, ok := ids[sensor.SensorKey]; ok {
			rv = append(rv, label)
		}
	}
	return rv
}

func thermalSensorLabels() []string {
	ids := loadIDs()
	sensors, err := host.SensorsTemperatures()
	if err != nil {
		log.Printf("error getting sensor list for temps: %s", err)
		return []string{}
	}
	rv := make([]string, 0, len(sensors))
	for _, sensor := range sensors {
		// 0-value sensors are not implemented
		if sensor.Temperature == 0 {
			continue
		}
		if _, ok := ids[sensor.SensorKey]; ok {
			rv = append(rv, sensor.SensorKey)
		}
	}
	return rv
}

//go:embed "smc.tsv"
var smcData []byte

// loadIDs parses the embedded smc.tsv data that maps Darwin SMC
// sensor IDs to their human-readable labels into an array and returns the
// array. The array keys are the 4-letter sensor keys; the values are the
// human labels.
func loadIDs() map[string]string {
	rv := make(map[string]string)
	parser := csv.NewReader(bytes.NewReader(smcData))
	parser.Comma = '\t'
	var line []string
	var err error
	for {
		if line, err = parser.Read(); err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error parsing SMC tags for temp widget: %s", err)
			break
		}
		// The line is malformed if len(line) != 2, but because the asset is static
		// it makes no sense to report the error to downstream users. This must be
		// tested at/around compile time.
		// FIXME assert all lines in smc.tsv have 2 columns during unit tests
		if len(line) == 2 {
			rv[line[0]] = line[1]
		}
	}
	return rv
}
