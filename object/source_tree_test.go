package object

import (
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/token"
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

func Test_LineLength(t *testing.T) {
	tst := ast.LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	}

	sl := NewSourceLine(tst.Token.Literal, 10)
	sl.AppendStatement(&tst)

	assert.EqualValues(t, 1, sl.LineLength())
}

func Test_NextStatement(t *testing.T) {
	tst := ast.LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	}
	tst2 := ast.LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "20"},
		Value: 20,
	}

	sl := NewSourceLine(tst.Token.Literal, 10)
	sl.AppendStatement(&tst)
	sl.AppendStatement(&tst2)

	assert.EqualValues(t, 2, sl.LineLength())

	r1 := sl.NextStatement()
	v1, ok := r1.(*ast.LineNumStmt)
	assert.True(t, ok)
	assert.EqualValues(t, tst.Value, v1.Value)

	r2 := sl.NextStatement()
	v2, ok := r2.(*ast.LineNumStmt)
	assert.True(t, ok)
	assert.EqualValues(t, tst2.Value, v2.Value)

	// check for end detection
	r3 := sl.NextStatement()

	assert.Nil(t, r3)
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

	st := InitSourceTree()
	for i, tt := range tests {

		assert.NotNil(t, st)
		sl := NewSourceLine(tt.src, tt.line)
		st.AddSourceLine(sl)

		assert.Equal(t, i+1, st.tree.Len())
	}
}

func Test_InitSourceTree(t *testing.T) {
	st := InitSourceTree()

	assert.NotNil(t, st)
}

func Test_JumpToLine(t *testing.T) {
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
		st := InitSourceTree()
		l1 := NewSourceLine(tt.src1, tt.lnum1)
		st.AddSourceLine(l1)

		l2 := NewSourceLine(tt.src2, tt.lnum2)
		st.AddSourceLine(l2)

		t2 := st.tree.Get(l2)
		assert.NotNil(t, t2)

		t1 := st.tree.Get(l1)
		assert.NotNil(t, t1)

		jl1 := st.JumpToLine(tt.lnum1)
		assert.Equal(t, tt.lnum1, jl1)

		t3 := st.JumpToLine(20)
		assert.NotNil(t, t3)

		// test for a bad goto
		jl2 := st.JumpToLine(100)
		assert.Equal(t, uint16(0), jl2)
	}
}

func Test_NextLine(t *testing.T) {
	tests := []struct {
		txt string
		num uint16
	}{
		{txt: `10 REM Comment`, num: 10},
		{txt: `50 END`, num: 50},
		{txt: `20 Print "Hello World`, num: 20},
		{txt: `40 Print "Goodbye`, num: 40},
		{txt: `30 REM Testing NextLine()`, num: 30},
	}

	// load all the source lines
	st := InitSourceTree()
	for _, tt := range tests {
		NewSourceLine(tt.txt, tt.num)
		sl := NewSourceLine(tt.txt, tt.num)
		st.AddSourceLine(sl)
	}

	// load all the lines in a random-ish order
	// make sure they come back in order
	for range tests {
		l := st.NextLine()

		if l != nil {
			t.Log(l.lineNum)
		}
	}

	assert.Nil(t, st.NextLine())

}
