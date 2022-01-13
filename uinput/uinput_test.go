package uinput

import (
	"testing"
	"unsafe"
)

func TestUinput(t *testing.T) {
	got := unsafe.Sizeof(Setup{})
	want := uintptr(setupStructSize)
	if got != want {
		t.Errorf("bad uinput_setup size: got %d want %d", got, want)
	}
}
