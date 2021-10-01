package object

import (
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

func Test_Array(t *testing.T) {
	arr := Array{}

	tp := arr.Type()
	assert.Equal(t, ObjectType("ARRAY"), tp)

	arr.Elements = append(arr.Elements, &String{Value: "First"})
	arr.Elements = append(arr.Elements, &String{Value: "Last"})
	tst := arr.Inspect()
	assert.Equal(t, "First, Last", tst)
}

func Test_BStr(t *testing.T) {
	tests := []struct {
		inp []byte
		out string
	}{
		{[]byte{0x41, 0x41}, "AA"},
		{[]byte{0x00, 0x41}, " A"},
		{[]byte{0x0d, 0x0e}, "  "},
	}

	for _, tt := range tests {
		bs := &BStr{Value: tt.inp}

		if strings.Compare(tt.out, bs.Inspect()) != 0 {
			t.Fatalf("expected %s, got %s", tt.out, bs.Inspect())
		}

		if bs.Type() != BSTR_OBJ {
			t.Fatalf("BSTR type not correct %v", bs.Type())
		}
	}
}

func Test_ClearCommon(t *testing.T) {
	env := newEnvironment()

	env.ClearCommon()
}

func Test_ClearFiles(t *testing.T) {
	env := newEnvironment()

	env.ClearFiles()
}

func Test_ClearVars(t *testing.T) {
	env := newEnvironment()

	env.ClearVars()
}

// test interface into the Code object
func Test_CodeInterface(t *testing.T) {
	env := newEnvironment()
	env.NewProgram()

	assert.NotNil(t, env.program, "Program failed to create")

	// first test statements

	env.AddStatement(&ast.LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "10"}, Value: 10})
	env.cont = &ast.Code{}
	env.AddStatement(&ast.StopStatement{})
	assert.Nil(t, env.cont, "continuation data failed to clear")

	itr := env.StatementIter()
	assert.NotNil(t, itr, "no statement iterator")
	l := itr.Len()
	assert.Equal(t, 2, l, "didn't find two statements")
	env.Parsed()

	// now test commands

	env.AddCmdStmt(&ast.RunCommand{})

	itr = env.CmdLineIter()
	assert.NotNil(t, itr, "no command line iterator")
	l = itr.Len()
	assert.Equal(t, 1, l, "didn't find my command")
	env.CmdParsed()
	env.CmdComplete()

	// check for constant data
	cd := env.ConstData()
	assert.NotNil(t, cd)
}

func Test_Integer(t *testing.T) {
	fv, _ := decimal.NewFromString("14.25")

	tests := []struct {
		obj Object
		exp string
		tp  ObjectType
	}{
		{obj: &Integer{Value: 5}, exp: "5", tp: "INTEGER"},
		{obj: &Fixed{Value: fv}, exp: "14.25", tp: "FIXED"},
		{obj: &String{Value: "Hello"}, exp: "Hello", tp: "STRING"},
		{obj: &Error{Message: "Error"}, exp: "ERROR: Error", tp: "ERROR"},
		{obj: &Builtin{}, exp: "builtin function", tp: "BUILTIN"},
		{obj: &Null{}, exp: "null", tp: "NULL"},
		{obj: &IntDbl{Value: 65999}, exp: "65999", tp: "INTDBL"},
		{obj: &FloatSgl{Value: 3.14159}, exp: "3.141590E+00", tp: "FLOATSGL"},
		{obj: &FloatDbl{Value: 3.14159}, exp: "3.141590E+00", tp: "FLOATDBL"},
	}

	for _, tt := range tests {
		if tt.obj.Inspect() != tt.exp {
			t.Errorf("Inspection found %s, expected %s", tt.obj.Inspect(), tt.exp)
		}

		if tt.obj.Type() != tt.tp {
			t.Errorf("integer object returned %s, expecting %s", tt.obj.Type(), tt.tp)
		}
	}
}

func Test_Environment(t *testing.T) {
	env := newEnvironment()
	encenv := NewEnclosedEnvironment(env)

	tests := []struct {
		setev *Environment
		getev *Environment
		item  string
		set   Object
		exp   Object
	}{
		{setev: nil, getev: env, item: "A", set: nil, exp: nil},
		{setev: env, getev: env, item: "B", set: &Integer{Value: 5}, exp: &Integer{Value: 5}},
		{setev: encenv, getev: env, item: "B", set: &Integer{Value: 6}, exp: &Integer{Value: 5}},
		{setev: env, getev: encenv, item: "D", set: &Integer{Value: 6}, exp: &Integer{Value: 6}},
	}

	for _, tt := range tests {
		if tt.setev != nil {
			tt.setev.Set(tt.item, tt.set)
		}
		testIntEnvGet(t, *tt.getev, tt.item, tt.exp)
	}
}

func Test_TermEnvironment(t *testing.T) {
	var trm mocks.MockTerm
	env := NewTermEnvironment(trm)

	if env.Terminal() == nil {
		t.Fatalf("Terminal failed to set!")
	}

	if env.GetTrace() || (env.GetAuto() != nil) || env.ProgramRunning() {
		t.Fatalf("env defaults not false, %t, %t, %t", env.GetTrace(), (env.autoOn != nil), env.ProgramRunning())
	}

	env.SetTrace(true)
	env.SetAuto(&ast.AutoCommand{})
	env.SetRun(true)

	if !env.GetTrace() || (env.GetAuto() == nil) || !env.ProgramRunning() || (env.GetClient() == nil) {
		t.Fatalf("env defaults not changed, %t, %t, %t, %t", env.GetTrace(), (env.GetAuto() == nil), env.ProgramRunning(), (env.GetClient() == nil))
	}
}

func testIntEnvGet(t *testing.T, env Environment, item string, exp Object) bool {
	obj, ok := env.Get(item)

	if !ok && exp == nil {
		// got nothing, and I wasn't suppose too
		return true
	}
	if !ok && exp != nil {
		t.Errorf("testIntEnvGet got nothing, but I should have %s", exp.Inspect())
		return false
	}
	if ok && exp == nil {
		t.Errorf("testIntEnvGet got something, wasn't expecting anything")
		return false
	}

	_, ok = obj.(*Integer)

	if !ok {
		// I didn't get an Integer object
		return false
	}

	if obj.Inspect() != exp.Inspect() {
		return false
	}

	return true
}

func TestRandom(t *testing.T) {
	tests := []struct {
		inp    int
		exp    float32
		rndMze int64
	}{
		{0, 0.61560816, 0},
		{0, 0.61560816, 0},
		{1, 0.123114005, 0},
		{0, 0.604660273, 1},
		{1, 0.940509081, 0},
	}

	env := newEnvironment()

	for _, tt := range tests {
		if tt.rndMze != 0 {
			env.Randomize(tt.rndMze)
		}
		rc := env.Random(tt.inp)

		if rc.Value != tt.exp {
			t.Fatalf("Random returned %.9f, expected %.9f!  That's too random!!!", rc.Value, tt.exp)
		}
	}
}

func Test_Function(t *testing.T) {
	tkBlk := token.Token{Literal: "{"}
	tkBep := token.Token{Literal: "BEEP"}
	stmt := ast.BeepStatement{Token: tkBep}
	fn := &Function{Body: &ast.BlockStatement{Token: tkBlk, Statements: []ast.Statement{&stmt}}}

	if fn.Type() != FUNCTION_OBJ {
		t.Fatalf("Function gave incorrect type %v", fn.Type())
	}
}

func Test_HaltSingal(t *testing.T) {
	hs := HaltSignal{}

	assert.Equal(t, ObjectType("HALT"), hs.Type(), "HaltSignal, incorrect type")

	assert.Equal(t, "HALT", hs.Inspect(), "HaltSignal, Inspect incorrect value")
}

func Test_Restart(t *testing.T) {
	tests := []struct {
		title string
		stmt  ast.Statement
		noval bool
	}{
		{title: "STOP", stmt: &ast.StopStatement{}},
		{title: "END", stmt: &ast.EndStatement{}},
	}

	for _, tt := range tests {
		mt := mocks.MockTerm{}
		env := NewTermEnvironment(mt)
		env.program.AddStatement(&ast.LineNumStmt{Token: token.Token{Type: token.LINENUM, Literal: "10"}, Value: 10})
		env.program.AddStatement(tt.stmt)
		itr := env.program.StatementIter()
		itr.Next()

		env.SaveRestart(itr)

		if tt.noval {
			itr = nil
			env.program.StatementIter()
		}

		itr2 := env.GetRestart()

		assert.Equal(t, itr, itr2, "%s got %T, wanted %T", tt.title, itr2, itr)
	}

}

func Test_TypedValue(t *testing.T) {
	tv := TypedVar{TypeID: TYPED_OBJ, Value: &Integer{Value: 5}}

	assert.Equal(t, ObjectType(TYPED_OBJ), tv.Type())
	assert.Equal(t, "5", tv.Inspect())
}
