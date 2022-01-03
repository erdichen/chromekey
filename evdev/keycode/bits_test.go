package keycode

import (
	"testing"
)

func TestKeyBits(t *testing.T) {
	var keys KeyBits
	for k := KEY_ESC; k < KEY_CNT; k++ {
		keys.Set(k, true)
		if got := keys.Get(k); !got {
			t.Errorf("Set key %v got %v want %v", k, got, true)
		}
	}
	for k := KEY_ESC; k < KEY_CNT; k++ {
		keys.Set(k, false)
		if got := keys.Get(k); got {
			t.Errorf("Set key %v got %v want %v", k, got, false)
		}
	}
}
