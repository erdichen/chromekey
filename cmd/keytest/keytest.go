// Binary keytest sends all keycodes to the virtual uinput device and reads them from the corresponding virtual evdev device for comparison.
package main

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/erdichen/chromekey/evdev"
	"github.com/erdichen/chromekey/evdev/keycode"
	"github.com/erdichen/chromekey/log"
	"github.com/erdichen/chromekey/remap"
	"github.com/erdichen/chromekey/uinput"
)

func main() {
	var keyBits keycode.KeyBits
	for key := keycode.Code_KEY_ESC; key < keycode.Code_KEY_MAX; key++ {
		keyBits.Set(key, true)
	}

	out, err := uinput.CreateDevice("/dev/uinput", &keyBits)
	if err != nil {
		log.Fatalf("faied to open uinput device: %v", err)
	}
	defer out.Close()

	time.Sleep(1 * time.Second)

	in, err := openInputDevice("/dev/input")
	if err != nil {
		log.Fatalf("faied to open evdev device: %v", err)
	}
	defer in.Close()

	// Grab the virtual device, otherwise the keys will do funny things!
	if err := in.Grab(); err != nil {
		log.Fatalf("faied to grab evdev device: %v", err)
	}
	defer in.Ungrab()

	ctx := context.Background()

	bad := 0
	matched := 0
	for key := keycode.Code_KEY_ESC; key < keycode.Code_KEY_MAX; key++ {
		if !keyBits.Get(key) {
			continue
		}

		want := remap.GenKey(key, 1)
		if out.WriteEvents(want); err != nil {
			log.Fatalf("failed to write events: %v", err)
		}
		got, err := in.ReadEvents(ctx)
		if err != nil {
			log.Fatalf("failed to read events: %v", err)
		}
		for i, v := range got {
			if !cmpEvent(v, want[i]) {
				log.Infof("EVENTS: got %#v\n, want %#v\n", v, want[i])
				bad++
			} else {
				matched++
			}
		}

		want = remap.GenKey(key, 0)
		if out.WriteEvents(want); err != nil {
			log.Fatalf("failed to write events: %v", err)
		}
		got, err = in.ReadEvents(ctx)
		if err != nil {
			log.Fatalf("failed to read events: %v", err)
		}
		for i, v := range got {
			if !cmpEvent(v, want[i]) {
				log.Infof("EVENTS: got %#v\n, want %#v\n", v, want[i])
				bad++
			} else {
				matched++
			}
		}
	}
	if bad == 0 {
		log.Errorf("Tested %d matched %d mismatch %d", matched+bad, matched, bad)
	} else {
		log.Infof("Tested %d matched %d mismatch %d", matched+bad, matched, bad)
	}
}

// cmpEvent compares two InputEvents for equality while ignoring the event time.
func cmpEvent(a, b evdev.InputEvent) bool {
	return a.Type == b.Type && a.Code == b.Code && a.Value == b.Value
}

// openInputDevice returns the virtual input device for testing.
func openInputDevice(devDir string) (*evdev.Device, error) {
	files, err := ioutil.ReadDir(devDir)
	if err != nil {
		return nil, err
	}

	for _, v := range files {
		if !strings.HasPrefix(v.Name(), "event") {
			continue
		}
		file := filepath.Join(devDir, v.Name())
		dev, e := evdev.OpenDevice(file)
		if e != nil {
			err = e
			continue
		}
		name, e := dev.GetName()
		if err != nil {
			err = e
			continue
		}
		log.Infof("Device name: %v", name)
		if strings.Contains(name, "Chromebook") {
			log.Infof("Opened keyboard input device: %v,", file)
			return dev, nil
		}
		log.Infof("Skipped non-test input device: %v,", file)
		if err := dev.Close(); err != nil {
			log.Errorf("failed to close an evdev device: %v", err)
		}
	}

	if err == nil {
		err = errors.New("found no input device")
	}

	return nil, err
}
