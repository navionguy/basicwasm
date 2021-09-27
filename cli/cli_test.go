package cli

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/object"
)

type mockTerm struct {
	row     int
	col     int
	strVal  string
	expChk  *Expector
	sawBeep *bool
}

func (mt mockTerm) Cls() {

}

func (mt mockTerm) Print(msg string) {
	mt.expChk.chkExpectations(msg)
	fmt.Print(msg)
}

func (mt mockTerm) Println(msg string) {
	mt.expChk.chkExpectations(msg)
	fmt.Println(msg)
}

func (mt mockTerm) Locate(int, int) {

}

func (mt mockTerm) GetCursor() (int, int) {
	return mt.row, mt.col
}

func (mt mockTerm) Read(col, row, len int) string {
	return mt.strVal
}

func (mt mockTerm) ReadKeys(count int) []byte {
	return nil
}

func (mt mockTerm) SoundBell() {
	fmt.Print("\x07")
	*mt.sawBeep = true
}

type Expector struct {
	failed bool
	exp    []string
}

func (ep *Expector) chkExpectations(msg string) {
	if len(ep.exp) == 0 {
		return // no expectations
	}
	if strings.Compare(msg, ep.exp[0]) != 0 {
		fmt.Print(("unexpected this is ->"))
		ep.failed = true
	}
	ep.exp = ep.exp[1:]
}

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

	var trm mockTerm
	var eChk Expector
	trm.expChk = &eChk
	env := object.NewTermEnvironment(trm)
	for _, tt := range tests {
		eChk.exp = tt.exp
		execCommand(tt.inp, env)
		if eChk.failed {
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
		var trm mockTerm
		var eChk Expector
		eChk.exp = tt.exp
		trm.expChk = &eChk
		env := object.NewTermEnvironment(trm)
		terr := errors.New(tt.inp)
		giveError(terr.Error(), env)
		if eChk.failed {
			t.Fatalf("GiveError didn't")
		}
	}

}
