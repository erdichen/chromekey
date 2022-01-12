package config

import keycode "erdi.us/chromekey/evdev/keycode"

func FromPBKeymap(from []*KeymapEntry) map[keycode.Code]keycode.Code {
	to := make(map[keycode.Code]keycode.Code)
	for _, v := range from {
		to[keycode.Code(v.From)] = keycode.Code(v.To)
	}
	return to
}

func ToPBKeymap(from map[keycode.Code]keycode.Code) (to []*KeymapEntry) {
	for k, v := range from {
		e := &KeymapEntry{
			From: keycode.Code(k),
			To:   keycode.Code(v),
		}
		to = append(to, e)
	}
	return
}

type RunConfig struct {
	FnENabled   bool                          `json:"fn_enabled"`
	FnKey       keycode.Code                  `json:"fn_key"`
	KeyMap      map[keycode.Code]keycode.Code `json:"key_map"`
	ShiftKeyMap map[keycode.Code]keycode.Code `json:"shift_key_map"`
}

func (cfg RunConfig) Clone() RunConfig {
	keyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.KeyMap {
		keyMap[k] = v
	}
	shiftKeyMap := make(map[keycode.Code]keycode.Code)
	for k, v := range cfg.ShiftKeyMap {
		shiftKeyMap[k] = v
	}
	return RunConfig{
		FnENabled:   cfg.FnENabled,
		FnKey:       cfg.FnKey,
		KeyMap:      keyMap,
		ShiftKeyMap: shiftKeyMap,
	}
}

func FromPBConfig(pb *KeymapConfig) RunConfig {
	return RunConfig{
		FnENabled:   pb.FnEnabled,
		FnKey:       pb.FnKey,
		KeyMap:      FromPBKeymap(pb.KeyMap),
		ShiftKeyMap: FromPBKeymap(pb.ShiftKeyMap),
	}
}

func ToPBConfig(cfg RunConfig) *KeymapConfig {
	pb := KeymapConfig{
		FnEnabled:   cfg.FnENabled,
		FnKey:       cfg.FnKey,
		KeyMap:      ToPBKeymap(cfg.KeyMap),
		ShiftKeyMap: ToPBKeymap(cfg.ShiftKeyMap),
	}
	return &pb
}

func DefaultRunConfig() RunConfig {
	return RunConfig{
		FnKey:       keycode.Code_KEY_F13,
		KeyMap:      defaultFnKeyMap(),
		ShiftKeyMap: defaultShiftKeyMap(),
	}
}
