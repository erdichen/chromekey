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

unsigned int _EVIOCGNAME(unsigned int len) {
	return EVIOCGNAME(len);
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

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

func EVIOCGNAME(len uint) uint {
	return uint(C._EVIOCGNAME(C.uint(len)))
}

func structSizeMismatch()

const EventSize = int(unsafe.Sizeof(*(*C.struct_input_event)(nil)))
const goEventSize = int(unsafe.Sizeof(*(*InputEvent)(nil)))

func init() {
	if goEventSize != EventSize {
		structSizeMismatch()
	}
}

func (ev *InputEvent) Marshal() []byte {
	cev := C.struct_input_event{
		time:  C.struct_timeval{tv_sec: C.long(ev.Sec), tv_usec: C.long(ev.Usec)},
		_type: C.ushort(ev.Type),
		code:  C.ushort(ev.Code),
		value: C.int(ev.Value),
	}
	return (*(*[EventSize]byte)(unsafe.Pointer(&cev)))[:]
}

func (ev *InputEvent) unmarshal(b []byte) (int, error) {
	if len(b) < EventSize {
		return 0, errors.New("not enough data")
	}
	p := (*C.struct_input_event)(unsafe.Pointer(&b[0]))
	ev.Sec = uint(p.time.tv_sec)
	ev.Usec = uint(p.time.tv_usec)
	ev.Type = uint16(p._type)
	ev.Code = uint16(p.code)
	ev.Value = int32(p.value)
	return EventSize, nil
}
