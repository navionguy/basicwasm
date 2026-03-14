package ast

import (
	"testing"

	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

// Create a new Code object and exercise all default values
func Test_InitCode(t *testing.T) {
	c := InitCode()

	assert.NotNil(t, c)
	assert.Equal(t, "", c.TokenLiteral())
	assert.Equal(t, "SourceCode", c.String())
	assert.NotNil(t, c.srcCode)
	assert.Zero(t, c.srcCode.tree.Len())
	assert.Zero(t, c.currLine)
	assert.Nil(t, c.err)
}

func Test_AddSrcLine(t *testing.T) {
	c := InitCode()
	c.AddSrcLine("10 REM A Comment", 10)

	// to avoid using the parser, manually load two statements
	c.currLine = c.srcCode.firstLine()
	c.currLine.statements = append(c.currLine.statements, &LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "10"}, Value: 10})
	c.currLine.statements = append(c.currLine.statements, &RemStatement{Token: token.Token{Type: token.REM, Literal: token.REM}, Comment: "A Comment"})
	// nil currLine so he will search the tree
	c.currLine = nil

	assert.NotZero(t, c.srcCode.tree.Len())

	s, _ := c.NextStmt()
	assert.NotNil(t, s)
	_, ok := s.(*LineNumStmt)
	assert.True(t, ok)

	s, _ = c.NextStmt()
	_, ok = s.(*RemStatement)
	assert.True(t, ok)

	s, _ = c.NextStmt()
	assert.Nil(t, s)

	// ask for the next source line, there shouldn't be one.
	c.currLine = c.srcCode.NextLine()
	assert.Nil(t, c.currLine)
}

func Test_GetReturnPoint(t *testing.T) {
	// build up some source lines without using the parser
	c := InitCode()
	c.AddSrcLine("10 GOSUB 200: REM A Comment", 10)
	c.currLine = c.srcCode.firstLine()
	c.currLine.statements = append(c.currLine.statements,
		&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "10"}, Value: 10})
	c.currLine.statements = append(c.currLine.statements,
		&GosubStatement{Token: token.Token{Type: token.GOSUB, Literal: "GOSUB"},
			Gosub: []token.Token{{Type: token.INT, Literal: "1000"}}})
	c.NextStmt()
	rp := c.GetReturnPoint()

	assert.NotNil(t, rp)
	assert.EqualValues(t, 10, rp.Line)
	assert.EqualValues(t, 2, rp.Stmt)
}

func Test_GetSrcLineCount(t *testing.T) {
	tests := []struct {
		src  string
		line uint16
		exp  uint16
	}{
		{src: "10 REM A Comment", line: 10, exp: 1},
		{src: `20 REM "Hello World!"`, line: 20, exp: 2},
		{src: "30 REM That's everything!", line: 30, exp: 3},
	}
	c := InitCode()
	for _, tt := range tests {
		c.AddSrcLine(tt.src, tt.line)

		assert.EqualValues(t, tt.exp, c.GetSrcLineCount())
	}
}

func Test_CodeJumpToLine(t *testing.T) {
	c := InitCode()
	c.AddSrcLine("10 REM A Comment", 10)
	c.AddSrcLine("20 REM Another Comment", 20)

	c.Goto(20)
	assert.NotNil(t, c.currLine)
	assert.EqualValues(t, 20, c.currLine.lineNum)

	// make sure it signals failure
	c.Goto(30)
	assert.Nil(t, c.currLine)
}

func Test_LineExists(t *testing.T) {
	c := InitCode()
	c.AddSrcLine("10 REM A Comment", 10)
	c.AddSrcLine("20 REM Another Comment", 20)

	assert.False(t, c.LineExists(30))
	assert.True(t, c.LineExists(20))

}

func Test_LineParsed(t *testing.T) {
	c := InitCode()
	c.AddSrcLine("10 REM A Comment", 10)
	c.currLine = c.srcCode.firstLine()
	assert.False(t, c.LineParsed())

	c.currLine.statements = append(c.currLine.statements, &RemStatement{Comment: "A Comment"})
	assert.True(t, c.LineParsed())
}
func Test_NextSrcLine(t *testing.T) {
	tests := []struct {
		src   string
		lnum  uint16
		stmts []Statement
	}{
		{src: "10 REM A Comment", lnum: 10, stmts: []Statement{
			&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "10"}, Value: 10},
			&RemStatement{Token: token.Token{Type: token.REM, Literal: "REM"}, Comment: "A Comment"},
		}},
		{src: "20 REM Another Comment", lnum: 20, stmts: []Statement{
			&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "20"}, Value: 20},
			&RemStatement{Token: token.Token{Type: token.REM, Literal: "REM"}, Comment: "Another Comment"},
		}},
	}

	c := InitCode()

	// load the initial source code
	for _, tt := range tests {
		c.AddSrcLine(tt.src, tt.lnum)
	}

	// make sure he detects no statements
	//c.FirstLine()
	s, l := c.NextStmt()
	assert.EqualValues(t, tests[0].lnum, l.lineNum)
	assert.Nil(t, s)

	// load the statements for each line
	for i, tt := range tests {
		// move to the first line
		if i == 0 {
			c.currLine = c.srcCode.firstLine()
		} else {
			_, c.currLine = c.nextSrcLine()
		}

		// add the statements
		for _, stmt := range tt.stmts {
			c.currLine.AddStatement(stmt)
		}
	}

	// now check to make sure I can read them out correctly
	c.currLine = c.srcCode.firstLine()
	_, c.currLine = c.NextStmt()
	for _, tt := range tests {
		//var s Statement
		assert.NotNil(t, c.srcCode.curLine)
		for range tt.stmts {
			if tt.lnum != c.currLine.lineNum {
				println("this is bad!")
			}
			assert.Equal(t, tt.lnum, c.currLine.lineNum)
			_, c.currLine = c.NextStmt()
		}
	}

	/*
	   c.currLine = c.currLine = c.srcCode.firstLine()
	   c.currLine.AddStatement(&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "10"}, Value: 10})
	   c.currLine.AddStatement(&RemStatement{Token: token.Token{Type: token.REM, Literal: "REM"}, Comment: "A Comment"})
	   assert.NotNil(t, c.srcCode.curLine)
	   assert.Equal(t, uint16(10), c.currLine.lineNum)
	   assert.Equal(t, uint16(10), c.srcCode.curLine)

	   var s Statement
	   s, c.currLine = c.NextStmt()
	   assert.NotNil(t, c.srcCode.curLine)
	   assert.EqualValues(t, uint16(10), c.srcCode.curLine)
	   assert.NotNil(t, s)

	   s, c.currLine = c.nextSrcLine()
	   c.currLine.AddStatement(&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "20"}, Value: 20})
	   c.currLine.AddStatement(&RemStatement{Token: token.Token{Type: token.REM, Literal: "REM"}, Comment: "Another Comment"})
	   assert.NotNil(t, c.srcCode.curLine)
	   assert.EqualValues(t, uint16(20), c.srcCode.curLine)
	*/
}
