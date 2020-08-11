package ast

import (
	"bytes"
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

type line struct {
	stmtIndx int16
	Value    string
}

type lines struct {
	stmts   map[int16]line
	current *Statement
	lastAdd int16
}

// Code allows iterating over the code lines subject to control transfer
type Code struct {
	statements []Statement
	lines      map[int16]int16
	currStmt   int16
	nextStmt   int16
	err        error
	pg         *Program
}

//Program holds the root of the AST (Abstract Syntax Tree)
type Program struct {
	code       Code
	maxLineNum int16
}

// New setups internal state
func (p *Program) New() {
	var err error
	p.code = Code{
		statements: []Statement{},
		currStmt:   0,
		nextStmt:   1,
		err:        err,
		pg:         p,
	}
	p.code.lines = make(map[int16]int16)
}

// AddStatement adds a new statement to the AST
func (p *Program) AddStatement(stmt Statement) string {
	p.code.statements = append(p.code.statements, stmt)

	lNum, ok := stmt.(*LineNumStmt)

	if !ok {
		return ""
	}

	lne, err := strconv.Atoi(lNum.Token.Literal)

	if err != nil {
		return fmt.Sprintf("invalid line number %s", lNum.Token.Literal)
	}
	line := int16(lne)

	if line <= p.maxLineNum {
		return fmt.Sprintf("line numbers not sequential after line %d", p.maxLineNum)
	}

	p.maxLineNum = line
	p.code.lines[line] = int16(len(p.code.statements))
	return ""
}

// StatementIter lets them iterate over lines
func (p *Program) StatementIter() *Code {
	return &p.code
}

// Next iter over the statments
func (cd *Code) Next() bool {
	if cd.nextStmt >= int16(len(cd.statements)) {
		return false
	}
	cd.currStmt = cd.nextStmt
	cd.nextStmt++
	return true
}

// Value sends the next statement
func (cd *Code) Value() Statement {
	return cd.statements[cd.currStmt]
}

// Len tells caller how many statements I have, used for unit tests
func (cd *Code) Len() int {
	return len(cd.statements)
}

// Jump to the target line in the AST
func (cd *Code) Jump(line int16) error {
	val, ok := cd.lines[line]
	if !ok {
		cd.nextStmt = int16(len(cd.statements))
		return fmt.Errorf("line %d does not exist", line)
	}
	cd.nextStmt = val
	return nil
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
	return i.Token.Literal
}

//TokenLiteral returns the literal for the root token
func (p *Program) TokenLiteral() string {
	if len(p.code.statements) > 0 {
		return p.code.statements[0].TokenLiteral()
	} else {
		return ""
	}
}

type FunctionLiteral struct {
	Token      token.Token // The 'DEF' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
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
	for _, s := range p.code.statements {
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
	return ls.Token.Literal
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
	Value int16
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
	return es.Token.Literal
}

// String returns text version of my expression
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

/*type DimVar struct {
	Name  Identifier // identifier like "A[]" or "DATAPAIRS[]"
	DData []int      // size of each dimension
}

func (dv *DimVar) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range dv.DData {
		params = append(params, strconv.Itoa(p))
	}

	out.WriteString(dv.Name.String())
	out.WriteString("[")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString("] ")

	return out.String()
}*/

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
func (il *StringLiteral) String() string { return il.Token.Literal }

type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
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
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// String returns my string representation
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString("then")
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
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
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
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
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
func (gsb *GosubStatement) TokenLiteral() string { return gsb.Token.Literal }
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
func (end *EndStatement) TokenLiteral() string { return end.Token.Literal }

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
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
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
func (pe *PrintStatement) TokenLiteral() string { return pe.Token.Literal }

func (pe *PrintStatement) String() string {
	var out bytes.Buffer

	out.WriteString(pe.Token.Literal)
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
func (cls *ClsStatement) TokenLiteral() string { return cls.Token.Literal }

func (cls *ClsStatement) String() string {
	if cls.Param == -1 {
		return "CLS"
	}

	return fmt.Sprintf("CLS %d", cls.Param)
}
