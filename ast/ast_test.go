package ast

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/token"
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
