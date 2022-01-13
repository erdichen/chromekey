//go:build !cgo

package evdev

import "erdi.us/chromekey/ioc"

var (
	EVIOCGRAB     = ioc.IOW('E', 0x90, 4)
	EVIOCREVOKE   = ioc.IOW('E', 0x91, 4)
	EVIOCSCLOCKID = ioc.IOW('E', 0xa0, 4)
)

func EVIOCGBIT(ev, len uint) uint {
	return ioc.IOC(ioc.Read, 'E', 0x20+ev, len)
}

func EVIOCGKEY(len uint) uint {
	return ioc.IOC(ioc.Read, 'E', 0x18, len)
}

func EVIOCGLED(len uint) uint {
	return ioc.IOC(ioc.Read, 'E', 0x19, len)
}
