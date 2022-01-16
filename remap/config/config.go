package config

import (
	"sort"

	keycode "github.com/erdichen/chromekey/evdev/keycode"
)

// FromPBKeymap converts a slice of key map entry protos to a Go map.
func FromPBKeymap(from []*KeymapEntry) map[keycode.Code]keycode.Code {
	to := make(map[keycode.Code]keycode.Code)
	for _, v := range from {
		to[keycode.Code(v.From)] = keycode.Code(v.To)
	}
	return to
}

// ToPBKeymap converts a Go map to a slice of key map entry protos.
func ToPBKeymap(from map[keycode.Code]keycode.Code) (to []*KeymapEntry) {
	for k, v := range from {
		e := &KeymapEntry{
			From: keycode.Code(k),
			To:   keycode.Code(v),
		}
		to = append(to, e)
	}
	sort.SliceStable(to, func(i, j int) bool { return to[i].From < to[j].From })
	return
}

// RunConfig is the runtime key remap configuration. We do not use the KeymapConfig proto
// directly because protobuf does not support a map with enum keys.
type RunConfig struct {
	FnEnabled   bool                          `json:"fn_enabled"`
	FnKey       keycode.Code                  `json:"fn_key"`
	KeyMap      map[keycode.Code]keycode.Code `json:"key_map"`
	ModKeyMap   map[keycode.Code]keycode.Code `json:"mod_key_map"`
	ShiftKeyMap map[keycode.Code]keycode.Code `json:"shift_key_map"`
	UseLED      keycode.LED                   `json:"use_led"`
}

// Clone returns a deep copy of a RunConfig.
func (cfg RunConfig) Clone() RunConfig {
	keyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.KeyMap {
		keyMap[k] = v
	}
	shiftKeyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.ShiftKeyMap {
		shiftKeyMap[k] = v
	}
	modKeyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.ShiftKeyMap {
		modKeyMap[k] = v
	}
	rc := cfg
	rc.KeyMap = keyMap
	rc.ShiftKeyMap = shiftKeyMap
	rc.ModKeyMap = modKeyMap
	return rc
}

// FromPBConfig creates a RunConfig from a KeymapConfig proto.
func FromPBConfig(pb *KeymapConfig) RunConfig {
	rc := RunConfig{
		FnEnabled:   pb.FnEnabled,
		FnKey:       pb.FnKey,
		KeyMap:      FromPBKeymap(pb.KeyMap),
		ModKeyMap:   FromPBKeymap(pb.ModKeyMap),
		ShiftKeyMap: FromPBKeymap(pb.ShiftKeyMap),
	}
	if pb.UseLed != nil {
		rc.UseLED = pb.GetUseLed()
	}
	return rc
}

// FromPBConfig creates a KeymapConfig proto from a RunConfig.
func ToPBConfig(cfg RunConfig) *KeymapConfig {
	pb := KeymapConfig{
		FnEnabled:   cfg.FnEnabled,
		FnKey:       cfg.FnKey,
		KeyMap:      ToPBKeymap(cfg.KeyMap),
		ModKeyMap:   ToPBKeymap(cfg.ModKeyMap),
		ShiftKeyMap: ToPBKeymap(cfg.ShiftKeyMap),
	}
	if cfg.UseLED <= keycode.LED_MAX {
		pb.UseLed = &cfg.UseLED
	}
	return &pb
}

// DefaultRunConfig a default RunConfig with Chromebook media key mappings.
func DefaultRunConfig() RunConfig {
	return RunConfig{
		FnKey:       keycode.Code_KEY_F13,
		KeyMap:      defaultFnKeyMap(),
		ModKeyMap:   map[keycode.Code]keycode.Code{},
		ShiftKeyMap: defaultShiftKeyMap(),
	}
}
