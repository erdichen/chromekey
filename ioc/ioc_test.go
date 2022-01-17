//go:build cgo

package ioc

import (
	"testing"
	"unsafe"

	"github.com/erdichen/chromekey/ioc/testdefs"
)

const ptrSize = uint(unsafe.Sizeof(uintptr(0)))

func TestIOC(t *testing.T) {
	testData := [...][2]uint{
		{IOW(testdefs.UINPUT_IOCTL_BASE, 100, 4), uint(testdefs.UI_SET_EVBIT)},      // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 101, 4), uint(testdefs.UI_SET_KEYBIT)},     // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 102, 4), uint(testdefs.UI_SET_RELBIT)},     // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 103, 4), uint(testdefs.UI_SET_ABSBIT)},     // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 104, 4), uint(testdefs.UI_SET_MSCBIT)},     // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 105, 4), uint(testdefs.UI_SET_LEDBIT)},     // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 106, 4), uint(testdefs.UI_SET_SNDBIT)},     // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 107, 4), uint(testdefs.UI_SET_FFBIT)},      // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 108, ptrSize), uint(testdefs.UI_SET_PHYS)}, // char*
		{IOW(testdefs.UINPUT_IOCTL_BASE, 109, 4), uint(testdefs.UI_SET_SWBIT)},      // int
		{IOW(testdefs.UINPUT_IOCTL_BASE, 110, 4), uint(testdefs.UI_SET_PROPBIT)},    // int
	}

	for i, d := range testData {
		got, want := d[0], d[1]
		if got != want {
			t.Errorf("test %d IOW got %#x want %#x", i, got, want)
		}
	}
}
