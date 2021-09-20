package devices

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/VictoriaMetrics/metrics"
)

type NVidia map[string]GPU

type GPU struct {
	Temps *Temperature
	Mems  Memory
	CPUs  *CPUs
}

// NewNVidia forks a thread to call the nvidia tool periodically and update the
// cached cpu, memory, and temperature values that are used by the update*()
// functions to return data to gotop.
//
// In this code, one call to the nvidia program returns *all* the data
// we're looking for, but gotop will call each update function during each
// cycle. This means that the nvidia program would be called 3 (or more)
// times per update, which isn't very efficient. Therefore, we make this
// code more complex to run a job in the background that runs the nvidia
// tool periodically and puts the results into hashes; the update functions
// then just sync data from those hashes into the return data.
//
// The refresh argument determines how frequently the metrics are updated.
func NewNVidia() (NVidia, error) {
	bs, err := exec.Command(
		"nvidia-smi",
		"--query-gpu=name,index,temperature.gpu,utilization.gpu,memory.total,memory.used",
		"--format=csv,noheader,nounits").Output()
	if err != nil {
		return NVidia{}, fmt.Errorf("NVidia GPU error during set-up: %s", err)
	}
	csvReader := csv.NewReader(bytes.NewReader(bs))
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	nv := make(map[string]GPU)
	for _, cols := range records {
		cpus := NewCPUs(cols[0], false)
		temp := NewTemperature()
		nv[cols[0]] = GPU{
			Temps: &temp,
			Mems:  NewMemory(),
			CPUs:  &cpus,
		}
	}
	return nv, nil
}

// Update calls the nvidia tool, parses the output, and caches the results
// in the various _* maps. The metric data parsed is: name, index,
// temperature.gpu, utilization.gpu, utilization.memory, memory.total,
// memory.free, memory.used
//
// This function returns an error if it can't call the `nvidia-smi` tool, or if
// there's a problem parsing the tool output.
func (nv NVidia) Update() error {
	bs, err := exec.Command(
		"nvidia-smi",
		"--query-gpu=name,index,temperature.gpu,utilization.gpu,memory.total,memory.used",
		"--format=csv,noheader,nounits").Output()
	if err != nil {
		return err
	}
	csvReader := csv.NewReader(bytes.NewReader(bs))
	csvReader.TrimLeadingSpace = true
	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}

	// Errors during parsing are recorded, but do not stop parsing.
	for i, cols := range records {
		errsFound := false
		aggErr := fmt.Errorf("error parsing line %d\n", i)

		name := cols[0]
		index := cols[1]
		idx, err := strconv.Atoi(index)
		if err != nil {
			idx = 0
		}
		gpu := nv[name]
		// 2 = GPU temp
		if gpu.Temps.temps[index], err = strconv.ParseFloat(cols[2], 64); err != nil {
			errsFound = true
			aggErr = fmt.Errorf("%v   parsing col %d: %v\n", aggErr, 2, err)
		}
		// 3 = GPU speed
		if idx >= len(gpu.CPUs.Data) {
			gpu.CPUs.Data = append(gpu.CPUs.Data, 0)
		}
		if gpu.CPUs.Data[idx], err = strconv.ParseFloat(cols[3], 64); err != nil {
			errsFound = true
			aggErr = fmt.Errorf("%v   parsing col %d: %v\n", aggErr, 3, err)
		}
		// 4 = total memory
		mt, err := strconv.Atoi(cols[4])
		if err != nil {
			errsFound = true
			aggErr = fmt.Errorf("%v   parsing col %d: %v\n", aggErr, 4, err)
		}
		// 5 = used memory
		mu, err := strconv.Atoi(cols[5])
		if err != nil {
			errsFound = true
			aggErr = fmt.Errorf("%v   parsing col %d: %v\n", aggErr, 5, err)
		}
		gpu.Mems[index] = &MemoryInfo{
			Total:       1048576 * uint64(mt),
			Used:        1048576 * uint64(mu),
			UsedPercent: (float64(mu) / float64(mt)) * 100.0,
		}
		if errsFound {
			return aggErr
		}
	}
	return nil
}

func (v NVidia) EnableMetrics(s *metrics.Set) {
	// NOP
}
