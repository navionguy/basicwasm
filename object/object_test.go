package object

import (
	"fmt"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

type mockTerm struct {
	row     int
	col     int
	strVal  string
	sawBeep *bool
}

func (mt mockTerm) Cls() {

}

func (mt mockTerm) Print(msg string) {
	fmt.Print(msg)
}

func (mt mockTerm) Println(msg string) {
	fmt.Println(msg)
}

func (mt mockTerm) Locate(int, int) {

}

func (mt mockTerm) GetCursor() (int, int) {
	return mt.row, mt.col
}

func (mt mockTerm) Read(col, row, len int) string {
	return mt.strVal
}

func (mt mockTerm) ReadKeys(count int) []byte {
	return nil
}

func (mt mockTerm) SoundBell() {
	fmt.Print("\x07")
	*mt.sawBeep = true
}

func TestBStr(t *testing.T) {
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

func TestInteger(t *testing.T) {
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

func TestEnvironment(t *testing.T) {
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

func TestTermEnvironment(t *testing.T) {
	var trm mockTerm
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

func TestFunction(t *testing.T) {
	tkBlk := token.Token{Literal: "{"}
	tkBep := token.Token{Literal: "BEEP"}
	stmt := ast.BeepStatement{Token: tkBep}
	fn := &Function{Body: &ast.BlockStatement{Token: tkBlk, Statements: []ast.Statement{&stmt}}}

	if fn.Type() != FUNCTION_OBJ {
		t.Fatalf("Function gave incorrect type %v", fn.Type())
	}
}

func TestReturnValue(t *testing.T) {
	fn := &ReturnValue{Value: &Integer{Value: 404}}

	if fn.Type() != RETURN_VALUE_OBJ {
		t.Fatalf("ReturnValue gave incorrect type %v", fn.Type())
	}

	assert.EqualValues(t, fn.Inspect(), "404", "Return value didn't set correctly.")
}
