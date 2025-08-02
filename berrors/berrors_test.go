package berrors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextForError(t *testing.T) {
	tests := []struct {
		inp int
		val int
		exp string
	}{
		{inp: CantContinue, val: 17, exp: "Can't continue"},
		{inp: DivByZero, val: 11, exp: "Division by zero"},
		{inp: FileNotFound, val: 53, exp: "File not found"},
		{inp: DeviceIOError, val: 57, exp: "Device I/O Error"},
		{inp: IllegalDirect, val: 12, exp: "Illegal direct"},
		{inp: IllegalFuncCallErr, val: 5, exp: "Illegal function call"},
		{inp: NextWithoutFor, val: 1, exp: "NEXT without FOR"},
		{inp: OutOfData, val: 4, exp: "Out of DATA"},
		{inp: Overflow, val: 6, exp: "Overflow"},
		{inp: ReturnWoGosub, val: 3, exp: "RETURN without GOSUB"},
		{inp: Syntax, val: 2, exp: "Syntax error"},
		{inp: TypeMismatch, val: 13, exp: "Type mismatch"},
		{inp: UndefinedFunction, val: 18, exp: "Undefined user function"},
		{inp: UnDefinedLineNumber, val: 8, exp: "Undefined line number"},
		{inp: PermissionDenied, val: 70, exp: "Permission Denied"},
		{inp: PathNotFound, val: 76, exp: "Path not found"},
		{inp: 100, val: 100, exp: "Unprintable error"},
		{inp: ServerError, val: 77, exp: "Server error"},
	}

	for _, tt := range tests {
		rc := TextForError(tt.inp)

		assert.EqualValuesf(t, tt.exp, rc, "TextForError(%d) got %s, wanted %s", tt.inp, rc, tt.exp)
		assert.EqualValues(t, tt.val, tt.inp)
	}
}
