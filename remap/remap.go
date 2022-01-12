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

type KeymapConfig struct {
	FnENabled   bool                          `json:"fn_enabled"`
	FnKey       keycode.Code                  `json:"fn_key"`
	KeyMap      map[keycode.Code]keycode.Code `json:"key_map"`
	ShiftKeyMap map[keycode.Code]keycode.Code `json:"shift_key_map"`
}

func (cfg KeymapConfig) Clone() KeymapConfig {
	keyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.KeyMap {
		keyMap[k] = v
	}
	shiftKeyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.ShiftKeyMap {
		shiftKeyMap[k] = v
	}
	return KeymapConfig{
		FnENabled:   cfg.FnENabled,
		FnKey:       cfg.FnKey,
		KeyMap:      keyMap,
		ShiftKeyMap: shiftKeyMap,
	}
}

type State struct {
	in  *evdev.Device
	out *uinput.Device
	evC chan []evdev.InputEvent

	grabbed  bool
	fnEnable bool
	lastKey  keycode.Code
	keys     keycode.KeyBits

	cfg KeymapConfig
}

func New(ctx context.Context, inputDev, outputDev string, fnKey keycode.Code, grab bool) (*State, error) {
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

	if fnKey == keycode.Code_KEY_RESERVED {
		fnKey = keycode.Code_KEY_F13
	}

	ok = true
	return &State{in: in, out: out, evC: evC,
		grabbed:  grab,
		fnEnable: true,
		cfg: KeymapConfig{
			FnKey:       fnKey,
			KeyMap:      defaultFnKeyMap(),
			ShiftKeyMap: defaultShiftKeyMap(),
		},
	}, nil
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

func (s *State) Config() KeymapConfig {
	return s.cfg.Clone()
}

func (s *State) SetConfig(cfg KeymapConfig) {
	s.cfg = cfg.Clone()
	s.fnEnable = cfg.FnENabled
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

func (s *State) genKey(key keycode.Code, value int32) []evdev.InputEvent {
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
			s.keys.Set(keycode.Code(ev.Code), ev.Value != 0)
			switch keycode.Code(ev.Code) {
			case s.cfg.FnKey:
				if ev.Value == 0 && s.lastKey == s.cfg.FnKey {
					s.fnEnable = !s.fnEnable
					if verbosity > 0 {
						log.Printf("FN %v", s.fnEnable)
					}
				}
				events[i].Code = uint16(keycode.Code_KEY_FN)
			default:
				isShiftDown := s.keys.Get(keycode.Code_KEY_LEFTSHIFT) || s.keys.Get(keycode.Code_KEY_RIGHTSHIFT)
				if isShiftDown {
					if key, ok := s.cfg.ShiftKeyMap[keycode.Code(ev.Code)]; ok {
						fnDown := s.keys.Get(s.cfg.FnKey)
						if fnDown {
							// Clear the shift keys to simulate the mapped key with the shift keys released.
							if s.keys.Get(keycode.Code_KEY_LEFTSHIFT) {
								pre = append(pre, s.genKey(keycode.Code_KEY_LEFTSHIFT, 0)...)
								post = append(post, s.genKey(keycode.Code_KEY_LEFTSHIFT, 1)...)
							}
							if s.keys.Get(keycode.Code_KEY_RIGHTSHIFT) {
								pre = append(pre, s.genKey(keycode.Code_KEY_RIGHTSHIFT, 0)...)
								post = append(post, s.genKey(keycode.Code_KEY_RIGHTSHIFT, 1)...)
							}
							events[i].Code = uint16(key)
							if verbosity > 0 {
								log.Printf("Map %v to %v", keycode.Code(ev.Code), key)
							}
						}
					}
				} else if key, ok := s.cfg.KeyMap[keycode.Code(ev.Code)]; ok {
					fnDown := s.keys.Get(s.cfg.FnKey)
					if s.fnEnable != fnDown {
						events[i].Code = uint16(key)
						if verbosity > 0 {
							log.Printf("Map %v to %v", keycode.Code(ev.Code), key)
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
