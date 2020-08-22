package ast

import (
	"bytes"
	"errors"
	"fmt"
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

//Program holds the root of the AST (Abstract Syntax Tree)
type Program struct {
	code Code
}

// New setups internal state
func (p *Program) New() {
	var err error
	p.code = Code{
		lines:    []codeLine{},
		currLine: 0,
		err:      err,
	}

	p.code.lines = append(p.code.lines, codeLine{lineNum: 0, curStmt: 0})
}

// TokenLiteral returns string representation of the program
func (p *Program) TokenLiteral() string { return p.code.Value().String() }

// AddStatement adds a new statement to the AST
func (p *Program) AddStatement(stmt Statement) {
	//p.code.statements = append(p.code.statements, stmt)

	lNum, ok := stmt.(*LineNumStmt)

	if ok {
		// we are starting a new line
		p.code.addLine(lNum.Value)
		p.code.currLine = lNum.Value
	}

	p.code.lines[p.code.currIndex].stmts = append(p.code.lines[p.code.currIndex].stmts, stmt)
}

// StatementIter lets them iterate over lines
func (p *Program) StatementIter() *Code {
	if p.code.currIndex != 0 {
		p.code.currIndex = 1
	}

	return &p.code
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
	if len(cd.lines) == 1 {
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

	if line.curStmt > len(line.stmts)-1 {
		if !cd.Next() {
			return nil
		}
		rc := cd.Value()
		return rc
	}

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

// Identifier holds the token for the identifier in the statement
type Identifier struct {
	Token token.Token // the token.IDENT token Value string
	Value string
	Type  string
	Index []*IndexExpression
	Array bool
}

func (i *Identifier) expressionNode() {}
func (i *Identifier) String() string  { return i.Value }

// TokenLiteral returns literal value of the identifier
func (i *Identifier) TokenLiteral() string {
	return strings.ToUpper(i.Token.Literal)
}

type FunctionLiteral struct {
	Token      token.Token // The 'DEF' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
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
	Value int
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

// DimStatement holds the dimension data for an Identifier
type DimStatement struct {
	Token token.Token // token.DIM
	Vars  []*Identifier
}

func (ds *DimStatement) statementNode() {}
func (ds *DimStatement) String() string { return ds.Token.Literal }

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

type DblIntegerLiteral struct {
	Token token.Token
	Value int32
}

func (dil *DblIntegerLiteral) expressionNode()      {}
func (dil *DblIntegerLiteral) TokenLiteral() string { return dil.Token.Literal }
func (dil *DblIntegerLiteral) String() string       { return fmt.Sprintf("%d", dil.Value) }

// Fixed Point number
type FixedLiteral struct {
	Token token.Token
	Value decimal.Decimal
}

func (fl *FixedLiteral) expressionNode()      {}
func (fl *FixedLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FixedLiteral) String() string       { return fl.Value.String() }

// FloatSingleLiteral holds a single precision floating point eg. "123.45"
type FloatSingleLiteral struct {
	Token token.Token
	Value float32
}

func (fs *FloatSingleLiteral) expressionNode()      {}
func (fs *FloatSingleLiteral) TokenLiteral() string { return fs.Token.Literal }
func (fs *FloatSingleLiteral) String() string       { return fs.Token.Literal }

type FLoatDoubleLiteral struct {
	Token token.Token
	Value float64
}

func (fd *FLoatDoubleLiteral) expressionNode()      {}
func (fd *FLoatDoubleLiteral) TokenLiteral() string { return fd.Token.Literal }
func (fd *FLoatDoubleLiteral) String() string       { return fd.Token.Literal }

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
	out.WriteString(il.Token.Literal)
	out.WriteString("\"")
	return out.String()
}

type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return strings.ToUpper(ie.Token.Literal) }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	//out.WriteString(ie.Index.String())
	out.WriteString("])")
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
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
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
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
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

//Bl
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

type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
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

// PrintStatement holds everything to control the output
type PrintStatement struct {
	Token      token.Token
	Items      []Expression
	Seperators []string
}

func (pe *PrintStatement) statementNode()       {}
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

// ClsStatement command to clear screen
type ClsStatement struct {
	Token token.Token
	Param int
}

func (cls *ClsStatement) statementNode()       {}
func (cls *ClsStatement) TokenLiteral() string { return strings.ToUpper(cls.Token.Literal) }

func (cls *ClsStatement) String() string {
	if cls.Param == -1 {
		return "CLS"
	}

	return fmt.Sprintf("CLS %d", cls.Param)
}

// RemStatement command to clear screen
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
