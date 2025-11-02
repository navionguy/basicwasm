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
	if src.itr > src.LineLength() {
		return nil
	}

	s := src.statements[src.itr]
	src.itr++

	return s
}

// Change the line number for the source line
// TODO: what does this do to the tree?
func (src *SourceLine) SetLineNumber(num uint16) {
	src.lineNum = num
}

// SourceTree is a binary tree of all the source code for the loaded program.
type SourceTree struct {
	tree *btree.BTree
}

// Initialize a new source tree for the environment
func initSourceTree() *SourceTree {
	t := &SourceTree{tree: btree.New(2)}

	return t
}

// Puts the passed source line into the tree.
// If it is replacing an existing line, ReplaceOrInsert()
// returns the line.  Callers to this function don't care.
func (src *SourceTree) AddSourceLine(sl *SourceLine) {
	src.tree.ReplaceOrInsert(sl)
}
