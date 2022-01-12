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
