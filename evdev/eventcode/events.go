package eventcode

/*
 * Event types
 */

//go:generate stringer -type=EventType
type EventType int

const (
	EV_SYN       EventType = 0x00
	EV_KEY       EventType = 0x01
	EV_REL       EventType = 0x02
	EV_ABS       EventType = 0x03
	EV_MSC       EventType = 0x04
	EV_SW        EventType = 0x05
	EV_LED       EventType = 0x11
	EV_SND       EventType = 0x12
	EV_REP       EventType = 0x14
	EV_FF        EventType = 0x15
	EV_PWR       EventType = 0x16
	EV_FF_STATUS EventType = 0x17
	EV_MAX       EventType = 0x1f
	EV_CNT       EventType = (EV_MAX + 1)
)

/*
 * Synchronization events.
 */

//go:generate stringer -type=SynEvent
type SynEvent int

const (
	SYN_REPORT    SynEvent = 0
	SYN_CONFIG    SynEvent = 1
	SYN_MT_REPORT SynEvent = 2
	SYN_DROPPED   SynEvent = 3
	SYN_MAX       SynEvent = 0xf
	SYN_CNT       SynEvent = (SYN_MAX + 1)
)

/*
 * Misc events
 */

//go:generate stringer -type=MiscEvent
type MiscEvent int

const (
	MSC_SERIAL    MiscEvent = 0x00
	MSC_PULSELED  MiscEvent = 0x01
	MSC_GESTURE   MiscEvent = 0x02
	MSC_RAW       MiscEvent = 0x03
	MSC_SCAN      MiscEvent = 0x04
	MSC_TIMESTAMP MiscEvent = 0x05
	MSC_MAX       MiscEvent = 0x07
	MSC_CNT       MiscEvent = (MSC_MAX + 1)
)

/*
 * LEDs
 */

//go:generate stringer -type=LEDEvent
type LEDEvent int

const (
	LED_NUML     LEDEvent = 0x00
	LED_CAPSL    LEDEvent = 0x01
	LED_SCROLLL  LEDEvent = 0x02
	LED_COMPOSE  LEDEvent = 0x03
	LED_KANA     LEDEvent = 0x04
	LED_SLEEP    LEDEvent = 0x05
	LED_SUSPEND  LEDEvent = 0x06
	LED_MUTE     LEDEvent = 0x07
	LED_MISC     LEDEvent = 0x08
	LED_MAIL     LEDEvent = 0x09
	LED_CHARGING LEDEvent = 0x0a
	LED_MAX      LEDEvent = 0x0f
	LED_CNT      LEDEvent = (LED_MAX + 1)
)
