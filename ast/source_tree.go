package ast

import (
	"strings"

	"github.com/google/btree"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/token"
)

/***************************************************************************/
// First up, the structures dealing with source code.
/***************************************************************************/

// renumberData holds information about a in progress renumber operation
type renumberData struct {
	tempTree *SourceTree
	mapping  map[uint16]uint16
}

// SourceTree is a binary tree of all the source code for the loaded program.
type SourceTree struct {
	curLine uint16       // current source line executing
	tree    *btree.BTree // Holds the source lines for a program
}

// The implementation for sourceTree
// Initialize a new source tree for the environment
func initSourceTree() *SourceTree {
	t := &SourceTree{
		curLine: 1,
		tree:    btree.New(2),
	}

	return t
}

// Puts the passed source line into the tree.
// If it is replacing an existing line, ReplaceOrInsert()
// returns the line.  Callers to this function don't care.
func (srcTree *SourceTree) addSourceLine(sl *SourceLine) Statement {
	// This should not happen
	if sl.lineNum == 0 {
		rc := ErrorStatement{
			Token:  token.Token{Type: token.ERROR, Literal: "ERROR"},
			ErrNum: &IntegerLiteral{Value: berrors.IllegalFuncCallErr},
		}
		return &rc
	}
	srcTree.tree.ReplaceOrInsert(sl)

	return nil
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

// Fetches the first line of code based on line number.
// The iterator for the line is zeroed before returning.
// If the tree is empty, returns a nill value.
func (srcTree *SourceTree) firstLine() *SourceLine {
	srcTree.curLine = 0
	return srcTree.NextLine()
}

// Fetches the next line of code to execute.
// Returns nil if there is no next line.
func (srcTree *SourceTree) NextLine() *SourceLine {
	var got []btree.Item
	l := NewSourceLine("", srcTree.curLine+1)

	srcTree.tree.AscendGreaterOrEqual(l, func(a btree.Item) bool {
		got = append(got, a)
		return true
	})

	// if nothing found, end of tree reached
	if len(got) == 0 {
		return nil
	}

	// Convert the Item to a SourceLine
	nl, _ := got[0].(*SourceLine)
	nl.itr = 0 // assume the statement array has been parsed out

	// remember which line number I'm on
	srcTree.curLine = nl.lineNum

	return nl
}

/***************************************************************************/
// The next section are the functions that work with the AST structures
/***************************************************************************/

// create a new structure and initialize the tree
func newRenumberData() renumberData {
	rd := renumberData{tempTree: initSourceTree(), mapping: make(map[uint16]uint16)}
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
THEN, ON...GOTO, ON...GOSUB, RESTORE, and RESUME statements
are identified as needing updates.  The CLI package will handle
parsing each line and then calling for an update to the various
statements that to be fixed.

Replaces the current source tree in the environment if successful.

If the command fails, returns an Error otherwise nil.
*/

func (srcTree *SourceTree) Renumber(new uint16, old uint16, inc uint16) Statement {
	// create data structure for processing the command
	rd := newRenumberData() // contains a SourceTree and a map of old#->new#

	// go build the new source tree
	rc := srcTree.buildRenumberedTree(&rd, new, old, inc)

	// rc should be nil, if not an error was encountered
	if rc != nil {
		return rc
	}

	// find all the lines and mark them for updating when they are parsed
	srcTree.findJumpLines(&rd)

	return rc
}

// Traverses the tree and renumbers the portion indicated by the "old" param.
func (srcTree *SourceTree) buildRenumberedTree(rd *renumberData, new uint16, old uint16, inc uint16) Statement {
	var rc Statement
	// create data structure for processing the command
	start := NewSourceLine("", old)

	// iterate over the old tree
	srcTree.tree.Ascend(func(a btree.Item) bool {
		// pull the SourceLine out of the item
		ol, _ := a.(*SourceLine)
		if a.Less(start) {
			// add the line as is
			rd.tempTree.addSourceLine(ol)
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
			rd.tempTree.addSourceLine(nl)

			new += inc
		}

		return true
	})

	return rc
}

// Now we need to find all the lines that jump to a line number
// This includes GOTO, ON GOTO, GOSUB, ON GOSUB, THEN, ELSE, RESTORE, RESUME
// We use this to build an array of lines that need to be fixed once the
// renumber operation is complete.
//
// If an invalid item is encountered, a Error object will be returned.
func (srcTree *SourceTree) findJumpLines(rd *renumberData) ([]*SourceLine, Statement) {
	var lines []*SourceLine
	var rc ErrorStatement

	rd.tempTree.tree.Ascend(func(a btree.Item) bool {
		l, ok := a.(*SourceLine)

		// This should not happen
		if !ok {
			rc = ErrorStatement{
				Token:  token.Token{Type: token.ERROR, Literal: "ERROR"},
				ErrNum: &IntegerLiteral{Value: berrors.IllegalFuncCallErr},
			}
			return false
		}

		if strings.Contains(l.source, "GOTO") ||
			strings.Contains(l.source, "GOSUB") ||
			strings.Contains(l.source, "THEN") ||
			strings.Contains(l.source, "ELSE") ||
			strings.Contains(l.source, "RESTORE") ||
			strings.Contains(l.source, "RESUME") {

			lines = append(lines, l)
		}

		return true
	})

	// update the tree pointer and we are done!
	srcTree.tree = rd.tempTree.tree

	if rc.ErrNum != nil {
		return lines, &rc
	}

	return lines, nil
}
