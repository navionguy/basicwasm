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

// BeepStatement triggers a beep, no parameters
type BeepStatement struct {
	Token token.Token
}

func (bp *BeepStatement) statementNode() {}

// TokenLiteral returns my token literal
func (bp *BeepStatement) TokenLiteral() string { return strings.ToUpper(bp.Token.Literal) }

func (bp *BeepStatement) String() string {
	return "BEEP"
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

// ChainStatement loads a program file
type ChainStatement struct {
	Token  token.Token
	Path   Expression // filespec for file to chain in
	Line   Expression // line number to start execution
	Range  Expression // range of line numbers to delete
	All    bool       // signals keep all variable values
	Delete bool       // specifies a range of lines deleted before Chaining to new overlay
	Merge  bool       // overlays current program with called progarm, files stay open
}

func (chn *ChainStatement) statementNode() {}

// TokenLiteral returns my token literal
func (chn *ChainStatement) TokenLiteral() string { return strings.ToUpper(chn.Token.Literal) }

func (chn *ChainStatement) String() string {
	lit := "CHAIN "
	if chn.Merge {
		lit = lit + "MERGE "
	}
	lit = lit + chn.Path.String()
	if chn.Line != nil {
		lit = lit + ", " + chn.Line.String()
	}
	if chn.All {
		lit = lit + ", ALL"
	}
	if chn.Delete {
		lit = lit + ", DELETE "
		if chn.Range != nil {
			lit = lit + chn.Range.String()
		}
	}
	return lit
}

// Clear clears all variables and closes open file
type ClearCommand struct {
	Token token.Token
	Exp   [3]Expression // parameters I don't support
}

func (clr *ClearCommand) statementNode() {}

func (clr *ClearCommand) TokenLiteral() string { return strings.ToUpper(clr.Token.Literal) }

func (clr *ClearCommand) String() string {
	rc := ""

	// I build the output from right to left
	// it just seemed easier
	for i := 2; i >= 0; i-- {
		if clr.Exp[i] != nil {
			rc = clr.Exp[i].String() + rc
			if i > 0 {
				rc = "," + rc
			}
		}
	}
	rc = "CLEAR " + rc

	return rc
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

// ColorStatement changes foreground/background colors
type ColorStatement struct {
	Token token.Token
	Parms []Expression // 1-3 parameter expressions
}

func (color *ColorStatement) statementNode()       {}
func (color *ColorStatement) TokenLiteral() string { return strings.ToUpper(color.Token.Literal) }
func (color *ColorStatement) String() string {
	var out bytes.Buffer

	out.WriteString("COLOR ")

	for i, e := range color.Parms {
		if i > 0 {
			out.WriteString(",")
		}

		if e != nil {
			out.WriteString(e.String())
		}
	}

	tmp := out.String()
	return tmp
}

type CommonStatement struct {
	Token token.Token
	Vars  []*Identifier
}

func (cmn *CommonStatement) statementNode() {}

func (cmn *CommonStatement) TokenLiteral() string { return strings.ToUpper(cmn.Token.Literal) }

func (cmn *CommonStatement) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, v := range cmn.Vars {
		params = append(params, v.String())
	}

	out.WriteString("COMMON ")
	out.WriteString(strings.Join(params, ", "))

	tmp := out.String()
	return tmp
}

// Cont command means restarting a stopped program
type ContCommand struct {
	Token token.Token
}

func (cnt *ContCommand) statementNode()       {}
func (cnt *ContCommand) TokenLiteral() string { return strings.ToUpper(cnt.Token.Literal) }

func (cnt *ContCommand) String() string { return "CONT" }

// CSRLIN variable serves up the curre
type Csrlin struct {
	Token token.Token
}

func (csr *Csrlin) expressionNode()      {}
func (csr *Csrlin) TokenLiteral() string { return strings.ToUpper(csr.Token.Literal) }
func (csr *Csrlin) String() string       { return csr.Token.Literal + " " }

type EOFExpression struct {
	Token token.Token
}

func (eof *EOFExpression) expressionNode()      {}
func (eof *EOFExpression) TokenLiteral() string { return strings.ToUpper(eof.Token.Literal) }
func (eof *EOFExpression) String() string       { return "" }

// FilesCommand gets list of files from basic server
type FilesCommand struct {
	Token token.Token
	Path  string
}

func (fls *FilesCommand) statementNode() {}

// TokenLiteral returns my token literal
func (fls *FilesCommand) TokenLiteral() string { return strings.ToUpper(fls.Token.Literal) }

func (fls *FilesCommand) String() string {

	fc := "FILES"
	if len(fls.Path) > 0 {
		fc = fc + ` "` + fls.Path + `"`
	}

	return strings.Trim(fc, " ")
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

	out.WriteString("DEF ")
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")=")
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
func (lns *LineNumStmt) TokenLiteral() string { return lns.Token.Literal }

func (lns *LineNumStmt) String() string {
	return fmt.Sprintf("%d ", lns.Value)
}

// Load command loads a source file and optionally starts it
type LoadCommand struct {
	Token    token.Token
	Path     Expression // program file to load
	KeppOpen bool       // keep all open files open
}

func (ld *LoadCommand) statementNode() {}

// TokenLiteral returns my token literal
func (ld *LoadCommand) TokenLiteral() string { return strings.ToUpper(ld.Token.Literal) }

func (ld *LoadCommand) String() string {

	lc := "LOAD " + ld.Path.String()

	if ld.KeppOpen {
		lc = lc + ",R"
	}

	return lc
}

// locate and optional configure the look of the cursor
type LocateStatement struct {
	Token token.Token
	Parms [5]Expression
}

const (
	Lct_row = iota
	Lct_col
	Lct_cursor
	Lct_start
	Lct_stop
)

func (lct *LocateStatement) statementNode() {}

func (lct *LocateStatement) TokenLiteral() string { return strings.ToUpper(lct.Token.Literal) }

func (lct *LocateStatement) String() string {
	stmt := ""

	for i := Lct_stop; i >= Lct_row; i-- {
		if lct.Parms[i] != nil {
			stmt = lct.Parms[i].String() + stmt
		}
		if (stmt != "") && (i > 0) {
			stmt = "," + stmt
		}
	}
	return "LOCATE " + stmt
}

// ColorPalette maps[GWBasicColor]XTermColor
type ColorPalette map[int16]int

// user wants to change the color palette
type PaletteStatement struct {
	Token  token.Token // token.PALETTE
	Attrib Expression  // index of attribute to change
	Color  Expression  // color value to use, array of values for PALETTE USING
	// values below will hold the active palette settings
	CurPalette  ColorPalette // current color mappings for screen
	BasePalette ColorPalette // base mapping of basic colors to xterm colors
}

func (plt *PaletteStatement) statementNode()       {}
func (plt *PaletteStatement) TokenLiteral() string { return strings.ToUpper(plt.Token.Literal) }
func (plt *PaletteStatement) String() string {
	var buf bytes.Buffer

	buf.WriteString("PALETTE")

	// attribute if I have one
	if plt.Attrib != nil {
		buf.WriteString(" " + plt.Attrib.String())
	}

	if plt.Color != nil {
		buf.WriteString("," + plt.Color.String())
	}
	return buf.String()
}

// NewCommand clears the program and variables
type NewCommand struct {
	Token token.Token // my Token
}

func (new *NewCommand) statementNode()       {}
func (new *NewCommand) TokenLiteral() string { return strings.ToUpper(new.Token.Literal) }
func (new *NewCommand) String() string       { return new.Token.Literal + " " }

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

func (ds *DataStatement) statementNode()       {}
func (ds *DataStatement) TokenLiteral() string { return ds.Token.Literal }

// String sends my contents
func (ds *DataStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ds.Token.Literal)
	out.WriteString(" ")
	for i, c := range ds.Consts {
		if i != 0 {
			out.WriteString(", ")
		}
		out.WriteString(c.String())
	}

	return out.String()
}

// DimStatement holds the dimension data for an Identifier
type DimStatement struct {
	Token token.Token // token.DIM
	Vars  []*Identifier
}

func (ds *DimStatement) statementNode()       {}
func (ds *DimStatement) TokenLiteral() string { return ds.Token.Literal }

// String displays the statment
func (ds *DimStatement) String() string {
	var out bytes.Buffer

	out.WriteString("DIM ")
	for i, v := range ds.Vars {
		if i != 0 {
			out.WriteString(", ")
		}
		out.WriteString(v.String())
	}

	return out.String()
}

// Identifier holds the token for the identifier in the statement
type Identifier struct {
	Token token.Token // the token.IDENT token Value string, arrays can be [] or ()
	Value string      // for an array, will always have []
	Type  string
	Index []*IndexExpression
	Array bool
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return strings.ToUpper(i.Token.Literal) }
func (i *Identifier) String() string {
	var out bytes.Buffer

	if i.Array {
		out.WriteString(i.Token.Literal[:len(i.Token.Literal)-1])
		for x, ind := range i.Index {
			out.WriteString(ind.String())

			if x+1 < len(i.Index) {
				out.WriteString(",")
			}
		}
		out.WriteString(i.Token.Literal[len(i.Token.Literal)-1:])
	} else {
		out.WriteString(i.Value)
	}

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

// ReturnStatement holds a return
type ReturnStatement struct {
	Token    token.Token // the 'return' token
	ReturnTo string      // in gwbasic, you can return to a line # rather thant the point of the GOSUB
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return strings.ToUpper(rs.Token.Literal) }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnTo != "" {
		out.WriteString(rs.ReturnTo)
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
	out.WriteString(`"`)
	out.WriteString(il.Value)
	out.WriteString(`"`)
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

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return strings.ToUpper(bs.Token.Literal) }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// GosubStatement call subroutine
type GosubStatement struct {
	Token token.Token
	Gosub int
}

func (gsb *GosubStatement) statementNode() {}

// TokenLiteral should return GOTO
func (gsb *GosubStatement) TokenLiteral() string { return strings.ToUpper(gsb.Token.Literal) }
func (gsb *GosubStatement) String() string {
	var out bytes.Buffer

	out.WriteString("GOSUB " + fmt.Sprintf("%d", gsb.Gosub))

	return out.String()
}

// GotoStatement triggers a jump
type GotoStatement struct {
	Token token.Token
	Goto  string
}

func (gt *GotoStatement) statementNode()       {}
func (gt *GotoStatement) TokenLiteral() string { return gt.Token.Literal }
func (gt *GotoStatement) String() string {
	var out bytes.Buffer

	out.WriteString(gt.TokenLiteral() + " " + gt.Goto)

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
	return strings.ToUpper(rem.Token.Literal) + " " + strings.TrimRight(rem.Comment, " ")
}

// RunCommand clears all variables and starts execution
// RUN linenum starts execution at linenum
type RunCommand struct {
	Token     token.Token
	StartLine int
	LoadFile  Expression
	KeepOpen  bool
}

func (run *RunCommand) statementNode() {}

// TokenLiteral should return RUN
func (run *RunCommand) TokenLiteral() string { return strings.ToUpper(run.Token.Literal) }

func (run *RunCommand) String() string {
	rc := "RUN"

	if run.LoadFile != nil {
		rc = rc + " " + run.LoadFile.String()
	}

	if run.StartLine != 0 {
		rc = rc + " " + strconv.Itoa(run.StartLine)
	}

	if run.KeepOpen {
		rc = rc + ",r"
	}

	return rc
}

// parameter indexs for ScreenStatement
const (
	ScrnMode        = iota // 0
	ScrnColorSwitch        // 1
	ScrnActivePage         // 2
	ScrnViewedPage         // 3
)

// Mode names for ScrnMode (at least the ones I support)
const (
	ScrnModeMDA = iota // 0
	ScrnModeCGA        // 1
)

type ScreenStatement struct {
	Token    token.Token  // stmt token
	Params   []Expression // Parser creates these
	Settings [4]int       // When executed, eval of expressions go here
}

func (scrn *ScreenStatement) statementNode()       {}
func (scrn *ScreenStatement) TokenLiteral() string { return strings.ToUpper(scrn.Token.Literal) }

// String returns the statement and any parameters as a string
func (scrn *ScreenStatement) String() string {
	var out bytes.Buffer

	out.WriteString("SCREEN ")
	for i, p := range scrn.Params {
		out.WriteString(p.String())

		if i+1 < len(scrn.Params) {
			out.WriteString(",")
		}
	}

	return out.String()
}

// InitValue returns the default settings for the screen
func (scrn *ScreenStatement) InitValue() {
	scrn.Settings[ScrnMode] = ScrnModeMDA // monochrome text mode
	scrn.Settings[ScrnColorSwitch] = 1    // color not allowed, 1 is false for MDA only
	scrn.Settings[ScrnActivePage] = 0     // apage ignored
	scrn.Settings[ScrnViewedPage] = 0     // vpage ignored
}

// Stop statement stops execution
type StopStatement struct {
	Token token.Token
}

func (stop *StopStatement) statementNode()       {}
func (stop *StopStatement) TokenLiteral() string { return strings.ToUpper(stop.Token.Literal) }
func (stop *StopStatement) String() string       { return strings.ToUpper(stop.Token.Literal) + " " }

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
