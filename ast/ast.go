// Defines all statements, commands, and expressions that form the Abstract Syntax Tree (AST)
package ast

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

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

// Expression defines interface for all expression nodes
type Expression interface {
	Node
	expressionNode()
	TrashCan
}

// TrashCan defines interface for all nodes that store parser trash
// This allows the evaluation loop to catch when ast.Node has trash
// and simply return a syntax error.
type TrashCan interface {
	HasTrash() bool
}

// AutoCommand turns on automatic line numbering during entry
// it comes in two forms
// AUTO [line number][,[increment]]
// AUTO .[,[increment]] where the '.' indicates start at current line
type AutoCommand struct {
	Token  token.Token
	Params []Expression
	On     bool
	Trash  []TrashStatement
}

func (ac *AutoCommand) statementNode()       {}
func (ac *AutoCommand) TokenLiteral() string { return strings.ToUpper(ac.Token.Literal) }
func (ac *AutoCommand) HasTrash() bool       { return len(ac.Trash) > 0 }
func (ac *AutoCommand) String() string {
	var out bytes.Buffer

	out.WriteString("AUTO")

	for i, p := range ac.Params {
		if i > 0 {
			out.WriteString(",")
		}
		out.WriteString(" " + p.String())
	}

	if ac.HasTrash() {
		out.WriteString(Trash(ac.Trash))
	}

	return out.String()
}

// BeepStatement triggers a beep, no parameters
type BeepStatement struct {
	Token token.Token
	Trash []TrashStatement
}

func (bp *BeepStatement) statementNode()       {}
func (bp *BeepStatement) TokenLiteral() string { return strings.ToUpper(bp.Token.Literal) }
func (bp *BeepStatement) HasTrash() bool       { return len(bp.Trash) > 0 }
func (bp *BeepStatement) String() string {
	var out bytes.Buffer
	out.WriteString("BEEP")

	if bp.HasTrash() {
		out.WriteString(Trash(bp.Trash))
	}

	return out.String()
}

// the expression that forms the user defined function
type BlockExpression struct {
	Token token.Token
	Exp   Expression
}

func (be *BlockExpression) statementNode()       {}
func (be *BlockExpression) TokenLiteral() string { return "" }

func (be *BlockExpression) String() string {
	var out bytes.Buffer

	out.WriteString(" " + be.Exp.String())

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

// holds a call to builtin function
type BuiltinExpression struct {
	Token  token.Token  // literal will hold the function name
	Params []Expression //
	Trash  []TrashStatement
}

func (bi *BuiltinExpression) expressionNode()      {}
func (bi *BuiltinExpression) TokenLiteral() string { return strings.ToUpper(bi.Token.Literal) }
func (bi *BuiltinExpression) HasTrash() bool       { return len(bi.Trash) > 0 }
func (bi *BuiltinExpression) String() string {
	var out bytes.Buffer

	out.WriteString(bi.TokenLiteral() + "(")

	for i, p := range bi.Params {
		if i > 0 {
			out.WriteString(",")
		}
		out.WriteString(p.String())
	}

	out.WriteString(")")
	out.WriteString(Trash(bi.Trash))

	return out.String()
}

// CallExpression is used when calling built in functions
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
	Trash     []TrashStatement
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return strings.ToUpper(ce.Token.Literal) }
func (ce *CallExpression) HasTrash() bool       { return len(ce.Trash) > 0 }

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

	out.WriteString(Trash(ce.Trash))

	return out.String()
}

// ChainStatement loads a program file
type ChainStatement struct {
	Token  token.Token
	Path   Expression       // filespec for file to chain in
	Line   Expression       // line number to start execution
	Range  Expression       // range of line numbers to delete
	All    bool             // signals keep all variable values
	Delete bool             // specifies a range of lines deleted before Chaining to new overlay
	Merge  bool             // overlays current program with called program, files stay open
	Trash  []TrashStatement // Stuff that could not be parsed
}

func (chn *ChainStatement) statementNode() {}

// TokenLiteral returns my token literal
func (chn *ChainStatement) TokenLiteral() string { return strings.ToUpper(chn.Token.Literal) }
func (chn *ChainStatement) HasTrash() bool       { return len(chn.Trash) > 0 }

func (chn *ChainStatement) String() string {
	var out bytes.Buffer

	out.WriteString(chn.TokenLiteral())

	if chn.Merge {
		out.WriteString(" MERGE")
	}

	if chn.Path != nil {
		out.WriteString(" " + chn.Path.String())
	}

	if chn.Line != nil {
		if !chn.Line.HasTrash() {
			out.WriteString(", ")
		} else {
			out.WriteString(",")
		}
		out.WriteString(chn.Line.String())
	}

	if chn.All {
		if chn.Line == nil {
			out.WriteString(",")
		}
		out.WriteString(", ALL")
	}
	if chn.Delete {
		out.WriteString(", DELETE")
		if chn.Range != nil {
			out.WriteString(" " + chn.Range.String())
		}
	}

	out.WriteString(Trash(chn.Trash))
	return out.String()
}

type ChDirStatement struct {
	Token token.Token
	Path  []Expression // should be the directory to change to
}

func (cd *ChDirStatement) statementNode()       {}
func (cd *ChDirStatement) TokenLiteral() string { return strings.ToUpper(cd.Token.Literal) }
func (cd *ChDirStatement) String() string {
	var out bytes.Buffer

	out.WriteString(cd.TokenLiteral() + " ")

	for _, it := range cd.Path {
		out.WriteString(it.String())
	}

	return out.String()
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

// CloseStatement closes an open file or COM port
type CloseStatement struct {
	Token token.Token  // Literal "CLOSE"
	Files []FileNumber // array of file numbers to close
}

func (cls *CloseStatement) statementNode()       {}
func (cls *CloseStatement) TokenLiteral() string { return cls.Token.Literal }
func (cls *CloseStatement) String() string {

	var out bytes.Buffer
	out.WriteString(cls.Token.Literal + " ")

	for i, f := range cls.Files {
		out.WriteString(f.String())
		// you can close more than one file at a time
		if i+1 < len(cls.Files) {
			out.WriteString(", ")
		}
	}

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

// ColorStatement changes foreground/background colors
type ColorStatement struct {
	Token token.Token
	Parms []Expression     // 1-3 parameter expressions
	Trash []TrashStatement // Stuff that could not be parsed
}

func (color *ColorStatement) statementNode()       {}
func (color *ColorStatement) TokenLiteral() string { return strings.ToUpper(color.Token.Literal) }
func (color *ColorStatement) HasTrash() bool       { return len(color.Trash) > 0 }
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

	out.WriteString(Trash(color.Trash))

	return out.String()
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

	return out.String()
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
	Trash []TrashStatement
}

func (csr *Csrlin) expressionNode()      {}
func (csr *Csrlin) TokenLiteral() string { return strings.ToUpper(csr.Token.Literal) }
func (csr *Csrlin) HasTrash() bool       { return len(csr.Trash) > 0 }
func (csr *Csrlin) String() string {
	var out bytes.Buffer

	out.WriteString(csr.Token.Literal + " ")

	out.WriteString(Trash(csr.Trash))

	return out.String()
}

// the end of the program file
type EOFExpression struct {
	Token token.Token
	Trash []TrashStatement
}

func (eof *EOFExpression) expressionNode()      {}
func (eof *EOFExpression) TokenLiteral() string { return strings.ToUpper(eof.Token.Literal) }
func (eof *EOFExpression) HasTrash() bool       { return len(eof.Trash) > 0 }
func (eof *EOFExpression) String() string       { return Trash(eof.Trash) }

// signal that an error has occurred
type ErrorStatement struct {
	Token  token.Token
	ErrNum Expression // should evaluate to an integer value
	Resume RetPoint   // set by OnError, where to resume at after handling the error
}

func (err *ErrorStatement) statementNode()       {}
func (err *ErrorStatement) TokenLiteral() string { return strings.ToUpper(err.Token.Literal) }
func (err *ErrorStatement) String() string {
	var out bytes.Buffer

	out.WriteString(err.TokenLiteral() + " " + err.ErrNum.String())

	return out.String()
}

// FileNumber holds the I/O identity of an open file
type FileNumber struct {
	Token token.Token
	Numbr Node
}

func (fn *FileNumber) expressionNode()      {}
func (fn *FileNumber) TokenLiteral() string { return fn.Token.Literal }
func (fn *FileNumber) String() string {
	var out bytes.Buffer

	out.WriteString(fn.TokenLiteral())
	if fn.Numbr != nil {
		out.WriteString(fn.Numbr.String())
	}

	return out.String()
}

// FilesCommand gets list of files from basic server
type FilesCommand struct {
	Token token.Token
	Path  []Expression
}

func (fls *FilesCommand) statementNode() {}

// TokenLiteral returns my token literal
func (fls *FilesCommand) TokenLiteral() string { return strings.ToUpper(fls.Token.Literal) }

func (fls *FilesCommand) String() string {

	fc := "FILES"
	if len(fls.Path) > 0 {
		for i, fp := range fls.Path {
			if i > 0 {
				fc = fc + `,`
			}
			fc = fc + ` ` + fp.String()
		}
	}

	return strings.Trim(fc, " ")
}

type ForStatement struct {
	Token token.Token
	Init  *LetStatement // assigns starting value
	Final []Expression  // loop ends when this value reached
	Step  []Expression  // value to increment /decrement
}

func (four *ForStatement) statementNode()       {}
func (four *ForStatement) TokenLiteral() string { return four.Token.Literal }
func (four *ForStatement) String() string {
	var out bytes.Buffer

	out.WriteString("FOR")
	out.WriteString(four.Init.String())
	out.WriteString(" TO ")
	out.WriteString(four.Final[0].String())
	if len(four.Step) > 0 {
		out.WriteString(" STEP ")
		out.WriteString(four.Step[0].String())
	}

	return out.String()
}

// FunctionLiteral starts the definition of a user function
type FunctionLiteral struct {
	Token      token.Token // The 'DEF' token
	Parameters []*Identifier
	Body       *BlockStatement
	Trash      []TrashStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return strings.ToUpper(fl.Token.Literal) }
func (fl *FunctionLiteral) HasTrash() bool       { return len(fl.Trash) > 0 }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(" ")
	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") =")
	if fl.Body != nil {
		out.WriteString(fl.Body.String())
	}

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

// holds the key settings in the environment settings
type KeySettings struct {
	Disp   bool              // if true, show current key values at bottom of screen
	Keys   map[string]string // scan code in hex maps to key macro
	OnKeys []byte            // keys that will be used in ON KEY statements
}

func (kset *KeySettings) statementNode()       {}
func (kset *KeySettings) TokenLiteral() string { return "KEYS" }
func (kset *KeySettings) String() string {
	var out bytes.Buffer

	for i := 1; i < 11; i++ {
		k := fmt.Sprintf("F%d", i)

		v := kset.Keys[k]

		out.WriteString(k + " " + v + "\r\n")
	}

	return out.String()
}

type KeyStatement struct {
	Token token.Token  // "KEY"
	Param Expression   // ON, OFF, 1...
	Data  []Expression // string to assign to key
}

func (key *KeyStatement) statementNode() {}

func (key *KeyStatement) TokenLiteral() string { return strings.ToUpper(key.Token.Literal) }

func (key *KeyStatement) String() string {
	var out bytes.Buffer

	out.WriteString(key.TokenLiteral() + " ")
	out.WriteString(key.Param.String())
	for _, d := range key.Data {
		out.WriteString(", ")
		if key.Data != nil {
			out.WriteString(d.String())
		}
	}

	return out.String()
}

// LetStatement holds the assignment expression
type LetStatement struct {
	Token token.Token // the token.LET token Name *Identifier
	Name  *Identifier
	Value Expression
	Trash []TrashStatement
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return strings.ToUpper(ls.Token.Literal) }
func (ls *LetStatement) HasTrash() bool       { return len(ls.Trash) > 0 }

func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(Trash(ls.Trash))

	return out.String()
}

// LineNumStmt holds the line number
type LineNumStmt struct {
	Token token.Token
	Value int32
	Trash []TrashStatement
}

func (lns *LineNumStmt) statementNode()       {}
func (lns *LineNumStmt) TokenLiteral() string { return lns.Token.Literal }
func (lns *LineNumStmt) HasTrash() bool       { return len(lns.Trash) > 0 }

func (lns *LineNumStmt) String() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%d ", lns.Value))

	out.WriteString(Trash(lns.Trash))
	return out.String()
}

// Load command loads a source file and optionally starts it
type LoadCommand struct {
	Token    token.Token
	Path     Expression // program file to load
	KeepOpen bool       // keep all open files open
	Trash    []TrashStatement
}

func (ld *LoadCommand) statementNode()       {}
func (ld *LoadCommand) TokenLiteral() string { return strings.ToUpper(ld.Token.Literal) }
func (ld *LoadCommand) HasTrash() bool       { return len(ld.Trash) > 0 }

func (ld *LoadCommand) String() string {
	var out bytes.Buffer

	out.WriteString("LOAD " + ld.Path.String())

	if ld.KeepOpen {
		out.WriteString(",R")
	}

	out.WriteString(Trash(ld.Trash))

	return out.String()
}

// locate and optional configure the look of the cursor
type LocateStatement struct {
	Token token.Token
	Parms []Expression
}

func (lct *LocateStatement) statementNode()       {}
func (lct *LocateStatement) TokenLiteral() string { return strings.ToUpper(lct.Token.Literal) }

func (lct *LocateStatement) String() string {
	var out bytes.Buffer
	out.WriteString("LOCATE ")

	for i, pp := range lct.Parms {
		if i > 0 {
			out.WriteString(",")
		}

		if pp != nil {
			out.WriteString(pp.String())
		}
	}

	return out.String()
}

// ColorPalette maps[GWBasicColor]XTermColor
type ColorPalette map[int16]string

// user wants to change the color palette
type PaletteStatement struct {
	Token  token.Token      // token.PALETTE
	Attrib Expression       // index of attribute to change
	Color  Expression       // color value to use, array of values for PALETTE USING
	Trash  []TrashStatement // any un-parsable stuff

	// values below will hold the active palette settings
	// defaults are set in evaluator.evalPaletteDefault()
	//
	Foreground     ColorPalette // current color mappings for screen foreground
	Background     ColorPalette // current color mappings for screen background
	BaseForeground ColorPalette // base mapping of basic colors to xterm colors
	BaseBackground ColorPalette // base mapping of basic background colors to xterm colors
}

func (plt *PaletteStatement) statementNode()       {}
func (plt *PaletteStatement) TokenLiteral() string { return strings.ToUpper(plt.Token.Literal) }
func (plt *PaletteStatement) HasTrash() bool       { return len(plt.Trash) > 0 }
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

	if plt.HasTrash() {
		buf.WriteString(Trash(plt.Trash))
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

// NextStatement
type NextStatement struct {
	Token token.Token
	Id    Identifier // for loop iterator id, not required
	Trash []TrashStatement
}

func (nxt *NextStatement) statementNode()       {}
func (nxt *NextStatement) TokenLiteral() string { return strings.ToUpper(nxt.Token.Literal) }
func (nxt *NextStatement) HasTrash() bool       { return len(nxt.Trash) > 0 }
func (nxt *NextStatement) String() string {
	var out bytes.Buffer
	out.WriteString(nxt.Token.Literal)

	if len(nxt.Id.Token.Literal) > 0 {
		out.WriteString(" " + nxt.Id.String())
	}

	out.WriteString(Trash(nxt.Trash))

	return out.String()
}

// OffExpression used as a param to KEY statement
type OffExpression struct {
	Token token.Token
	Trash []TrashStatement
}

func (off *OffExpression) expressionNode()      {}
func (off *OffExpression) TokenLiteral() string { return off.Token.Literal }
func (off *OffExpression) HasTrash() bool       { return len(off.Trash) > 0 }
func (off *OffExpression) String() string       { return "OFF" + Trash(off.Trash) }

// OnExpression used as a param to KEY statement
type OnExpression struct {
	Token token.Token
	Trash []TrashStatement
}

func (on *OnExpression) expressionNode()      {}
func (on *OnExpression) TokenLiteral() string { return on.Token.Literal }
func (on *OnExpression) HasTrash() bool       { return len(on.Trash) > 0 }
func (on *OnExpression) String() string       { return "ON" + Trash(on.Trash) }

// OnErrorGoto statement transfers execution when an error occurs
type OnErrorGoto struct {
	Token token.Token // "ON ERROR GOTO"
	Jump  int         // line number to continue from
}

func (oer *OnErrorGoto) statementNode()       {}
func (oer *OnErrorGoto) TokenLiteral() string { return oer.Token.Literal }

// serialize into a string
func (oer *OnErrorGoto) String() string {
	var out bytes.Buffer
	out.WriteString(oer.Token.Literal)
	if oer.Jump > 0 {
		out.WriteString(fmt.Sprintf(" %d", oer.Jump))
	}

	return out.String()
}

// OnGoStatement handles both GOSUB and GOSUB
type OnGoStatement struct {
	Token  token.Token // should be the "GO"
	Exp    Expression  // expression to evaluate
	MidTok token.Token // could be "GOSUB" or "GOTO"
	Jumps  []Expression
}

func (og *OnGoStatement) statementNode()       {}
func (og *OnGoStatement) TokenLiteral() string { return og.Token.Literal }

// serialize into a string
func (og *OnGoStatement) String() string {
	var out bytes.Buffer
	out.WriteString(og.TokenLiteral() + " ")
	out.WriteString(og.Exp.String() + " ")
	out.WriteString(og.MidTok.Literal + " ")
	for i, j := range og.Jumps {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(fmt.Sprint(j))
	}
	return out.String()
}

// ExpressionStatement holds an expression
type ExpressionStatement struct {
	Token      token.Token      // the first token of the expression
	Expression Expression       // the parsed expression
	Trash      []TrashStatement // Stuff that could not be parsed
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return strings.ToUpper(es.Token.Literal) }
func (es *ExpressionStatement) HasTrash() bool       { return len(es.Trash) > 0 }

// String returns text version of my expression
func (es *ExpressionStatement) String() string {
	var out bytes.Buffer

	// expression statement handles trash a little differently
	if es.HasTrash() {
		for i, tr := range es.Trash {
			if i > 0 {
				out.WriteString(" ")
			}
			out.WriteString(tr.String())
		}

		return out.String()
	}

	if es.Token.Literal != "" {
		out.WriteString(es.Token.Literal)
	}

	if es.Expression != nil {
		_, ok := es.Expression.(*FunctionLiteral)

		if !ok {
			out.WriteString(" = ")
		}
		out.WriteString(es.Expression.String())
	}
	return out.String()
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
	Trash []TrashStatement
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return strings.ToUpper(i.Token.Literal) }
func (i *Identifier) HasTrash() bool       { return len(i.Trash) > 0 }
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

	if len(i.Trash) > 0 {
		out.WriteString(Trash(i.Trash))
	}

	return out.String()
}

// IntegerLiteral holds an IntegerLiteral eg. "5"
type IntegerLiteral struct {
	Token token.Token
	Value int16
	Trash []TrashStatement
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) HasTrash() bool       { return len(il.Trash) > 0 }

// String returns value as an integer
func (il *IntegerLiteral) String() string {
	if !il.HasTrash() {
		return fmt.Sprintf("%d", il.Value)
	}

	return Trash(il.Trash)
}

// DblIntegerLiteral holds a 32bit integer
type DblIntegerLiteral struct {
	Token token.Token
	Value int32
	Trash []TrashStatement
}

func (dil *DblIntegerLiteral) expressionNode()      {}
func (dil *DblIntegerLiteral) TokenLiteral() string { return dil.Token.Literal }
func (dil *DblIntegerLiteral) HasTrash() bool       { return len(dil.Trash) > 0 }

// String returns value as an integer
func (dil *DblIntegerLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%d", dil.Value))

	if dil.HasTrash() {
		out.WriteString(Trash(dil.Trash))
	}

	return out.String()
}

// FixedLiteral is a Fixed Point number
type FixedLiteral struct {
	Token token.Token
	Value token.Token
	Trash []TrashStatement
}

func (fl *FixedLiteral) expressionNode()      {}
func (fl *FixedLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FixedLiteral) HasTrash() bool       { return len(fl.Trash) > 0 }

// return the value as a string
func (fl *FixedLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(fl.Value.Literal)

	out.WriteString(Trash(fl.Trash))
	return out.String()
}

// FloatSingleLiteral holds a single precision floating point eg. "123.45"
type FloatSingleLiteral struct {
	Token token.Token
	Value float32
	Trash []TrashStatement
}

func (fs *FloatSingleLiteral) expressionNode()      {}
func (fs *FloatSingleLiteral) TokenLiteral() string { return fs.Token.Literal }
func (fs *FloatSingleLiteral) HasTrash() bool       { return len(fs.Trash) > 0 }

// String returns value as an integer
func (fs *FloatSingleLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(fs.Token.Literal)

	if fs.HasTrash() {
		out.WriteString(Trash(fs.Trash))
	}
	return out.String()
}

// FloatDoubleLiteral 64 bit floating point number
type FloatDoubleLiteral struct {
	Token token.Token
	Value float64
	Trash []TrashStatement
}

func (fd *FloatDoubleLiteral) expressionNode()      {}
func (fd *FloatDoubleLiteral) TokenLiteral() string { return fd.Token.Literal }
func (fd *FloatDoubleLiteral) HasTrash() bool       { return len(fd.Trash) > 0 }

// String returns value as a string
func (fd *FloatDoubleLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(fd.Token.Literal)

	if fd.HasTrash() {
		out.WriteString(Trash(fd.Trash))
	}
	return out.String()
}

// HexConstant holds values in the from &H76 &H32F
type HexConstant struct {
	Token token.Token
	Value string
	Trash []TrashStatement
}

func (hc *HexConstant) expressionNode()      {}
func (hc *HexConstant) TokenLiteral() string { return hc.Token.Literal }
func (hc *HexConstant) HasTrash() bool       { return len(hc.Trash) > 0 }

// String returns literal as a string
func (hc *HexConstant) String() string {
	var out bytes.Buffer

	out.WriteString(hc.Token.Literal)
	out.WriteString(hc.Value)

	if hc.HasTrash() {
		out.WriteString(Trash(hc.Trash))
	}
	return out.String()
}

// TrashStatement holds stuff the parser couldn't make sense out of
type TrashStatement struct {
	Token token.Token
}

func (nse *TrashStatement) statementNode()       {}
func (nse *TrashStatement) TokenLiteral() string { return nse.Token.Literal }

func (nse *TrashStatement) String() string { return nse.Token.Literal }

func Trash(Trashes []TrashStatement) string {
	var out bytes.Buffer

	if len(Trashes) == 0 {
		return ""
	}

	for _, Trash := range Trashes {
		switch Trash.Token.Type {
		case token.COMMA, token.COLON:
			out.WriteString(Trash.String())
		case token.STRING:
			out.WriteString(` "` + Trash.String() + `"`)
		default:
			out.WriteString(` ` + Trash.String())
		}
	}

	return out.String()
}

// TrashExpression when a parseExpression can't make any sense of the input
type TrashExpression struct {
	Token token.Token
	Trash []TrashStatement
}

func (trash *TrashExpression) expressionNode()      {}
func (trash *TrashExpression) TokenLiteral() string { return trash.Token.Literal }
func (trash *TrashExpression) HasTrash() bool       { return len(trash.Trash) > 0 }

func (trash *TrashExpression) String() string { return Trash(trash.Trash) }

// OctalConstant has two forms &37 or &O37
type OctalConstant struct {
	Token token.Token
	Value string
	Trash []TrashStatement
}

func (oc *OctalConstant) expressionNode()      {}
func (oc *OctalConstant) TokenLiteral() string { return oc.Token.Literal }
func (oc *OctalConstant) HasTrash() bool       { return len(oc.Trash) > 0 }

// String gives printable version of me
func (oc *OctalConstant) String() string {
	var out bytes.Buffer

	out.WriteString(oc.Token.Literal)
	out.WriteString(oc.Value)
	out.WriteString(Trash(oc.Trash))

	return out.String()
}

// OpenStatement opens a data file or com port
// comes in two flavors
// OPEN filename [FOR mode][ACCESS access][lock] AS [#]file number [LEN=reclen]
// OPEN mode,[#]file number,filename[,reclen]

type OpenStatement struct {
	Token    token.Token // OPEN
	FileName string      // filename to open
	//	FileNameSep string           // seperator before FileName
	FileNumber FileNumber       // file number associated with file
	FileNumSep string           // seperator before FileNum
	Mode       string           // access mode, read, write append...
	Access     string           //	read, write, or read/write
	Lock       string           // access for other processes, share mode
	Trash      []TrashStatement // Stuff I was unable to parse
	RecLen     string           // record length for fixed len records
	Verbose    bool             // true means the long syntax version of open
}

func (opn *OpenStatement) statementNode()       {}
func (opn *OpenStatement) TokenLiteral() string { return opn.Token.Literal }
func (opn *OpenStatement) HasTrash() bool       { return len(opn.Trash) > 0 }

func (opn *OpenStatement) String() string {
	var out bytes.Buffer

	out.WriteString(opn.Token.Literal)

	if opn.Verbose {
		out.WriteString(` "` + opn.FileName + `"`)

		if len(opn.Mode) > 0 {
			out.WriteString(` FOR ` + opn.Mode)
		}

		if len(opn.Access) > 0 {
			out.WriteString(` ACCESS ` + opn.Access)
		}

		if len(opn.Lock) > 0 {
			out.WriteString(` ` + opn.Lock)
		}

		if len(opn.FileNumber.String()) > 0 {
			out.WriteString(` AS ` + opn.FileNumSep + opn.FileNumber.String())
		}

		if len(opn.RecLen) > 0 {
			out.WriteString(` LEN = ` + opn.RecLen)
		}
	} else { // non verbose form
		if len(opn.Mode) > 0 {
			out.WriteString(` "` + opn.Mode + `"`)
		}

		if len(opn.FileNumSep) > 0 {
			out.WriteString(`, ` + opn.FileNumSep)
		} else {
			if len(opn.FileNumber.String()) > 0 {
				out.WriteString(`, `)
			}
		}

		if len(opn.FileNumber.String()) > 0 {
			out.WriteString(opn.FileNumber.String())
		}

		if len(opn.FileName) > 0 {
			out.WriteString(`, "` + opn.FileName + `"`)
		}

		if len(opn.RecLen) > 0 {
			out.WriteString(`,` + opn.RecLen)
		}
	}

	// if there was any Trash tokens, print them
	out.WriteString(Trash(opn.Trash))
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
	Trash []TrashStatement
}

func (rs *RestoreStatement) statementNode()       {}
func (rs *RestoreStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *RestoreStatement) HasTrash() bool       { return len(rs.Trash) > 0 }

// String sends the original code
func (rs *RestoreStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.Token.Literal)
	if rs.Line > 0 {
		out.WriteString(" ")
		out.WriteString(strconv.Itoa((rs.Line)))
	}

	out.WriteString(Trash(rs.Trash))

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
	Trash []TrashStatement
}

func (il *StringLiteral) expressionNode()      {}
func (il *StringLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *StringLiteral) HasTrash() bool       { return len(il.Trash) > 0 }

// String returns literal as a string
func (il *StringLiteral) String() string {
	var out bytes.Buffer
	out.WriteString(`"`)
	out.WriteString(il.Value)
	out.WriteString(`"`)

	out.WriteString(Trash(il.Trash))
	return out.String()
}

// IndexExpression contains the index into an array
type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
	Trash []TrashStatement
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return strings.ToUpper(ie.Token.Literal) }
func (ie *IndexExpression) HasTrash() bool       { return len(ie.Trash) > 0 }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ie.Index.String())
	out.WriteString(Trash(ie.Trash))
	return out.String()
}

// PrefixExpression the big one here is - as in -5
type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
	Trash    []TrashStatement
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) HasTrash() bool       { return len(pe.Trash) > 0 }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(Trash(pe.Trash))
	return out.String()
}

// InfixExpression things like 5 + 6
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
	Trash    []TrashStatement
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) HasTrash() bool {
	if len(ie.Trash) > 0 {
		return true
	}

	if ie.Left != nil {
		if ie.Left.HasTrash() {
			return true
		}
	}

	if ie.Right != nil {
		if ie.Right.HasTrash() {
			return true
		}
	}

	return false
}

// String the readable version of me
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	if ie.Left != nil {
		out.WriteString(ie.Left.String() + " ")
	}
	if len(ie.Operator) > 0 {
		out.WriteString(ie.Operator + " ")
	}
	if ie.Right != nil {
		out.WriteString(ie.Right.String())
	}

	out.WriteString(Trash(ie.Trash))
	return out.String()
}

// GroupedExpression is enclosed in parentheses
type GroupedExpression struct {
	Token token.Token
	Exp   Expression
	Trash []TrashStatement
}

func (ge *GroupedExpression) expressionNode()      {}
func (ge *GroupedExpression) TokenLiteral() string { return ge.Token.Literal }
func (ge *GroupedExpression) HasTrash() bool       { return len(ge.Trash) > 0 }

// String the readable version of me
func (ge *GroupedExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	if ge.Exp != nil {
		out.WriteString(ge.Exp.String())
	}

	out.WriteString(")")
	out.WriteString(Trash(ge.Trash))

	return out.String()
}

// IfStatement holds an If statement
type IfStatement struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence Statement
	Alternative Statement
}

func (ifs *IfStatement) statementNode()       {}
func (ifs *IfStatement) TokenLiteral() string { return strings.ToUpper(ifs.Token.Literal) }

// String returns my string representation
func (ifs *IfStatement) String() string {
	var out bytes.Buffer

	out.WriteString("IF ")
	if ifs.Condition != nil {
		out.WriteString(ifs.Condition.String())
	}
	out.WriteString(" THEN")
	if ifs.Consequence != nil {
		_, ok := ifs.Consequence.(*EndStatement)
		if ok {
			out.WriteString(" ")
		}
		out.WriteString(ifs.Consequence.String())
	}

	if ifs.Alternative != nil {
		out.WriteString(" ELSE")
		s := ifs.Alternative
		_, ok := s.(*EndStatement)
		if ok {
			out.WriteString(" ")
		}
		out.WriteString(s.String())
	}

	return out.String()
}

// GosubStatement call subroutine
type GosubStatement struct {
	Token token.Token
	Gosub []token.Token
}

func (gsb *GosubStatement) statementNode() {}

// TokenLiteral should return GOSUB
func (gsb *GosubStatement) TokenLiteral() string { return strings.ToUpper(gsb.Token.Literal) }
func (gsb *GosubStatement) String() string {
	var out bytes.Buffer

	out.WriteString(" GOSUB ")
	for _, t := range gsb.Gosub {
		out.WriteString(t.Literal)
	}

	return out.String()
}

// GotoStatement triggers a jump
type GotoStatement struct {
	Token token.Token
	JmpTo []token.Token
}

func (gt *GotoStatement) statementNode()       {}
func (gt *GotoStatement) TokenLiteral() string { return gt.Token.Literal }
func (gt *GotoStatement) String() string {
	var out bytes.Buffer

	out.WriteString(" ")
	if len(gt.TokenLiteral()) > 0 {
		out.WriteString(gt.TokenLiteral() + " ")
	}

	for _, t := range gt.JmpTo {
		out.WriteString(t.Literal)
	}

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
	out.WriteString(end.TokenLiteral())
	return out.String()
}

// ListExpression when LIST is a parameter to a Statement
type ListExpression struct {
	Token token.Token
	Trash []TrashStatement
}

func (lst *ListExpression) expressionNode()      {}
func (lst *ListExpression) TokenLiteral() string { return strings.ToUpper(lst.Token.Literal) }
func (lst *ListExpression) HasTrash() bool       { return len(lst.Trash) > 0 }
func (lst *ListExpression) String() string       { return lst.TokenLiteral() + Trash(lst.Trash) }

// ListStatement command to clear screen
type ListStatement struct {
	Token  token.Token
	Start  string //starting line number
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

// Resume execution after recovering from an error
type ResumeStatement struct {
	Token   token.Token  // "RESUME"
	ResmDir []Expression // 0, NEXT or line ##### from source
	ResmPt  RetPoint     // location that caused the error
}

func (resm *ResumeStatement) statementNode()       {}
func (resm *ResumeStatement) TokenLiteral() string { return strings.ToUpper(resm.Token.Literal) }
func (resm *ResumeStatement) String() string {
	var out bytes.Buffer

	out.WriteString(resm.TokenLiteral())

	for _, rd := range resm.ResmDir {
		out.WriteString(" " + rd.String())
	}

	return out.String()
}

// RunCommand clears all variables and starts execution
// RUN linenum starts execution at linenum
type RunCommand struct {
	Token     token.Token
	StartLine int
	LoadFile  Expression
	KeepOpen  bool
	Trash     []TrashStatement
}

func (run *RunCommand) statementNode()       {}
func (run *RunCommand) TokenLiteral() string { return strings.ToUpper(run.Token.Literal) }
func (run *RunCommand) HasTrash() bool       { return len(run.Trash) > 0 }

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
	Token    token.Token      // stmt token
	Params   []Expression     // Parser creates these
	Settings [4]int           // When executed, eval of expressions go here
	Trash    []TrashStatement // any trash found while parsing
}

func (scrn *ScreenStatement) statementNode()       {}
func (scrn *ScreenStatement) TokenLiteral() string { return strings.ToUpper(scrn.Token.Literal) }
func (scrn *ScreenStatement) HasTrash() bool       { return len(scrn.Trash) > 0 }

// String returns the statement and any parameters as a string
func (scrn *ScreenStatement) String() string {
	var out bytes.Buffer

	out.WriteString("SCREEN ")
	for i, p := range scrn.Params {
		out.WriteString(p.String())

		if i+1 < len(scrn.Params) {
			out.WriteString(",")
		}

		if scrn.HasTrash() {
			out.WriteString(Trash(scrn.Trash))
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

type ToStatement struct {
	Token token.Token
}

func (to *ToStatement) statementNode()       {}
func (to *ToStatement) TokenLiteral() string { return strings.ToUpper(to.Token.Literal) }
func (to *ToStatement) String() string       { return " " + strings.ToUpper(to.Token.Literal) + " " }

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

// Using Expression, part of the PRINT statement
type UsingExpression struct {
	Token  token.Token
	Format Expression
	Sep    string
	Items  []Expression
	Seps   []string
	Trash  []TrashStatement
}

func (us *UsingExpression) expressionNode()      {}
func (us *UsingExpression) TokenLiteral() string { return strings.ToUpper(us.Token.Literal) }
func (us *UsingExpression) HasTrash() bool       { return len(us.Trash) > 0 }
func (us *UsingExpression) String() string {
	var out bytes.Buffer

	out.WriteString(us.TokenLiteral() + " ")

	// if I actually have a format, send it
	if us.Format != nil {
		out.WriteString(us.Format.String())
	}

	if len(us.Sep) != 0 {
		out.WriteString(us.Sep)
	}

	for i, item := range us.Items {
		if item != nil {
			out.WriteString(item.String())
		}
		if len(us.Seps) > i {
			out.WriteString(us.Seps[i])
		}
	}

	// just print what I got, evaluator will decide if it is legal
	if len(us.Seps) > len(us.Items) {
		for i := len(us.Items); i < len(us.Seps); i++ {
			out.WriteString(us.Seps[i])
		}
	}

	out.WriteString(Trash(us.Trash))

	return out.String()
}

// View Statement changes the viewport size for graphics
type ViewStatement struct {
	Token token.Token
	Parms []Node
}

func (vw *ViewStatement) statementNode()       {}
func (vw *ViewStatement) TokenLiteral() string { return strings.ToUpper(vw.Token.Literal) }
func (vw *ViewStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vw.TokenLiteral() + " ")

	for _, pm := range vw.Parms {
		out.WriteString(pm.String())
	}

	return out.String()
}

type ViewPrintStatement struct {
	Token token.Token
	Parms []Node // top, "TO", bottom
}

func (vw *ViewPrintStatement) statementNode()       {}
func (vw *ViewPrintStatement) TokenLiteral() string { return strings.ToUpper(vw.Token.Literal) }
func (vw *ViewPrintStatement) String() string {
	var out bytes.Buffer

	out.WriteString(vw.TokenLiteral() + " ")

	for _, pm := range vw.Parms {
		out.WriteString(pm.String())
	}

	return out.String()
}
