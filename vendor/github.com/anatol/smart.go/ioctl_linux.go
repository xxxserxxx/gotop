package smart

import (
	"golang.org/x/sys/unix"
)

const (
	directionNone  = 0
	directionWrite = 1
	directionRead  = 2

	numberBits    = 8
	typeBits      = 8
	sizeBits      = 14
	directionBits = 2

	numberMask    = (1 << numberBits) - 1
	typeMask      = (1 << typeBits) - 1
	sizeMask      = (1 << sizeBits) - 1
	directionMask = (1 << directionBits) - 1

	numberShift    = 0
	typeShift      = numberShift + numberBits
	sizeShift      = typeShift + typeBits
	directionShift = sizeShift + sizeBits
)

// ioc calculates the ioctl command for the specified direction, type, number and size
func ioc(dir, t, nr, size uintptr) uintptr {
	return (dir << directionShift) | (t << typeShift) | (nr << numberShift) | (size << sizeShift)
}

// ior calculates the ioctl command for a read-ioctl of the specified type, number and size
func ior(t, nr, size uintptr) uintptr {
	return ioc(directionRead, t, nr, size)
}

// iow calculates the ioctl command for a write-ioctl of the specified type, number and size
func iow(t, nr, size uintptr) uintptr {
	return ioc(directionWrite, t, nr, size)
}

// iowr calculates the ioctl command for a read/write-ioctl of the specified type, number and size
func iowr(t, nr, size uintptr) uintptr {
	return ioc(directionWrite|directionRead, t, nr, size)
}

// ioctl executes an ioctl command on the specified file descriptor
func ioctl(fd, cmd, ptr uintptr) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, fd, cmd, ptr)
	if errno != 0 {
		return errno
	}
	return nil
}
