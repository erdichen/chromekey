package keycode

import (
	"testing"
)

func TestKeyBits(t *testing.T) {
	var keys KeyBits
	for k := Code_KEY_ESC; k < Code_KEY_CNT; k++ {
		keys.Set(k, true)
		if got := keys.Get(k); !got {
			t.Errorf("Set key %v got %v want %v", k, got, true)
		}
	}
	for k := Code_KEY_ESC; k < Code_KEY_CNT; k++ {
		keys.Set(k, false)
		if got := keys.Get(k); got {
			t.Errorf("Set key %v got %v want %v", k, got, false)
		}
	}
}
