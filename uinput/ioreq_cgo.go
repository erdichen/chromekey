//go:build cgo

package uinput

/*
#include <linux/uinput.h>

int uinput_setup_size = sizeof(struct uinput_setup);
*/
import "C"
import "unsafe"

const MaxNameSize = C.UINPUT_MAX_NAME_SIZE

var setupStructSize = C.uinput_setup_size

var (
	UI_DEV_CREATE  = uint(C.UI_DEV_CREATE)
	UI_DEV_DESTROY = uint(C.UI_DEV_DESTROY)

	UI_DEV_SETUP = uint(C.UI_DEV_SETUP)

	UI_SET_EVBIT   = uint(C.UI_SET_EVBIT)
	UI_SET_KEYBIT  = uint(C.UI_SET_KEYBIT)
	UI_SET_RELBIT  = uint(C.UI_SET_RELBIT)
	UI_SET_ABSBIT  = uint(C.UI_SET_ABSBIT)
	UI_SET_MSCBIT  = uint(C.UI_SET_MSCBIT)
	UI_SET_LEDBIT  = uint(C.UI_SET_LEDBIT)
	UI_SET_SNDBIT  = uint(C.UI_SET_SNDBIT)
	UI_SET_FFBIT   = uint(C.UI_SET_FFBIT)
	UI_SET_PHYS    = uint(C.UI_SET_PHYS)
	UI_SET_SWBIT   = uint(C.UI_SET_SWBIT)
	UI_SET_PROPBIT = uint(C.UI_SET_PROPBIT)
)

func structSizeMismatch()

func init() {
	const sz = int(unsafe.Sizeof(*(*Setup)(unsafe.Pointer(uintptr(0)))))
	if sz != (8 + MaxNameSize + 4) {
		structSizeMismatch()
	}
}
