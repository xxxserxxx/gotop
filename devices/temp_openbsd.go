//go:build openbsd
// +build openbsd

package devices

// loosely based on https://github.com/openbsd/src/blob/master/sbin/sysctl/sysctl.c#L2517

// #include <sys/time.h>
// #include <sys/sysctl.h>
// #include <sys/sensors.h>
import "C"

import (
	"strconv"
	"syscall"
	"unsafe"
)

func (t *Temperature) Update() error {
	mib := []C.int{0, 1, 2, 3, 4}

	var snsrdev C.struct_sensordev
	var len C.ulong = C.sizeof_struct_sensordev

	mib[0] = C.CTL_HW
	mib[1] = C.HW_SENSORS
	mib[3] = C.SENSOR_TEMP

	tmps := make(map[string]float64)
	var i C.int
	for i = 0; ; i++ {
		mib[2] = i
		if v, e := C.sysctl(&mib[0], 3, unsafe.Pointer(&snsrdev), &len, nil, 0); v == -1 {
			if e == syscall.ENXIO {
				continue
			}
			if e == syscall.ENOENT {
				break
			}
		}
		t.getTemp(mib, 4, &snsrdev, 0, tmps)
	}
	t.temps = tmps
	return nil
}

func (t Temperature) getTemp(mib []C.int, mlen int, snsrdev *C.struct_sensordev, index int, tmps map[string]float64) {
	switch mlen {
	case 4:
		k := mib[3]
		var numt C.int
		for numt = 0; numt < snsrdev.maxnumt[k]; numt++ {
			mib[4] = numt
			t.getTemp(mib, mlen+1, snsrdev, int(numt), tmps)
		}
	case 5:
		var snsr C.struct_sensor
		var slen C.size_t = C.sizeof_struct_sensor

		if v, _ := C.sysctl(&mib[0], 5, unsafe.Pointer(&snsr), &slen, nil, 0); v == -1 {
			return
		}

		if slen > 0 && (snsr.flags&C.SENSOR_FINVALID) == 0 {
			key := C.GoString(&snsrdev.xname[0]) + ".temp" + strconv.Itoa(index)
			temp := (snsr.value - 273150000.0) / 1000000.0

			if _, ok := temps[key]; ok {
				tmps[key] = temp
			}
		}
	}
}

func thermalSensorNames() []string {
	t.NewTemperature()
	t.Update()
	rv := make([]string, 0)
	for n, _ := range t.temps {
		rv = append(rv, n)
	}
	return rv
}

func thermalSensorLabels() []string {
	return thermalSensorNames()
}
