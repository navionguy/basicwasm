package gwtoken

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/navionguy/basicwasm/object"
	"github.com/stretchr/testify/assert"
)

type mockTerm struct {
	row     *int
	col     *int
	strVal  *string
	sawStr  *string
	sawCls  *bool
	sawBeep *bool
}

func initMockTerm(mt *mockTerm) {
	mt.row = new(int)
	*mt.row = 0

	mt.col = new(int)
	*mt.col = 0

	mt.strVal = new(string)
	*mt.strVal = ""

	mt.sawCls = new(bool)
	*mt.sawCls = false
}

func (mt mockTerm) Cls() {
	*mt.sawCls = true
}

func (mt mockTerm) Print(msg string) {
	fmt.Print(msg)
}

func (mt mockTerm) Println(msg string) {
	fmt.Println(msg)
	if mt.sawStr != nil {
		*mt.sawStr = *mt.sawStr + msg
	}
}

func (mt mockTerm) SoundBell() {
	fmt.Print("\x07")
	*mt.sawBeep = true
}

func (mt mockTerm) Locate(int, int) {
}

func (mt mockTerm) GetCursor() (int, int) {
	return *mt.row, *mt.col
}

func (mt mockTerm) Read(col, row, len int) string {
	// make sure your test is correct
	trim := (row-1)*80 + (col - 1)

	tstr := *mt.strVal

	newstr := tstr[trim : trim+len]

	return newstr
}

func (mt mockTerm) ReadKeys(count int) []byte {
	if mt.strVal == nil {
		return nil
	}

	bt := []byte(*mt.strVal)

	if count >= len(bt) {
		mt.strVal = nil
		return bt
	}

	v := (*mt.strVal)[:count]
	mt.strVal = &v

	return bt[:count]
}

func Test_ParseFile(t *testing.T) {
	tests := []struct {
		prg   []byte
		ascii string
	}{
		{prg: []byte{}},
		{prg: []byte{0x00}},
		{prg: []byte{0xFF, 0x82}},
		{prg: []byte{0xFF, 0x82, 0x12, 0x0a, 0x00}},       // test case line with no statements
		{prg: []byte{0xFF, 0x82, 0x12, 0x0a, 0x00, 0x3a}}, // test case line with no statements
		{prg: []byte{0xFF, 0x82, 0x12, 0x00, 0x00, 0x91, 0x00, 0x00, 0x00, 0x1A}}, // test case line with zero linenum
		{prg: []byte{0xFF, 0x82, 0x12, 0x0A, 0x00, 0x91, 0x20, 0x22, 0x48, 0x65,
			0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, 0x22, 0x00, 0x00, 0x00, 0x1A}},
		{prg: []byte{0xFF, 0x82, 0x12, 0x0A, 0x00, 0x91, 0x00, 0x00, 0x00, 0x1A}},
	}

	for _, tt := range tests {
		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		rdr := bufio.NewReader(bytes.NewReader(tt.prg))
		ParseFile(rdr, env)
	}
}

func Test_TokenTable(t *testing.T) {
	tests := []struct {
		inp []byte
		exp []string
	}{
		{inp: []byte{colon_TOK}, exp: []string{":"}},              // special case, colon followed by nothing is legal
		{inp: []byte{colon_TOK, else_TOK}, exp: []string{"ELSE"}}, // special case, colon followed by else is ELSE
		{inp: []byte{int1Byte_TOK, 0x0a}, exp: []string{"10"}},
		{inp: []byte{int1Byte_TOK}, exp: []string{""}}, // error case
		{inp: []byte{dblQuote_TOK}, exp: []string{`""`}},
		{inp: []byte{eol_TOK, end_TOK, stop_TOK, width_TOK, colon_TOK, data_TOK, dblQuote_TOK, 'A', dblQuote_TOK, dim_TOK, else_TOK, end_TOK, for_TOK, goto_TOK, input_TOK},
			exp: []string{"", "END", "STOP", "WIDTH", ":", "DATA", `"A"`, "DIM", "ELSE", "END", "FOR", "GOTO", "INPUT"}},
		{inp: []byte{let_TOK, next_TOK, read_TOK}, exp: []string{"LET", "NEXT", "READ"}},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		rdr := progRdr{src: src}
		for i := range tt.exp {
			res := rdr.readToken()
			assert.Equal(t, tt.exp[i], res, "TokenTable expected %s, got %s", tt.exp[i], res)

		}
	}
}

func Test_DecryptByte(t *testing.T) {
	tests := []struct {
		inp []byte
		out []byte
	}{
		{inp: []byte{0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4},
			out: []byte{0x82, 0x12, 0x0a, 0x00, 0x91, 0x20, 0x22, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x22, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		pr := protReader{}
		for i, bt := range tt.inp {
			bt = pr.decryptByte(bt)

			assert.Equal(t, tt.out[i], bt, "byte decrypt failed")
		}
	}
}

func Test_DecryptBytes(t *testing.T) {
	tests := []struct {
		inp []byte
		out []byte
	}{
		{inp: []byte{0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4},
			out: []byte{0x82, 0x12, 0x0a, 0x00, 0x91, 0x20, 0x22, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x22, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		pr := protReader{}
		bts := pr.decryptBytes(tt.inp)
		for i, bt := range bts {
			assert.Equal(t, tt.out[i], bt, "bytes decrypt failed")
		}
	}
}

func Test_ReadProtProg(t *testing.T) {
	tests := []struct {
		inp   []byte
		stmts int
	}{
		{inp: []byte{0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4}, stmts: 2},
		{inp: []byte{0xCB, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x1C, 0x18, 0x7B, 0x02, 0xC8, 0x87, 0x78, 0xC5, 0x19, 0xCF, 0x94, 0x74, 0x87, 0xA4, 0x6C, 0x03}, stmts: 4},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		pr := protReader{src: src}
		rdr := progRdr{src: &pr}

		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		rdr.readProg(env)
		itr := env.Program.StatementIter()

		assert.Equal(t, tt.stmts, itr.Len(), "Test_readProgProtected expected %d statements, got %d", tt.stmts, itr.Len())
	}
}

func Test_ParseProtProg(t *testing.T) {
	tests := []struct {
		inp   []byte
		stmts int
	}{
		{inp: []byte{0xFE, 0xE3, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x1C, 0x18, 0x7B, 0x02, 0xC8}, stmts: 2}, // test case colon at eol
		{inp: []byte{0xFE, 0xE3, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD}, stmts: 2},                                     // test case dblquote at eol
		{inp: []byte{}, stmts: 0},
		{inp: []byte{0xff}, stmts: 0},
		{inp: []byte{0xfe, 0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4}, stmts: 2},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))

		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		ParseProtectedFile(src, env)

		itr := env.Program.StatementIter()

		assert.Equal(t, tt.stmts, itr.Len(), "Test_ParseProtectedFile")
	}
}
