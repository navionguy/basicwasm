package object

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/settings"
)

// GWBasic color values for screen work,https://hwiegman.home.xs4all.nl/gw-man/SCREENS.html
const (
	GWBlack     = iota // 0
	GWBlue             // 1
	GWGreen            // 2
	GWCyan             // 3
	GWRed              // 4
	GWMagenta          // 5
	GWBrown            // 6
	GWWhite            // 7
	GWGray             // 8
	GWLtBlue           // 9
	GWLtGreen          // 10
	GWLtCyan           // 11
	GWLtRed            // 12
	GWLtMagenta        // 13
	GWYellow           // 14
	GWBrtWhite         // 15
)

// XTermjs color directives,https://xtermjs.org/docs/api/vtfeatures/
const (
	XBlack   = iota + 90 //90
	XRed                 // 91
	XGreen               // 92
	XYellow              // 93
	XBlue                // 94
	XMagenta             // 95
	XCyan                // 96
	XWhite               // 97
)

// Console defines how to collect input and display output
type Console interface {
	// Cls clears the screen contents
	Cls()
	// Print outputs the passed string at the curent cursor position
	Print(string)
	// Println prints the string followed by a CR/LF
	Println(string)

	// Locate moves the cursor to the desired (row, col)
	Locate(int, int)
	// Log string to browser debug console
	Log(string)
	// GetCursor, return cursor location(row, col)
	GetCursor() (int, int)
	// Read, return contents of screen range
	Read(col, row, len int) string
	// ReadKeys reads up to (count) keycode values
	ReadKeys(count int) []byte
	// SoundBell emits facsimile of a console beep
	SoundBell()
	// BreakCheck returns true if a ctrl-c was entered
	BreakCheck() bool
}

// HttpClient allows me to mock an http.Client, minimally
type HttpClient interface {
	//Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

// Environment holds my variables and possibly an outer environment
type Environment struct {
	store    map[string]Object   // variables and other program data
	settings map[string]ast.Node // environment settings
	outer    *Environment        // possibly a tempory containing environment
	program  *ast.Program        // current Abstract Syntax Tree
	term     Console             // the terminal console object

	// The following hold "state" information controlled by commands/statements
	autoOn  *ast.AutoCommand // is auto line numbering turned on
	cwd     string           // current working directory
	rnd     *rand.Rand       // random number generator
	rndVal  float32          // most recent generated value
	traceOn bool             // is tracing turned on
	client  HttpClient       // for making server requests
	run     bool             // program is currently execute, if false, a command is executing
	stack   []ast.Code       // return addresses for GOSUB/RETURN
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

// Fetch a runtime setting
func (e *Environment) GetSetting(name string) ast.Node {
	return e.settings[name]
}

// Save a runtime settting
func (e *Environment) SaveSetting(name string, obj ast.Node) {
	e.settings[name] = obj
}

// Push an address, returns stack size
func (e *Environment) Push(cd ast.Code) int {
	e.stack = append(e.stack, cd)
	return len(e.stack)
}

// Pop a return address, nil means stack is empty
func (e *Environment) Pop() *ast.Code {
	l := len(e.stack)
	if l == 0 {
		return nil
	}

	cd := e.stack[l-1]
	e.stack = e.stack[:l-1]

	return &cd
}

// ClearVars empties the map of environment objects
func (e *Environment) ClearVars() {
	e.store = make(map[string]Object)
}

// ClearFiles closes all open files
func (e *Environment) ClearFiles() {
	// ToDo: add support for files
}

// ClearCommon variables
func (e *Environment) ClearCommon() {
	// ToDo: implement run time support for Common variables
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

// Functions below talk to my program object

// Add a statement to the ast
func (e *Environment) AddStatement(stmt ast.Statement) {
	delete(e.settings, settings.Restart) // clear any restart point since the ast is changing

	e.program.AddStatement(stmt)
}

func (e *Environment) StatementIter() *ast.Code {
	return e.program.StatementIter()
}

// Signals that the program has been parsed
func (e *Environment) Parsed() {
	e.program.Parsed()
}

func (e *Environment) AddCmdStmt(stmt ast.Statement) {
	e.program.AddCmdStmt(stmt)
}

func (e *Environment) CmdLineIter() *ast.Code {
	return e.program.CmdLineIter()
}

func (e *Environment) CmdComplete() {
	e.program.CmdComplete()
}

// Command line has been parsed
func (e *Environment) CmdParsed() {
	e.program.CmdParsed()
}

// return the programs constant data
func (e *Environment) ConstData() *ast.ConstData {
	return e.program.ConstData()
}

// NewProgram makes sure the program has been initialized
func (e *Environment) NewProgram() {
	e.program = &ast.Program{}
	e.program.New() // make sure to initialize the new program
}
