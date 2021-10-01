package ast

import (
	"bytes"
	"errors"

	"github.com/navionguy/basicwasm/berrors"
)

type codeLine struct {
	lineNum int
	stmts   []Statement
	curStmt int
}

func (cl codeLine) String() string {
	var out bytes.Buffer
	for i := range cl.stmts {
		out.WriteString(cl.stmts[i].String())

		if (i+1 < len(cl.stmts)) && (i > 0) {
			out.WriteString(" : ")
		}
	}
	return out.String()
}

// Code allows iterating over the code lines subject to control transfer
type Code struct {
	lines     []codeLine // array of code lines sorted by ascending line number
	currIndex int        // index into lines
	currLine  int        // current line excuting
	err       error
}

// ConstData provides access to DATA elements
type ConstData struct {
	code *Code // pointer to the current lines of code
	line int   // index into code.lines[]
	stmt int   // index into code.lines[line].stmts

	data *DataStatement // the data statment I'm working from
	exp  int            // index into data.exp[]
}

//Program holds the root of the AST (Abstract Syntax Tree)
type Program struct {
	code    *Code
	cmdLine *Code
	data    *ConstData
}

// New setups internal state
func (p *Program) New() {
	var err error
	p.code = &Code{
		lines:    []codeLine{},
		currLine: 0,
		err:      err,
	}
	p.cmdLine = &Code{
		lines:    []codeLine{},
		currLine: 0,
		err:      err,
	}

	//p.code.lines = append(p.code.lines, codeLine{lineNum: 0, curStmt: 0})
	//p.cmdLine.lines = append(p.cmdLine.lines, codeLine{lineNum: 0, curStmt: 0})
	p.code.lines = nil
	p.cmdLine.lines = nil
}

// Parsed lets me know the parser has finished and I should expect the next input from the command line
func (p *Program) Parsed() {
	p.code.lines[0].curStmt = 0
}

// CmdParsed gets the command line ready to execute
func (p *Program) CmdParsed() {
	p.cmdLine.lines[0].curStmt = 0
}

// CmdComplete execution is complete, empty the command line
func (p *Program) CmdComplete() {
	p.cmdLine.lines = nil
}

// TokenLiteral returns string representation of the program
func (p *Program) TokenLiteral() string { return "GWBasic" }

// AddStatement adds a new statement to the AST
func (p *Program) AddStatement(stmt Statement) {
	lNum, ok := stmt.(*LineNumStmt)

	if ok {
		// we are starting a new line
		p.code.addLine(int(lNum.Value))
		p.code.currLine = int(lNum.Value)
	}

	if len(p.code.lines) == 0 {
		p.code.err = errors.New("invalid line number")
		return
	}

	p.code.lines[p.code.currIndex].stmts = append(p.code.lines[p.code.currIndex].stmts, stmt)
}

// AddCmdStmt adds a statement to the command line
// he only ever has one line
func (p *Program) AddCmdStmt(stmt Statement) {
	if len(p.cmdLine.lines) == 0 {
		p.cmdLine.addLine(0)
	}
	p.cmdLine.lines[0].stmts = append(p.cmdLine.lines[0].stmts, stmt)
}

// StatementIter lets them iterate over lines
func (p *Program) StatementIter() *Code {
	p.code.currIndex = 0

	return p.code
}

// CmdLineIter iterates over the command line
func (p *Program) CmdLineIter() *Code {
	if p.cmdLine.Len() > 0 {
		p.cmdLine.lines[0].curStmt = 0
	}
	return p.cmdLine
}

// ConstData returns the ConstData object
func (p *Program) ConstData() *ConstData {
	if p.data == nil {
		var cd ConstData
		p.data = &cd
		p.data.code = p.code
		p.data.data = nil
	}

	return p.data
}

// going to add, or possibly replace, a line of code
func (cd *Code) addLine(lineNum int) {
	// create a new codeLine struct
	nl := codeLine{
		lineNum: lineNum,
	}

	// *most* of the time, adding to the end of the program
	if lineNum > cd.MaxLineNum() {
		cd.lines = append(cd.lines, nl)
		cd.currIndex = len(cd.lines) - 1
		cd.currLine = lineNum
		return
	}

	i, found := cd.findLine(lineNum)

	if found {
		cd.lines[i] = nl
		cd.currIndex = i
		cd.currLine = lineNum
		return
	}

	// insert it into the array
	cd.lines = append(cd.lines[:i], append([]codeLine{nl}, cd.lines[i:]...)...)
	cd.currIndex = i
	cd.currLine = lineNum
}

// CurLine returns the current executing line number or zero if there isn't one
func (cd *Code) CurLine() int {
	return cd.currLine
}

// tries to find the requested line number in the array of lines
// returns index into lines and true if found
// returns index to insert it and false if not found
func (cd *Code) findLine(lNum int) (int, bool) {
	if len(cd.lines) == 0 {
		return 0, false
	}

	// todo: come up with a clever way to do this faster
	for i := range cd.lines {
		if cd.lines[i].lineNum == lNum {
			return i, true //found him!  Just replace with the new version
		}

		if cd.lines[i].lineNum > lNum {
			return i, false // time to insert a new line
		}
	}

	// line doesn't exist
	return 0, false
}

// MaxLineNum finds the highest line number currently in Code
func (cd *Code) MaxLineNum() int {
	// if array of code lines is empty
	if len(cd.lines) == 0 {
		return 0 //return zero
	}
	return cd.lines[len(cd.lines)-1].lineNum
}

// Next tries to move to the next statment
// if I can't find one, returns false
func (cd *Code) Next() bool {

	if cd.currIndex > len(cd.lines)-1 {
		return false
	}

	line := &cd.lines[cd.currIndex]
	line.curStmt++

	if line.curStmt > len(line.stmts)-1 {
		line.curStmt = 0 // reset to start of the line
		cd.currIndex++   // move to the next line

		if cd.currIndex > len(cd.lines)-1 {
			return false
		}
		line = &cd.lines[cd.currIndex]
		line.curStmt = 0

		return (line.curStmt <= len(line.stmts)-1)
	}

	return true
}

// Value sends the next statement
func (cd *Code) Value() Statement {
	if cd.currIndex > len(cd.lines)-1 {
		return nil
	}
	line := &cd.lines[cd.currIndex]

	rc := line.stmts[line.curStmt]
	return rc
}

// Len tells caller how many statements I have, used for unit tests
func (cd *Code) Len() int {
	i := 0

	for _, ln := range cd.lines {
		i += len(ln.stmts)
	}
	return i
}

// Exists just tell you if I could find it
func (cd *Code) Exists(target int) bool {
	_, ok := cd.findLine(target)

	return ok
}

// Jump to the target line in the AST
func (cd *Code) Jump(target int) string {
	i, ok := cd.findLine(target)

	if ok {
		cd.currIndex = i
		return ""
	}
	// stop execution
	cd.currIndex = cd.Len()

	return berrors.TextForError(berrors.UnDefinedLineNumber)
}

// Next returns the next constant data item
func (data *ConstData) Next() *Expression {
	if data.data == nil {
		exp := data.findNextData()

		return exp
	}

	// can I just increment?
	data.exp++
	if data.exp < len(data.data.Consts) {
		// all good
		exp := &data.data.Consts[data.exp]
		return exp
	}

	// go look for more consts
	data.nextStmt()
	exp := data.findNextData()

	return exp
}

// Restore the const scanner to the first data element
func (data *ConstData) Restore() {
	data.exp = 0
	data.line = 0
	data.stmt = 0
}

// RestoreTo a particular point in the constant data
// based on a line number.
// The line number passed has to exist, but doesn't
// have to start with, or even contain a DATA statement
func (data *ConstData) RestoreTo(line int) bool {
	data.Restore()
	index, found := data.code.findLine(line)

	if !found {
		return found
	}

	data.line = index
	return true
}

func (data *ConstData) findNextData() *Expression {
	for ok := false; !ok; {
		stmt := data.value()

		if stmt == nil {
			return nil
		}

		ds, ok := (*stmt).(*DataStatement)

		if ok {
			// found him
			data.data = ds
			data.exp = 0
			return &ds.Consts[0]
		}
		data.nextStmt()
	}
	return nil
}

func (data *ConstData) value() *Statement {
	if data == nil {
		return nil
	}

	if data.line >= len(data.code.lines) {
		return nil
	}

	if data.stmt >= len(data.code.lines[data.line].stmts) {
		// end of line, go to the next one
		data.stmt = 0
		data.line++

		if data.line >= len(data.code.lines) {
			return nil
		}

	}
	return &data.code.lines[data.line].stmts[data.stmt]
}

func (data *ConstData) nextStmt() {
	data.stmt++

	if data.stmt < len(data.code.lines[data.line].stmts) {
		return
	}

	// have to move to the next line
	data.stmt = 0
	data.line++
}
