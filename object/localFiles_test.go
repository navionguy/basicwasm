package object

import (
	"testing"

	"github.com/navionguy/basicwasm/mocks"
	"github.com/stretchr/testify/assert"
)

func Test_CreateFileStore(t *testing.T) {
	lf := CreateFileStore()

	assert.NotNil(t, lf, "CreateFileStore returned nil")
}

func Test_OpenFile(t *testing.T) {
	tests := []struct {
		inp   string
		inp2  string
		lock2 LockMode
		fail  bool
	}{
		{inp: "driveC/data.txt", inp2: "driveC/data.txt"},
		{inp: "driveC/data.txt", inp2: "driveC/data.txt", lock2: LockRead},
		{inp: "driveC/data.txt", fail: true},
	}

	for _, tt := range tests {
		lf := CreateFileStore()

		if len(tt.inp2) > 0 {
			af := aFile{locked: tt.lock2}
			lf.localFiles[tt.inp2] = &af
		}

		var mt mocks.MockTerm
		//initMockTerm(&mt)
		env := NewTermEnvironment(mt)
		file, _ := lf.OpenLocalReadOnly(tt.inp, env)

		if tt.fail {
			assert.Nil(t, file, "OpenFile didn't fail")
		} else {
			assert.NotNil(t, file, "OpenFile didn't succeed")
		}

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
