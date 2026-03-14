package ast

import (
	"github.com/google/btree"
)

// ConstData provides access to DATA elements
type ConstData struct {
	code *Code  // pointer to the current lines of code
	line uint16 // index into code.lines[]
	stmt uint16 // index into code.lines[]
	// index into code.lines[line].stmts

	data *DataStatement // the data statment I'm working from
	exp  uint16         // index into data.exp[]
}

func (cd *ConstData) statementNode()       {}
func (cd *ConstData) TokenLiteral() string { return "DATA" }
func (cd *ConstData) String() string       { return "DATA" }

// TODO this logic is just stubbed in
func (cd *ConstData) Next() Expression {
	return cd.data.Consts[cd.stmt]
}

// Move back to the beginning of the static data
func (cd *ConstData) Restore() {
	cd.line = 0
	cd.stmt = 0
	cd.exp = 0
}

// Move back to a specific point in the data
// TODO need to check if line number is valid, return false if not
func (cd *ConstData) RestoreTo(l uint16) bool {
	cd.line = l
	cd.stmt = 0
	cd.exp = 0

	return true
}

type constDataValues struct {
	line  uint16 // line number for this data
	start uint16 // index of first data value statement
	end   uint16 // index of final data value statement
}

// Test if the passed data value is a lower line number
// than this line.  A BTree required interface
func (cdv *constDataValues) Less(than btree.Item) bool {
	switch d := than.(type) {
	case *constDataValues:
		return (cdv.line < d.line)
	}
	return false
}
