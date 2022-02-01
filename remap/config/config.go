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
	FnEnabled        bool                          `json:"fn_enabled"`
	FnKey            keycode.Code                  `json:"fn_key"`
	KeyMap           map[keycode.Code]keycode.Code `json:"key_map"`
	ModKeyMap        map[keycode.Code]keycode.Code `json:"mod_key_map"`
	ThirdLevelKeyMap map[keycode.Code]keycode.Code `json:"third_level_key_map"`
	UseLED           keycode.LED                   `json:"use_led"`
	ThirdLevelKey    []keycode.Code                `json:"third_level_key"`
}

// Clone returns a deep copy of a RunConfig.
func (cfg RunConfig) Clone() RunConfig {
	keyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.KeyMap {
		keyMap[k] = v
	}
	thirdLevelKeyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.ThirdLevelKeyMap {
		thirdLevelKeyMap[k] = v
	}
	modKeyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.ThirdLevelKeyMap {
		modKeyMap[k] = v
	}
	rc := cfg
	rc.KeyMap = keyMap
	rc.ThirdLevelKeyMap = thirdLevelKeyMap
	rc.ModKeyMap = modKeyMap
	return rc
}

// FromPBConfig creates a RunConfig from a KeymapConfig proto.
func FromPBConfig(pb *KeymapConfig) RunConfig {
	rc := RunConfig{
		FnEnabled:        pb.FnEnabled,
		FnKey:            pb.FnKey,
		KeyMap:           FromPBKeymap(pb.KeyMap),
		ModKeyMap:        FromPBKeymap(pb.ModKeyMap),
		ThirdLevelKeyMap: FromPBKeymap(pb.ThirdLevelKeyMap),
	}
	if pb.UseLed != nil {
		rc.UseLED = pb.GetUseLed()
	}
	rc.ThirdLevelKey = append([]keycode.Code{}, pb.GetThirdLevelKey()...)
	return rc
}

// FromPBConfig creates a KeymapConfig proto from a RunConfig.
func ToPBConfig(cfg RunConfig) *KeymapConfig {
	pb := KeymapConfig{
		FnEnabled:        cfg.FnEnabled,
		FnKey:            cfg.FnKey,
		KeyMap:           ToPBKeymap(cfg.KeyMap),
		ModKeyMap:        ToPBKeymap(cfg.ModKeyMap),
		ThirdLevelKeyMap: ToPBKeymap(cfg.ThirdLevelKeyMap),
	}
	if cfg.UseLED <= keycode.LED_MAX {
		useLED := cfg.UseLED
		pb.UseLed = &useLED
	}
	pb.ThirdLevelKey = append([]keycode.Code{}, cfg.ThirdLevelKey...)
	return &pb
}

// DefaultRunConfig a default RunConfig with Chromebook media key mappings.
func DefaultRunConfig() RunConfig {
	return RunConfig{
		FnKey:            keycode.Code_KEY_F13,
		KeyMap:           defaultFnKeyMap(),
		ModKeyMap:        defaultModKeyMap(),
		ThirdLevelKeyMap: defaultThirdLevelKeyMap(),
		ThirdLevelKey:    []keycode.Code{keycode.Code_KEY_LEFTSHIFT, keycode.Code_KEY_RIGHTSHIFT},
	}
}
