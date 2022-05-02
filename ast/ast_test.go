package ast

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

func Test_AutoCommand(t *testing.T) {
	auto := AutoCommand{Token: token.Token{Type: token.AUTO, Literal: "AUTO"}, Start: -1, Increment: -1, Curr: true}

	auto.statementNode()
	assert.Equal(t, "AUTO", auto.TokenLiteral(), "AUTO gave wrong token literal")

	assert.Equalf(t, "AUTO .", auto.String(), "auto.String returned %s wanted %s", auto.String(), "AUTO .")
}

func Test_BeepStatement(t *testing.T) {
	beep := BeepStatement{Token: token.Token{Type: token.BEEP, Literal: "BEEP"}}

	beep.statementNode()
	assert.Equal(t, "BEEP", beep.TokenLiteral(), "BEEP gave wrong token literal")

	assert.Equalf(t, "BEEP", beep.String(), "beep.String returned %s wanted %s", beep.String(), "BEEP")
}

func Test_BuiltinExpression(t *testing.T) {
	builtin := BuiltinExpression{Token: token.Token{Type: token.BUILTIN, Literal: "INSTR"},
		Params: []Expression{&StringLiteral{Value: "FooBar"}, &StringLiteral{Value: "Bar"}, &DblIntegerLiteral{Value: 3}}}

	builtin.expressionNode()

	assert.Equal(t, "INSTR", builtin.TokenLiteral())
	assert.Equal(t, `INSTR("FooBar","Bar",3)`, builtin.String())
}

func Test_BlockStatement(t *testing.T) {
	blk := BlockStatement{Token: token.Token{Type: token.LBRACE, Literal: "{"}}

	blk.statementNode()

	assert.Equal(t, "{", blk.TokenLiteral())
}

func Test_CallExpression(t *testing.T) {
	call := CallExpression{Token: token.Token{Type: token.LPAREN, Literal: "ABS"}}

	call.expressionNode()

	assert.Equalf(t, "ABS", call.TokenLiteral(), "call gave wrong token literal")
}

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

func Test_ChainStatement(t *testing.T) {
	tests := []struct {
		cmd ChainStatement
		exp string
	}{
		{cmd: ChainStatement{Token: token.Token{Type: token.CHAIN, Literal: "CHAIN"},
			Path: &StringLiteral{Token: token.Token{Type: token.STRING, Literal: "HIWORLD.BAS"}, Value: "HIWORLD.BAS"}},
			exp: `CHAIN "HIWORLD.BAS"`},
		{cmd: ChainStatement{Token: token.Token{Type: token.CHAIN, Literal: "CHAIN"},
			Path: &StringLiteral{Token: token.Token{Type: token.STRING, Literal: "HIWORLD.BAS"}, Value: "HIWORLD.BAS"},
			Line: &IntegerLiteral{Value: 100}, All: true, Delete: true, Merge: true,
			Range: &InfixExpression{Token: token.Token{Type: token.MINUS, Literal: "-"},
				Left: &IntegerLiteral{Value: 100}, Operator: "-", Right: &IntegerLiteral{Value: 500}}},
			exp: `CHAIN MERGE "HIWORLD.BAS", 100, ALL, DELETE 100 - 500`},
	}

	for _, tt := range tests {
		tt.cmd.statementNode()

		assert.Equal(t, tt.cmd.Token.Literal, tt.cmd.TokenLiteral(), "(%s) Token.Literal and TokenLiteral() mismatch", tt.exp, tt.cmd.Token.Literal, tt.cmd.TokenLiteral())

		assert.Equal(t, tt.exp, tt.cmd.String(), "(%s) came back as %s", tt.exp, tt.cmd.String())
	}
}

func Test_ChDir(t *testing.T) {
	cd := &ChDirStatement{Token: token.Token{Type: token.CHDIR, Literal: "CHDIR"}, Path: []Expression{&StringLiteral{Value: `D:\`}}}

	cd.statementNode()
	assert.Equal(t, "CHDIR", cd.TokenLiteral())
	assert.Equal(t, `CHDIR "D:\"`, cd.String())
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
		stmt.TokenLiteral()
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

	if len(err) > 0 {
		t.Fatalf("code.Jump to line 10 failed with %s!", err)
	}

	err = program.code.Jump(400)

	if len(err) == 0 {
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

func Test_CodeRetPoint(t *testing.T) {
	code := Code{currIndex: 1, currLine: 100, lines: []codeLine{{}, {lineNum: 10, curStmt: 5}}}

	rp := code.GetReturnPoint()

	assert.Equal(t, 1, rp.currIndex, "GetReturnPoint gave index %d, expected, 1", rp.currIndex)
	assert.Equal(t, 5, rp.currStmt, "GetReturnPointgave stmt %d, expected 5", rp.currStmt)

	code.JumpToRetPoint(rp)
}

func Test_ColorStatement(t *testing.T) {
	tests := []struct {
		prms []Expression
		exp  string
	}{
		{exp: "COLOR "},
		{prms: []Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}}, exp: "COLOR 1"},
		{prms: []Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 2}}, exp: "COLOR 1,2"},
		{prms: []Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 2}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 3}}, exp: "COLOR 1,2,3"},
		{prms: []Expression{nil, &IntegerLiteral{
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

func Test_ContCommand(t *testing.T) {
	cmd := ContCommand{Token: token.Token{Type: token.CONT, Literal: "CONT"}}

	cmd.statementNode()
	assert.Equal(t, "CONT", cmd.TokenLiteral())
	assert.Equal(t, "CONT", cmd.String())
}

func Test_NoLineNum(t *testing.T) {
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
		assert.Equal(t, "", cd.TokenLiteral())
		assert.Equal(t, "The Code", cd.String())

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

func Test_DataStatement(t *testing.T) {
	dt := DataStatement{Token: token.Token{Type: token.DATA, Literal: "DATA"},
		Consts: []Expression{
			&IntegerLiteral{Value: 12},
			&StringLiteral{Value: "Fred"},
		}}

	dt.statementNode()

	assert.Equal(t, "DATA", dt.TokenLiteral())
	assert.Equal(t, `DATA 12, "Fred"`, dt.String())
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

func Test_DimStatement(t *testing.T) {
	id1 := Identifier{Token: token.Token{Type: token.IDENT, Literal: "T[]"}, Value: "[]", Type: "", Index: []*IndexExpression{
		{Left: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "10"}, Value: 5},
			Index: &Identifier{Value: "10"},
		},
	}, Array: true}
	id2 := Identifier{Token: token.Token{Type: token.IDENT, Literal: "X[]"}, Value: "[]", Type: "", Index: []*IndexExpression{
		{Left: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "10"}, Value: 5},
			Index: &Identifier{Value: "10"},
		},
		{Left: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "20"}, Value: 5},
			Index: &Identifier{Value: "20"},
		},
	}, Array: true}
	dim := DimStatement{Token: token.Token{Type: token.DIM, Literal: "DIM"}, Vars: []*Identifier{&id1, &id2}}

	dim.statementNode()

	assert.Equal(t, "DIM", dim.TokenLiteral())
	assert.Equal(t, "DIM T[10], X[10,20]", dim.String())
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
		Token: token.Token{Type: token.IDENT, Literal: "X"},
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
	// 10 AUTO 10, 10 : X = ABS(1) : CLS 1 : LET myVar = anotherVar
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

func Test_Csrlin(t *testing.T) {
	csr := Csrlin{Token: token.Token{Literal: "csrlin", Type: token.CSRLIN}}

	csr.expressionNode()
	assert.Equal(t, "CSRLIN", csr.TokenLiteral())
	assert.Equal(t, "csrlin ", csr.String())
}

func Test_DblIntegerLiteral(t *testing.T) {
	dint := &DblIntegerLiteral{Token: token.Token{Type: token.TYPE_DBL, Literal: "#"}, Value: 375}

	dint.expressionNode()

	assert.Equal(t, "#", dint.TokenLiteral())
	assert.Equal(t, "375", dint.String())
}

func Test_EndStatement(t *testing.T) {
	end := EndStatement{Token: token.Token{Type: token.END, Literal: "END"}}

	end.statementNode()
	assert.Equal(t, "END", end.TokenLiteral())
	assert.Equal(t, "END ", end.String())
}

func Test_EOFExpression(t *testing.T) {
	eof := &EOFExpression{Token: token.Token{Type: token.EOF, Literal: ""}}

	eof.expressionNode()
	assert.Equal(t, "", eof.TokenLiteral())
	assert.Equal(t, "", eof.String())
}

func Test_ExpressionStatement(t *testing.T) {
	exp := ExpressionStatement{Token: token.Token{Type: token.EQ, Literal: "X"},
		Expression: &InfixExpression{Token: token.Token{Type: token.ASTERISK}, Left: &Identifier{Value: "X"}, Operator: "*", Right: &Identifier{Value: "Y"}}}

	exp.statementNode()
	assert.Equal(t, "X", exp.TokenLiteral())
	assert.Equal(t, "X = X * Y", exp.String())
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

func Test_FixedLiteral(t *testing.T) {
	d := decimal.New(12345, -2)
	fx := &FixedLiteral{Token: token.Token{Type: token.FIXED, Literal: "123.45"}, Value: d}

	fx.expressionNode()

	assert.Equal(t, "123.45", fx.TokenLiteral())
	assert.Equal(t, "123.45", fx.String())
}

func Test_FloatDoubleLiteral(t *testing.T) {
	fdbl := &FloatDoubleLiteral{Token: token.Token{Type: token.FLOAT, Literal: "1.09432D-06"}}

	fdbl.expressionNode()
	assert.Equal(t, "1.09432D-06", fdbl.TokenLiteral())
	assert.Equal(t, "1.09432D-06", fdbl.String())
}

func Test_FloatSingleLiteral(t *testing.T) {
	fsng := &FloatSingleLiteral{Token: token.Token{Type: token.FLOAT, Literal: "3.14159E02"}, Value: 314.159}

	fsng.expressionNode()
	assert.Equal(t, "3.14159E02", fsng.TokenLiteral())
	assert.Equal(t, "3.14159E02", fsng.String())
}

func Test_ForStatement(t *testing.T) {
	four := ForStatment{Token: token.Token{Type: token.FOR, Literal: "FOR"}, Init: &LetStatement{
		Token: token.Token{Type: token.LET, Literal: ""},
		Name: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "myVar"},
			Value: "myVar",
		},
		Value: &Identifier{
			Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
			Value: "anotherVar",
		}}, Final: []Expression{&IntegerLiteral{Value: 10}}, Step: []Expression{&IntegerLiteral{Value: 2}}}

	four.statementNode()
	assert.Equal(t, "FOR", four.TokenLiteral())
	assert.Equal(t, "FOR myVar = anotherVar TO 10 STEP 2", four.String())
}

func Test_FunctionLiteral(t *testing.T) {
	fn := &FunctionLiteral{Token: token.Token{Type: token.DEF, Literal: "FNMUL"}, Parameters: []*Identifier{{Value: "X"}, {Value: "Y"}},
		Body: &BlockStatement{Statements: []Statement{&ExpressionStatement{Expression: &InfixExpression{Token: token.Token{Type: token.ASTERISK},
			Left: &Identifier{Value: "X"}, Operator: "*", Right: &Identifier{Value: "Y"}}}}}}

	fn.expressionNode()

	assert.Equal(t, "FNMUL", fn.TokenLiteral())
	assert.Equal(t, "DEF FNMUL(X, Y) = X * Y", fn.String())
}

func Test_GosubStatement(t *testing.T) {
	gsb := GosubStatement{Token: token.Token{Type: token.GOSUB, Literal: "GOSUB"}, Gosub: 1000}

	gsb.statementNode()

	assert.Equal(t, "GOSUB", gsb.TokenLiteral())
	assert.Equal(t, "GOSUB 1000", gsb.String())
}

func Test_GotoStatement(t *testing.T) {
	gto := GotoStatement{Token: token.Token{Type: token.GOTO, Literal: "GOTO"}, Goto: "1000"}

	gto.statementNode()

	assert.Equal(t, "GOTO", gto.TokenLiteral())
	assert.Equal(t, "GOTO 1000", gto.String())
}

func Test_GroupedExpression(t *testing.T) {
	grp := GroupedExpression{Token: token.Token{Type: token.LPAREN, Literal: "("}, Exp: &IntegerLiteral{Value: 5}}

	grp.expressionNode()

	assert.Equal(t, "(", grp.TokenLiteral())

	assert.Equal(t, "(5)", grp.String())
}

func Test_HexConstant(t *testing.T) {
	hx := &HexConstant{Token: token.Token{Type: token.HEX, Literal: "&H"}, Value: "cf"}

	hx.expressionNode()

	assert.Equal(t, "&H", hx.TokenLiteral())
	assert.Equal(t, "&Hcf", hx.String())
}

func Test_Identifier(t *testing.T) {
	tests := []struct {
		id  Identifier
		lit string
		exp string
	}{
		{id: Identifier{Token: token.Token{Type: token.IDENT, Literal: "[]"}, Array: true,
			Index: []*IndexExpression{{Left: &IntegerLiteral{Value: 5}, Index: &IntegerLiteral{Value: 0}},
				{Left: &IntegerLiteral{Value: 6}, Index: &IntegerLiteral{Value: 1}},
			}}, lit: "[]", exp: "[0,1]"},
		{id: Identifier{Token: token.Token{Type: token.IDENT, Literal: "X"}, Value: "5"}, lit: "X", exp: "5"},
	}

	for _, tt := range tests {
		tt.id.expressionNode()
		assert.Equal(t, tt.lit, tt.id.TokenLiteral())
		assert.Equal(t, tt.exp, tt.id.String())
	}
}

func Test_IfExpression(t *testing.T) {

	ife := &IfExpression{Token: token.Token{Type: token.IF, Literal: "IF"},
		Condition:   &InfixExpression{Left: &Identifier{Value: "X"}, Operator: "!=", Right: &IntegerLiteral{Value: 5}},
		Consequence: &GosubStatement{Gosub: 200},
		Alternative: &GotoStatement{Goto: "1000"}}

	ife.expressionNode()
	assert.Equal(t, "IF", ife.TokenLiteral())

	assert.Equal(t, "IF X != 5 THEN GOSUB 200 ELSE 1000", ife.String())
}

func Test_IndexExpression(t *testing.T) {
	ind := &IndexExpression{Token: token.Token{Type: token.INT, Literal: "[]"}, Index: &IntegerLiteral{Value: 5}}

	ind.expressionNode()

	assert.Equal(t, "[]", ind.TokenLiteral())

	assert.Equal(t, "5", ind.String())
}

// exercise the InfixExpression structure
func Test_InfixExpression(t *testing.T) {
	tests := []struct {
		exp   string
		typ   token.TokenType
		lit   string
		left  Expression
		right Expression
	}{
		{exp: "100 - 1000", typ: token.MINUS, lit: "-", left: &IntegerLiteral{Value: 100}, right: &IntegerLiteral{Value: 1000}},
	}

	for _, tt := range tests {
		exp := InfixExpression{Token: token.Token{Type: tt.typ, Literal: tt.lit}, Left: tt.left, Right: tt.right, Operator: tt.lit}

		exp.expressionNode()

		assert.Equalf(t, tt.lit, exp.TokenLiteral(), "%s returned literal %s", tt.exp, exp.TokenLiteral())

		assert.Equalf(t, tt.exp, exp.String(), "exp %s got %s instead", tt.exp, exp.String())
	}
}

func Test_IntegerLiteral(t *testing.T) {
	il := &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "13"}, Value: 13}

	il.expressionNode()

	assert.Equal(t, "13", il.TokenLiteral())
	assert.Equal(t, "13", il.String())
}

func Test_KeyStatement(t *testing.T) {
	key := &KeyStatement{Token: token.Token{Type: token.KEY, Literal: "KEY"}, Param: token.Token{Type: token.INT, Literal: "1"}, Data: []Expression{&StringLiteral{Value: "FILES"}}}

	key.statementNode()

	assert.Equal(t, "KEY", key.TokenLiteral())
	assert.Equal(t, `KEY 1, "FILES"`, key.String())
}

func Test_ListStatement(t *testing.T) {
	list := ListStatement{Token: token.Token{Type: token.LIST, Literal: "LIST"}, Start: "10", Lrange: "-", Stop: "100"}

	list.statementNode()

	assert.Equal(t, "LIST", list.TokenLiteral())
	assert.Equal(t, "LIST 10-100", list.String())
}

func Test_LocateStatement(t *testing.T) {
	tests := []struct {
		parms []Expression
		exp   string
	}{
		{parms: []Expression{nil, &IntegerLiteral{Value: 12}}, exp: "LOCATE ,12"},
		{parms: []Expression{&IntegerLiteral{Value: 12}}, exp: "LOCATE 12"},
		{parms: []Expression{&IntegerLiteral{Value: 1}, &IntegerLiteral{Value: 12}}, exp: "LOCATE 1,12"},
		{parms: []Expression{&IntegerLiteral{Value: 1},
			&IntegerLiteral{Value: 12},
			&IntegerLiteral{Value: 1},
			&IntegerLiteral{Value: 3},
			&IntegerLiteral{Value: 30}},
			exp: "LOCATE 1,12,1,3,30"},
	}

	for _, tt := range tests {
		stmt := LocateStatement{Token: token.Token{Type: token.LOCATE, Literal: "LOCATE"}}
		for _, p := range tt.parms {
			stmt.Parms = append(stmt.Parms, p)
		}
		stmt.statementNode()

		assert.Equal(t, "LOCATE", stmt.TokenLiteral(), "Locate statement has incorrect TokenLiteral")
		assert.Equal(t, tt.exp, stmt.String(), "Locate statement didn't build string correctly")
	}
}

func Test_LoadCommand(t *testing.T) {

	tests := []struct {
		cmd LoadCommand
		exp string
	}{
		{cmd: LoadCommand{Token: token.Token{Type: token.LOAD, Literal: "LOAD"},
			Path: &StringLiteral{Token: token.Token{Type: token.STRING, Literal: "HIWORLD.BAS"}, Value: "HIWORLD.BAS"}},
			exp: `LOAD "HIWORLD.BAS"`},
		{cmd: LoadCommand{Token: token.Token{Type: token.LOAD, Literal: "LOAD"},
			Path: &StringLiteral{Token: token.Token{Type: token.STRING, Literal: "HIWORLD.BAS"}, Value: "HIWORLD.BAS"}, KeppOpen: true},
			exp: `LOAD "HIWORLD.BAS",R`},
		//{cmd: LoadCommand{Token: token.Token{Type: token.LOAD, Literal: "LOAD"}, Path: `HIWORLD.BAS`}, exp: `LOAD "HIWORLD.BAS"`},
		//{cmd: LoadCommand{Token: token.Token{Type: token.LOAD, Literal: "LOAD"}, Path: `HIWORLD.BAS`, KeppOpen: true}, exp: `LOAD "HIWORLD.BAS",R`},
	}

	for _, tt := range tests {
		tt.cmd.statementNode()

		assert.Equal(t, tt.cmd.Token.Literal, tt.cmd.TokenLiteral(), "Load command has incorrect TokenLiteral")
		assert.Equal(t, tt.exp, tt.cmd.String(), "Load command didn't build string correctly")

	}
}

func Test_OctalConstant(t *testing.T) {
	oct := OctalConstant{Token: token.Token{Type: token.OCTAL, Literal: "&"}, Value: "37"}

	oct.expressionNode()

	assert.Equal(t, "&", oct.TokenLiteral())
	assert.Equal(t, "&37", oct.String())
}

func Test_NewCommand(t *testing.T) {
	cmd := NewCommand{Token: token.Token{Type: token.NEW, Literal: "NEW"}}

	cmd.statementNode()

	assert.Equal(t, "NEW", cmd.TokenLiteral())
	assert.Equal(t, "NEW ", cmd.String())
}

func Test_NextStatement(t *testing.T) {
	nxt := NextStatement{Token: token.Token{Type: token.NEXT, Literal: "NEXT"}}

	nxt.statementNode()
	assert.Equal(t, "NEXT", nxt.TokenLiteral())
	assert.Equal(t, "NEXT ", nxt.String())
}

func Test_PaletteStatement(t *testing.T) {
	tests := []struct {
		stmt PaletteStatement
		exp  string
	}{
		{stmt: PaletteStatement{Token: token.Token{Type: token.PALETTE, Literal: "PALETTE"}}, exp: "PALETTE"},
		{stmt: PaletteStatement{Token: token.Token{Type: token.PALETTE, Literal: "PALETTE"},
			Attrib: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "INT"},
				Value: 1}, Color: &IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "INT"}, Value: 2}}, exp: "PALETTE 1,2"},
	}

	for _, tt := range tests {
		tt.stmt.statementNode()

		assert.Equal(t, token.PALETTE, tt.stmt.TokenLiteral())
		if len(tt.exp) != 0 {
			assert.Equal(t, tt.exp, tt.stmt.String())
		}
	}
}

func Test_PrefixExpression(t *testing.T) {
	tests := []struct {
		exp string
		typ token.TokenType
		lit string
		val int16
	}{
		{exp: "-37", typ: token.MINUS, lit: "-", val: 37},
	}

	for _, tt := range tests {
		exp := PrefixExpression{Token: token.Token{Type: tt.typ, Literal: tt.lit}, Operator: tt.lit, Right: &IntegerLiteral{Value: tt.val}}

		exp.expressionNode()

		assert.Equalf(t, tt.lit, exp.TokenLiteral(), "%s returned literal %s", tt.exp, exp.TokenLiteral())

		assert.Equalf(t, tt.exp, exp.String(), "expected %s got %s", tt.exp, exp.String())
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

func Test_ReadStatement(t *testing.T) {
	rd := &ReadStatement{Token: token.Token{Type: token.READ, Literal: "READ"}, Vars: []Expression{&Identifier{Value: "X"}, &Identifier{Value: "Y"}}}

	rd.statementNode()

	assert.Equal(t, "READ", rd.TokenLiteral())

	assert.Equal(t, "READ X, Y", rd.String())
}

func Test_RemStatement(t *testing.T) {
	stmt := &RemStatement{Token: token.Token{Type: token.REM, Literal: "REM"}, Comment: "A Comment"}

	stmt.statementNode()

	assert.Equal(t, "REM", stmt.TokenLiteral(), "Rem statement has incorrect TokenLiteral")
	assert.Equal(t, stmt.String(), "REM A Comment", "Rem statement didn't build string correctly")
}

func Test_RestoreStatement(t *testing.T) {
	rstr := &RestoreStatement{Token: token.Token{Type: token.RESTORE, Literal: "RESTORE"}, Line: 200}

	rstr.statementNode()

	assert.Equal(t, "RESTORE", rstr.TokenLiteral())
	assert.Equal(t, "RESTORE 200", rstr.String())
}

func Test_ReturnStatement(t *testing.T) {
	stmt := &ReturnStatement{Token: token.Token{Type: token.RETURN, Literal: "RETURN"}, ReturnTo: "1000"}

	stmt.statementNode()
	assert.Equal(t, "RETURN", stmt.TokenLiteral())
	assert.Equal(t, "RETURN 1000", stmt.String())
}

func Test_RunCommand(t *testing.T) {
	tests := []struct {
		cmd RunCommand
		exp string
	}{
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}}, exp: "RUN"},
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}, StartLine: 100}, exp: "RUN 100"},
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}, LoadFile: &StringLiteral{Value: `START.BAS`}}, exp: `RUN "START.BAS"`},
		{cmd: RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}, LoadFile: &StringLiteral{Value: `START.BAS`}, KeepOpen: true}, exp: `RUN "START.BAS",r`},
	}

	for _, tt := range tests {
		cmd := &tt.cmd
		cmd.statementNode()

		assert.Equal(t, "RUN", cmd.TokenLiteral(), "Run command has incorrect TokenLiteral")
		assert.Equal(t, tt.exp, cmd.String(), "Run command didn't build string correctly")
	}
}

func Test_ScreenStatement(t *testing.T) {
	tests := []struct {
		prms []Expression // array of parameter expressions
		exp  string       // what the string output should look like
	}{
		{prms: []Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}}, exp: "SCREEN 1"},
		{prms: []Expression{&IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 1}, &IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: "INT"},
			Value: 2}}, exp: "SCREEN 1,2"},
	}

	for _, tt := range tests {
		scrn := ScreenStatement{Token: token.Token{Type: token.SCREEN, Literal: "SCREEN"}, Params: tt.prms}

		scrn.statementNode()

		assert.Equal(t, token.SCREEN, scrn.TokenLiteral())
		assert.Equal(t, tt.exp, scrn.String())
	}

	scrn := ScreenStatement{}
	scrn.InitValue()
	assert.Equal(t, 0, scrn.Settings[0])
	assert.Equal(t, 1, scrn.Settings[1])
	assert.Equal(t, 0, scrn.Settings[2])
	assert.Equal(t, 0, scrn.Settings[3])
}

func Test_StopStatement(t *testing.T) {
	stop := StopStatement{Token: token.Token{Type: token.STOP, Literal: "STOP"}}

	stop.statementNode()

	assert.Equal(t, token.STOP, stop.TokenLiteral())
	assert.Equal(t, "STOP ", stop.String())
}

func Test_StringLiteral(t *testing.T) {
	str := &StringLiteral{Token: token.Token{Type: token.STRING, Literal: "STRING"}, Value: `Test String`}

	str.expressionNode()

	assert.Equal(t, "STRING", str.TokenLiteral())
	assert.Equal(t, "\"Test String\"", str.String())
}

func Test_ToStatement(t *testing.T) {
	to := &ToStatement{Token: token.Token{Type: token.TO, Literal: "TO"}}

	to.statementNode()
	assert.Equal(t, "TO", to.TokenLiteral())
	assert.Equal(t, " TO ", to.String())
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

func Test_ViewStatement(t *testing.T) {
	vw := &ViewStatement{Token: token.Token{Type: token.VIEW, Literal: "VIEW"},
		Parms: []Node{&Identifier{Value: "("}, &IntegerLiteral{Value: 3}, &Identifier{Value: ","}, &IntegerLiteral{Value: 24}, &Identifier{Value: ")"},
			&Identifier{Value: " - "}, &Identifier{Value: "("}, &IntegerLiteral{Value: 100}, &Identifier{Value: ","}, &IntegerLiteral{Value: 100}, &Identifier{Value: ")"}}}

	vw.statementNode()
	assert.Equal(t, "VIEW", vw.TokenLiteral())
	assert.Equal(t, "VIEW (3,24) - (100,100)", vw.String())
}

func Test_ViewPrintStatement(t *testing.T) {
	vwp := &ViewPrintStatement{Token: token.Token{Type: token.VIEW, Literal: "VIEW PRINT"}, Parms: []Node{&IntegerLiteral{Value: 3}, &ToStatement{Token: token.Token{Type: token.TO, Literal: "TO"}}, &IntegerLiteral{Value: 24}}}

	vwp.statementNode()
	assert.Equal(t, "VIEW PRINT", vwp.TokenLiteral())
	assert.Equal(t, "VIEW PRINT 3 TO 24", vwp.String())
}
