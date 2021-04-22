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

func Test_PrintStmt(t *testing.T) {
	tests := []struct {
		prg   []byte
		ascii string
	}{
		/*{prg: []byte{}},
		{prg: []byte{0x00}},
		{prg: []byte{0xFF, 0x82}},
		{prg: []byte{0xFF, 0x82, 0x12, 0x00, 0x00, 0x91, 0x00, 0x00, 0x00, 0x1A}},*/
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
		exp string
	}{
		{inp: []byte{0x81}, exp: "END"},
		{inp: []byte{0x90}, exp: "STOP"},
		{inp: []byte{0xa0}, exp: "WIDTH"},
		{inp: []byte{0x3a}, exp: ":"},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		rdr := progRdr{src: src}
		res := rdr.readToken()
		assert.Equal(t, tt.exp, res, "TokenTable expected %s, got %s", tt.exp, res)
	}
}
