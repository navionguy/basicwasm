package mocks

import (
	"encoding/hex"
	"fmt"
	"math"
	"strings"
)

type Expector struct {
	Failed bool
	Exp    []string
}

func (ep *Expector) chkExpectations(msg string) {
	if (ep == nil) || (ep.Exp == nil) {
		return // no Expectations
	}
	if !strings.EqualFold(msg, ep.Exp[0]) {
		fmt.Printf("Got    -> %s\n", hex.EncodeToString([]byte(msg)))
		fmt.Printf("wanted -> %s\n", hex.EncodeToString([]byte(ep.Exp[0])))
		ep.Failed = true
	}
	if len(ep.Exp) > 1 {
		ep.Exp = ep.Exp[1:]
		return
	}

	ep.Exp = nil
}

type MockTerm struct {
	Row      *int
	Col      *int
	StrVal   *string
	SawStr   *string
	SawCls   *bool
	SawBeep  *bool
	SawBreak *bool
	ExpMsg   *Expector
	Delay    *int    // tells me to wait before printing a msg
	DMsg     *string // delayed message
}

func InitMockTerm(mt *MockTerm) {
	mt.Row = new(int)
	*mt.Row = 0

	mt.Col = new(int)
	*mt.Col = 0

	mt.StrVal = new(string)
	*mt.StrVal = ""

	mt.SawCls = new(bool)
	*mt.SawCls = false

	mt.Delay = new(int)
	*mt.Delay = 0

	mt.DMsg = new(string)
	*mt.DMsg = ""
}

func (mt MockTerm) Cls() {
	*mt.SawCls = true
}

func (mt MockTerm) Print(msg string) {
	if *mt.Delay > 0 {
		*mt.DMsg = msg
		return
	}
	fmt.Print(msg)
	mt.ExpMsg.chkExpectations(msg)
	*mt.Col += len(msg)
	if strings.Contains(msg, "\r") {
		*mt.Col = 0
	}
	if strings.Contains(msg, "\n") {
		*mt.Row++
	}
	if *mt.Col > 80 {
		*mt.Row++
		*mt.Col = *mt.Col - 80
	}
}

func (mt MockTerm) Println(msg string) {
	fmt.Println(msg)
	if mt.SawStr != nil {
		*mt.SawStr = *mt.SawStr + msg
	}
	mt.ExpMsg.chkExpectations(msg)
	*mt.Row++
	*mt.Col = 0
}

func (mt MockTerm) SoundBell() {
	fmt.Print("\x07")
	*mt.SawBeep = true
}

func (mt MockTerm) Log(msg string) {
	fmt.Println(msg)
}

func (mt MockTerm) Locate(int, int) {
}

func (mt MockTerm) GetCursor() (int, int) {
	// if there is a string ready for display
	// and delay is > 0, decrement delay and
	// print the string if delay == 0

	if len(*mt.DMsg) > 0 {
		if *mt.Delay > 0 {
			*mt.Delay--
			if *mt.Delay == 0 {
				mt.Print(*mt.DMsg)
				*mt.DMsg = ""
			}
		}
	}
	return *mt.Row, *mt.Col
}

func (mt MockTerm) Read(col, row, length int) string {
	tstr := []byte(*mt.StrVal)

	l := int(math.Min(float64(length), float64(len(tstr))))
	c := int(math.Max(float64(col-1), float64(0)))

	newstr := tstr[c : l+c]

	return string(newstr)
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

func (mt MockTerm) BreakCheck() bool {
	if mt.SawBreak == nil {
		return false
	}

	return *mt.SawBreak
}
