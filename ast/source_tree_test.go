package ast

import (
	"testing"

	"github.com/google/btree"
	"github.com/stretchr/testify/assert"
)

func Test_AddSourceLine(t *testing.T) {
	tests := []struct {
		src  string
		line uint16
	}{
		{src: "10 REM A simple program", line: 10},
		{src: `20 PRINT "Hello World!"`, line: 20},
	}

	st := InitSourceTree()
	for i, tt := range tests {

		assert.NotNil(t, st)
		sl := NewSourceLine(tt.src, tt.line)
		st.AddSourceLine(sl)

		assert.Equal(t, i+1, st.tree.Len())
	}
}

func Test_InitSourceTree(t *testing.T) {
	st := InitSourceTree()

	assert.NotNil(t, st)
}

func Test_JumpToLine(t *testing.T) {
	tests := []struct {
		src1  string
		lnum1 uint16
		src2  string
		lnum2 uint16
		less  bool
	}{
		{`10 PRINT "Hello World"`, 10, `20 REM Comment`, 20, true},
		{`20 REM Comment`, 20, `10 PRINT "Hello World"`, 10, false},
	}

	for _, tt := range tests {
		st := InitSourceTree()
		l1 := NewSourceLine(tt.src1, tt.lnum1)
		st.AddSourceLine(l1)

		l2 := NewSourceLine(tt.src2, tt.lnum2)
		st.AddSourceLine(l2)

		t2 := st.tree.Get(l2)
		assert.NotNil(t, t2)

		t1 := st.tree.Get(l1)
		assert.NotNil(t, t1)

		jl1 := st.JumpToLine(tt.lnum1)
		assert.Equal(t, tt.lnum1, jl1)

		t3 := st.JumpToLine(20)
		assert.NotNil(t, t3)

		// test for a bad goto
		jl2 := st.JumpToLine(100)
		assert.Equal(t, uint16(0), jl2)
	}
}

func Test_NextLine(t *testing.T) {
	tests := []struct {
		txt string
		num uint16
	}{
		{txt: `10 REM Comment`, num: 10},
		{txt: `50 END`, num: 50},
		{txt: `20 Print "Hello World`, num: 20},
		{txt: `40 Print "Goodbye`, num: 40},
		{txt: `30 REM Testing NextLine()`, num: 30},
	}

	// load all the source lines
	st := InitSourceTree()
	for _, tt := range tests {
		sl := NewSourceLine(tt.txt, tt.num)
		st.AddSourceLine(sl)
	}

	// load all the lines in a random-ish order
	// make sure they come back in order
	for range tests {
		l := st.NextLine()

		if l != nil {
			t.Log(l.lineNum)
		}
	}

	assert.Nil(t, st.NextLine())
}

func Test_Renumber(t *testing.T) {
	lines := []struct {
		txt string
		num uint16
	}{
		{txt: `10 REM Comment`, num: 10},
		{txt: `50 END`, num: 50},
		{txt: `20 Print "Hello World"`, num: 20},
		{txt: `40 Print "Goodbye"`, num: 40},
		{txt: `30 REM Testing NextLine()`, num: 30},
	}

	tests := []struct {
		new uint16
		old uint16
		inc uint16
		err bool
	}{
		{new: 15, old: 0, inc: 10, err: false},
		{new: 35, old: 30, inc: 10, err: false},
		// now we force an error
		{new: 15, old: 0, inc: 10, err: true},
	}

	exp := []ExpectedValues{
		{src: []string{
			`15 REM Comment`,
			`25 Print "Hello World"`,
			`35 REM Testing NextLine()`,
			`45 Print "Goodbye"`,
			`55 END`,
		}, lines: []uint16{15, 25, 35, 45, 55}},
		{src: []string{
			`10 REM Comment`,
			`20 Print "Hello World"`,
			`35 REM Testing NextLine()`,
			`45 Print "Goodbye"`,
			`55 END`,
		}, lines: []uint16{10, 20, 35, 45, 55}},
	}

	for i, tt := range tests {
		// load all the source lines
		st := InitSourceTree()

		for _, l := range lines {
			sl := NewSourceLine(l.txt, l.num)
			st.AddSourceLine(sl)
		}

		if tt.err {
			sl := NewSourceLine(`100 REM Forced Error`, 200)
			st.AddSourceLine(sl)
		}
		rc := st.Renumber(tt.new, tt.old, tt.inc)

		_, ok := rc.(*ErrorStatement)
		if !tt.err {
			assert.True(t, ok)

			exp[i].checkFinalLines(t, st)
		} else {
			assert.False(t, ok)
		}
	}
}

func Test_findJumpLines(t *testing.T) {
	tests := []struct {
		inp string
		exp []uint16
		err bool
	}{
		{inp: `10 GOTO 20`, exp: []uint16{25}, err: false},
		{inp: `20 GOTO 10`, exp: []uint16{25}, err: true},
	}

	lineMap := []struct {
		old uint16
		new uint16
	}{
		{20, 25},
	}

	rd := newRenumberData()
	for _, line := range lineMap {
		rd.mapping[line.old] = line.new
	}

	for _, tt := range tests {
		sl := NewSourceLine(tt.inp, 10)
		st := InitSourceTree()
		rd.tempTree.AddSourceLine(sl)
		if tt.err {
			bad := &badTestItem{bad: 0}
			rd.tempTree.tree.ReplaceOrInsert(bad)
		}
		sla, rc := st.findJumpLines(&rd)

		_, ok := rc.(*ErrorStatement)

		if ok {
			assert.True(t, tt.err)
		} else {
			assert.False(t, tt.err)
			assert.Equal(t, sla[0], tt.exp[0])
		}
	}
}

// helper struct to test the final contents of a tree
type ExpectedValues struct {
	src   []string
	lines []uint16
}

// a non-SourceLine Item to push into the tree
type badTestItem struct{ bad uint16 }

func (btt *badTestItem) Less(than btree.Item) bool {
	return true
}

// make sure source code matches expectations
func (ev *ExpectedValues) checkFinalLines(t *testing.T, tree *SourceTree) {
	// first check the line counts match
	assert.Equal(t, len(ev.src), tree.tree.Len())
	assert.Equal(t, len(ev.lines), tree.tree.Len())

	// now validate all the values

	i := uint16(0)
	tree.tree.Ascend(func(a btree.Item) bool {
		l, ok := a.(*SourceLine)

		if !ok {
			return false
		}

		// check for expected values
		assert.Equal(t, l.source, ev.src[i])
		assert.Equal(t, l.lineNum, ev.lines[i])

		i++
		return true
	})
}
