package object

import (
	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
)

/*
type codeLine struct {
	lineNum int
	stmts   []ast.Statement
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
}*/

// New sets up internal state
func NewProgram() *ast.Program {
	var err error
	code := ast.Code{
		SrcCode:  InitSourceTree(),
		CurrLine: nil,
		Err:      err,
	}
	cmdLine := ast.CmdLine{
		SrcLine: NewSourceLine("", 0),
		Err:     err,
	}

	p := ast.Program{
		Code:    &code,
		CmdLine: &cmdLine,
	}

	return &p
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
	if p.code.Len() > 0 {
		p.code.lines[0].curStmt = 0
	}

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

// CurLine returns the current executing line number or zero if there isn't one
func (cd *Code) CurLine() int {
	if cd.currIndex > len(cd.lines)-1 {
		return 0
	}
	return cd.lines[cd.currIndex].lineNum
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

// returns true if line number exists
func (cd *Code) Exists(target int) bool {
	_, ok := cd.findLine(target)

	return ok
}

// Jump to the target line in the AST
func (cd *Code) Jump(target int) int {
	i, ok := cd.findLine(target)

	if ok {
		cd.currIndex = i
		return 0
	}
	// stop execution
	cd.currIndex = cd.Len()

	return berrors.UnDefinedLineNumber
}

// GetReturnPoint sends back the current position in the code
func GetReturnPoint(cd *ast.Code) ast.RetPoint {
	return ast.RetPoint{currIndex: cd.currIndex, currStmt: cd.lines[cd.currIndex].curStmt}
}

// JumpToRetPoint puts us back at the passed RetPoint
func (cd *ast.Code) JumpToRetPoint(rp RetPoint) {
	cd.currIndex = rp.currIndex
	cd.currLine = cd.lines[cd.currIndex].lineNum
	cd.lines[cd.currIndex].curStmt = rp.currStmt
}

// Need to jump to the instruction prior to the target
func (cd *Code) JumpBeforeRetPoint(rp RetPoint) {
	// first move to the return point
	cd.JumpToRetPoint(rp)

	// now back up one statement
	if rp.currStmt > 0 {
		// easy peasy, just decrement it to prior statement
		cd.lines[cd.currIndex].curStmt = rp.currStmt - 1
		return
	}

	// have to move to the previous line
	if cd.currIndex == 0 {
		// not possible, don't know how this could happen but.....
		return
	}

	// last statement of previous line
	cd.currIndex--
	cd.lines[cd.currIndex].curStmt = len(cd.lines[cd.currIndex].stmts) - 1
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
			break
		}

		ds, ok := stmt.(*DataStatement)

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

// value returns the statement that evaluates to a value
// to be READ into a variable
func (data *ConstData) value() Statement {
	if data == nil {
		// shouldn't happen, but if the data item is nil, is the value
		return nil
	}

	if data.line >= len(data.code.lines) {
		// READ has consumed all the data items
		return nil
	}

	if data.stmt >= len(data.code.lines[data.line].stmts) {
		// end of line, go to the next one
		data.stmt = 0
		data.line++

		if data.line >= len(data.code.lines) {
			// no more data to be found
			return nil
		}

	}
	return data.code.lines[data.line].stmts[data.stmt]
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
