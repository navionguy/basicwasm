package ast

import (
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

	if program.String() != "10 LET myVar = anotherVar" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}

func TestLineNum(t *testing.T) {
	tests := []struct {
		line   Statement
		output string
	}{
		{&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "10"}, Value: 10}, ""},
		{&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "20"}, Value: 20}, ""},
		{&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "15"}, Value: 15}, "line numbers not sequential after line 20"},
		{&LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "XX"}, Value: 0}, "invalid line number XX"},
	}

	var program Program

	program.New()
	for _, tt := range tests {
		rc := program.AddStatement(tt.line)

		if rc != tt.output {
			t.Errorf("program.AddStatement() gave %s, wanted %s", rc, tt.output)
		}
	}

}
