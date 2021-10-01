package keybuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SaveKeyStroke(t *testing.T) {
	tests := []struct {
		inp   string
		size  int
		read  int
		write int
	}{
		{"f", 1, 0, 0},
		{"longer", 6, 0, 0},
		{"wrap test", 9, ringsize - 4, ringsize - 4},
		{"buffer full test", ringsize - 1, 1, 0},
	}

	for _, tt := range tests {
		read = tt.read
		write = tt.write
		SaveKeyStroke([]byte(tt.inp))
		if size() != tt.size {
			t.Errorf("test %s failed, got %d wanted %d", tt.inp, write-read, tt.size)
		}
	}
}

func Test_SawBreak(t *testing.T) {
	var tt []byte
	tt = append(tt, 0x03)
	SaveKeyStroke(tt)

	assert.True(t, sig_break, "Ctrl-C missed!")
	assert.True(t, BreakSeen(), "Break not seen")
	ClearBreak()
	assert.False(t, BreakSeen(), "Flag not reset")
}

func Test_ReadByte(t *testing.T) {
	tests := []struct {
		inp   []byte
		fail  bool
		read  int
		write int
	}{
		{[]byte("foo"), false, 0, 0},
		{nil, true, 0, 0},
		{[]byte("wrap test"), false, ringsize - 4, ringsize - 4},
	}

	for _, tt := range tests {
		read = tt.read
		write = tt.write
		SaveKeyStroke(tt.inp)

		for i := range tt.inp {
			bt, ok := ReadByte()

			if !ok && tt.fail {
				break
			}

			if bt != tt.inp[i] {
				t.Errorf("Test %s failed, got %x wanted %x", string(tt.inp), bt, tt.inp[i])
			}
		}

		// make sure I've drained the buffer

		bt, ok := ReadByte()

		if ok {
			t.Errorf("Test %s got more than I wrote %x", string(tt.inp), bt)
		}
	}
}
