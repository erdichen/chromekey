package uinput

import (
	"bytes"
	"encoding/binary"
	"log"
	"os"
	"syscall"
	"unsafe"

	"erdi.us/chromekey/evdev"
	"erdi.us/chromekey/evdev/eventcode"
	"erdi.us/chromekey/evdev/keycode"
	"erdi.us/chromekey/ioc"
	"golang.org/x/sys/unix"
)

const UInputMaxNameSize = 80

const (
	UINPUT_IOCTL_BASE = 'U'
)

var (
	UI_DEV_CREATE  = ioc.IO(UINPUT_IOCTL_BASE, 1)
	UI_DEV_DESTROY = ioc.IO(UINPUT_IOCTL_BASE, 2)

	UI_DEV_SETUP = ioc.IOW(UINPUT_IOCTL_BASE, 3, 8+80+4)

	// UI_ABS_SETUP            iow(UINPUT_IOCTL_BASE, 4, struct uinput_abs_setup)

	UI_SET_EVBIT   = ioc.IOW(UINPUT_IOCTL_BASE, 100, 4)
	UI_SET_KEYBIT  = ioc.IOW(UINPUT_IOCTL_BASE, 101, 4)
	UI_SET_RELBIT  = ioc.IOW(UINPUT_IOCTL_BASE, 102, 4)
	UI_SET_ABSBIT  = ioc.IOW(UINPUT_IOCTL_BASE, 103, 4)
	UI_SET_MSCBIT  = ioc.IOW(UINPUT_IOCTL_BASE, 104, 4)
	UI_SET_LEDBIT  = ioc.IOW(UINPUT_IOCTL_BASE, 105, 4)
	UI_SET_SNDBIT  = ioc.IOW(UINPUT_IOCTL_BASE, 106, 4)
	UI_SET_FFBIT   = ioc.IOW(UINPUT_IOCTL_BASE, 107, 4)
	UI_SET_PHYS    = ioc.IOW(UINPUT_IOCTL_BASE, 108, 8)
	UI_SET_SWBIT   = ioc.IOW(UINPUT_IOCTL_BASE, 109, 4)
	UI_SET_PROPBIT = ioc.IOW(UINPUT_IOCTL_BASE, 110, 4)
)

type Setup struct {
	ID   evdev.InputID
	Name [UInputMaxNameSize]byte

	FFEffectsMax uint32
}

func (dev *Setup) Marshal() []byte {
	b := &bytes.Buffer{}
	b.Write(dev.ID.Marshal())
	b.Write(dev.Name[:])
	binary.Write(b, binary.LittleEndian, &dev.FFEffectsMax)
	return b.Bytes()
}

type Device struct {
	f *os.File
}

func CreateDevice(device string, keyBits *keycode.KeyBits) (*Device, error) {
	f, err := os.OpenFile(device, os.O_RDWR|unix.O_NONBLOCK, 0644)
	if err != nil {
		return nil, err
	}
	evTypes := []eventcode.EventType{eventcode.EV_SYN, eventcode.EV_KEY, eventcode.EV_MSC}
	for _, t := range evTypes {
		if err := unix.IoctlSetInt(int(f.Fd()), UI_SET_EVBIT, int(t)); err != nil {
			return nil, err
		}
	}

	if err := unix.IoctlSetInt(int(f.Fd()), UI_SET_MSCBIT, int(eventcode.MSC_SCAN)); err != nil {
		return nil, err
	}

	keyBits.Set(keycode.KEY_BACK, true)
	keyBits.Set(keycode.KEY_FORWARD, true)
	keyBits.Set(keycode.KEY_REFRESH, true)
	keyBits.Set(keycode.KEY_SEARCH, true)
	keyBits.Set(keycode.KEY_BRIGHTNESSDOWN, true)
	keyBits.Set(keycode.KEY_BRIGHTNESSUP, true)
	keyBits.Set(keycode.KEY_KBDILLUMDOWN, true)
	keyBits.Set(keycode.KEY_KBDILLUMUP, true)
	keyBits.Set(keycode.KEY_MUTE, true)
	keyBits.Set(keycode.KEY_VOLUMEDOWN, true)
	keyBits.Set(keycode.KEY_VOLUMEUP, true)

	for i, b := range keyBits {
		for j := 0; j < 8; j++ {
			if (b & (1 << j)) != 0 {
				k := i*8 + j
				// log.Printf("Event code %v\n", keycode.Key(k))
				if err := unix.IoctlSetInt(int(f.Fd()), UI_SET_KEYBIT, k); err != nil {
					return nil, err
				}
			}
		}
	}

	setup := Setup{
		ID: evdev.InputID{BusType: 3, Vendor: 1, Product: 1, Version: 9999},
	}
	copy(setup.Name[:], "Chromebook keyboard remap")
	_, _, e1 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(f.Fd()), uintptr(UI_DEV_SETUP), uintptr(unsafe.Pointer(&setup)))
	if e1 != 0 {
		return nil, e1
	}

	if _, err := unix.IoctlRetInt(int(f.Fd()), UI_DEV_CREATE); err != nil {
		return nil, err
	}

	return &Device{f: f}, nil
}

func CreateFromDevice(device string, in *evdev.Device) (*Device, error) {
	bits, err := in.GetKeyBits()
	if err != nil {
		log.Fatal(err)
	}
	return CreateDevice(device, bits)
}

func (dev *Device) Close() error {
	if _, err := unix.IoctlRetInt(int(dev.f.Fd()), UI_DEV_DESTROY); err != nil {
		return err
	}
	return dev.f.Close()
}

func (dev *Device) WriteEvents(events []evdev.InputEvent) error {
	b := make([]byte, evdev.EventSize*len(events))
	for i, e := range events {
		copy(b[i*evdev.EventSize:], e.Marshal())
	}
	_, err := dev.f.Write(b)
	return err
}
