package config

import "github.com/erdichen/chromekey/evdev/keycode"

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
	}
}

func defaultModKeyMap() map[keycode.Code]keycode.Code {
	return map[keycode.Code]keycode.Code{
		keycode.Code_KEY_MINUS:     keycode.Code_KEY_F11,
		keycode.Code_KEY_EQUAL:     keycode.Code_KEY_F12,
		keycode.Code_KEY_BACKSPACE: keycode.Code_KEY_DELETE,
		keycode.Code_KEY_DOT:       keycode.Code_KEY_INSERT,
		keycode.Code_KEY_UP:        keycode.Code_KEY_PAGEUP,
		keycode.Code_KEY_LEFT:      keycode.Code_KEY_HOME,
		keycode.Code_KEY_RIGHT:     keycode.Code_KEY_END,
		keycode.Code_KEY_DOWN:      keycode.Code_KEY_PAGEDOWN,
		keycode.Code_KEY_LEFTMETA:  keycode.Code_KEY_CAPSLOCK,
		keycode.Code_KEY_TAB:       keycode.Code_KEY_NUMLOCK,
	}
}

func defaultThirdLevelKeyMap() map[keycode.Code]keycode.Code {
	return map[keycode.Code]keycode.Code{
		keycode.Code_KEY_F6: keycode.Code_KEY_KBDILLUMDOWN,
		keycode.Code_KEY_F7: keycode.Code_KEY_KBDILLUMUP,
	}
}
