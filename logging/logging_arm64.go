// +build !freebsd,arm64, !darwin

package logging

import (
	"os"
	"syscall"
)

func stderrToLogfile(logfile *os.File) {
	syscall.Dup3(int(logfile.Fd()), 2, 0)
}
