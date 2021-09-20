package devices

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/VictoriaMetrics/metrics"
)

type Remote struct {
	Name    string
	URL     string
	Refresh time.Duration

	CPUData  *CPUs
	TempData *Temperature
	NetData  *Network
	DiskData Disk
	MemData  Memory
}

// TODO remote network & disk aren't reported
func NewRemote(n, u string, r time.Duration) Remote {
	cpus := NewCPUs(n, false)
	net := NewNetwork()
	temps := NewTemperature()
	rm := Remote{
		Name:     n,
		URL:      u,
		Refresh:  r,
		CPUData:  &cpus,
		TempData: &temps,
		NetData:  &net,
		DiskData: NewDisk(),
		MemData:  NewMemory(),
	}

	_, err := url.Parse(u)
	if err != nil {
		log.Printf("bad remote URL %s", u)
		return Remote{}
	}
	rm.Update()
	return rm
}

func (rm Remote) Update() error {
	res, err := http.Get(rm.URL)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == http.StatusOK {
			bi := bufio.NewScanner(res.Body)
			rm.process(bi)
		} else {
			return fmt.Errorf("unsuccessful connection to %s: http status %s.", rm.URL, res.Status)
		}
	} else {
		return fmt.Errorf("error pulling remote gotop: %v", err)
	}
	return nil
}

func (r *Remote) process(data *bufio.Scanner) {
	temps := make(map[string]float64)
	r.NetData.RecentBytesRecv = 0
	r.NetData.RecentBytesSent = 0
	for data.Scan() {
		line := data.Text()
		if len(line) < 6 {
			continue
		}
		switch {
		case strings.HasPrefix(line, "cpu_"): // cpu_0 INT
			parts := strings.Split(line[4:], " ") // 0 INT
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			i, err := strconv.Atoi(parts[0])
			if parts[0] == "avg" {
				r.CPUData.Average = val
			} else {
				if i >= len(r.CPUData.Data) {
					r.CPUData.Data = append(r.CPUData.Data, val)
				} else {
					r.CPUData.Data[i] = val
				}
			}
		case strings.HasPrefix(line, "temp_"): // int temp_LABEL
			parts := strings.Split(line[5:], " ")
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			temps[parts[0]] = val
		case strings.HasPrefix(line, "net_"): // int net_IFACE_RECV
			parts := strings.Split(line[4:], " ")
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			subparts := strings.Split(parts[0], "_") // IFACE_RECV
			if len(subparts) < 2 {
				log.Printf(`bad network data; expected net_IFACE_RECV, got "%s"`, line)
				continue
			}
			iface := r.NetData.Data[subparts[0]]
			log.Printf("%s %d", subparts[0], r.NetData.RecentBytesRecv)
			iface.IFace = subparts[0]
			// val is the total data received on this interface
			if subparts[1] == "recv" {
				recent := uint64(val) - iface.BytesRecv
				iface.BytesRecv = uint64(val)
				// recent is current value - previous value
				if recent > 0 {
					r.NetData.RecentBytesRecv += recent
					r.NetData.TotalBytesRecv += recent
				}
			} else {
				recent := uint64(val) - r.NetData.TotalBytesSent
				iface.BytesSent = uint64(val)
				// recent is current value - previous value
				if recent > 0 {
					r.NetData.RecentBytesSent += recent
					r.NetData.TotalBytesSent += recent
				}
			}
			r.NetData.Data[subparts[0]] = iface
		case strings.HasPrefix(line, "disk_"): // disk_freepc_:dev:mmcblk0p1 (freepc, free, read, write)
			parts := strings.Split(line[5:], " ")
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			var part *Partition
			var ok bool
			subparts := strings.Split(parts[0], "_") // [free,PART]
			if part, ok = r.DiskData[subparts[1]]; !ok {
				part = &Partition{
					Device: strings.ReplaceAll(subparts[1], ":", "/"),
				}
			}
			switch subparts[0] {
			case "freepc":
				part.UsedPercent = val
			case "free":
				part.Free = uint64(val)
			case "read":
				part.BytesReadRecently = uint64(val) - part.BytesRead
				part.BytesRead = uint64(val)
			case "write":
				part.BytesWrittenRecently = uint64(val) - part.BytesWritten
				part.BytesWritten = uint64(val)
			default:
			}
			r.DiskData[subparts[1]] = part
		case strings.HasPrefix(line, "memory_"): // memory_TOTAL_LABEL floatBytes (memory_used_Main 1000)
			parts := strings.Split(line[7:], " ")
			if len(parts) < 2 {
				log.Printf(`bad data; not enough columns in "%s"`, line)
				continue
			}
			val, err := strconv.ParseFloat(parts[1], 64)
			if err != nil {
				log.Print(err)
				continue
			}
			subparts := strings.Split(parts[0], "_") // TOTAL_LABEL
			if len(subparts) < 2 {
				log.Printf(`bad memory data; expected e.g. memory_used_Main, got "%s"`, line)
				continue
			}
			tu := subparts[0]
			mn := subparts[1]
			var mi *MemoryInfo
			var ok bool
			if mi, ok = r.MemData[mn]; !ok {
				mi = &MemoryInfo{}
			}
			if tu == "total" {
				mi.Total = uint64(val)
				mi.UsedPercent = (float64(mi.Used) / float64(mi.Total)) * 100.0
			} else if tu == "used" {
				mi.Used = uint64(val)
				mi.UsedPercent = (float64(mi.Used) / float64(mi.Total)) * 100.0
			}
			r.MemData[mn] = mi
		default:
			// NOP!  This is a metric we don't care about.
		}
	}
	r.TempData.temps = temps

}

func ParseConfig(vars map[string]string) map[string]Remote {
	rv := make(map[string]Remote)
	for key, value := range vars {
		if strings.HasPrefix(key, "remote-") {
			parts := strings.Split(key, "-")
			if len(parts) == 2 {
				log.Printf("malformed Remote extension configuration '%s'; must be 'remote-NAME-url' or 'remote-NAME-refresh'", key)
				continue
			}
			name := parts[1]
			remote, ok := rv[name]
			if !ok {
				remote = Remote{}
			}
			if parts[2] == "url" {
				remote.URL = value
			} else if parts[2] == "refresh" {
				sleep, err := strconv.Atoi(value)
				if err != nil {
					log.Printf("illegal Remote extension value for %s: '%s'.  Must be a duration in seconds, e.g. '2'", key, value)
					continue
				}
				remote.Refresh = time.Duration(sleep) * time.Second
			} else {
				log.Printf("bad configuration option for Remote extension: '%s'; must be 'remote-NAME-url' or 'remote-NAME-refresh'", key)
				continue
			}
			rv[name] = remote
		}
	}
	return rv
}

func (v *Remote) EnableMetrics(s *metrics.Set) {
	// NOP
}

// TODO add sensor names from remote machines
// TODO add partitions from remote machines
