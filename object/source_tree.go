package object

import (
	"github.com/google/btree"
	"github.com/navionguy/basicwasm/ast"
)

// SourceLine gets created with just the line number and the text
// The first time the line executes, the AST for the line is parsed.
// Subsequent execution goes direct to the eval step.
// If the source line is edited by the user, a new struct replaces
// the old one in the btree.
type SourceLine struct {
	lineNum    uint16 // Basic actually has a max line number of 65529
	source     string // text of the source line
	statements []ast.Statement
	itr        uint16 //used to traverse the statements
}

func (sl *SourceLine) Inspect() string { return sl.source }
func (sl *SourceLine) Value() uint16   { return sl.lineNum }

// BTree required interface
func (sl *SourceLine) Less(than btree.Item) bool {
	switch nl := than.(type) {
	case *SourceLine:
		if sl.Value() < nl.Value() {
			return true
		}
	}
	return false
}

// create a new SourceLine
func NewSourceLine(src string, lNumber uint16) *SourceLine {
	return &SourceLine{source: src, lineNum: lNumber, itr: 0}
}

// AppendStatement grows the list of statements on a source line
func (src *SourceLine) AppendStatement(stmt ast.Statement) {
	src.statements = append(src.statements, stmt)
}

// LineLength returns the count of statements in the line
func (src *SourceLine) LineLength() uint16 {
	return uint16(len(src.statements))
}

func (src *SourceLine) NextStatement() ast.Statement {
	if src.itr >= src.LineLength() {
		return nil
	}

	s := src.statements[src.itr]
	src.itr++

	return s
}

// SourceTree is a binary tree of all the source code for the loaded program.
type SourceTree struct {
	cur_line uint16       // current source line executing
	tree     *btree.BTree // Holds the source lines for a program
}

// Initialize a new source tree for the environment
func InitSourceTree() *SourceTree {
	t := &SourceTree{
		cur_line: 1,
		tree:     btree.New(2),
	}

	return t
}

// Puts the passed source line into the tree.
// If it is replacing an existing line, ReplaceOrInsert()
// returns the line.  Callers to this function don't care.
func (srcTree *SourceTree) AddSourceLine(sl *SourceLine) {
	srcTree.tree.ReplaceOrInsert(sl)
}

// Perform a "GOTO" jump to a new source line.
// Unlike a "GOSUB", I don't have to remember how to get back.
func (srcTree *SourceTree) JumpToLine(l uint16) uint16 {
	test := srcTree.tree.Get(NewSourceLine("", l))
	line, ok := test.(*SourceLine)

	if !ok {
		return 0 // line not found
	}

	// only if I found the line
	return line.lineNum
}

// Fetches the next line of code to execute.
func (srcTree *SourceTree) NextLine() *SourceLine {
	var got []btree.Item
	l := NewSourceLine("", srcTree.cur_line+1)

	srcTree.tree.AscendGreaterOrEqual(l, func(a btree.Item) bool {
		got = append(got, a)
		return true
	})

	// if nothing found, end of tree reached
	if len(got) == 0 {
		return nil
	}

	// Try to convert the Item to a SourceLine
	nl, _ := got[0].(*SourceLine)

	// remember which line number I'm on
	srcTree.cur_line = nl.lineNum

	return nl
}
