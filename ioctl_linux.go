// +build linux

package joystick

import (
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"unsafe"
)

const (
	IOC_NRBITS   = 8
	IOC_TYPEBITS = 8
	IOC_SIZEBITS = 14
	IOC_DIRBITS  = 2

	IOC_NRMASK   = ((1 << IOC_NRBITS) - 1)
	IOC_TYPEMASK = ((1 << IOC_TYPEBITS) - 1)
	IOC_SIZEMASK = ((1 << IOC_SIZEBITS) - 1)
	IOC_DIRMASK  = ((1 << IOC_DIRBITS) - 1)

	IOC_NRSHIFT   = 0
	IOC_TYPESHIFT = (IOC_NRSHIFT + IOC_NRBITS)
	IOC_SIZESHIFT = (IOC_TYPESHIFT + IOC_TYPEBITS)
	IOC_DIRSHIFT  = (IOC_SIZESHIFT + IOC_SIZEBITS)

	IOC_NONE  = 0
	IOC_WRITE = 1
	IOC_READ  = 2
)

func _IOC(dir int, t int, nr int, size int) int {
	return (dir << IOC_DIRSHIFT) | (t << IOC_TYPESHIFT) | (nr << IOC_NRSHIFT) | (size << IOC_SIZESHIFT)
}

func _IOR(t int, nr int, size int) int {
	return _IOC(IOC_READ, t, nr, size)
}

func _IOW(t int, nr int, size int) int {
	return _IOC(IOC_WRITE, t, nr, size)
}

func Ioctl(f *os.File, req int, ptr unsafe.Pointer) syscall.Errno {
	_, _, err := unix.Syscall(unix.SYS_IOCTL, uintptr(f.Fd()), uintptr(req), uintptr(ptr))
	return err
}
