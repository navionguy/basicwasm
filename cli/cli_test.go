package cli

import (
	"errors"
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/settings"
)

func Test_EvalKeyCodes(t *testing.T) {
	tests := []struct {
		inp  string
		key  []byte
		exp  []string
		auto bool
	}{
		{inp: "", key: []byte("\r")},
		{inp: "PRINT", key: []byte("\r"), exp: []string{"\r\n", "", "OK"}},           // output a blank line
		{inp: "Down arrow", key: []byte{0x7f}, exp: []string{"\x1b[P"}},              // move cursor down
		{inp: "F", key: []byte("F"), exp: []string{"F"}},                             // just echo the key
		{inp: "ctrl-c", key: []byte{0x03}, exp: []string{""}},                        // nothing visibile, need to check state ToDo
		{inp: "ctrl-c auto", key: []byte{0x03}, exp: []string{"", "OK"}, auto: true}, // should turn off auto ToDo check that
	}

	for _, tt := range tests {
		trm := mocks.MockTerm{}
		mocks.InitMockTerm(&trm)
		*trm.Row = 1
		*trm.Col = len(tt.inp)
		trm.StrVal = &tt.inp
		trm.ExpMsg = &mocks.Expector{}
		if len(tt.exp) > 0 {
			trm.ExpMsg.Exp = tt.exp
		}
		env := object.NewTermEnvironment(trm)
		if tt.auto {
			ato := ast.AutoCommand{Params: []ast.Expression{&ast.DblIntegerLiteral{Value: int32(10)}, &ast.DblIntegerLiteral{Value: int32(10)}}}
			env.SaveSetting(settings.Auto, &ato)
		}
		evalKeyCodes(tt.key, env)
		if len(tt.exp) > 0 {
			if trm.ExpMsg.Failed {
				t.Fatalf("%s didn't expect that!", tt.inp)
			}

			if len(trm.ExpMsg.Exp) != 0 {
				t.Fatalf("%s expected %s but didn't get it", tt.inp, trm.ExpMsg.Exp[0])

			}

		}
	}
}

func Test_ExecCommand(t *testing.T) {
	tests := []struct {
		inp string
		exp []string
	}{
		{inp: "\n"},
		{inp: `10 PRINT X`},
		{inp: "RESTORE X", exp: []string{"Syntax error", "OK"}},
		{inp: "CHAIN", exp: []string{"Syntax error", "OK"}},
		{inp: `PRINT "HELLO"`, exp: []string{"HELLO", "", "OK"}},
	}

	for _, tt := range tests {
		trm := mocks.MockTerm{}
		mocks.InitMockTerm(&trm)
		trm.ExpMsg = &mocks.Expector{}
		if len(tt.exp) > 0 {
			trm.ExpMsg.Exp = tt.exp
		}
		env := object.NewTermEnvironment(trm)
		execCommand(tt.inp, env)
		if len(tt.exp) > 0 {
			if trm.ExpMsg.Failed {
				t.Fatalf("%s didn't expect that!", tt.inp)
			}

			if len(trm.ExpMsg.Exp) != 0 {
				t.Fatalf("%s expected %s but didn't get it", tt.inp, trm.ExpMsg.Exp[0])

			}

		}
	}
}

func Test_GiveError(t *testing.T) {
	tests := []struct {
		inp string
		exp []string
	}{
		{inp: "Syntax Error", exp: []string{"Syntax Error", "OK"}},
	}

	for _, tt := range tests {
		var trm mocks.MockTerm
		mocks.InitMockTerm(&trm)
		trm.ExpMsg = &mocks.Expector{}
		trm.ExpMsg.Exp = tt.exp
		env := object.NewTermEnvironment(trm)
		terr := errors.New(tt.inp)
		giveError(terr.Error(), env)
		if trm.ExpMsg.Failed {
			t.Fatalf("GiveError didn't")
		}
	}

}
