package object

import (
	"github.com/google/btree"
)

// SourceLine gets created with just the line number and the text
// The first time the line executes, the AST for the line is parsed.
// Subsequent execution goes direct to the eval step.
// If the source line is edited by the user, a new struct replaces
// the old one in the btree.
type SourceLine struct {
	lineNum uint16 // Basic actually has a max line number of 65529
	source  string // text of the source line
}

func (sl *SourceLine) Inspect() string { return sl.source }
func (sl *SourceLine) Value() uint16   { return sl.lineNum }

// BTree required interface
func (sl *SourceLine) Less(than btree.Item) bool {
	switch nl := than.(type) {
	case *SourceLine:
		return sl.Value() < nl.Value()
	}
	return false
}

// create a new SourceLine
func NewSourceLine(src string, lNumber uint16) *SourceLine {
	sl := &SourceLine{source: src, lineNum: lNumber}

	return sl
}

type sourceTree struct {
	tree *btree.BTree
}

// Initialize a new source tree for the environment
func initSourceTree() *sourceTree {
	t := &sourceTree{tree: btree.New(2)}

	return t
}

func (src *sourceTree) addSourceLine(sl *SourceLine) {
	src.tree.ReplaceOrInsert(sl)
}
