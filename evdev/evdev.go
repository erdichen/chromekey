package evdev

import (
	"context"
	"encoding/binary"
	"errors"
	"os"
	"unsafe"

	"erdi.us/chromekey/evdev/eventcode"
	"erdi.us/chromekey/evdev/keycode"
	"erdi.us/chromekey/ioc"
	"golang.org/x/sys/unix"
)

const EventSize = 24

type InputEvent struct {
	Sec  uint
	Usec uint

	Type  uint16
	Code  uint16
	Value int32
}

func (ev *InputEvent) Marshal() []byte {
	b := [EventSize]byte{}
	binary.LittleEndian.PutUint64(b[:], uint64(ev.Sec))
	binary.LittleEndian.PutUint64(b[8:], uint64(ev.Usec))
	binary.LittleEndian.PutUint16(b[16:], ev.Type)
	binary.LittleEndian.PutUint16(b[18:], ev.Code)
	binary.LittleEndian.PutUint32(b[20:], uint32(ev.Value))
	return b[:]
}

func (ev *InputEvent) unmarshal(b []byte) (int, error) {
	if len(b) < EventSize {
		return 0, errors.New("not enough data")
	}
	ev.Sec = uint(binary.LittleEndian.Uint64(b))
	ev.Usec = uint(binary.LittleEndian.Uint64(b[8:]))
	ev.Type = uint16(binary.LittleEndian.Uint16(b[16:]))
	ev.Code = uint16(binary.LittleEndian.Uint16(b[18:]))
	ev.Value = int32(binary.LittleEndian.Uint32(b[20:]))
	return EventSize, nil
}

type Device struct {
	f *os.File
}

func OpenDevice(device string) (*Device, error) {
	f, err := os.OpenFile(device, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Device{f: f}, nil
}

func (in *Device) Close() error {
	return in.f.Close()
}

func (in *Device) ReadEvents2(ctx context.Context, c chan InputEvent) error {
	buf := [EventSize * 64]byte{}

	for {
		n, err := in.f.Read(buf[:])
		if err != nil {
			return err
		}

		b := buf[:n]
		for len(b) >= EventSize {
			var ev InputEvent
			n, err := ev.unmarshal(b[:n])
			if err != nil {
				return err
			}
			select {
			case c <- ev:
			case <-ctx.Done():
				return nil
			}
			b = b[n:]
		}
	}
}

func (in *Device) ReadEvents(ctx context.Context) ([]InputEvent, error) {
	buf := [EventSize * 64]byte{}

	n, err := in.f.Read(buf[:])
	if err != nil {
		return nil, err
	}

	cnt := n / EventSize
	events := make([]InputEvent, cnt)

	b := buf[:n]
	for i := 0; len(b) >= EventSize; i++ {
		n, err := events[i].unmarshal(b[:n])
		if err != nil {
			return nil, err
		}
		b = b[n:]
	}
	return events, nil
}

func (in *Device) GetKeyBits() (*keycode.KeyBits, error) {
	var bits keycode.KeyBits
	err := ioc.Ioctl(int(in.f.Fd()), EVIOCGBIT(uint(eventcode.EV_KEY), uint(len(bits))), uintptr(unsafe.Pointer(&bits[0])))
	if err != nil {
		return nil, err
	}
	return &bits, nil
}

func (in *Device) GetKeyStates(bits []byte) error {
	err := ioc.Ioctl(int(in.f.Fd()), EVIOCGKEY(uint(len(bits))), uintptr(unsafe.Pointer(&bits[0])))
	if err != nil {
		return err
	}
	return nil
}

func (in *Device) Grab() error {
	return unix.IoctlSetInt(int(in.f.Fd()), uint(EVIOCGRAB), 1)
}

func (in *Device) Ungrab() error {
	return unix.IoctlSetInt(int(in.f.Fd()), uint(EVIOCGRAB), 0)
}

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

type InputID struct {
	BusType uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

func (id *InputID) Marshal() []byte {
	b := [8]byte{}
	binary.LittleEndian.PutUint16(b[:], id.BusType)
	binary.LittleEndian.PutUint16(b[2:], id.Vendor)
	binary.LittleEndian.PutUint16(b[4:], id.Product)
	binary.LittleEndian.PutUint16(b[6:], id.Version)
	return b[:]
}
