package remap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"erdi.us/chromekey/evdev"
	"erdi.us/chromekey/evdev/eventcode"
	"erdi.us/chromekey/evdev/keycode"
	"erdi.us/chromekey/uinput"
)

type State struct {
	in  *evdev.Device
	out *uinput.Device
	evC chan []evdev.InputEvent

	grabbed  bool
	fnEnable bool
	fnKey    keycode.Key
	lastKey  keycode.Key
	keys     keycode.KeyBits
}

func New(ctx context.Context, inputDev, outputDev string, fnEnable bool, fnKey keycode.Key, grab bool) (*State, error) {
	ok := false

	in, err := evdev.OpenDevice(inputDev)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if !ok {
			in.Close()
		}
	}()

	if err := waitAllKeyReleased(in); err != nil {
		return nil, err
	}

	if grab {
		if err := in.Grab(); err != nil {
			return nil, err
		}
	}

	out, err := uinput.CreateFromDevice(outputDev, in)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			out.Close()
		}
	}()

	evC := startReadEventsLoop(ctx, in)

	ok = true
	return &State{in: in, out: out, evC: evC, fnEnable: fnEnable, fnKey: fnKey, grabbed: grab}, nil
}

func (s *State) Close() error {
	if err := s.in.Ungrab(); err != nil {
		return err
	}
	if err := s.in.Close(); err != nil {
		return err
	}
	if err := s.out.Close(); err != nil {
		return err
	}
	return nil
}

func (s *State) Start(ctx context.Context, sigC chan os.Signal, timeout time.Duration) error {
	t := time.NewTimer(timeout)
	if timeout == 0 {
		t.Stop()
	}

	done := false
	for !done {
		select {
		case <-sigC:
			done = true
		case <-ctx.Done():
			done = true
		case <-t.C:
			done = true
		case events := <-s.evC:
			events = s.handleEvents(events)
			if err := s.out.WriteEvents(events); err != nil {
				log.Print(err)
				break
			}
			if timeout > 0 {
				t.Reset(timeout)
			}
		}
	}
	return nil
}

var verbosity = 0

func SetVerbosity(v int) {
	verbosity = v
}

var fnKeys = map[keycode.Key]keycode.Key{
	keycode.KEY_F1:  keycode.KEY_BACK,
	keycode.KEY_F2:  keycode.KEY_FORWARD,
	keycode.KEY_F3:  keycode.KEY_REFRESH,
	keycode.KEY_F4:  keycode.KEY_F11,
	keycode.KEY_F5:  keycode.KEY_SEARCH,
	keycode.KEY_F6:  keycode.KEY_BRIGHTNESSDOWN,
	keycode.KEY_F7:  keycode.KEY_BRIGHTNESSUP,
	keycode.KEY_F8:  keycode.KEY_MUTE,
	keycode.KEY_F9:  keycode.KEY_VOLUMEDOWN,
	keycode.KEY_F10: keycode.KEY_VOLUMEUP,

	keycode.KEY_BRIGHTNESSDOWN: keycode.KEY_KBDILLUMDOWN,
	keycode.KEY_BRIGHTNESSUP:   keycode.KEY_KBDILLUMUP,

	keycode.KEY_VOLUMEDOWN: keycode.KEY_KBDILLUMDOWN,
	keycode.KEY_VOLUMEUP:   keycode.KEY_KBDILLUMUP,
}

var fnShiftKeys = map[keycode.Key]keycode.Key{
	keycode.KEY_F6: keycode.KEY_KBDILLUMDOWN,
	keycode.KEY_F7: keycode.KEY_KBDILLUMUP,

	keycode.KEY_BRIGHTNESSDOWN: keycode.KEY_KBDILLUMDOWN,
	keycode.KEY_BRIGHTNESSUP:   keycode.KEY_KBDILLUMUP,
}

func (s *State) genKey(key keycode.Key, value int32) []evdev.InputEvent {
	return []evdev.InputEvent{
		{
			Type:  uint16(eventcode.EV_MSC),
			Code:  uint16(eventcode.MSC_SCAN),
			Value: int32(key),
		},
		{
			Type:  uint16(eventcode.EV_KEY),
			Code:  uint16(key),
			Value: value,
		},
		{
			Type: uint16(eventcode.EV_SYN),
			Code: uint16(eventcode.SYN_REPORT),
		},
	}
}

func (s *State) handleEvents(events []evdev.InputEvent) []evdev.InputEvent {
	var pre, post []evdev.InputEvent
	for i, ev := range events {
		if verbosity > 1 {
			fmt.Printf("%s\n", eventString(&ev))
		}
		switch eventcode.EventType(ev.Type) {
		case eventcode.EV_KEY:
			s.keys.Set(keycode.Key(ev.Code), ev.Value != 0)
			switch keycode.Key(ev.Code) {
			case s.fnKey:
				if ev.Value == 0 && s.lastKey == s.fnKey {
					s.fnEnable = !s.fnEnable
					if verbosity > 0 {
						log.Printf("FN %v", s.fnEnable)
					}
				}
				events[i].Code = uint16(keycode.KEY_FN)
			default:
				isShiftDown := s.keys.Get(keycode.KEY_LEFTSHIFT) || s.keys.Get(keycode.KEY_RIGHTSHIFT)
				if isShiftDown {
					if key, ok := fnShiftKeys[keycode.Key(ev.Code)]; ok {
						fnDown := s.keys.Get(s.fnKey)
						if fnDown {
							// Clear the shift keys to simulate the mapped key with the shift keys released.
							if s.keys.Get(keycode.KEY_LEFTSHIFT) {
								pre = append(pre, s.genKey(keycode.KEY_LEFTSHIFT, 0)...)
								post = append(post, s.genKey(keycode.KEY_LEFTSHIFT, 1)...)
							}
							if s.keys.Get(keycode.KEY_RIGHTSHIFT) {
								pre = append(pre, s.genKey(keycode.KEY_RIGHTSHIFT, 0)...)
								post = append(post, s.genKey(keycode.KEY_RIGHTSHIFT, 1)...)
							}
							events[i].Code = uint16(key)
							if verbosity > 0 {
								log.Printf("Map %v to %v", keycode.Key(ev.Code), key)
							}
						}
					}
				} else if key, ok := fnKeys[keycode.Key(ev.Code)]; ok {
					fnDown := s.keys.Get(s.fnKey)
					if s.fnEnable != fnDown {
						events[i].Code = uint16(key)
						if verbosity > 0 {
							log.Printf("Map %v to %v", keycode.Key(ev.Code), key)
						}
					}
				}
			}
			s.lastKey = keycode.Key(ev.Code)
		}
	}
	if pre != nil {
		events = append(pre, events...)
	}
	if post != nil {
		events = append(events, post...)
	}
	return events
}

func startReadEventsLoop(ctx context.Context, in *evdev.Device) chan []evdev.InputEvent {
	evC := make(chan []evdev.InputEvent)
	go func(ctx context.Context) {
		for {
			events, err := in.ReadEvents(ctx)
			if err != nil {
				log.Print(err)
				break
			}
			select {
			case evC <- events:
			case <-ctx.Done():
				return
			}
		}
	}(ctx)

	return evC
}

func waitAllKeyReleased(in *evdev.Device) error {
	bits := make([]byte, (keycode.KEY_CNT+7)/8)
	for {
		if err := in.GetKeyStates(bits); err != nil {
			return err
		}
		hasDown := false
		for _, b := range bits {
			hasDown = hasDown || b != 0
		}
		if !hasDown {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}
