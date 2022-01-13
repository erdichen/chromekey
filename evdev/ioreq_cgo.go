//go:build cgo

package evdev

/*
#include <linux/input.h>

unsigned int _EVIOCGBIT(unsigned int ev, unsigned int len) {
	return EVIOCGBIT(ev, len);
}

unsigned int _EVIOCGKEY(unsigned int len) {
	return EVIOCGKEY(len);
}

unsigned int _EVIOCGLED(unsigned int len) {
	return EVIOCGLED(len);
}
*/
import "C"

var (
	EVIOCGRAB     = C.EVIOCGRAB
	EVIOCREVOKE   = C.EVIOCREVOKE
	EVIOCSCLOCKID = C.EVIOCSCLOCKID
)

func EVIOCGBIT(ev, len uint) uint {
	return uint(C._EVIOCGBIT(C.uint(ev), C.uint(len)))
}

func EVIOCGKEY(len uint) uint {
	return uint(C._EVIOCGKEY(C.uint(len)))
}

func EVIOCGLED(len uint) uint {
	return uint(C._EVIOCGLED(C.uint(len)))
}
