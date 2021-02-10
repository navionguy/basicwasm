package ast

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/token"
)

// Node defines interface for all node types
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement defines the interface for all statement nodes
type Statement interface {
	Node
	statementNode()
}

//Expression defines interface for all expression nodes
type Expression interface {
	Node
	expressionNode()
}

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
func (cd *Code) Jump(target int) error {
	i, ok := cd.findLine(target)

	if ok {
		cd.currIndex = i
		return nil
	}
	// stop execution
	cd.currIndex = cd.Len()

	return errors.New("Undefined line number")
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

// AutoCommand turns on automatic line numbering during entry
// it comes in two forms
// AUTO [line number][,[increment]]
// AUTO .[,[increment]] where the '.' indicates start at current line
type AutoCommand struct {
	Token     token.Token
	Start     int
	Increment int
	Curr      bool
}

func (ac *AutoCommand) statementNode() {}

// TokenLiteral returns my token literal
func (ac *AutoCommand) TokenLiteral() string { return strings.ToUpper(ac.Token.Literal) }
func (ac *AutoCommand) String() string {
	var out bytes.Buffer

	out.WriteString("AUTO")

	if ac.Start != -1 {
		out.WriteString(fmt.Sprintf(" %d", ac.Start))
	}

	if ac.Curr {
		out.WriteString(" .")
	}

	if ac.Increment != -1 {
		out.WriteString(fmt.Sprintf(", %d", ac.Increment))
	}

	return out.String()
}

// CallExpression is used when calling built in functions
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns my token literal
func (ce *CallExpression) TokenLiteral() string { return strings.ToUpper(ce.Token.Literal) }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

// ClsStatement command to clear screen
type ClsStatement struct {
	Token token.Token
	Param int
}

func (cls *ClsStatement) statementNode() {}

// TokenLiteral returns my token literal
func (cls *ClsStatement) TokenLiteral() string { return strings.ToUpper(cls.Token.Literal) }

func (cls *ClsStatement) String() string {
	if cls.Param == -1 {
		return "CLS"
	}

	return fmt.Sprintf("CLS %d", cls.Param)
}

// Identifier holds the token for the identifier in the statement
type Identifier struct {
	Token token.Token // the token.IDENT token Value string
	Value string
	Type  string
	Index []*IndexExpression
	Array bool
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string {
	var out bytes.Buffer

	if i.Array {
		out.WriteString(i.Value[:len(i.Value)-1])
		for x, ind := range i.Index {
			out.WriteString(ind.String())

			if x+1 < len(i.Index) {
				out.WriteString(",")
			}
		}
		out.WriteString("]")
	} else {
		out.WriteString(i.Value)
	}

	return out.String()
}

// TokenLiteral returns literal value of the identifier
func (i *Identifier) TokenLiteral() string {
	return strings.ToUpper(i.Token.Literal)
}

// FunctionLiteral starts the definition of a user function
type FunctionLiteral struct {
	Token      token.Token // The 'DEF' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode() {}

// TokenLiteral returns my literal
func (fl *FunctionLiteral) TokenLiteral() string { return strings.ToUpper(fl.Token.Literal) }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())

	return out.String()
}

// String returns the program as a string
func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.code.lines {
		out.WriteString(s.String())
	}
	return out.String()
}

// LetStatement holds the assignment expression
type LetStatement struct {
	Token token.Token // the token.LET token Name *Identifier
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode() {}

// TokenLiteral returns literal value of the statement
func (ls *LetStatement) TokenLiteral() string {
	return strings.ToUpper(ls.Token.Literal)
}
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	return out.String()
}

// LineNumStmt holds the line number
type LineNumStmt struct {
	Token token.Token
	Value int32
}

func (lns *LineNumStmt) statementNode() {}

// TokenLiteral returns the literal value
func (lns *LineNumStmt) TokenLiteral() string {
	return lns.Token.Literal
}

func (lns *LineNumStmt) String() string {
	return fmt.Sprintf("%d ", lns.Value)
}

// ExpressionStatement holds an expression
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns my literal
func (es *ExpressionStatement) TokenLiteral() string {
	return strings.ToUpper(es.Token.Literal)
}

// String returns text version of my expression
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// DataStatement how basic did constants
type DataStatement struct {
	Token  token.Token // token.DATA
	Consts []Expression
}

func (ds *DataStatement) statementNode() {}

// String sends my contents
func (ds *DataStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ds.Token.Literal)
	out.WriteString(" ")
	for _, c := range ds.Consts {
		out.WriteString(c.String())
		out.WriteString(", ")
	}

	return out.String()
}

// TokenLiteral returns my token
func (ds *DataStatement) TokenLiteral() string {

	return ds.Token.Literal
}

// DimStatement holds the dimension data for an Identifier
type DimStatement struct {
	Token token.Token // token.DIM
	Vars  []*Identifier
}

func (ds *DimStatement) statementNode() {}

// String displays the statment
func (ds *DimStatement) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, v := range ds.Vars {
		params = append(params, v.String())
	}

	out.WriteString("DIM ")
	out.WriteString(strings.Join(params, ", "))

	tmp := out.String()
	return tmp
}

// TokenLiteral displays the full Dim statement
func (ds *DimStatement) TokenLiteral() string {
	var out bytes.Buffer

	params := []string{}
	for _, v := range ds.Vars {
		params = append(params, v.String())
	}

	out.WriteString(ds.TokenLiteral())
	out.WriteString(" ")
	out.WriteString(strings.Join(params, ", "))

	return out.String()
}

// IntegerLiteral holds an IntegerLiteral eg. "5"
type IntegerLiteral struct {
	Token token.Token
	Value int16
}

func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns literal value
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// String returns value as an integer
func (il *IntegerLiteral) String() string { return fmt.Sprintf("%d", il.Value) }

// DblIntegerLiteral holds a 32bit integer
type DblIntegerLiteral struct {
	Token token.Token
	Value int32
}

func (dil *DblIntegerLiteral) expressionNode() {}

// TokenLiteral returns literal value
func (dil *DblIntegerLiteral) TokenLiteral() string { return dil.Token.Literal }

// String returns value as an integer
func (dil *DblIntegerLiteral) String() string { return fmt.Sprintf("%d!", dil.Value) }

// FixedLiteral is a Fixed Point number
type FixedLiteral struct {
	Token token.Token
	Value decimal.Decimal
}

func (fl *FixedLiteral) expressionNode() {}

// TokenLiteral returns literal value
func (fl *FixedLiteral) TokenLiteral() string { return fl.Token.Literal }

// String returns value as an integer
func (fl *FixedLiteral) String() string { return fl.Value.String() }

// FloatSingleLiteral holds a single precision floating point eg. "123.45"
type FloatSingleLiteral struct {
	Token token.Token
	Value float32
}

func (fs *FloatSingleLiteral) expressionNode() {}

// TokenLiteral returns literal value
func (fs *FloatSingleLiteral) TokenLiteral() string { return fs.Token.Literal }

// String returns value as an integer
func (fs *FloatSingleLiteral) String() string { return fs.Token.Literal }

// FloatDoubleLiteral 64 bit floating point number
type FloatDoubleLiteral struct {
	Token token.Token
	Value float64
}

func (fd *FloatDoubleLiteral) expressionNode() {}

// TokenLiteral returns literal value
func (fd *FloatDoubleLiteral) TokenLiteral() string { return fd.Token.Literal }

// String returns value as an integer
func (fd *FloatDoubleLiteral) String() string { return fd.Token.Literal }

// HexConstant holds values in the from &H76 &H32F
type HexConstant struct {
	Token token.Token
	Value string
}

func (hc *HexConstant) expressionNode() {}

// TokenLiteral returns my literal
func (hc *HexConstant) TokenLiteral() string { return hc.Token.Literal }

// String returns literal as a string
func (hc *HexConstant) String() string {
	var out bytes.Buffer

	out.WriteString(hc.Token.Literal)
	out.WriteString(hc.Value)
	return out.String()
}

// OctalConstant has two form &37 or &O37
type OctalConstant struct {
	Token token.Token
	Value string
}

func (oc *OctalConstant) expressionNode() {}

// TokenLiteral throws back my literal
func (oc *OctalConstant) TokenLiteral() string { return oc.Token.Literal }

// String gives printable version of me
func (oc *OctalConstant) String() string {
	var out bytes.Buffer

	out.WriteString(oc.Token.Literal)
	out.WriteString(oc.Value)
	return out.String()
}

// ReadStatement fills variables from constaint DATA elements
type ReadStatement struct {
	Token token.Token
	Vars  []Expression
}

func (rd *ReadStatement) statementNode() {}

// TokenLiteral returns my literal
func (rd *ReadStatement) TokenLiteral() string { return rd.Token.Literal }

// String sends my contents
func (rd *ReadStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rd.Token.Literal)
	out.WriteString(" ")
	for i, id := range rd.Vars {
		out.WriteString(id.String())
		if (i + 1) < len(rd.Vars) {
			out.WriteString(", ")
		}
	}

	return out.String()
}

// RestoreStatement resets the DATA constant scanner to
// either the beginning or to a specified line number
type RestoreStatement struct {
	Token token.Token
	Line  int
}

func (rs *RestoreStatement) statementNode() {}

// TokenLiteral returns my literal
func (rs *RestoreStatement) TokenLiteral() string { return rs.Token.Literal }

// String sends the original code
func (rs *RestoreStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.Token.Literal)
	if rs.Line > 0 {
		out.WriteString(" ")
		out.WriteString(strconv.Itoa((rs.Line)))
	}

	return out.String()
}

// StringLiteral holds an StringLiteral eg. "Hello World"
type StringLiteral struct {
	Token token.Token
	Value string
}

func (il *StringLiteral) expressionNode() {}

// TokenLiteral returns literal value
func (il *StringLiteral) TokenLiteral() string { return il.Token.Literal }

// String returns literal as a string
func (il *StringLiteral) String() string {
	var out bytes.Buffer
	out.WriteString("\"")
	out.WriteString(il.Value)
	out.WriteString("\"")
	return out.String()
}

// IndexExpression contains the index into an array
type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode() {}

// TokenLiteral returns my literal
func (ie *IndexExpression) TokenLiteral() string { return strings.ToUpper(ie.Token.Literal) }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ie.Index.String())
	return out.String()
}

//PrefixExpression the big one here is - as in -5
type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

//TokenLiteral returns read string of Token
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	return out.String()
}

// InfixExpression things like 5 + 6
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

//TokenLiteral my token
func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

// String the readable version of me
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	return out.String()
}

// GroupedExpression is enclosed in parentheses
type GroupedExpression struct {
	Token token.Token
	Exp   Expression
}

func (ge *GroupedExpression) expressionNode() {}

// TokenLiteral sends back my token
func (ge *GroupedExpression) TokenLiteral() string {
	return ge.Token.Literal
}

// String the readable version of me
func (ge *GroupedExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ge.Exp.String())
	out.WriteString(")")

	return out.String()
}

// IfExpression holds an If expression
type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence Statement
	Alternative Statement
}

func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns my literal
func (ie *IfExpression) TokenLiteral() string { return strings.ToUpper(ie.Token.Literal) }

// String returns my string representation
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("IF")
	out.WriteString(ie.Condition.String())
	out.WriteString("THEN")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		s := ie.Alternative
		out.WriteString(s.String())
	}

	return out.String()
}

// BlockStatement holds a block statement
type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns my literal
func (bs *BlockStatement) TokenLiteral() string { return strings.ToUpper(bs.Token.Literal) }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// ReturnStatement holds a return
type ReturnStatement struct {
	Token    token.Token // the 'return' token
	ReturnTo string      // in gwbasic, you can return to a line # rather thant the point of the GOSUB
}

func (rs *ReturnStatement) statementNode() {}

// TokenLiteral should be RETURN
func (rs *ReturnStatement) TokenLiteral() string { return strings.ToUpper(rs.Token.Literal) }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnTo != "" {
		out.WriteString(rs.ReturnTo)
	}

	return out.String()
}

// GotoStatement triggers a jump
type GotoStatement struct {
	Token token.Token
	Goto  string
}

func (gt *GotoStatement) statementNode() {}

// TokenLiteral should return GOTO
func (gt *GotoStatement) TokenLiteral() string { return gt.Token.Literal }
func (gt *GotoStatement) String() string {
	var out bytes.Buffer

	out.WriteString(gt.TokenLiteral() + " " + gt.Goto)

	return out.String()
}

// GosubStatement call subroutine
type GosubStatement struct {
	Token token.Token
	Gosub string
}

func (gsb *GosubStatement) statementNode() {}

// TokenLiteral should return GOTO
func (gsb *GosubStatement) TokenLiteral() string { return strings.ToUpper(gsb.Token.Literal) }
func (gsb *GosubStatement) String() string {
	var out bytes.Buffer

	out.WriteString(gsb.TokenLiteral() + " " + gsb.Gosub)

	return out.String()
}

// EndStatement signals it is time to quit
type EndStatement struct {
	Token token.Token
}

func (end *EndStatement) statementNode() {}

// TokenLiteral is END
func (end *EndStatement) TokenLiteral() string { return strings.ToUpper(end.Token.Literal) }

// String just prettier TokenLiteral
func (end *EndStatement) String() string {
	var out bytes.Buffer
	out.WriteString(end.TokenLiteral() + " ")
	return out.String()
}

// ListStatement command to clear screen
type ListStatement struct {
	Token  token.Token
	Start  string
	Lrange string
	Stop   string
}

func (lst *ListStatement) statementNode() {}

// TokenLiteral should return LIST
func (lst *ListStatement) TokenLiteral() string { return strings.ToUpper(lst.Token.Literal) }

func (lst *ListStatement) String() string {

	return fmt.Sprintf("LIST %s%s%s", lst.Start, lst.Lrange, lst.Stop)
}

// PrintStatement holds everything to control the output
type PrintStatement struct {
	Token      token.Token
	Items      []Expression
	Seperators []string
}

func (pe *PrintStatement) statementNode() {}

// TokenLiteral returns my token literal
func (pe *PrintStatement) TokenLiteral() string { return strings.ToUpper(pe.Token.Literal) }

func (pe *PrintStatement) String() string {
	var out bytes.Buffer

	out.WriteString(pe.TokenLiteral())
	out.WriteString(" ")

	for i, s := range pe.Items {
		out.WriteString(s.String() + pe.Seperators[i])
	}

	return out.String()
}

// RemStatement holds a comment about the program
type RemStatement struct {
	Token   token.Token
	Comment string
}

func (rem *RemStatement) statementNode() {}

// TokenLiteral should return REM
func (rem *RemStatement) TokenLiteral() string { return strings.ToUpper(rem.Token.Literal) }

func (rem *RemStatement) String() string {
	rc := strings.TrimRight(rem.Comment, " ")
	return fmt.Sprint(rc)
}

// RunCommand clears all variables and starts execution
// RUN linenum starts execution at linenum
type RunCommand struct {
	Token     token.Token
	StartLine int
	LoadFile  string
}

func (run *RunCommand) statementNode() {}

// TokenLiteral should return RUN
func (run *RunCommand) TokenLiteral() string { return strings.ToUpper(run.Token.Literal) }

func (run *RunCommand) String() string {
	if run.StartLine != 0 {
		return fmt.Sprintf("RUN %d", run.StartLine)
	}

	if len(run.LoadFile) > 0 {
		return fmt.Sprintf("RUN %s", run.LoadFile)
	}
	return "RUN"
}

// TroffCommand turns off tracing
type TroffCommand struct {
	Token token.Token
}

func (tof *TroffCommand) statementNode() {}

// TokenLiteral returns my token literal
func (tof *TroffCommand) TokenLiteral() string { return strings.ToUpper(tof.Token.Literal) }

func (tof *TroffCommand) String() string { return tof.TokenLiteral() }

// TronCommand turns on tracing
type TronCommand struct {
	Token token.Token
}

func (ton *TronCommand) statementNode() {}

// TokenLiteral returns my token literal
func (ton *TronCommand) TokenLiteral() string { return strings.ToUpper(ton.Token.Literal) }

func (ton *TronCommand) String() string { return ton.TokenLiteral() }
