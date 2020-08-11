package object

import (
	"testing"

	"github.com/navionguy/basicwasm/decimal"
)

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
