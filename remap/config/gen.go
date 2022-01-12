package config

//go:generate protoc -I.:../../evdev/keycode --go_out=module=erdi.us/chromekey:../.. config.proto
