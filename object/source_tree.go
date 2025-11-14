package object

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/btree"
	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
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

// When a line of source code gets renumbered
// we have to update the display string
func (src *SourceLine) fixLineNumber(old uint16) Object {
	ol := strconv.Itoa(int(old))
	nsrc, ok := strings.CutPrefix(src.source, ol)

	if ok {
		src.source = strconv.Itoa(int(src.lineNum)) + nsrc
	} else {
		e := &Error{Code: berrors.IllegalFuncCallErr,
			Message: berrors.TextForError(berrors.IllegalFuncCallErr)}

		return e
	}

	return nil
}

// get the next statement in the source line
// reutrn nil to signal end of line
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

// renumber_data holds information about a in progress renumber operation
type renumber_data struct {
	tempTree *SourceTree
	mapping  map[uint16]uint16
}

// create a new structure and initialize the tree
func new_renumber_data() renumber_data {
	rd := renumber_data{tempTree: InitSourceTree(), mapping: make(map[uint16]uint16)}
	return rd
}

/*
Renumber takes three optional parameters:

	new is the first line number in the new sequence, defaults to 10
	old is the line in the current program where renumber will begin
	inc is the increment to be used in the new sequence, defaults to 10

Command Examples:

	RENUM	Renumbers the entire program first new line number will
			be 10, values will increment by 10.

	RENUM 300,,50	Renumbers the entire program.  The first new line
					number will be 300.  Lines increment by 50.

	RENUm 1000,900,20	Renumbers the lines from 900 up so they start
						with line number 1000 and increment by 20.

Once the renumbered tree is scanned so that all ELSE, GOTO, GOSUB,
THEN, ON...GOTO, ON...GOSUB, RESTORE, RESUME, and ERL statements
reflect the new line numbers.

Replaces the current source tree in the environment if successful.

If the command fails, returns an Error otherwise nil.
*/

func (srcTree *SourceTree) Renumber(new uint16, old uint16, inc uint16) *Object {
	var rc Object
	// create data structure for processing the command
	rd := new_renumber_data()
	start := NewSourceLine("", old)

	// iterate over the old tree
	srcTree.tree.Ascend(func(a btree.Item) bool {
		// pull the SourceLine out of the item
		ol, _ := a.(*SourceLine)
		if a.Less(start) {
			// add the line as is
			rd.tempTree.AddSourceLine(ol)
		} else {
			// create a SourceLine with the existing source but new line number
			nl := NewSourceLine(ol.source, new)
			// save the mapping from old line # to new
			rd.mapping[ol.lineNum] = new
			// go fix the line number in the source line
			rc = nl.fixLineNumber(ol.lineNum)
			if rc != nil {
				return false
			}
			// add tree to the new tree
			rd.tempTree.AddSourceLine(nl)

			new += inc
		}

		return true
	})

	// if no error occurred, go update the source code tree
	if rc == nil {
		srcTree.updateTree(rd)
	}

	return &rc
}

// Now we need to clean up all the various transfers to a line number
func (rd renumber_data) fixUpJumps() {
	rd.tempTree.tree.Ascend(func(a btree.Item) bool {
		l, ok := a.(*SourceLine)

		// This should not happen
		if !ok {
			return true
		}

		if strings.Contains(l.source, "GOTO") ||
			strings.Contains(l.source, "GOSUB") ||
			strings.Contains(l.source, "THEN") ||
			strings.Contains(l.source, "ELSE") ||
			strings.Contains(l.source, "RESTORE") ||
			strings.Contains(l.source, "RESUME") {
			fmt.Println("got em!")
		}

		return true
	})
}

// Once we have renumbered the lines, and fixed all the jumps
// It is time to replace the old tree with the new one
func (srcTree *SourceTree) updateTree(rd renumber_data) {
	srcTree.tree.Clear(false) // ToDo: Add a FreeList to the source tree

	rd.tempTree.tree.Ascend(func(a btree.Item) bool {
		l, ok := a.(*SourceLine)

		if !ok {
			return true
		}
		srcTree.AddSourceLine(l)
		return true
	})
}
