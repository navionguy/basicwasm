package cli

import (
	"errors"
	"testing"

	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
)

func TestExecCommand(t *testing.T) {
	tests := []struct {
		inp string
		exp []string
	}{
		{inp: "STOP", exp: []string{"Break", "OK"}},
		{inp: "10 CLS", exp: []string{"OK"}},
		{inp: "20 PRINT X * Y", exp: []string{"OK"}},
		{"LIST", []string{"10 CLS", "20 PRINT X * Y ", "OK"}},
		{"nerf", []string{"Syntax error", "OK"}},
		{"LET X = 5", []string{"OK"}},
		{"LET Y = 2", []string{"OK"}},
		{"PRINT X", []string{"5", "", "OK"}},
		{"PRINT 45.2 / 3.4", []string{"13.29412", "", "OK"}},
		{"CLS : LIST", []string{"10 CLS", "20 PRINT X * Y ", "OK"}},
		{"GOTO 10", []string{"10", "", "OK"}},
		{"AUTO 10", []string{"10*"}},
	}

	var trm mocks.MockTerm
	mocks.InitMockTerm(&trm)
	env := object.NewTermEnvironment(trm)
	for _, tt := range tests {
		trm.ExpMsg.Exp = tt.exp
		execCommand(tt.inp, env)
		if trm.ExpMsg.Failed {
			t.Fatalf("didn't expect that!")
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
		trm.ExpMsg.Exp = tt.exp
		env := object.NewTermEnvironment(trm)
		terr := errors.New(tt.inp)
		giveError(terr.Error(), env)
		if trm.ExpMsg.Failed {
			t.Fatalf("GiveError didn't")
		}
	}

}
