//go:build !cgo

package evdev

import (
	"encoding/binary"
	"errors"
	"unsafe"

	"github.com/erdichen/chromekey/ioc"
)

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

func EVIOCGNAME(len uint) uint {
	return ioc.IOC(ioc.Read, 'E', 0x06, len)
}

const EventSize = int(unsafe.Sizeof(*(*InputEvent)(nil)))

func (ev *InputEvent) Marshal() []byte {
	b := [EventSize + 8]byte{}
	if PtrSize == 8 {
		binary.LittleEndian.PutUint64(b[:], uint64(ev.Sec))
		binary.LittleEndian.PutUint64(b[8:], uint64(ev.Usec))
		binary.LittleEndian.PutUint16(b[16:], ev.Type)
		binary.LittleEndian.PutUint16(b[18:], ev.Code)
		binary.LittleEndian.PutUint32(b[20:], uint32(ev.Value))
	} else {
		binary.LittleEndian.PutUint64(b[:], uint64(ev.Sec))
		binary.LittleEndian.PutUint64(b[4:], uint64(ev.Usec))
		binary.LittleEndian.PutUint16(b[8:], ev.Type)
		binary.LittleEndian.PutUint16(b[10:], ev.Code)
		binary.LittleEndian.PutUint32(b[12:], uint32(ev.Value))
	}
	return b[:EventSize]
}

func (ev *InputEvent) unmarshal(b []byte) (int, error) {
	if len(b) < EventSize {
		return 0, errors.New("not enough data")
	}
	if PtrSize == 8 {
		ev.Sec = uint(binary.LittleEndian.Uint64(b))
		ev.Usec = uint(binary.LittleEndian.Uint64(b[8:]))
		ev.Type = uint16(binary.LittleEndian.Uint16(b[16:]))
		ev.Code = uint16(binary.LittleEndian.Uint16(b[18:]))
		ev.Value = int32(binary.LittleEndian.Uint32(b[20:]))
	} else {
		ev.Sec = uint(binary.LittleEndian.Uint32(b))
		ev.Usec = uint(binary.LittleEndian.Uint32(b[4:]))
		ev.Type = uint16(binary.LittleEndian.Uint16(b[8:]))
		ev.Code = uint16(binary.LittleEndian.Uint16(b[10:]))
		ev.Value = int32(binary.LittleEndian.Uint32(b[12:]))
	}
	return EventSize, nil
}
