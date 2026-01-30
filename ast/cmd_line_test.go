package ast

import (
	"testing"

	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

func Test_NewCmdLine(t *testing.T) {
	inp := "GOTO 10"
	cl := NewCmdLine(inp)

	assert.NotNil(t, cl)
	assert.Equal(t, cl.input, inp)
	assert.Zero(t, cl.itr)
	assert.Nil(t, cl.NextStatement())
	assert.Nil(t, cl.err)
}

func Test_AddStatement(t *testing.T) {
	tests := []struct {
		inp   string
		stmts []Statement
		count int
	}{
		{inp: "RUN", stmts: []Statement{&RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}}}, count: 1},
		{inp: `LOAD "HELLOWORLD.BAS": RUN`, stmts: []Statement{
			&LoadCommand{Token: token.Token{Type: token.LOAD, Literal: "LOAD"},
				Path: &StringLiteral{Token: token.Token{Type: token.STRING, Literal: `HELLOWORLD.BAS`}},
			},
			&RunCommand{Token: token.Token{Type: token.RUN, Literal: "RUN"}},
		}, count: 2}}

	for _, tt := range tests {
		cl := NewCmdLine(tt.inp)
		for _, st := range tt.stmts {
			cl.AddStatement(st)
		}
		assert.Equal(t, tt.count, len(cl.stmts))
		// also check against the function
		assert.Equal(t, uint16(tt.count), cl.LineLength())

		for _, st := range tt.stmts {
			s := cl.NextStatement()
			assert.Equal(t, s.String(), st.String())
		}

		// make sure the iterator sees the end of the command
		assert.Nil(t, cl.NextStatement())
	}
}

func Test_CmdComplete(t *testing.T) {
	cl := NewCmdLine("RUN")
	cl.itr = 5

	cl.CmdComplete()
}
