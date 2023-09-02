package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CreateFileStore(t *testing.T) {
	lf := CreateFileStore()

	assert.NotNil(t, lf, "CreateFileStore returned nil")
}

func Test_OpenFile(t *testing.T) {
	tests := []struct {
		inp  string
		inp2 string
		exp  bool
	}{
		{inp: "driveC/data.txt", inp2: "driveC/data.txt", exp: true},
		{inp: "driveC/data.txt", exp: false},
	}

	for _, tt := range tests {
		lf := CreateFileStore()

		if len(tt.inp2) > 0 {
			af := aFile{}
			lf.localFiles[tt.inp2] = &af
		}

		assert.Equal(t, tt.exp, lf.CheckLocal(tt.inp))

	}
}

func Test_CloseFIle(t *testing.T) {
	tests := []struct {
		start []int
		num   int
	}{
		{num: 0},
		{num: -1},
	}

	for _, tt := range tests {
		lf := CreateFileStore()

		for _, fn := range tt.start {
			f := aFile{}
			lf.openFiles[fn] = &f
		}

		lf.CloseFile(tt.num)
	}
}
