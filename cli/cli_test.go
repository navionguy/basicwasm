package cli

import (
	"fmt"
	"testing"

	"github.com/navionguy/basicwasm/object"
)

type mockTerm struct {
	row    int
	col    int
	strVal string
}

func (mt mockTerm) Cls() {

}

func (mt mockTerm) Print(msg string) {
	fmt.Print(msg)
}

func (mt mockTerm) Println(msg string) {
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

func TestExecCommand(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{"10 CLS"},
		{"LIST"},
		{"LET X = 5"},
		{"PRINT X"},
		{"PRINT 45.2 / 3.4"},
	}

	var trm mockTerm

	env := object.NewTermEnvironment(trm)
	for _, tt := range tests {
		execCommand(tt.inp, env)
	}
}
