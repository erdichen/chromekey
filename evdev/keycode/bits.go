package keycode

import "fmt"

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

func (kb *KeyBits) IsZero() bool {
	for _, v := range kb {
		if v != 0 {
			return false
		}
	}
	return true
}

type BitField struct {
	Data  []byte
	count uint
}

func NewBitField(cnt uint) BitField {
	return BitField{
		Data:  make([]byte, (cnt+7)/8),
		count: cnt,
	}
}

func (bf *BitField) Get(k uint) (bool, error) {
	if k >= bf.count {
		return false, fmt.Errorf("bit index %d exceeded count %d", k, bf.count)
	}
	bits := bf.Data
	bytes := int(k) / 8
	v := byte(1 << (int(k) & 7))
	return (bits[bytes] & v) != 0, nil
}

func (bf *BitField) GetDefault(k uint, def bool) bool {
	if k >= bf.count {
		return def
	}
	bits := bf.Data
	bytes := int(k) / 8
	v := byte(1 << (int(k) & 7))
	return (bits[bytes] & v) != 0
}

func (bf *BitField) Set(k uint, b bool) error {
	if k >= bf.count {
		return fmt.Errorf("bit index %d exceeded count %d", k, bf.count)
	}
	bits := bf.Data
	bytes := int(k) / 8
	v := byte(1 << (int(k) & 7))
	if b {
		bits[bytes] |= v
	} else {
		bits[bytes] &= ^v
	}
	return nil
}
