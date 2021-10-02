package object

import (
	"math/rand"
	"net/http"
	"strings"

	"github.com/navionguy/basicwasm/ast"
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

	BreakCheck() bool
}

// HttpClient allows me to mock an http.Client, minimally
type HttpClient interface {
	//Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

// Environment holds my variables and possibly an outer environment
type Environment struct {
	store   map[string]Object // variables and other program data
	outer   *Environment      // possibly a tempory containing environment
	program *ast.Program      // current Abstract Syntax Tree
	term    Console           // the terminal console object

	// The following hold "state" information controlled by commands/statements
	autoOn  *ast.AutoCommand // is auto line numbering turned on
	cwd     string           // current working directory
	rnd     *rand.Rand       // random number generator
	rndVal  float32          // most recent generated value
	traceOn bool             // is tracing turned on
	client  HttpClient       // for making server requests
	run     bool             // program is currently execute, if false, a command is executing
	cont    *ast.Code        // save program iterator due to STOP, END or ctrl-C to allow continue command CONT
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

// save a restart point
func (e *Environment) SaveRestart(cd *ast.Code) {
	rcd := *cd    // make a copy of code struct
	e.cont = &rcd // save it for later
}

// get a restart point, nil if there is none
func (e *Environment) GetRestart() *ast.Code {
	cd := e.cont
	e.cont = nil
	return cd
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
	if e.cont != nil {
		e.cont = nil
	}

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
