package remap

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/erdichen/chromekey/evdev"
	"github.com/erdichen/chromekey/evdev/eventcode"
	"github.com/erdichen/chromekey/evdev/keycode"
	"github.com/erdichen/chromekey/log"
	"github.com/erdichen/chromekey/remap/config"
	"github.com/erdichen/chromekey/uinput"
)

// State is the data of a key remapper that simulates the FN key that can remap function keys to media keys.
type State struct {
	in  *evdev.Device
	out *uinput.Device
	evC chan []evdev.InputEvent

	fnEnable bool
	lastKey  keycode.Code
	keys     keycode.KeyBits

	cfg config.RunConfig
}

// New returns new a key remapper.
func New(ctx context.Context, in *evdev.Device, outputDev string, cfg config.RunConfig, grab bool) (*State, error) {
	ok := false

	defer func() {
		if !ok {
			in.Close()
		}
	}()

	// Grabbing an input device will cause any pressed key to stuck in the pressed state.
	if err := waitForAllKeysReleased(in); err != nil {
		return nil, err
	}

	if grab {
		if err := in.Grab(); err != nil {
			return nil, err
		}
	}

	// Create an virtual device that replicates the capabilities and keys of the give input device.
	out, err := uinput.CreateFromDevice(outputDev, in)
	if err != nil {
		return nil, err
	}
	defer func() {
		if !ok {
			out.Close()
		}
	}()

	ok = true
	return &State{in: in, out: out, fnEnable: cfg.FnEnabled, cfg: cfg}, nil
}

// Close closes a remapper and its input and output devices.
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

// Config returns a copy of a remapper's RunConfig.
func (s *State) Config() config.RunConfig {
	return s.cfg.Clone()
}

// SetConfig load a RunConfig into a remapper's internal state.
func (s *State) SetConfig(cfg config.RunConfig) {
	s.cfg = cfg.Clone()
	s.fnEnable = cfg.FnEnabled
}

// Start runs the execution loop that forwards input events from the real keyboard to the virtual keyboard, remapping keys when necessary.
func (s *State) Start(ctx context.Context, sigC chan os.Signal, evC chan []evdev.InputEvent, timeout time.Duration) error {
	t := time.NewTimer(timeout)
	if timeout == 0 {
		t.Stop()
	}

	// Wait for the system to respond to the new input device.
	ledTimer := time.NewTimer(2 * time.Second)
	again := true

	done := false
	for !done {
		select {
		case sig := <-sigC:
			switch sig {
			case syscall.SIGTSTP:
				fmt.Printf("·Suspend breaks keyboard input. Press Ctrl-C to exit!\n")
			case syscall.SIGCONT:
			default:
				done = true
			}
		case <-ctx.Done():
			done = true
		case <-t.C:
			done = true
		case <-ledTimer.C:
			s.setFnLED()
			if again {
				// Set twice in case the desktop envinrment's state is out of sync with the hardware.
				again = false
				ledTimer.Reset(250 * time.Millisecond)
			}
		case events, ok := <-evC:
			if !ok {
				done = true
				break
			}
			events = s.handleEvents(events)
			if err := s.out.WriteEvents(events); err != nil {
				log.Errorf("failed to write events to uinput device: %v", err)
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

// SetVerbosity sets logging verbosity.
func SetVerbosity(v int) {
	verbosity = v
}

// genKey returns a sequence of input events that simulates a key press/release.
func GenKey(key keycode.Code, value int32) []evdev.InputEvent {
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

// setFnLED uses one the keyboard's LEDs to indicate FN key lock.
func (s *State) setFnLED() {
	key := keycode.Code_KEY_RESERVED
	switch s.cfg.UseLED {
	case keycode.LED_NUML:
		key = keycode.Code_KEY_NUMLOCK
	case keycode.LED_CAPSL:
		key = keycode.Code_KEY_CAPSLOCK
	case keycode.LED_SCROLLL:
		key = keycode.Code_KEY_SCROLLLOCK
	case keycode.LED_COMPOSE:
		key = keycode.Code_KEY_COMPOSE
	default:
		return
	}

	leds, err := s.in.GetLED()
	if err != nil {
		log.Errorf("failed get evdev device LED status: %v", err)
		return
	}
	if leds[s.cfg.UseLED] == s.fnEnable {
		return
	}

	v := 0
	if s.fnEnable {
		v = 1
	}

	events := GenKey(key, 1)
	events = append(events, GenKey(key, 0)...)
	events = append(events, evdev.InputEvent{
		Type:  uint16(eventcode.EV_LED),
		Code:  uint16(keycode.LED_NUML),
		Value: int32(v),
	})
	if err := s.out.WriteEvents(events); err != nil {
		log.Errorf("failed to write num lock key events: %v", err)
	}
}

// handleEvents converts key events to mapped key events if it matches the mapping rules.
func (s *State) handleEvents(events []evdev.InputEvent) []evdev.InputEvent {
	var pre, post []evdev.InputEvent
	for i, ev := range events {
		if verbosity > 1 {
			fmt.Printf("%s\n", ev.String())
		}
		switch eventcode.EventType(ev.Type) {
		case eventcode.EV_KEY:
			s.keys.Set(keycode.Code(ev.Code), ev.Value != 0)
			switch keycode.Code(ev.Code) {
			case s.cfg.FnKey:
				// Toogle FN key lock state only if it is pressed by itself.
				// Ignore these two cases:
				//   1. FN is last key released, but another key was released while FN is down.
				//   2. FN released with at least 1 key still down.
				if ev.Value == 0 && s.lastKey == s.cfg.FnKey && s.keys.IsZero() {
					s.fnEnable = !s.fnEnable
					if verbosity > 0 {
						log.Infof("FN %v", s.fnEnable)
					}
					s.setFnLED()
				}
				events[i].Code = uint16(keycode.Code_KEY_FN)
			default:
				isShiftDown := s.keys.Get(keycode.Code_KEY_LEFTSHIFT) || s.keys.Get(keycode.Code_KEY_RIGHTSHIFT)
				if isShiftDown {
					// Handle FN+Shift+ key map.
					if key, ok := s.cfg.ShiftKeyMap[keycode.Code(ev.Code)]; ok {
						fnDown := s.keys.Get(s.cfg.FnKey)
						if fnDown {
							// Clear the shift keys to simulate the mapped key with the shift keys released.
							if s.keys.Get(keycode.Code_KEY_LEFTSHIFT) {
								pre = append(pre, GenKey(keycode.Code_KEY_LEFTSHIFT, 0)...)
								post = append(post, GenKey(keycode.Code_KEY_LEFTSHIFT, 1)...)
							}
							if s.keys.Get(keycode.Code_KEY_RIGHTSHIFT) {
								pre = append(pre, GenKey(keycode.Code_KEY_RIGHTSHIFT, 0)...)
								post = append(post, GenKey(keycode.Code_KEY_RIGHTSHIFT, 1)...)
							}
							events[i].Code = uint16(key)
							if verbosity > 0 {
								log.Infof("shift map %v to %v", keycode.Code(ev.Code), key)
							}
						}
					}
				} else if s.keys.Get(s.cfg.FnKey) {
					if key, ok := s.cfg.ModKeyMap[keycode.Code(ev.Code)]; ok {
						events[i].Code = uint16(key)
						if verbosity > 0 {
							log.Infof("mod map %v to %v", keycode.Code(ev.Code), key)
						}
					} else if key, ok := s.cfg.KeyMap[keycode.Code(ev.Code)]; ok && !s.fnEnable {
						events[i].Code = uint16(key)
						if verbosity > 0 {
							log.Infof("fn mod map %v to %v", keycode.Code(ev.Code), key)
						}
					}
				} else if key, ok := s.cfg.KeyMap[keycode.Code(ev.Code)]; ok {
					if s.fnEnable {
						events[i].Code = uint16(key)
						if verbosity > 0 {
							log.Infof("fn map %v to %v", keycode.Code(ev.Code), key)
						}
					}
				}
			}
			s.lastKey = keycode.Code(ev.Code)
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

// StartReadEventsLoop loops reading input events and sends them to a channel.
func StartReadEventsLoop(ctx context.Context, in *evdev.Device) chan []evdev.InputEvent {
	evC := make(chan []evdev.InputEvent)
	go func(ctx context.Context) {
		defer close(evC)
		for {
			events, err := in.ReadEvents(ctx)
			if err != nil {
				log.Errorf("failed to read from evdev input: %v", err)
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

// waitForAllKeysReleased returns after all pressed keys have been released.
func waitForAllKeysReleased(in *evdev.Device) error {
	bits := make([]byte, (keycode.Code_KEY_CNT+7)/8)
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
