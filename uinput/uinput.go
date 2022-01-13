package uinput

import (
	"bytes"
	"encoding/binary"
	"os"
	"syscall"
	"unsafe"

	"erdi.us/chromekey/evdev"
	"erdi.us/chromekey/evdev/eventcode"
	"erdi.us/chromekey/evdev/keycode"
	"erdi.us/chromekey/log"
	"golang.org/x/sys/unix"
)

// Setup is a uinput device setup request info struct.
type Setup struct {
	ID   evdev.InputID
	Name [MaxNameSize]byte

	FFEffectsMax uint32
}

// Marshal converts a Setup struct to binary bytes for ioctl use.
func (dev *Setup) Marshal() []byte {
	b := &bytes.Buffer{}
	b.Write(dev.ID.Marshal())
	b.Write(dev.Name[:])
	binary.Write(b, binary.LittleEndian, &dev.FFEffectsMax)
	return b.Bytes()
}

// Device is a virtual keyboard device.
type Device struct {
	f *os.File
}

// CreateDevice creates a virtual keyboard device with the keycodes set in keyBits.
func CreateDevice(device string, keyBits *keycode.KeyBits) (*Device, error) {
	f, err := os.OpenFile(device, os.O_RDWR|unix.O_NONBLOCK, 0644)
	if err != nil {
		return nil, err
	}
	evTypes := []eventcode.EventType{eventcode.EV_SYN, eventcode.EV_KEY, eventcode.EV_MSC, eventcode.EV_LED}
	for _, t := range evTypes {
		if err := unix.IoctlSetInt(int(f.Fd()), UI_SET_EVBIT, int(t)); err != nil {
			return nil, err
		}
	}

	if err := unix.IoctlSetInt(int(f.Fd()), UI_SET_MSCBIT, int(eventcode.MSC_SCAN)); err != nil {
		return nil, err
	}

	for i := 0; i < int(keycode.LED_CNT); i++ {
		if err := unix.IoctlSetInt(int(f.Fd()), UI_SET_LEDBIT, i); err != nil {
			return nil, err
		}
	}

	keyBits.Set(keycode.Code_KEY_BACK, true)
	keyBits.Set(keycode.Code_KEY_FORWARD, true)
	keyBits.Set(keycode.Code_KEY_REFRESH, true)
	keyBits.Set(keycode.Code_KEY_SEARCH, true)
	keyBits.Set(keycode.Code_KEY_BRIGHTNESSDOWN, true)
	keyBits.Set(keycode.Code_KEY_BRIGHTNESSUP, true)
	keyBits.Set(keycode.Code_KEY_KBDILLUMDOWN, true)
	keyBits.Set(keycode.Code_KEY_KBDILLUMUP, true)
	keyBits.Set(keycode.Code_KEY_MUTE, true)
	keyBits.Set(keycode.Code_KEY_VOLUMEDOWN, true)
	keyBits.Set(keycode.Code_KEY_VOLUMEUP, true)

	for i, b := range keyBits {
		for j := 0; j < 8; j++ {
			if (b & (1 << j)) != 0 {
				k := i*8 + j
				// log.Infof("event code %v\n", keycode.Code(k))
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

// CreateFromDevice creates a virtual keyboard devices that supports the same set of keycodes as an input device.
func CreateFromDevice(device string, in *evdev.Device) (*Device, error) {
	bits, err := in.GetKeyBits()
	if err != nil {
		log.Errorf("failed to get key bits from evdev device: %v", err)
		return nil, err
	}
	return CreateDevice(device, bits)
}

// Close closes the virtual keyboard devices.
func (dev *Device) Close() error {
	if _, err := unix.IoctlRetInt(int(dev.f.Fd()), UI_DEV_DESTROY); err != nil {
		return err
	}
	return dev.f.Close()
}

// WriteEvents writes input events to a virtual keyboard device.
func (dev *Device) WriteEvents(events []evdev.InputEvent) error {
	b := make([]byte, evdev.EventSize*len(events))
	for i, e := range events {
		copy(b[i*evdev.EventSize:], e.Marshal())
	}
	_, err := dev.f.Write(b)
	return err
}
