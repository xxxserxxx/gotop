package devices

import (
	"log"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/xxxserxxx/gotop/v4"
)

type Device interface {
	Update() error
	EnableMetrics(*metrics.Set)
}

// Startup is after configuration has been parsed, and initializes devices.
//
// devices is a list of the devices to spin up; any specific device means a sensor
// on the current machine. `nvidia` means NVidia GPU sensors, and any remote gotop
// instances defined in c.Remotes will also be connected.
//
// Startup attempts to start everything and continues when it encounters errors; any
// collected errors are returned in the error array, and devices which have errors
// will not be included in the returned Device array.
func Startup(devices []string, c gotop.Config) (map[string]Device, []error) {
	devs := make(map[string]Device)
	for _, d := range devices {
		switch d {
		case "batt", "power":
			bat := LocalBatteries()
			bat.Update()
			devs["batt"] = &bat
		case "cpu":
			cpu := LocalCPUs(c.PercpuLoad)
			cpu.Update()
			devs[d] = &cpu
		case "disk":
			disk := LocalDisk()
			disk.Update()
			devs[d] = disk
		case "mem":
			mem := LocalMemory()
			mem.Update()
			devs[d] = mem
		case "net":
			net := LocalNetwork(c.NetInterface, false)
			net.Update()
			devs[d] = &net
		case "temp":
			tmp := LocalTemperature(c.Temps)
			tmp.Update()
			devs[d] = &tmp
		case "procs":
			// TODO procs not yet implemented as local device
		case "remote", "nvidia":
			// NOP handled below
		default:
			log.Printf(c.Tr.Value("error.unknowndevice", d))
		}
	}
	if c.Nvidia {
		dev, err := NewNVidia()
		dev.Update()
		if err != nil {
			log.Print(err)
		} else {
			devs["nvidia"] = &dev
		}
	}
	for _, r := range c.Remotes {
		dev := NewRemote(r.Name, r.URL, r.Refresh)
		dev.Update()
		devs["remote-"+r.Name] = &dev
	}
	return devs, nil
}

// Spawn spins up threads for updating the devices.
func Spawn(devs map[string]Device, c gotop.Config) {
	for name, dev := range devs {
		if c.ExportPort != "" {
			dev.EnableMetrics(c.Metrics)
		}
		go func(n string, d Device) {
			for range time.NewTicker(c.UpdateInterval).C {
				err := d.Update()
				if err != nil {
					log.Print(err)
					break
				}
			}
		}(name, dev)
	}
}

func Domains() []string {
	ad := gotop.AllDevices()
	rv := make([]string, len(ad))
	for i, d := range ad {
		rv[i] = strings.Title(d)
	}
	return rv
}

// Return a list of devices registered under domain, where `domain` is one of the
// defined constants in `devices`, e.g., devices.Temperatures.  The
// `enabledOnly` flag determines whether all devices are returned (false), or
// only the ones that have been enabled for the domain.
func Devices(domain string, all bool) []string {
	switch domain {
	case "Temperatures":
		return thermalSensorNames()
	case "Disk":
		ps, _ := partitions()
		return ps
	case "Network":
		is, _ := interfaces()
		return is
	default:
		return []string{}
	}
}
