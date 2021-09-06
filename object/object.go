// Package object how the interpretor holds objects during execution
package object

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"

	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/decimal"
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
	SERVER_URL = "$$mom"
	WORK_DRIVE = "$$path"
)

const (
	ERROR_OBJ        = "ERROR"
	INTEGER_OBJ      = "INTEGER"
	INTEGER_DBL      = "INTDBL"
	FIXED_OBJ        = "FIXED"
	FLOATSGL_OBJ     = "FLOATSGL"
	FLOATDBL_OBJ     = "FLOATDBL"
	STRING_OBJ       = "STRING"
	BSTR_OBJ         = "BSTR"
	NULL_OBJ         = "NULL"
	BUILTIN_OBJ      = "BUILTIN"
	FUNCTION_OBJ     = "FUNCTION"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ARRAY_OBJ        = "ARRAY"
	TYPED_OBJ        = "TYPED"
)

// Console defines how to collect input and display output
type Console interface {
	Cls()
	Print(string)
	Println(string)

	Locate(int, int)
	GetCursor() (int, int)
	Read(col, row, len int) string
	ReadKeys(count int) []byte
	SoundBell()
}

// HttpClient allows me to mock an http.Client, minimally
type HttpClient interface {
	//Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

// NewEnclosedEnvironment allows variables during function calls
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := newEnvironment()
	env.outer = outer
	env.term = outer.term
	return env
}

// NewEnvironment creates a place to store variables
func newEnvironment() *Environment {
	s := make(map[string]Object)
	e := &Environment{store: s}
	if e.Program == nil {
		e.Program = &ast.Program{}
	}
	e.Program.New()

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

// Environment holds my variables and possibly an outer environment
type Environment struct {
	store   map[string]Object // variables and other program data
	outer   *Environment      // possibly a tempory containing environment
	Program *ast.Program      // current Abstract Syntax Tree
	term    Console           // the terminal console object

	// The following hold "state" information controlled by commands/statements
	autoOn  *ast.AutoCommand // is auto line numbering turned on
	cwd     string           // current working directory
	rnd     *rand.Rand       // random number generator
	rndVal  float32          // most recent generated value
	traceOn bool             // is tracing turned on
	client  HttpClient       // for making server requests
	run     bool             // program is currently execute, if false, a command is executing
}

// Get attempts to retrieve an object from the environment
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[strings.ToUpper(name)]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set stores an object in the environment
func (e *Environment) Set(name string, val Object) Object {
	e.store[strings.ToUpper(name)] = val
	return val
}

// Terminal allows access to the termianl console
func (e *Environment) Terminal() Console {
	return e.term
}

// SetTrace turns it on or off
func (e *Environment) SetTrace(on bool) {
	e.traceOn = on
}

// GetTrace returns true if we are tracing
func (e *Environment) GetTrace() bool {
	return e.traceOn
}

// SetAuto saves the line numbering parameters
func (e *Environment) SetAuto(auto *ast.AutoCommand) {
	e.autoOn = auto
}

// GetAuto returns the line numbering parameters
func (e *Environment) GetAuto() *ast.AutoCommand {
	return e.autoOn
}

// GetClient returns my http client
func (e *Environment) GetClient() HttpClient {
	return e.client
}

// SetClient setter for the client element
// mostly used for testing
func (e *Environment) SetClient(cl HttpClient) {
	e.client = cl
}

// SetRun controls the "a program is running"
func (e *Environment) SetRun(run bool) {
	e.run = run
}

// Quick test to see if program is currently running
func (e *Environment) ProgramRunning() bool {
	return e.run
}

// Random returns a random number between 0 and 1
// if x is greater than zero, a new random number is generated
// otherwise, the current rndVal is returned
func (e *Environment) Random(x int) *FloatSgl {
	if x > 0 {
		e.rndVal = e.rnd.Float32()
	}
	return &FloatSgl{Value: e.rndVal}
}

// Randomize takes in a new seed and starts a new random series
func (e *Environment) Randomize(seed int64) {
	e.rnd = rand.New(rand.NewSource(seed))
	e.rndVal = e.rnd.Float32()
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

type TypedVar struct {
	Value  Object
	TypeID string
}

func (tv *TypedVar) Type() ObjectType { return TYPED_OBJ }
func (tv *TypedVar) Inspect() string {
	return tv.Value.Inspect()
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
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return "ERROR: " + e.Message }

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
	out.WriteString("foo")

	return out.String()
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
