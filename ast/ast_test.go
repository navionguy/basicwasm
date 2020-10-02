package ast

import (
	"strconv"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/token"
)

func TestString(t *testing.T) {
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
	//program.AddStatement()

	rc := program.String()
	if rc != "10 LET myVar = anotherVar" {
		t.Errorf("program.String() wrong. got=%q", program.String())
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

	sz = strings.Compare(stmt.String(), "10 ")
	if sz != 0 {
		t.Fatalf("expected 10, got %s", stmt.String())
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

		p.code.lines = tt.inp

		//p.ConstData().Next()

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
