// +build linux,!arm64,!loong64,!riscv64 openbsd,!arm64 freebsd darwin

package logging

import (
	"os"
	"syscall"
)

func stderrToLogfile(logfile *os.File) {
	syscall.Dup2(int(logfile.Fd()), 2)
}
