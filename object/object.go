// Package object how the interpretor holds objects during execution
package object

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"

	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/settings"
	"github.com/navionguy/basicwasm/token"
)

// BuiltinFunction is a function defined by gwbasic
type BuiltinFunction func(env *Environment, fn *Builtin, args ...Object) Object

// ObjectType can always be displayed as a string
type ObjectType string
type Object interface {
	Type() ObjectType
	Inspect() string
}

// some internal environment variables
const (
	SERVER_URL = "mom"
	WORK_DRIVE = "path"
)

const (
	ARRAY_OBJ      = "ARRAY"
	AUTO_OBJ       = "AUTO"
	BSTR_OBJ       = "BSTR"
	BUILTIN_OBJ    = "BUILTIN"
	FIXED_OBJ      = "FIXED"
	FLOATSGL_OBJ   = "FLOATSGL"
	FLOATDBL_OBJ   = "FLOATDBL"
	FUNCTION_OBJ   = "FUNCTION"
	ERROR_OBJ      = "ERROR"
	HALT_SIGNAL    = "HALT"
	INTEGER_OBJ    = "INTEGER"
	INTEGER_DBL    = "INTDBL"
	NULL_OBJ       = "NULL"
	RESTART_SIGNAL = "RESTART"
	STRING_OBJ     = "STRING"
	TYPED_OBJ      = "TYPED"
)

// NewEnclosedEnvironment allows variables during function calls
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := newEnvironment()
	env.outer = outer
	env.term = outer.term
	return env
}

// NewEnvironment creates a place to store variables and settings
func newEnvironment() *Environment {
	e := &Environment{settings: make(map[string]ast.Node)}
	e.ClearCommon()
	e.ClearFiles()
	e.ClearVars()
	if e.program == nil {
		e.program = &ast.Program{}
	}
	e.program.New()

	// initialize my random number generator
	e.rnd = rand.New(rand.NewSource(37))
	e.rndVal = e.rnd.Float32()
	dc := http.DefaultClient
	e.SetClient(dc)
	return e
}

// NewTermEnvironment creates an environment with a terminal front-end
func NewTermEnvironment(term Console) *Environment {
	env := newEnvironment()
	env.term = term
	return env
}

type Array struct {
	Elements []Object
	TypeID   string
}

func (ao *Array) Type() ObjectType { return ARRAY_OBJ }
func (ao *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range ao.Elements {
		if e != nil {
			elements = append(elements, e.Inspect())
		}
	}
	out.WriteString(strings.Join(elements, ", "))
	return out.String()
}

// Integer values
type Integer struct {
	Value int16
}

// Type returns my type
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// Inspect returns value as a string
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// IntDbl values
type IntDbl struct {
	Value int32 // 32bit value
}

// Type returns my type
func (id *IntDbl) Type() ObjectType { return INTEGER_DBL }

// Inspect returns value as a string
func (id *IntDbl) Inspect() string { return fmt.Sprintf("%d", id.Value) }

// Single precision floats
type FloatSgl struct {
	Value float32 // value of the float
}

func (fs *FloatSgl) Type() ObjectType { return FLOATSGL_OBJ }
func (fs *FloatSgl) Inspect() string {
	return fmt.Sprintf("%E", fs.Value)
}

// Double precision floats
type FloatDbl struct {
	Value float64
}

func (fd *FloatDbl) Type() ObjectType { return FLOATDBL_OBJ }
func (fd *FloatDbl) Inspect() string  { return fmt.Sprintf("%E", fd.Value) }

// Fixed decimal point value
type Fixed struct {
	Value decimal.Decimal
}

func (f *Fixed) Type() ObjectType { return FIXED_OBJ }
func (f *Fixed) Inspect() string  { return f.Value.String() }

type Error struct {
	Message string // text error message
	Code    int    // basic internal error code
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return e.Message }

// StdError returns both the GWBASIC message, and error code for a basic error
func StdError(env *Environment, err int) *Error {
	// get the base error message
	e := Error{Code: err, Message: berrors.TextForError(err)}
	env.SaveSetting(settings.ERR, &ast.DblIntegerLiteral{Value: int32(err)})

	erl := 0
	if env.ProgramRunning() {
		tk := env.Get(token.LINENUM)

		if tk != nil {
			erl = int(tk.(*IntDbl).Value)
			e.Message += fmt.Sprintf(" in %d", erl)

			// and save the variable
			env.SaveSetting(settings.ERL, &ast.DblIntegerLiteral{Value: int32(erl)})
		}
	}

	// create or update the ERL & ERR variables
	env.Set(strings.ToUpper(settings.ERL), &Integer{Value: int16(erl)})
	env.Set(strings.ToUpper(settings.ERR), &Integer{Value: int16(err)})

	return &e
}

type ForBlock struct {
	Code ast.RetPoint     // the location in the AST of the FOR statement
	Four *ast.ForStatment // the actual statment
}

// String values
type String struct {
	Value string
}

// Type returns my type
func (i *String) Type() ObjectType { return STRING_OBJ }

// Inspect returns value as a string
func (i *String) Inspect() string { return i.Value }

// BStr is a byte backed string
// not COMmonly used
// parser won't generate one
// they only occur at run-time
type BStr struct {
	Value []byte
}

// Type returns my type BSTR_OBJ
func (bs *BStr) Type() ObjectType { return BSTR_OBJ }

// Inspect returns a displayable string
func (bs *BStr) Inspect() string {
	var out bytes.Buffer
	for _, bt := range bs.Value {
		if bt < 0x20 {
			out.WriteRune(' ')
		} else {
			out.WriteByte(bt)
		}
	}
	return out.String()
}

// When auto is on, this holds the current state
type Auto struct {
	Next int // the next line number to use
	Step int // the step to add to next each time
}

func (a *Auto) Type() ObjectType { return AUTO_OBJ }
func (a *Auto) Inspect() string  { return "AUTO" }

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("DEF ")
	out.WriteString(f.Body.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(f.Body.String())

	return out.String()
}

// HaltSignal tells the eval loop to stop executing
type HaltSignal struct {
	Msg string
}

func (hs *HaltSignal) Type() ObjectType { return HALT_SIGNAL }
func (hs *HaltSignal) Inspect() string  { return "HALT" }

// RestartSignal restarts the eval loop after a chain
type RestartSignal struct{}

func (rs *RestartSignal) Type() ObjectType { return RESTART_SIGNAL }
func (rs *RestartSignal) Inspect() string  { return "RESTART" }

type TypedVar struct {
	Value  Object
	TypeID string
}

func (tv *TypedVar) Type() ObjectType { return ObjectType(tv.TypeID) }
func (tv *TypedVar) Inspect() string {
	return tv.Value.Inspect()
}
