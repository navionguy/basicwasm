package object

import (
	"bytes"
	"math/rand"
	"net/http"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/keybuffer"
	"github.com/navionguy/basicwasm/settings"
	"golang.org/x/text/encoding/charmap"
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
	//	XBlack   = iota + 90 // 90
	XRed     = iota + 91 // 91
	XGreen               // 92
	XYellow              // 93
	XBlue                // 94
	XMagenta             // 95
	XCyan                // 96
	XWhite               // 97
	XBlack   = 30
)

// size of arrays that haven't been DIM'd
const DefaultDimSize = 10

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
	// BreakCheck returns true if a ctrl-c was entere
	BreakCheck() bool
}

// HttpClient allows me to mock an http.Client, minimally
type HttpClient interface {
	//Do(req *http.Request) (*http.Response, error)
	Get(url string) (*http.Response, error)
}

// file access modes
const (
	inputFile = 1
	outputFile
	rndFile
	apndFile
)

// in-memory implementation of data files
type aFile struct {
	mode int          // access mode for file
	data bytes.Buffer // the local storage for the file
}

// Environment holds my variables and possibly an outer environment
type Environment struct {
	ForLoops []ForBlock           // any For Loops that are active
	store    map[string]*variable // variables and other program data
	common   map[string]*variable // variables that live through a CHAIN
	files    map[int16]*aFile     // currently open files
	dir      map[string]*aFile    // locally cached files by full name
	settings map[string]ast.Node  // environment settings
	readOnly map[string]bool      // my read only environment variables
	outer    *Environment         // possibly a tempory containing environment
	program  *ast.Program         // current Abstract Syntax Tree
	term     Console              // the terminal console object

	// The following hold "state" information controlled by commands/statements
	client  HttpClient     // for making server requests
	rnd     *rand.Rand     // random number generator
	rndVal  float32        // most recent generated value
	run     bool           // program is currently executing, if false, a command is executing
	stack   []ast.RetPoint // return addresses for GOSUB/RETURN
	traceOn bool           // is tracing turned on
}

type variable struct {
	value Object // the variable object
}

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
	e.dir = make(map[string]*aFile)
	e.ClearCommon()
	e.CloseAllFiles()
	e.ClearVars()
	if e.program == nil {
		e.program = &ast.Program{}
	}
	e.program.New()
	e.setDefaults()
	e.setReadOnlys()

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

// set defaults for all the settings that have defaults
func (e *Environment) setDefaults() {
	// I always start on driveC
	e.SaveSetting(settings.WorkDrive, &ast.StringLiteral{Value: `C:\`})

	// setup default function key macros
	kys := ast.KeySettings{Disp: true}
	kys.Keys = make(map[string]string)
	kys.Keys["F1"] = "LIST"
	kys.Keys["F2"] = "RUN"
	kys.Keys["F3"] = `LOAD "`
	kys.Keys["F4"] = `SAVE "`
	kys.Keys["F5"] = "CONT\r"
	kys.Keys["F6"] = ", \"LPT1:\" \r"
	kys.Keys["F7"] = "TRON\r"
	kys.Keys["F8"] = "TROFF\r"
	kys.Keys["F9"] = "KEY"
	kys.Keys["F10"] = "SCREEN 0,0,0\r"
	kys.Keys["F11"] = "\x1b[A" // Up Arrow
	kys.Keys["F12"] = "\x1b[D" // Left Arrow
	kys.Keys["F13"] = "\x1b[C" // Right Arrow
	kys.Keys["F14"] = "\x1b[B" // Down Arrow
	e.SaveSetting(settings.KeyMacs, &kys)
}

// define all the variables that are read only
func (e *Environment) setReadOnlys() {
	e.readOnly = make(map[string]bool)

	e.readOnly["CSRLIN"] = true
	e.readOnly["ERDEV"] = true
	e.readOnly["ERDEV$"] = true
	e.readOnly["ERL"] = true
	e.readOnly["ERR"] = true
	e.readOnly["INKEY$"] = true
}

// preserve a variable across a chain
func (e *Environment) Common(name string) {
	// everything stores in upper case
	name = strings.ToUpper(name)

	// is he already common
	cv, exists := e.common[name]

	// does he already have a value
	v, ok := e.store[name]

	// if he is already common and doesn't exist in the store
	// put his variable into the store

	if exists && !ok {
		e.store[name] = cv
		return
	}

	//
	if !ok {
		e.Set(name, e.getDefaultValue(name))
		v = e.store[name]
	}

	// save the variable into common map
	e.common[name] = v
}

// Get attempts to retrieve an object from the environment, nil if not found
func (e *Environment) Get(name string) Object {
	name = strings.ToUpper(name)
	v, ok := e.store[name]

	// if I found him, send the value
	if ok {
		return v.value
	}

	// check for my special case
	if strings.EqualFold(name, "INKEY$") {
		bt, err := keybuffer.GetKeyBuffer().ReadByte()
		if err != nil {
			return &String{Value: ""}
		}
		return &String{Value: string(bt)}
	}

	// am I in an enclosed environment?
	if !ok && e.outer != nil {
		return e.outer.Get(name)
	}

	// no value to return
	return e.getDefaultValue(name)
}

// Variable isn't in memory, create it with correct default value
func (e *Environment) getDefaultValue(name string) Object {
	// check for type indicators
	// if single char name, no type default to integer
	if len(name) == 1 {
		return &Integer{Value: 0}
	}

	// it *may* have a type
	switch e.getType(name) {
	case '$': // string
		return &String{Value: ""}
	case '%', '!': // single precesion
		return &Integer{Value: 0}
	case '#': // double precision
		return &IntDbl{Value: 0}
	case ']': // array of something
		parts := strings.Split(name, "[")
		return e.buildDefaultArray(parts[0])
	}

	// the default case
	return &Integer{Value: 0}
}

// Build an array of default values
// Warning, here lies hidden recursion!
// Strap on your miners hat and prepare to descend
func (e *Environment) buildDefaultArray(name string) Object {
	def := Array{TypeID: "[]"}

	for i := 0; i < DefaultDimSize; i++ {
		def.Elements = append(def.Elements, e.getDefaultValue(name))
	}

	return &def
}

// determine type for variable
// TODO: implement defined type name ranges eg. DEFDBL
func (e *Environment) getType(name string) byte {

	if len(name) > 1 {
		t := name[len(name)-1]
		switch t {
		case '%', '!', '$', '#', ']': // single precesion
			return t
		}
	}

	return 0x00
}

// Set stores an object in the environment
func (e *Environment) Set(name string, val Object) Object {
	// don't store a nil
	if val == nil {
		return StdError(e, berrors.Syntax)
	}
	// I always store in upper case
	name = strings.ToUpper(name)

	// check for the read only variables
	if e.readOnly[name] {
		return StdError(e, berrors.Syntax)
	}

	// is he already saved?
	t, ok := e.store[name]

	if ok {
		t.value = val
		return nil
	}

	// create and store a variable to hold the value
	v := &variable{value: val}
	e.store[name] = v

	return nil
}

// clear a setting
func (e *Environment) ClrSetting(name string) {
	e.settings[name] = nil
}

// Fetch a runtime setting
func (e *Environment) GetSetting(name string) ast.Node {
	return e.settings[name]
}

// Save a runtime settting
func (e *Environment) SaveSetting(name string, obj ast.Node) {
	e.settings[name] = obj

	// check for a special setting

	if strings.EqualFold(name, settings.KeyMacs) {
		ks, ok := obj.(*ast.KeySettings)

		if !ok {
			return
		}
		keybuffer.GetKeyBuffer().KeySettings = ks
	}
}

// Push an address, returns stack size
func (e *Environment) Push(ret ast.RetPoint) int {
	e.stack = append(e.stack, ret)
	return len(e.stack)
}

// Pop a return address, nil means stack is empty
func (e *Environment) Pop() *ast.RetPoint {
	l := len(e.stack)
	if l == 0 {
		return nil
	}

	ret := e.stack[l-1]
	e.stack = e.stack[:l-1]

	return &ret
}

// ClearVars empties the map of environment objects
func (e *Environment) ClearVars() {
	e.store = make(map[string]*variable)
}

// CloseAllFiles closes all open files
func (e *Environment) CloseAllFiles() {
	e.files = make(map[int16]*aFile)
}

// CloseFile closes a file based on its handle
func (e *Environment) CloseFile(f int16) bool {
	if e.files[f] == nil {
		return false
	}
	e.files[f] = nil

	return true
}

// ClearCommon variables
func (e *Environment) ClearCommon() {
	e.common = make(map[string]*variable)
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

// check if a variable name is defined read only
func (e *Environment) ReadOnly(v string) bool {
	return e.readOnly[strings.ToUpper(v)]
}

// convert the CP437 values to a strings
func DecodeBytes(bts []byte) string {
	var r []rune

	for _, b := range bts {
		r = append(r, charmap.CodePage437.DecodeByte(b))
	}

	return string(r)
}
