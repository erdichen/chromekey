package evdev

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/erdichen/chromekey/evdev/eventcode"
	"github.com/erdichen/chromekey/evdev/keycode"
	"github.com/erdichen/chromekey/ioc"
	"github.com/erdichen/chromekey/log"
	"golang.org/x/sys/unix"
)

const PtrSize = uint(unsafe.Sizeof(uintptr(0)))

type InputEvent struct {
	Sec  uint
	Usec uint

	Type  uint16
	Code  uint16
	Value int32
}

func (ev *InputEvent) ToBytes() []byte {
	return (*(*[EventSize]byte)(unsafe.Pointer(ev)))[:]
}

type Device struct {
	f       *os.File
	grabbed bool
}

func OpenDevice(device string) (*Device, error) {
	f, err := os.OpenFile(device, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return &Device{f: f}, nil
}

func (in *Device) Close() error {
	f := in.f
	in.f = nil
	return f.Close()
}

func (in *Device) IsKeyboard() bool {
	if _, err := in.GetKeyBits(); err != nil {
		return false
	}
	if _, err := in.GetLED(); err != nil {
		return false
	}
	return true
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

func (in *Device) GetName() (string, error) {
	var buf [256]byte
	err := ioc.Ioctl(int(in.f.Fd()), EVIOCGNAME(uint(len(buf))), uintptr(unsafe.Pointer(&buf[0])))
	if err != nil {
		return "", err
	}
	sz := 0
	for i, v := range buf {
		if v == 0 {
			sz = i
			break
		}
	}
	return string(buf[:sz]), nil
}

func (in *Device) Grab() error {
	if in.grabbed {
		return nil
	}
	err := unix.IoctlSetInt(int(in.f.Fd()), uint(EVIOCGRAB), 1)
	if err != nil {
		return err
	}
	in.grabbed = true
	return nil
}

func (in *Device) Ungrab() error {
	if !in.grabbed {
		return nil
	}
	err := unix.IoctlSetInt(int(in.f.Fd()), uint(EVIOCGRAB), 0)
	if err != nil {
		return err
	}
	in.grabbed = false
	return nil
}

func (in *Device) GetLED() (map[keycode.LED]bool, error) {
	bits := keycode.NewBitField(uint(keycode.LED_CNT))
	err := ioc.Ioctl(int(in.f.Fd()), EVIOCGLED(uint(len(bits.Data))), uintptr(unsafe.Pointer(&bits.Data[0])))
	if err != nil {
		return nil, err
	}
	leds := map[keycode.LED]bool{}
	for i := 0; i < int(keycode.LED_CNT); i++ {
		if bits.GetDefault(uint(i), false) {
			leds[keycode.LED(i)] = true
		}
	}
	return leds, nil
}

func (in *Device) SetLED(code keycode.LED, on bool) error {
	v := int32(0)
	if on {
		v = 1
	}
	ev := InputEvent{
		Type:  uint16(eventcode.EV_LED),
		Code:  uint16(code),
		Value: v,
	}
	_, err := in.f.Write(ev.ToBytes())
	return err
}

type InputID struct {
	BusType uint16
	Vendor  uint16
	Product uint16
	Version uint16
}

// OpenByName opens event devices in a directory that contains kbdName.
func OpenByName(devDir string, kbdName string, verbosity int) (*Device, error) {
	files, err := ioutil.ReadDir(devDir)
	if err != nil {
		return nil, err
	}

	var file string
	var name string
	var dev *Device
	var errList []error

	for _, v := range files {
		if !strings.HasPrefix(v.Name(), "event") {
			continue
		}
		file = filepath.Join(devDir, v.Name())
		d, err := OpenDevice(file)
		if err != nil {
			errList = append(errList, err)
			continue
		}
		// defer d.Close()
		name, err = d.GetName()
		if err != nil {
			errList = append(errList, err)
			continue
		}
		if kbdName != "" {
			// Match by name
			if strings.Contains(name, kbdName) {
				dev = d
				d = nil
				break
			}
		} else if d.IsKeyboard() {
			// Or use first keyboard device
			dev = d
			d = nil
			break
		}
		if verbosity > 1 {
			log.Infof("Skipped non-keyboard input device: %v : %v", file, name)
		}
		if err := d.Close(); err != nil {
			log.Errorf("failed to close an evdev device: %v", err)
		}
	}

	if dev == nil {
		if len(errList) == 0 {
			return nil, errors.New("found no input device")
		}
		return nil, fmt.Errorf("found no input device: %v", errList)
	}

	if verbosity > 1 {
		log.Infof("Opened keyboard input device: %v : %v", file, name)
	}
	return dev, nil
}
