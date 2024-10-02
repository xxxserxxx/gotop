// +build !freebsd,!darwin,arm64 linux,loong64 linux,riscv64

package logging

import (
	"os"
	"syscall"
)

func stderrToLogfile(logfile *os.File) {
	syscall.Dup3(int(logfile.Fd()), 2, 0)
}
