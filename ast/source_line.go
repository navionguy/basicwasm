package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/btree"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/token"
)

// SourceLine gets created with just the line number and the text
// The first time the line executes, the AST for the line is parsed.
// Subsequent execution goes direct to the eval step.
// If the source line is edited by the user, a new struct replaces
// the old one in the btree.
type SourceLine struct {
	lineNum    uint16 // Basic actually has a max line number of 65529
	source     string // text of the source line
	statements []Statement
	itr        uint16 //used to traverse the statements
}

func (sl *SourceLine) TokenLiteral() string { return fmt.Sprint(sl.lineNum) }
func (sl *SourceLine) String() string       { return sl.source }
func (sl *SourceLine) Value() uint16        { return sl.lineNum }

// BTree required interface
func (sl *SourceLine) Less(than btree.Item) bool {
	switch nl := than.(type) {
	case *SourceLine:
		if sl.TokenLiteral() < nl.TokenLiteral() {
			return true
		}
	}
	return false
}

// create a new SourceLine
func NewSourceLine(src string, lNumber uint16) *SourceLine {
	return &SourceLine{source: src, lineNum: lNumber, itr: 0}
}

// AddStatement grows the list of statements on a source line
func (src *SourceLine) AddStatement(stmt Statement) {
	src.statements = append(src.statements, stmt)
}

// When a line of source code gets renumbered
// we have to update the display string
func (src *SourceLine) fixLineNumber(old uint16) Statement {
	ol := strconv.Itoa(int(old))
	nsrc, ok := strings.CutPrefix(src.source, ol)

	if ok {
		src.source = strconv.Itoa(int(src.lineNum)) + nsrc
	} else {
		err := ErrorStatement{Token: token.Token{Type: token.ERROR, Literal: "ERROR"}, ErrNum: &IntegerLiteral{Value: berrors.IllegalFuncCallErr}}

		return &err
	}

	return nil
}

// LineLength returns the count of statements in the line
func (src *SourceLine) LineLength() uint16 {
	return uint16(len(src.statements))
}

// get the next statement in the source line
// return nil to signal end of line
func (src *SourceLine) NextStatement() Statement {
	if src.itr >= src.LineLength() {
		return nil
	}

	s := src.statements[src.itr]
	src.itr++

	return s
}
