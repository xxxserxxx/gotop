package devices

import (
	"strings"

	"github.com/VictoriaMetrics/metrics"
)

type Temperature struct {
	temps map[string]float64
}

func NewTemperature() Temperature {
	return Temperature{temps: make(map[string]float64)}
}

// LocalTemperature sets up tracking for a filtered list of thermal sensors.
// `filter` contains the list filter:
//
// 1. Included sensors are globs matching the sensor name, e.g. "coretemp*"
// 2. Excluded interfaces are prefixed by `!`, e.g. "!nvme*"
// 3. If the list contains *only* exclusions, then all sensors not excluded are included
// 4. If the list contains any non-exclusions, then only those sensors are included
// 5. Exclusion overrides inclusion
func LocalTemperature(filter []string) Temperature {
	nt := Temperature{
		temps: make(map[string]float64),
	}
	if len(filter) != 0 {
		excludes := make(map[string]bool)
		includes := make(map[string]bool)
		for _, f := range filter {
			if strings.HasPrefix(f, "!") {
				excludes[strings.TrimPrefix(f, "!")] = true
			} else {
				includes[f] = true
			}
		}
		for _, s := range thermalSensorNames() {
			// If is in exclude, exclude
			if _, exclude := excludes[s]; exclude {
				continue
			}
			// If includes is not empty & not listed, exclude
			if _, ok := includes[s]; len(includes) != 0 && !ok {
				continue
			}
			// If we make it all the way to here, include it
			nt.temps[s] = 0.0
		}
	} else {
		// Include everything (except maybe VPNs)
		for _, s := range thermalSensorNames() {
			nt.temps[s] = 0.0
		}
	}
	return nt
}

func (t Temperature) EnableMetrics(s *metrics.Set) {
	for k := range t.temps {
		kc := k
		s.NewGauge(makeName("temp", k), func() float64 {
			tmps := t.temps
			rv := tmps[kc]
			return rv
		})
	}
}

func (t Temperature) Temps() map[string]float64 {
	rv := t.temps
	return rv
}
