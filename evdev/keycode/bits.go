package keycode

type KeyBits [(Code_KEY_CNT + 7) / 8]byte

func (kb *KeyBits) Get(k Code) bool {
	bytes := int(k) / 8
	if len(kb) < bytes {
		return false
	}
	v := byte(1 << (int(k) & 7))
	return (kb[bytes] & v) != 0
}

func (kb *KeyBits) Set(k Code, b bool) {
	bytes := int(k) / 8
	if len(kb) < bytes {
		return
	}
	v := byte(1 << (int(k) & 7))
	if b {
		kb[bytes] |= v
	} else {
		kb[bytes] &= ^v
	}
}
