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
		{inp: CantContinue, exp: "Can't continue"},
		{inp: DivByZero, exp: "Division by zero"},
		{inp: FileNotFound, exp: "File not found"},
		{inp: IllegalDirect, exp: "Illegal direct"},
		{inp: NextWithoutFor, exp: "NEXT without FOR"},
		{inp: OutOfData, exp: "Out of DATA"},
		{inp: Overflow, exp: "Overflow"},
		{inp: ReturnWoGosub, exp: "RETURN without GOSUB"},
		{inp: Syntax, exp: "Syntax error"},
		{inp: TypeMismatch, exp: "Type mismatch"},
		{inp: UndefinedFunction, exp: "Undefined user function"},
		{inp: UnDefinedLineNumber, exp: "Undefined line number"},
		{inp: PathNotFound, exp: "Path not found"},
		{inp: 100, exp: "Unprintable error"},
		{inp: ServerError, exp: "Server error"},
	}

	for _, tt := range tests {
		rc := TextForError(tt.inp)

		assert.EqualValuesf(t, tt.exp, rc, "TextForError(%d) got %s, wanted %s", tt.inp, rc, tt.exp)
	}
}
