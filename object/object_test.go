package object

import (
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/keybuffer"
	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/settings"
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

func Test_Auto(t *testing.T) {
	at := Auto{Next: 100, Step: 10}

	assert.EqualValues(t, "AUTO", at.Inspect(), "auto failed inspection")
	assert.EqualValues(t, 100, at.Next, "auto next failed to set")
	assert.EqualValues(t, 10, at.Step, "auto step failed to set")
	assert.EqualValues(t, AUTO_OBJ, at.Type(), "auto type is wrong")
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
	env.SaveSetting(settings.Restart, &ast.LineNumStmt{})
	env.AddStatement(&ast.StopStatement{})
	//assert.Nil(t, env.cont, "continuation data failed to clear")

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

func Test_Common(t *testing.T) {
	env := newEnvironment()

	assert.Zero(t, len(env.common), "New environment didn't zero common map")

	env.Common("I")

	// did he create the common item
	assert.Equalf(t, 1, len(env.common), "Expected one common item, got %d", len(env.common))
	// should also have created a place holder value
	assert.Equalf(t, 1, len(env.store), "Place holder variable not created")

	// now assign a value to the variable
	env.Set("I", &Integer{Value: 16})

	// Clear the variable space
	env.ClearVars()

	// make him common again, like after a CHAIN
	env.Common("I")

	// he should have recognized it was already there
	assert.Equalf(t, 1, len(env.common), "Second COMMON resulted in %d common items", len(env.common))
	assert.NotNil(t, env.common["I"].value, "Second COMMON lost value")
}

func Test_DefaultKeys(t *testing.T) {
	tests := []struct {
		key string
		val string
	}{
		{key: `F1`, val: `LIST`},
	}

	var mt mocks.MockTerm
	env := NewTermEnvironment(mt)
	obj := env.GetSetting(settings.KeyMacs)

	kys, ok := obj.(*ast.KeySettings)

	assert.Truef(t, ok, "KeyMacs didn't default to a KeySettings object")

	for _, tt := range tests {
		assert.EqualValuesf(t, tt.val, kys.Keys[tt.key], "DefaultKeys expected %s, got %s", tt.val, kys.Keys[tt.key])
	}
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
		{obj: &Error{Message: "Error"}, exp: "Error", tp: "ERROR"},
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
		setev  *Environment
		getev  *Environment
		item   string
		set    Object
		seterr Object
		exp    Object
	}{
		{setev: env, getev: env, item: "A$[]", set: nil, exp: &String{Value: ""}},
		{setev: env, getev: env, item: "A[]", set: nil, exp: &Integer{Value: 0}},
		{setev: env, getev: env, item: "A#", set: nil, exp: &IntDbl{Value: 0}},
		{setev: env, getev: env, item: "A%", set: nil, exp: &Integer{Value: 0}},
		{setev: env, getev: env, item: "A$", set: nil, exp: &String{Value: ""}},
		{setev: env, getev: env, item: "INDEX", set: nil, exp: &Integer{Value: 0}},
		{setev: env, getev: env, item: "A", set: nil, exp: &Integer{Value: 0}},
		{setev: env, getev: env, item: "B", set: &Integer{Value: 5}, exp: &Integer{Value: 5}},
		{setev: env, getev: env, item: "INKEY$", set: &Integer{Value: 5}, exp: &String{Value: ""}, seterr: &Error{Message: "Syntax error", Code: 5}},
		{setev: encenv, getev: env, item: "B", set: &Integer{Value: 6}, exp: &Integer{Value: 5}}, // this test depends on var set in previous test!!!
		{setev: env, getev: encenv, item: "D", set: &Integer{Value: 6}, exp: &Integer{Value: 6}},
	}

	var se Object
	for _, tt := range tests {
		if tt.setev != nil {
			se = tt.setev.Set(tt.item, tt.set)
		}
		obj := tt.getev.Get(tt.item)

		assert.NotNil(t, obj, "Environment.Get(%s) returned nil", tt.item)

		// if he is an array, get the first element
		arr, ok := obj.(*Array)
		if ok {
			obj = arr.Elements[0]
		}

		assert.True(t, strings.EqualFold(obj.Inspect(), tt.exp.Inspect()), "Get of %s differed %s | %s", tt.item, obj.Inspect(), tt.exp.Inspect())

		if tt.seterr != nil {
			assert.True(t, strings.EqualFold(se.Inspect(), tt.seterr.Inspect()), "Test_Environment failed to get %s", tt.seterr.Inspect())
		}
	}
}

func Test_TermEnvironment(t *testing.T) {
	var trm mocks.MockTerm
	env := NewTermEnvironment(trm)

	if env.Terminal() == nil {
		t.Fatalf("Terminal failed to set!")
	}

	if env.GetTrace() || env.ProgramRunning() {
		t.Fatalf("env defaults not false, %t, %t", env.GetTrace(), env.ProgramRunning())
	}

	env.SetTrace(true)
	env.SetRun(true)

	if !env.GetTrace() || !env.ProgramRunning() || (env.GetClient() == nil) {
		t.Fatalf("env defaults not changed, %t, %t, %t", env.GetTrace(), env.ProgramRunning(), (env.GetClient() == nil))
	}
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

func TestReadOnly(t *testing.T) {
	env := newEnvironment()

	assert.True(t, env.ReadOnly("ERL"), "TestReadOnly failed")
	assert.True(t, env.ReadOnly("erl"), "TestReadOnly failed")

}

func Test_Function(t *testing.T) {
	id := ast.Identifier{Value: "X"}
	tkBlk := token.Token{Literal: "FNDBL"}
	tkBep := token.Token{Literal: "= X * 2"}
	stmt := ast.BeepStatement{Token: tkBep}
	fn := &Function{Body: &ast.BlockStatement{Token: tkBlk, Statements: []ast.Statement{&stmt}}, Parameters: []*ast.Identifier{&id}}

	assert.Equalf(t, ObjectType("FUNCTION"), fn.Type(), "Expeced type FUNCTION but got %s", fn.Type())

	assert.Equal(t, "DEF FNDBL(X) BEEP", fn.Inspect(), "Function object didn't create properly.")
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

		env.SaveSetting(settings.Restart, itr)

		if tt.noval {
			itr = nil
			env.program.StatementIter()
		}

		itr2 := env.GetSetting(settings.Restart)

		assert.Equal(t, itr, itr2, "%s got %T, wanted %T", tt.title, itr2, itr)
	}

}

func Test_RestartSignal(t *testing.T) {
	rs := RestartSignal{}

	assert.Equal(t, ObjectType("RESTART"), rs.Type(), "Restart signal invalid type")
	assert.Equal(t, "RESTART", rs.Inspect(), "Restart Inspect() returned %s", rs.Inspect())
}

func Test_Settings(t *testing.T) {
	name := "test"
	env := newEnvironment()

	env.SaveSetting(name, &ast.StringLiteral{Value: name})
	tst := env.GetSetting(name)

	assert.NotNil(t, tst, "setting didn't save")

	env.ClrSetting(name)
	tst = env.GetSetting(name)

	assert.Nil(t, tst, "setting didn't clear")
}

func Test_SettingKeyMac(t *testing.T) {
	tests := []struct {
		fail bool
		sett ast.Node
	}{
		{fail: false, sett: &ast.KeySettings{}},
		{fail: true, sett: &ast.KeyStatement{}},
	}

	for _, tt := range tests {
		env := newEnvironment()

		env.SaveSetting(settings.KeyMacs, tt.sett)
		ks := keybuffer.GetKeyBuffer().KeySettings

		if tt.fail {
			assert.NotEqualValuesf(t, tt.sett, ks, "KeyMacs setting saved to KeyBuffer when it shouldn't have")
		} else {
			assert.EqualValuesf(t, tt.sett, ks, "KeyMacs setting didn't save to KeyBuffer")
		}
	}
}

func Test_Stack(t *testing.T) {
	tests := []struct {
		pushCount int
		popCount  int
		expNil    bool
	}{
		{pushCount: 3, popCount: 4, expNil: true},
		{pushCount: 3, popCount: 3, expNil: false},
	}

	for _, tt := range tests {
		env := newEnvironment()
		for i := 0; i < tt.pushCount; i++ {
			ret := ast.RetPoint{}
			env.Push(ret)
		}

		nilSeen := false
		for i := 0; i < tt.popCount; i++ {
			rc := env.Pop()

			if rc == nil {
				nilSeen = true
				if !tt.expNil {
					t.Fatalf("push(%d), pop(%d) resulted in nil result", tt.pushCount, tt.popCount)
				}
			}
		}

		if nilSeen != tt.expNil {
			assert.Equal(t, tt.expNil, nilSeen)
		}
	}
}

func Test_StdError(t *testing.T) {
	tests := []struct {
		errNum  int    // error I want to get to get back
		expMsg  string // expected message string
		running bool   // should running flag be set in environment
	}{
		{errNum: berrors.NextWithoutFor, expMsg: "NEXT without FOR"},
		{errNum: berrors.Syntax, expMsg: "Syntax error in 10", running: true},
		{errNum: berrors.PathNotFound, expMsg: "Path not found"},
	}

	for _, tt := range tests {
		env := newEnvironment()
		if tt.running {
			ln := &IntDbl{Value: 10}
			env.Set(token.LINENUM, ln)
			env.run = true
		}
		err := StdError(env, tt.errNum)

		assert.Equal(t, tt.errNum, err.Code)
		assert.Equal(t, tt.expMsg, err.Message)
	}
}

func Test_TypedValue(t *testing.T) {
	tv := TypedVar{TypeID: TYPED_OBJ, Value: &Integer{Value: 5}}

	assert.Equal(t, ObjectType(TYPED_OBJ), tv.Type())
	assert.Equal(t, "5", tv.Inspect())
}
