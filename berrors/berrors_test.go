package berrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextForError(t *testing.T) {
	tests := []struct {
		inp int
		exp string
	}{
		{inp: NextWithoutFor, exp: "NEXT without FOR"},
		{inp: Syntax, exp: "Syntax error"},
		{inp: ReturnWoGosub, exp: "RETURN without GOSUB"},
		{inp: OutOfData, exp: "Out of DATA"},
		{inp: IllegalDirect, exp: "Illegal direct"},
		{inp: TypeMismatch, exp: "Type mismatch"},
		{inp: FileNotFound, exp: "File not found"},
		{inp: 100, exp: "Unprintable error"},
		{inp: UnDefinedLineNumber, exp: "Undefined line number"},
	}

	for _, tt := range tests {
		rc := TextForError(tt.inp)

		assert.EqualValuesf(t, tt.exp, rc, "TextForError(%d) got %s, wanted %s", tt.inp, rc, tt.exp)
	}
}
