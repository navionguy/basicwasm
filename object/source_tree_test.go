package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SourceLine(t *testing.T) {
	tests := []struct {
		src  string
		lnum uint16
	}{
		{`10 Print"Hello World!"`, 10},
	}

	for _, tt := range tests {
		sl := NewSourceLine(tt.src, tt.lnum)

		assert.NotNil(t, sl)
		assert.EqualValues(t, tt.src, sl.source)
		assert.EqualValues(t, sl.source, sl.Inspect())
	}
}

func Test_SourceLineLess(t *testing.T) {
	tests := []struct {
		src1  string
		lnum1 uint16
		src2  string
		lnum2 uint16
		less  bool
	}{
		{`10 PRINT "Hello World"`, 10, `20 REM Comment`, 20, true},
		{`20 REM Comment`, 20, `10 PRINT "Hello World"`, 10, false},
	}

	for _, tt := range tests {
		l1 := NewSourceLine(tt.src1, tt.lnum1)
		l2 := NewSourceLine(tt.src2, tt.lnum2)

		assert.Equal(t, tt.less, l1.Less(l2))
	}
}

func Test_AddSourceLine(t *testing.T) {
	tests := []struct {
		src  string
		line uint16
	}{
		{src: "10 REM A simple program", line: 10},
		{src: `20 PRINT "Hello World!"`, line: 20},
	}

	st := initSourceTree()
	for i, tt := range tests {

		assert.NotNil(t, st)
		sl := NewSourceLine(tt.src, tt.line)
		st.addSourceLine(sl)

		assert.Equal(t, i+1, st.tree.Len())
	}
}

func Test_InitSourceTree(t *testing.T) {
	st := initSourceTree()

	assert.NotNil(t, st)
}
