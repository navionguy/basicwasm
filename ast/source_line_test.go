package ast

import (
	"fmt"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/berrors"
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
		tok := fmt.Sprint(tt.lnum)
		assert.EqualValues(t, tok, sl.TokenLiteral())
		assert.EqualValues(t, tt.src, sl.String())
		assert.EqualValues(t, tt.lnum, sl.Value())
	}
}

func Test_LineLength(t *testing.T) {
	tst := LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	}

	sl := NewSourceLine(tst.Token.Literal, 10)
	sl.AppendStatement(&tst)

	assert.EqualValues(t, 1, sl.LineLength())
}

func Test_FixLineNumber(t *testing.T) {
	tests := []struct {
		src      string
		line     uint16
		old_line uint16
		exp      string
		rc       *ErrorStatement
	}{
		{src: "10 REM A comment",
			line: 20, old_line: 10,
			exp: "20 REM A comment", rc: &ErrorStatement{}},
		{src: "10 REM A comment",
			line: 20, old_line: 30,
			exp: "10 REM A comment",
			rc:  &ErrorStatement{Token: token.Token{Type: token.ERROR, Literal: "ERROR"}, ErrNum: &IntegerLiteral{Value: berrors.IllegalFuncCallErr}}},
	}

	for _, tt := range tests {
		sl := NewSourceLine(tt.src, tt.line)
		r := sl.fixLineNumber(tt.old_line)

		assert.True(t, strings.EqualFold(tt.exp, sl.source))

		if tt.rc == nil {
			assert.Nil(t, r)
		} else {
			e, ok := r.(*ErrorStatement)

			assert.True(t, ok)
			assert.Equal(t, tt.rc, e)
		}
	}
}

func Test_NextStatement(t *testing.T) {
	tst := LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	}
	tst2 := LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "20"},
		Value: 20,
	}

	sl := NewSourceLine(tst.Token.Literal, 10)
	sl.AppendStatement(&tst)
	sl.AppendStatement(&tst2)

	assert.EqualValues(t, 2, sl.LineLength())

	r1 := sl.NextStatement()
	v1, ok := r1.(*LineNumStmt)
	assert.True(t, ok)
	assert.EqualValues(t, tst.Value, v1.Value)

	r2 := sl.NextStatement()
	v2, ok := r2.(*LineNumStmt)
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
