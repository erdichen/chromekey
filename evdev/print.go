package evdev

import (
	"bytes"
	"fmt"

	"github.com/erdichen/chromekey/evdev/eventcode"
	"github.com/erdichen/chromekey/evdev/keycode"
)

// String returns a string representation of InputEvent similar to evtest's output.
func (ev *InputEvent) String() string {
	b := &bytes.Buffer{}
	fmt.Fprintf(b, "Event: time %d.%06d, ", ev.Sec, ev.Usec)
	if ev.Type == uint16(eventcode.EV_SYN) {
		if ev.Code == uint16(eventcode.SYN_MT_REPORT) {
			fmt.Fprintf(b, "++++++++++++++ %s ++++++++++++", codename(ev.Type, ev.Code))
		} else if ev.Code == uint16(eventcode.SYN_DROPPED) {
			fmt.Fprintf(b, ">>>>>>>>>>>>>> %s <<<<<<<<<<<<", codename(ev.Type, ev.Code))
		} else {
			fmt.Fprintf(b, "-------------- %s ------------", codename(ev.Type, ev.Code))
		}
	} else {
		fmt.Fprintf(b, "type %d (%s), code %d (%s), ", ev.Type, typename(ev.Type), ev.Code, codename(ev.Type, ev.Code))
		if ev.Type == uint16(eventcode.EV_MSC) && ev.Code == uint16(eventcode.MSC_RAW) || ev.Code == uint16(eventcode.MSC_SCAN) {
			fmt.Fprintf(b, "value %02x", ev.Value)
		} else {
			fmt.Fprintf(b, "value %d", ev.Value)
		}
	}
	return b.String()
}

func typename(typ uint16) string {
	return eventcode.EventType(typ).String()
}

func codename(typ uint16, code uint16) string {
	switch eventcode.EventType(typ) {
	case eventcode.EV_KEY:
		return keycode.Code(code).String()
	case eventcode.EV_MSC:
		return eventcode.MiscEvent(code).String()
	case eventcode.EV_SYN:
		return eventcode.SynEvent(code).String()
	default:
		return "?"
	}
}
