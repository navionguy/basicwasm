package mocks

import "fmt"

type MockTerm struct {
	Row     *int
	Col     *int
	StrVal  *string
	SawStr  *string
	SawCls  *bool
	SawBeep *bool
}

func initMockTerm(mt *MockTerm) {
	mt.Row = new(int)
	*mt.Row = 0

	mt.Col = new(int)
	*mt.Col = 0

	mt.StrVal = new(string)
	*mt.StrVal = ""

	mt.SawCls = new(bool)
	*mt.SawCls = false
}

func (mt MockTerm) Cls() {
	*mt.SawCls = true
}

func (mt MockTerm) Print(msg string) {
	fmt.Print(msg)
}

func (mt MockTerm) Println(msg string) {
	fmt.Println(msg)
	if mt.SawStr != nil {
		*mt.SawStr = *mt.SawStr + msg
	}
}

func (mt MockTerm) SoundBell() {
	fmt.Print("\x07")
	*mt.SawBeep = true
}

func (mt MockTerm) Locate(int, int) {
}

func (mt MockTerm) GetCursor() (int, int) {
	return *mt.Row, *mt.Col
}

func (mt MockTerm) Read(col, row, len int) string {
	// make sure your test is correct
	trim := (row-1)*80 + (col - 1)

	tstr := *mt.StrVal

	newstr := tstr[trim : trim+len]

	return newstr
}

func (mt MockTerm) ReadKeys(count int) []byte {
	if mt.StrVal == nil {
		return nil
	}

	bt := []byte(*mt.StrVal)

	if count >= len(bt) {
		mt.StrVal = nil
		return bt
	}

	v := (*mt.StrVal)[:count]
	mt.StrVal = &v

	return bt[:count]
}
