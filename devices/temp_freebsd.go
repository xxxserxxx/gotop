//go:build freebsd
// +build freebsd

package devices

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xxxserxxx/gotop/v4/utils"
)

var sensorOIDS = map[string]string{
	"dev.cpu.0.temperature":           "CPU 0 ",
	"hw.acpi.thermal.tz0.temperature": "Thermal zone 0",
}

func (t *Temperature) Update() error {
	errors := make([]string, 0)

	temps := make(map[string]float64)
	for k, v := range sensorOIDS {
		if _, ok := t.temps[k]; !ok {
			continue
		}
		output, err := exec.Command("sysctl", "-n", k).Output()
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		s1 := strings.Replace(string(output), "C", "", 1)
		s2 := strings.TrimSuffix(s1, "\n")
		convertedOutput := utils.ConvertLocalizedString(s2)
		value, err := strconv.ParseFloat(convertedOutput, 64)
		if err != nil {
			errors = append(errors, err.Error())
			continue
		}

		temps[v] = value
	}

	t.temps = temps
	if len(errors) == 0 {
		return nil
	}
	return fmt.Errorf(strings.Join(errors, "; "))
}

func thermalSensorNames() []string {
	rv := make([]string, 0, len(sensorOIDS))
	// Check that thermal sensors are really available; they aren't in VMs
	bs, err := exec.Command("sysctl", "-a").Output()
	if err != nil {
		log.Printf("%v", err)
		//log.Printf(tr.Value("error.fatalfetch", "temp", err.Error()))
		return []string{}
	}
	for k, _ := range sensorOIDS {
		idx := strings.Index(string(bs), k)
		if idx >= 0 {
			rv = append(rv, k)
		}
	}
	if len(rv) == 0 {
		oids := make([]string, 0, len(sensorOIDS))
		for k, _ := range sensorOIDS {
			oids = append(oids, k)
		}
		log.Printf("%v", err)
		//log.Printf(tr.Value("error.nodevfound", strings.Join(oids, ", ")))
	}
	return rv
}

func thermalSensorLabels() []string {
	rv := make([]string, 0, len(sensorOIDS))
	ns := thermalSensorNames()
	for _, n := range ns {
		rv = append(rv, sensorOIDS[n])
	}
	return rv
}
