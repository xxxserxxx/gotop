// +build netbsd

package widgets

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xxxserxxx/gotop/v4/utils"
)

const (
	// Define column widths for ps output used in Procs()
	five  = "12345"
	ten   = five + five
	fifty = ten + ten + ten + ten + ten
)

func getProcs() ([]Proc, error) {
	// NetBSD ps does not support comma-separated keywords in a single -o argument:
	// it treats everything after the first '=' as the column header, so commas are
	// included in the header rather than parsed as keyword separators.
	// Use separate -o flags instead to get all five columns.
	//
	// Column layout with these headers (0-indexed):
	//   [0:10]   pid  (10 chars, right-justified)
	//   [11:61]  comm (50 chars, left-justified)
	//   [62:67]  pcpu (5 chars, right-justified)
	//   [68:73]  pmem (5 chars, right-justified)
	//   [74:]    args
	output, err := exec.Command("ps", "-caxo", "pid="+ten,
		"-o", "comm="+fifty,
		"-o", "pcpu="+five,
		"-o", "pmem="+five,
		"-o", "args").Output()
	if err != nil {
		return nil, fmt.Errorf("%s", tr.Value("widget.proc.err.ps", err.Error()))
	}

	// converts to []string and removes the header
	linesOfProcStrings := strings.Split(strings.TrimSpace(string(output)), "\n")[1:]
	procs := []Proc{}
	for _, line := range linesOfProcStrings {
		if len(line) < 74 {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(line[0:10]))
		if err != nil {
			log.Println(tr.Value("widget.proc.err.pidconv", err.Error(), line))
		}
		cpu, err := strconv.ParseFloat(utils.ConvertLocalizedString(strings.TrimSpace(line[62:67])), 64)
		if err != nil {
			log.Println(tr.Value("widget.proc.err.cpuconv", err.Error(), line))
		}
		mem, err := strconv.ParseFloat(utils.ConvertLocalizedString(strings.TrimSpace(line[68:73])), 64)
		if err != nil {
			log.Println(tr.Value("widget.proc.err.memconv", err.Error(), line))
		}
		proc := Proc{
			Pid:         pid,
			CommandName: strings.TrimSpace(line[11:61]),
			CPU:         cpu,
			Mem:         mem,
			FullCommand: strings.TrimSpace(line[74:]),
		}
		procs = append(procs, proc)
	}

	return procs, nil
}
