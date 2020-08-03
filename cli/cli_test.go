package cli

import (
	"testing"

	"github.com/navionguy/basicwasm/object"
)

func TestExecCommand(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{"LET X = 5"},
		{"PRINT X"},
	}

	env := object.NewEnvironment()
	for _, tt := range tests {
		execCommand(tt.inp, env)
	}
}
