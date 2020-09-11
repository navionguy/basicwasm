package evaluator

import (
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/object"
)

func compareObjects(inp string, evald object.Object, want interface{}, t *testing.T) {
	switch exp := want.(type) {
	case int:
		testIntegerObject(t, evald, int16(exp))
	case *object.Integer:
		testIntegerObject(t, evald, int16(exp.Value))
	case *object.Fixed:
		fx, ok := evald.(*object.Fixed)

		if !ok {
			t.Fatalf("object is not Fixed from %s, got %T", inp, evald)
		}

		if fx.Value.Cmp(exp.Value) != 0 {
			t.Fatalf("at %s, expected %s, got %s", inp, exp.Value.String(), fx.Value.String())
		}
	case *object.FloatSgl:
		flt, ok := evald.(*object.FloatSgl)

		if !ok {
			t.Fatalf("object is not FloatSgl from %s, got %T", inp, evald)
		}

		if flt.Value != exp.Value {
			t.Fatalf("%s got %f, expected %f", inp, flt.Value, exp.Value)
		}
	case *object.FloatDbl:
		flt, ok := evald.(*object.FloatDbl)

		if !ok {
			t.Fatalf("object is not FloatDbl from %s, got %T", inp, evald)
		}

		if flt.Value != exp.Value {
			t.Fatalf("%s got %f, expected %f", inp, flt.Value, exp.Value)
		}
	case *object.String:
		def, ok := evald.(*object.String)

		if !ok {
			t.Fatalf("object is not String from %s, got %T", inp, evald)
		}

		if strings.Compare(def.Value, exp.Value) != 0 {
			t.Fatalf("%s got %s, expected %s", inp, def.Value, exp.Value)
		}
	case *object.Error:
		err, ok := evald.(*object.Error)

		if !ok {
			t.Fatalf("object is not an error from %s, got %T", inp, evald)
		}

		if strings.Compare(err.Message, exp.Message) != 0 {
			t.Fatalf("%s got %s, expected %s", inp, err.Message, exp.Message)
		}
	default:
		t.Fatalf("compareObjects got unsupported type %T", exp)
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 ABS(1)`, &object.Integer{Value: 1}},
		{`20 ABS(-20)`, &object.Integer{Value: 20}},
		{`30 ABS(30.14)`, &object.Fixed{Value: decimal.New(3014, -2)}},
		{`40 ABS(-40.25)`, &object.Fixed{Value: decimal.New(4025, -2)}},
		{`50 ABS(5.05E+4)`, &object.FloatSgl{Value: float32(50500)}},
		{`60 ABS(-6.05E+4)`, &object.FloatSgl{Value: float32(60500)}},
		{`70 ABS(7.05D+4)`, &object.FloatDbl{Value: float64(70500)}},
		{`80 ABS(-8.05D+4)`, &object.FloatDbl{Value: float64(80500)}},
		{`90 ABS( "Foo" )`, &object.String{Value: "Foo"}},
		{`100 ABS( "Foo", "Bar" )`, &object.Error{Message: "Syntax error in 100"}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestAsc(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 ASC("Alpha")`, 65},
		{`20 ASC("")`, &object.Error{Message: "Illegal Function Call in 20"}},
		{`30 ASC("FRED", "Joe")`, &object.Error{Message: "Syntax error in 30"}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestAtn(t *testing.T) {

}

func TestLen(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 LEN("")`, 0},
		{`20 LEN("four")`, 4},
		{`30 LEN("hello world")`, 11},
		{`40 LEN(1)`, &object.Error{Message: "Type mismatch in 40"}},
		{`50 LEN("one", "two")`, &object.Error{Message: "Syntax error in 50"}},
		{`70 LEN("four" / "five")`, &object.Error{Message: "unknown operator: STRING / STRING in 70"}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}
