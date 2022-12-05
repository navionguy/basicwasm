package keybuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SaveKeyStroke(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{"f"},
		{"longer"},
	}

	for _, tt := range tests {
		bytes := []byte(tt.inp)
		bf := GetKeyBuffer()
		bf.SaveKeyStroke(bytes)

		assert.NotNil(t, bf.keycodes, "SaveKeyStroke failed to open channel")

		rc := <-bf.keycodes

		assert.EqualValuesf(t, bytes, rc, "SaveKeyStroke got %V wanted %V", rc, bytes)
	}
}

func Test_SawBreak(t *testing.T) {
	var tt []byte
	tt = append(tt, 0x03)
	buff := new(KeyBuffer)
	buff.SaveKeyStroke(tt)

	assert.True(t, buff.sig_break, "Ctrl-C missed!")
	assert.True(t, buff.BreakSeen(), "Break not seen")
	buff.ClearBreak()
	assert.False(t, buff.BreakSeen(), "Flag not reset")
}

func Test_ReadByte(t *testing.T) {
	tests := []struct {
		inp  []byte
		fail bool
	}{
		{[]byte("foo"), false},
		{nil, true},
		{[]byte("wrap test"), false},
	}

	for _, tt := range tests {
		buff := new(KeyBuffer)
		buff.SaveKeyStroke(tt.inp)

		for i := range tt.inp {
			bt, err := buff.ReadByte()

			if (err != nil) && tt.fail {
				break
			}

			if bt != tt.inp[i] {
				t.Errorf("Test %s failed, got %x wanted %x", string(tt.inp), bt, tt.inp[i])
			}
		}

		// make sure I've drained the buffer

		bt, err := buff.ReadByte()

		if err == nil {
			t.Errorf("Test %s got more than I wrote %x", string(tt.inp), bt)
		}
	}
}

func Test_EarlyRead(t *testing.T) {
	buff := new(KeyBuffer)
	bt, err := buff.ReadByte()

	if err == nil {
		assert.Failf(t, "An early ReadByte return %b", string([]byte{bt}))
	}
}
