package devices

import (
	"fmt"
	"log"
	"strings"

	"github.com/VictoriaMetrics/metrics"
	"github.com/shirou/gopsutil/v3/disk"
)

// Partition represents a partition on a disk.
type Partition struct {
	// Device is the operating system's name for the raw device
	Device string
	// MountPoint is where the operating systems mounts the device
	MountPoint string
	// BytesRead is the total number of bytes read through the device
	BytesRead uint64
	// BytesWritten is the total number of bytes written through the device
	BytesWritten uint64
	// BytesReadRecently is the number of bytes read since the last call to Update()
	BytesReadRecently uint64
	// BytesWrittenRecently is the number of bytes written since the last call to Update()
	BytesWrittenRecently uint64
	// UsedPercent is in [0,100]
	UsedPercent float64
	// Free is how many bytes are free on the partition
	Free uint64
}

// Disk is a set of all partitions, keyed by the partition's Device field
type Disk map[string]*Partition

func NewDisk() Disk {
	return make(Disk)
}

// LocalDisk creates a new local disk device
func LocalDisk() Disk {
	d := NewDisk()
	d.Update()
	return d
}

// Update refreshes partition information, adding newly discovered partitions
// and removing ones that have disappeared.
//
// Recoverable errors get logged, not returned
func (dsk Disk) Update() error {
	ps, err := disk.Partitions(false)
	if err != nil {
		return fmt.Errorf("disk partitions setup error %v", err.Error())
	}
	// add partition if it's new
	for _, p := range ps {
		// don't show loop devices
		if strings.HasPrefix(p.Device, "/dev/loop") {
			continue
		}
		// don't show docker container filesystems
		if strings.HasPrefix(p.Mountpoint, "/var/lib/docker/") {
			continue
		}
		// check if partition doesn't already exist in our list
		if _, ok := dsk[p.Device]; !ok {
			dsk[p.Device] = &Partition{
				Device:     p.Device,
				MountPoint: p.Mountpoint,
			}
		}

		// Update the usage stats
		// We add 0.5 to all values to make sure the truncation rounds
		part := dsk[p.Device]
		usage, err := disk.Usage(part.MountPoint)
		if err != nil {
			log.Printf("recoverable error fetching disk usage for partition %s: %v", part.MountPoint, err.Error())
			continue
		}
		part.UsedPercent = usage.UsedPercent
		part.Free = usage.Free

		ioCounters, err := disk.IOCounters(part.Device)
		if err != nil {
			log.Printf("recoverable error fetching IO counters for partition %s: %v", part.Device, err.Error())
			continue
		}
		ioCounter := ioCounters[strings.Replace(part.Device, "/dev/", "", -1)]
		bytesRead, bytesWritten := ioCounter.ReadBytes, ioCounter.WriteBytes
		// FIXME this is wrong if the update isn't ever second -- need to divide by refresh
		if part.BytesRead != 0 { // if this isn't the first update
			part.BytesReadRecently = bytesRead - part.BytesRead
			part.BytesWrittenRecently = bytesWritten - part.BytesWritten
		} else {
			part.BytesReadRecently = 0.0
			part.BytesWrittenRecently = 0.0
		}
		part.BytesRead, part.BytesWritten = bytesRead, bytesWritten
	}

	// delete a partition if it no longer exists
	if len(ps) != len(dsk) {
	top:
		for dev, _ := range dsk {
			for _, p := range ps {
				if dev == p.Device {
					continue top
				}
			}
			delete(dsk, dev)
		}
	}

	return nil
}

// EnableMetrics creates new percent-usage metrics gauges for all disk
// partitions. If Update() returns an error, no gauges are created.
func (d Disk) EnableMetrics(s *metrics.Set) {
	if d.Update() != nil {
		return
	}
	for key, part := range d {
		pc := part
		s.NewGauge(makeName("disk", "freepc", strings.ReplaceAll(key, "/", ":")), func() float64 {
			return pc.UsedPercent
		})
		s.NewGauge(makeName("disk", "free", strings.ReplaceAll(key, "/", ":")), func() float64 {
			return float64(pc.Free)
		})
		s.NewGauge(makeName("disk", "read", strings.ReplaceAll(key, "/", ":")), func() float64 {
			return float64(pc.BytesRead)
		})
		s.NewGauge(makeName("disk", "write", strings.ReplaceAll(key, "/", ":")), func() float64 {
			return float64(pc.BytesWritten)
		})
	}
}

func partitions() ([]string, error) {
	ps, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("disk partitions setup error %v", err.Error())
	}
	rv := make([]string, len(ps))
	for i, p := range ps {
		rv[i] = p.Device
	}
	return rv, nil
}

func mountPoints() ([]string, error) {
	ps, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("disk partitions setup error %v", err.Error())
	}
	rv := make([]string, len(ps))
	for i, p := range ps {
		rv[i] = p.Mountpoint
	}
	return rv, nil
}
