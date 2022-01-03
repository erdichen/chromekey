package ioc

import "syscall"

const (
	None  = 0
	Write = 1
	Read  = 2

	iocNRBits   = 8
	iocTypeBits = 8
	iocSizeBits = 14
	iocDirBits  = 2

	iocNRMask   = (1 << iocNRBits) - 1
	iocTypeMask = (1 << iocTypeBits) - 1
	iocSizeMask = (1 << iocSizeBits) - 1
	iocDirMask  = (1 << iocDirBits) - 1

	iocNRShift   = 0
	iocTypeShift = iocNRShift + iocNRBits
	iocSizeShift = iocTypeShift + iocTypeBits
	iocDirShift  = iocSizeShift + iocSizeBits
)

func IOC(dir, typ, nr, size uint) uint {
	return dir<<iocDirShift |
		typ<<iocTypeShift |
		nr<<iocNRShift |
		size<<iocSizeShift
}

func IO(typ, nr uint) uint {
	return IOC(None, typ, nr, 0)
}

func IOR(typ, nr, size uint) uint {
	return IOC(Read, typ, nr, size)
}

func IOW(typ, nr, size uint) uint {
	return IOC(Write, typ, nr, size)
}

func IORW(typ, nr, size uint) uint {
	return IOC(Read|Write, typ, nr, size)
}

func Ioctl(fd int, req uint, arg uintptr) (err error) {
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), uintptr(req), uintptr(arg))
	if e1 != 0 {
		err = e1
	}
	return
}
