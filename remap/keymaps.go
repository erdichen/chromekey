package remap

import "erdi.us/chromekey/evdev/keycode"

func defaultFnKeyMap() map[keycode.Code]keycode.Code {
	return map[keycode.Code]keycode.Code{
		keycode.Code_KEY_F1:  keycode.Code_KEY_BACK,
		keycode.Code_KEY_F2:  keycode.Code_KEY_FORWARD,
		keycode.Code_KEY_F3:  keycode.Code_KEY_REFRESH,
		keycode.Code_KEY_F4:  keycode.Code_KEY_F11,
		keycode.Code_KEY_F5:  keycode.Code_KEY_SEARCH,
		keycode.Code_KEY_F6:  keycode.Code_KEY_BRIGHTNESSDOWN,
		keycode.Code_KEY_F7:  keycode.Code_KEY_BRIGHTNESSUP,
		keycode.Code_KEY_F8:  keycode.Code_KEY_MUTE,
		keycode.Code_KEY_F9:  keycode.Code_KEY_VOLUMEDOWN,
		keycode.Code_KEY_F10: keycode.Code_KEY_VOLUMEUP,

		keycode.Code_KEY_BRIGHTNESSDOWN: keycode.Code_KEY_KBDILLUMDOWN,
		keycode.Code_KEY_BRIGHTNESSUP:   keycode.Code_KEY_KBDILLUMUP,

		keycode.Code_KEY_VOLUMEDOWN: keycode.Code_KEY_KBDILLUMDOWN,
		keycode.Code_KEY_VOLUMEUP:   keycode.Code_KEY_KBDILLUMUP,
	}
}

func defaultShiftKeyMap() map[keycode.Code]keycode.Code {
	return map[keycode.Code]keycode.Code{
		keycode.Code_KEY_F6: keycode.Code_KEY_KBDILLUMDOWN,
		keycode.Code_KEY_F7: keycode.Code_KEY_KBDILLUMUP,

		keycode.Code_KEY_BRIGHTNESSDOWN: keycode.Code_KEY_KBDILLUMDOWN,
		keycode.Code_KEY_BRIGHTNESSUP:   keycode.Code_KEY_KBDILLUMUP,
	}
}
