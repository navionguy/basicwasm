package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CodeNew(t *testing.T) {
	cd := &Code{}

	assert.Equal(t, "", cd.TokenLiteral(), "Code object has wrong literal")
	assert.Equal(t, "The Code", cd.String(), "Code.String() gave unexpected value")
	assert.Equal(t, 0, cd.currIndex, "New Code object not initialized with zero index")
	assert.Equal(t, 0, len(cd.lines), "New Code object initialized with a line")
}
func Test_CurLine(t *testing.T) {
	tests := []struct {
		lnum int
		stmt *Statement
		exp  int
	}{
		{lnum: 0, stmt: nil, exp: 0},
	}

	for _, tt := range tests {
		cd := &Code{}

		if tt.stmt != nil {
			cd.addLine(tt.lnum)
			cd.lines[0].stmts = append(cd.lines[0].stmts, *tt.stmt)
		}

		assert.Equal(t, tt.exp, cd.CurLine(), "cd.Curline() return unexpected value")
	}
}

func Test_JumpBeforeRetPoint(t *testing.T) {
	tests := []struct {
		rp RetPoint
	}{
		{rp: RetPoint{}},
	}

	for _, tt := range tests {
		cd := &Code{}
		cd.addLine(10)

		cd.JumpBeforeRetPoint(tt.rp)
	}
}

func Test_Restart(t *testing.T) {
	cd := &Code{}
	stmt := &RemStatement{}

	cd.addLine(10)
	cd.lines[0].stmts = append(cd.lines[0].stmts, stmt)
	cd.currIndex = 1
	cd.lines[0].curStmt = 1
	cd.Restart()

	assert.EqualValues(t, 0, cd.currIndex, "Restart failed to zero index")
	assert.EqualValues(t, 0, cd.lines[0].curStmt, "Restart failed to set curStmt")
}

func Test_Restore(t *testing.T) {
	data := &ConstData{exp: 1, line: 2, stmt: 3}

	data.Restore()

	assert.Equal(t, 0, data.exp, "data.Restore() failed to zero expression index")
	assert.Equal(t, 0, data.line, "data.Restore() failed to zero line index")
	assert.Equal(t, 0, data.stmt, "data.Restore() failed to zero statement index")
}

func Test_RestoreTo(t *testing.T) {
	tests := []struct {
		inp int
		exp int
	}{
		{inp: 20, exp: 1},
		{inp: 30, exp: 0},
	}

	for _, tt := range tests {
		cd := &Code{}
		data := &ConstData{exp: 1, line: 2, stmt: 3, code: cd}
		stmt := &RemStatement{}

		cd.addLine(10)
		cd.lines[0].stmts = append(cd.lines[0].stmts, stmt)
		cd.addLine(20)
		cd.lines[1].stmts = append(cd.lines[1].stmts, stmt)
		cd.lines[1].stmts = append(cd.lines[1].stmts, stmt)
		cd.lines[1].curStmt = 1

		data.RestoreTo(tt.inp)

		assert.Equal(t, tt.exp, data.line, "RestoreTo failed")
	}
}

func Test_Value(t *testing.T) {
	lit := &ExpressionStatement{Expression: &IntegerLiteral{Value: 5}}

	tests := []struct {
		cd  *ConstData
		res Statement
	}{
		{cd: nil},
		{cd: &ConstData{code: &Code{}}},
		{cd: &ConstData{code: &Code{lines: []codeLine{{lineNum: 10, stmts: []Statement{lit}}}}, stmt: 1}},
		{cd: &ConstData{code: &Code{lines: []codeLine{{lineNum: 10, stmts: []Statement{lit}}}}}, res: lit},
	}

	for _, tt := range tests {
		res := tt.cd.value()

		if tt.res != nil {
			_, ok := res.(*ExpressionStatement)

			assert.True(t, ok, "cd.Value didn't get an expression")
		}
		assert.Equal(t, tt.res, res, "cd.value() unexpected result")
	}
}
