package ast

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

func TestStringAndToken(t *testing.T) {
	var program Program

	program.New()
	program.AddStatement(&LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	})
	program.AddStatement(&LetStatement{
		Token: token.Token{Type: token.LET, Literal: "LET"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
			Value: "anotherVar",
		},
	})

	program.code.lines[0].curStmt = 42
	program.Parsed()

	if program.code.lines[0].curStmt != 0 {
		t.Fatalf("code statement ptr failed to reset!")
	}

	rc := program.String()
	if rc != "10 LET myVar = anotherVar" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}

	rc = program.TokenLiteral()
	if rc != "GWBasic" {
		t.Errorf("program.TokenLiteral() wrong. got=%q", program.TokenLiteral())
	}
}

func Test_ClsStatement(t *testing.T) {
	tests := []struct {
		inp ClsStatement
		exp string
	}{
		{inp: ClsStatement{Token: token.Token{Type: token.CLS, Literal: "CLS"}, Param: 1}, exp: "CLS 1"},
		{inp: ClsStatement{Token: token.Token{Type: token.CLS, Literal: "CLS"}, Param: -1}, exp: "CLS"},
	}

	for _, tt := range tests {
		cls := tt.inp

		cls.statementNode()

		assert.Equal(t, "CLS", cls.TokenLiteral(), "Cls command has incorrect TokenLiteral")
		assert.Equal(t, tt.exp, cls.String(), "Clear command didn't build string correctly")
	}
}

func TestCmdLineProgramSwitches(t *testing.T) {
	var program Program

	program.New()
	program.AddCmdStmt(&ClsStatement{
		Token: token.Token{Type: token.CLS, Literal: "CLS"},
		Param: 0,
	})

	cmdl := program.CmdLineIter()

	if len(cmdl.lines) != 1 {
		t.Fatalf("AddCmdStmt() failed, got %d, wanted 1", len(program.cmdLine.lines))
	}

	program.cmdLine.lines[0].curStmt = 37
	program.CmdParsed()

	if program.cmdLine.lines[0].curStmt != 0 {
		t.Fatalf("cmdLine statement ptr failed to reset!")
	}

	program.CmdComplete()

	if program.cmdLine.lines != nil {
		t.Fatalf("cmdLine failed to clear lines!")
	}
}

func TestCodeMultiLines(t *testing.T) {
	var program Program

	program.New()
	program.AddStatement(&LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	})
	program.AddStatement(&LetStatement{
		Token: token.Token{Type: token.LET, Literal: "LET"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
			Value: "anotherVar",
		},
	})
	program.AddStatement(&LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "20"},
		Value: 20,
	})
	program.AddStatement(&LetStatement{
		Token: token.Token{Type: token.LET, Literal: "LET"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "X"},
			Value: "X",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "6"},
			Value: "6",
		},
	})

	it := program.StatementIter()
	sz := it.Len()

	if sz != 4 {
		t.Fatalf("expected 4 statements, got %d", sz)
	}

	stmt := it.Value()

	tests := []struct {
		exp string
	}{
		{"10 "},
		{"LET myVar = anotherVar"},
		{"20 "},
		{"LET X = 6"},
	}

	for _, tt := range tests {
		stmt.statementNode()
		sz = strings.Compare(stmt.String(), tt.exp)
		if sz != 0 {
			t.Fatalf("expected %s, got %s", tt.exp, stmt.String())
		}
		it.Next()
		stmt = it.Value()
	}

	if !program.code.Exists(10) {
		t.Fatal("Code.Exists failed to find line 10!")
	}

	err := program.code.Jump(10)

	if err != nil {
		t.Fatal("code.Jump to line 10 failed!")
	}

	err = program.code.Jump(400)

	if err == nil {
		t.Fatal("code.Jump to non-existant line succeeded!")
	}
}

func TestCodeMultiStmts(t *testing.T) {
	var program Program

	program.New()
	program.AddStatement(&LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	})
	program.AddStatement(&LetStatement{
		Token: token.Token{Type: token.LET, Literal: "LET"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
			Value: "anotherVar",
		},
	})
	program.AddStatement(&LetStatement{
		Token: token.Token{Type: token.LET, Literal: "LET"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "X"},
			Value: "X",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "6"},
			Value: "6",
		},
	})

	it := program.StatementIter()
	sz := it.Len()

	if sz != 3 {
		t.Fatalf("expected 3 statements, got %d", sz)
	}

	stmt := it.Value()

	tests := []struct {
		exp string
	}{
		{"10 "},
		{"LET myVar = anotherVar"},
		{"LET X = 6"},
	}

	for _, tt := range tests {
		sz = strings.Compare(stmt.String(), tt.exp)
		if sz != 0 {
			t.Fatalf("expected %s, got %s", tt.exp, stmt.String())
		}
		it.Next()
		stmt = it.Value()
	}
}

func Test_ColorStatement(t *testing.T) {
	tests := []struct {
		prms [3]Expression
		exp  string
	}{
		{exp: "COLOR"},
		{prms: [3]Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}}, exp: "COLOR 1"},
		{prms: [3]Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 2}}, exp: "COLOR 1,2"},
		{prms: [3]Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 2}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 3}}, exp: "COLOR 1,2,3"},
		{prms: [3]Expression{nil, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 2}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 3}}, exp: "COLOR ,2,3"},
	}

	for _, tt := range tests {
		stmt := ColorStatement{Token: token.Token{Type: token.LINENUM, Literal: "COLOR"}, Parms: tt.prms}

		stmt.statementNode()

		assert.Equal(t, token.COLOR, stmt.TokenLiteral(), "Color statement returned wrong token")
		assert.Equal(t, tt.exp, stmt.String(), "Color command didn't build string correctly")
	}
}

func Test_CommonStatement(t *testing.T) {
	tests := []struct {
		vars []*Identifier
		exp  string
	}{
		{vars: []*Identifier{{Token: token.Token{Type: token.IDENT, Literal: token.IDENT}, Value: "X"},
			{Token: token.Token{Type: token.IDENT, Literal: token.IDENT}, Value: "Y"}},
			exp: "COMMON X, Y"},
	}

	for _, tt := range tests {
		stmt := CommonStatement{Token: token.Token{Type: token.COMMON, Literal: "COMMON"}}

		for _, id := range tt.vars {
			stmt.Vars = append(stmt.Vars, id)
		}

		stmt.statementNode()

		assert.Equal(t, token.COMMON, stmt.TokenLiteral(), "Common statement returned wrong token")
		assert.Equal(t, tt.exp, stmt.String(), "Common command didn't build string correctly")
	}
}

func TestNoLineNum(t *testing.T) {
	var program Program

	program.New()
	program.AddStatement(&LetStatement{
		Token: token.Token{Type: token.LET, Literal: "LET"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
			Value: "anotherVar",
		},
	})

	if program.code.err == nil {
		t.Fatal("failed to detect no line number on line")
	}
}

func TestCodeAdd(t *testing.T) {
	tests := []struct {
		lines    []int
		expected []int
	}{
		{lines: []int{10, 20, 30}, expected: []int{10, 20, 30}},
		{lines: []int{10, 20, 30, 40, 20}, expected: []int{10, 20, 30, 40}},
		{lines: []int{10, 20, 30, 40, 25}, expected: []int{10, 20, 25, 30, 40}},
	}

	for _, tt := range tests {
		var p Program
		p.New()

		cd := p.code
		for _, ln := range tt.lines {
			cd.addLine(ln)

			if p.code.CurLine() != ln {
				t.Fatalf("expected line %d, got %d", ln, p.code.currLine)
			}

		}

		for i, ln := range tt.expected {
			if cd.lines[i].lineNum != ln { // offset by one do to command line slot at 0
				t.Fatalf("test %d, got %d, expected %d", i, cd.lines[i+1].lineNum, ln)
			}
		}
	}
}

func TestCodeIterValue(t *testing.T) {
	tests := []struct {
		lines    []int
		expected []int
	}{
		{lines: []int{10, 20, 30}, expected: []int{10, 20, 30}},
		{lines: []int{10, 20, 30, 40, 20}, expected: []int{10, 20, 30, 40}},
		{lines: []int{10, 20, 30, 40, 25}, expected: []int{10, 20, 25, 30, 40}},
	}

	for _, tt := range tests {
		var p Program
		p.New()

		itr := p.StatementIter()

		for _, ln := range tt.lines {
			itr.addLine(ln)
			stmt := &GotoStatement{Token: token.Token{Type: token.GOTO, Literal: "GOTO"}, Goto: strconv.Itoa(int(ln))}
			itr.lines[itr.currIndex].stmts = append(itr.lines[itr.currIndex].stmts, stmt)
		}

		itr = p.StatementIter()

		for _, ln := range tt.expected {
			res := itr.Value()

			jmp, ok := res.(*GotoStatement)

			if !ok {
				t.Fatalf("expected goto statment, got %T", res)
			}

			if strings.Compare(jmp.Goto, strconv.Itoa(int(ln))) != 0 {
				t.Fatalf("expected line %d, got %s", ln, jmp.Goto)
			}
			itr.Next()
		}
		itr.Value()
		itr.Next()
	}
}

func TestData(t *testing.T) {
	tests := []struct {
		inp []codeLine
		exp []Expression
	}{
		{inp: []codeLine{
			{lineNum: 10, stmts: []Statement{
				&RemStatement{Comment: "Hi"},
				&DataStatement{Consts: []Expression{
					&IntegerLiteral{Value: 12},
					&StringLiteral{Value: "Fred"},
				}},
			}},
			{lineNum: 20, stmts: []Statement{
				&RemStatement{Comment: "Hi"},
				&DataStatement{Consts: []Expression{
					&IntegerLiteral{Value: 21},
					&StringLiteral{Value: "George"},
				}},
			}},
		},
			exp: []Expression{
				&IntegerLiteral{Value: 12},
				&StringLiteral{Value: "Fred"},
				&IntegerLiteral{Value: 21},
				&StringLiteral{Value: "George"},
			}},
	}

	for _, tt := range tests {
		var p Program
		p.New()

		// make sure he handles having no data
		if p.data.findNextData() != nil {
			t.Fatal("He found data without any code!")
		}

		p.code.lines = tt.inp

		for _, exp := range tt.exp {
			got := p.ConstData().Next()

			switch eVal := exp.(type) {
			case *IntegerLiteral:
				gVal, ok := (*got).(*IntegerLiteral)

				if !ok {
					t.Fatalf("expected type %T, but got %T", exp, got)
				}

				if gVal.Value != eVal.Value {
					t.Fatalf("expected value %d but got %d", eVal.Value, gVal.Value)
				}
			case *StringLiteral:
				gVal, ok := (*got).(*StringLiteral)

				if !ok {
					t.Fatalf("expected type %T, but got %T", exp, got)
				}

				if gVal.Value != eVal.Value {
					t.Fatalf("expected value %s but got %s", eVal.Value, gVal.Value)
				}
			}
		}
	}
}

// a long, dump test case
func ExampleStatement() {
	var program Program

	program.New()
	program.AddStatement(&LineNumStmt{
		Token: token.Token{Type: token.LINENUM, Literal: "10"},
		Value: 10,
	})
	program.AddStatement(&AutoCommand{
		Token:     token.Token{Type: token.AUTO, Literal: "AUTO"},
		Start:     10,
		Increment: 10,
		Curr:      false,
	})
	program.AddStatement(&ExpressionStatement{
		Token: token.Token{Type: token.IDENT, Literal: "ABS"},
		Expression: &CallExpression{
			Token:    token.Token{Type: token.LPAREN, Literal: "("},
			Function: &Identifier{Token: token.Token{Type: token.IDENT, Literal: "ABS"}, Value: "ABS"},
			Arguments: []Expression{&IntegerLiteral{
				Token: token.Token{Type: token.INT, Literal: "INT"},
				Value: 1}},
		},
	})
	program.AddStatement(&ClsStatement{
		Token: token.Token{Type: token.CLS, Literal: "CLS"},
		Param: 1,
	})
	program.AddStatement(&LetStatement{
		Token: token.Token{Type: token.LET, Literal: "LET"},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
			Value: "anotherVar",
		},
	})

	program.code.lines[0].curStmt = 42
	program.Parsed()

	fmt.Println(program.String())

	// Output:
	// 10 AUTO 10, 10 : ABS(1) : CLS 1 : LET myVar = anotherVar
}

func Test_ClearCommand(t *testing.T) {

	cmd := &ClearCommand{
		Token: token.Token{Type: token.CLEAR, Literal: "CLEAR"},
		Exp:   [3]Expression{&IntegerLiteral{Value: 1}, &IntegerLiteral{Value: 2}, &IntegerLiteral{Value: 3}},
	}

	cmd.statementNode()

	assert.Equal(t, "CLEAR", cmd.TokenLiteral(), "Clear command has incorrect TokenLiteral")
	assert.Equal(t, cmd.String(), "CLEAR 1,2,3", "Clear command didn't build string correctly")
}

func Test_FilesCommand(t *testing.T) {

	tests := []struct {
		cmd FilesCommand
		exp string
	}{
		{cmd: FilesCommand{Token: token.Token{Type: token.FILES, Literal: "FILES"}, Path: ""}, exp: `FILES`},
		{cmd: FilesCommand{Token: token.Token{Type: token.FILES, Literal: "FILES"}, Path: `C:\MENU`}, exp: `FILES "C:\MENU"`},
	}

	for _, tt := range tests {
		tt.cmd.statementNode()

		assert.Equal(t, "FILES", tt.cmd.TokenLiteral(), "Files command has incorrect TokenLiteral")
		assert.Equal(t, tt.exp, tt.cmd.String(), "Files command didn't build string correctly")

	}
}

func Test_LocateStatement(t *testing.T) {
	tests := []struct {
		parms [Lct_stop + 1]Expression
		exp   string
	}{
		{parms: [Lct_stop + 1]Expression{nil, &IntegerLiteral{Value: 12}}, exp: "LOCATE ,12"},
		{parms: [Lct_stop + 1]Expression{&IntegerLiteral{Value: 12}}, exp: "LOCATE 12"},
		{parms: [Lct_stop + 1]Expression{&IntegerLiteral{Value: 1}, &IntegerLiteral{Value: 12}}, exp: "LOCATE 1,12"},
		{parms: [Lct_stop + 1]Expression{&IntegerLiteral{Value: 1},
			&IntegerLiteral{Value: 12},
			&IntegerLiteral{Value: 1},
			&IntegerLiteral{Value: 3},
			&IntegerLiteral{Value: 30}},
			exp: "LOCATE 1,12,1,3,30"},
	}

	for _, tt := range tests {
		stmt := LocateStatement{Token: token.Token{Type: token.LOCATE, Literal: "LOCATE"}}
		for i := 0; i < len(tt.parms); i++ {
			if tt.parms[i] != nil {
				stmt.Parms[i] = tt.parms[i]
			}
		}
		stmt.statementNode()

		assert.Equal(t, "LOCATE", stmt.TokenLiteral(), "Locate statement has incorrect TokenLiteral")
		assert.Equal(t, tt.exp, stmt.String(), "Locate statement didn't build string correctly")
	}
}

func Test_PrintStatement(t *testing.T) {
	prt := &PrintStatement{Token: token.Token{Type: token.PRINT, Literal: "PRINT"},
		Items:      []Expression{&IntegerLiteral{Value: 12}, &StringLiteral{Value: "Fred"}},
		Seperators: []string{",", ";"}}

	prt.statementNode()

	assert.Equal(t, "PRINT", prt.TokenLiteral(), "Print statement has incorrect TokenLiteral")
	assert.Equal(t, `PRINT 12,"Fred";`, prt.String(), "Print statement didn't build string correctly")
}

func Test_RemStatement(t *testing.T) {
	stmt := &RemStatement{Token: token.Token{Type: token.REM, Literal: "REM"}, Comment: "A Comment"}

	stmt.statementNode()

	assert.Equal(t, "REM", stmt.TokenLiteral(), "Rem statement has incorrect TokenLiteral")
	assert.Equal(t, stmt.String(), "REM A Comment", "Rem statement didn't build string correctly")
}

func Test_RunCommand(t *testing.T) {
	tests := []struct {
		cmd RunCommand
		exp string
	}{
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}}, exp: "RUN"},
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}, StartLine: 100}, exp: "RUN 100"},
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}, LoadFile: `"START.BAS"`}, exp: `RUN "START.BAS"`},
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}, LoadFile: `"START.BAS"`, KeepOpen: true}, exp: `RUN "START.BAS",r`},
	}

	for _, tt := range tests {
		cmd := &tt.cmd
		cmd.statementNode()

		assert.Equal(t, "RUN", cmd.TokenLiteral(), "Run command has incorrect TokenLiteral")
		assert.Equal(t, tt.exp, cmd.String(), "Run command didn't build string correctly")
	}
}

func Test_TronTroffCommands(t *testing.T) {

	cmd := &TroffCommand{
		Token: token.Token{Type: token.TROFF, Literal: "TROFF"},
	}

	cmd.statementNode()

	assert.Equal(t, "TROFF", cmd.TokenLiteral(), "TROFF command has incorrect TokenLiteral")
	assert.Equal(t, cmd.String(), "TROFF", "TROFF command didn't build string correctly")

	cmd2 := &TronCommand{
		Token: token.Token{Type: token.TRON, Literal: "TRON"},
	}

	cmd2.statementNode()

	assert.Equal(t, "TRON", cmd2.TokenLiteral(), "TRON command has incorrect TokenLiteral")
	assert.Equal(t, cmd2.String(), "TRON", "TRON command didn't build string correctly")
}
