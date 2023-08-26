package localfiles

import (
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/stretchr/testify/assert"
)

func Test_CreateFileStore(t *testing.T) {
	lf := CreateFileStore()

	assert.NotNil(t, lf, "CreateFileStore returned nil")
}

func Test_OpenFile(t *testing.T) {
	tests := []struct {
		inp  *ast.OpenStatement
		inp2 *ast.OpenStatement
	}{
		{inp: &ast.OpenStatement{FileName: "driveC/data.txt"},
			inp2: &ast.OpenStatement{FileName: "driveC/data.txt"}},
	}

	for _, tt := range tests {
		lf := CreateFileStore()

		if tt.inp2 != nil {
			lf.Open(tt.inp2)
		}
		lf.Open(tt.inp)

	}
}
