package builtins

import (
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

func compareObjects(inp string, evald object.Object, want interface{}, t *testing.T) {
	if evald == nil {
		t.Fatalf("(%sd) got nil return value!", inp)
	}

	// if I got back a typed variable, I really care about what's inside

	inner, ok := evald.(*object.TypedVar)

	if ok {
		evald = inner.Value
	}

	switch exp := want.(type) {
	case int:
		testIntegerObject(t, evald, int16(exp))
	case *object.Integer:
		testIntegerObject(t, evald, int16(exp.Value))
	case *object.IntDbl:
		id, ok := evald.(*object.IntDbl)

		if !ok {
			t.Fatalf("object is not IntegerDbl from %s, got %T", inp, evald)
		}

		if id.Value != exp.Value {
			t.Fatalf("at %s, expected %d, got %d", inp, exp.Value, id.Value)
		}
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
			t.Fatalf("%s got %.9f, expected %.9f", inp, flt.Value, exp.Value)
		}
	case *object.FloatDbl:
		flt, ok := evald.(*object.FloatDbl)

		if !ok {
			t.Fatalf("object is not FloatDbl from %s, got %T", inp, evald)
		}

		if flt.Value != exp.Value {
			t.Fatalf("%s got %.16f, expected %.16f", inp, flt.Value, exp.Value)
		}
	case *object.String:
		def, ok := evald.(*object.String)

		if !ok {
			t.Fatalf("object is not String from %s, got %T", inp, evald)
		}

		if strings.Compare(def.Value, exp.Value) != 0 {
			t.Fatalf("%s got %s, expected %s", inp, def.Value, exp.Value)
		}
	case *object.BStr:
		bs, ok := evald.(*object.BStr)

		if !ok {
			t.Fatalf("object is not a BStr from %s, got %T", inp, evald)
		}

		if len(bs.Value) != len(exp.Value) {
			t.Fatalf("expected length %d, got length %d", len(exp.Value), len(bs.Value))
		}

		for i := range exp.Value {
			if exp.Value[i] != bs.Value[i] {
				t.Fatalf("difference in byte %d, expected %x, got %x", i, int(exp.Value[i]), int(bs.Value[i]))
			}
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

func testIntegerObject(t *testing.T, obj object.Object, expected int16) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

type test struct {
	cmd  string
	lnum int
	inp  []object.Object
	exp  interface{}
	scrn string
}

func runTests(t *testing.T, bltin string, tests []test) {

	for _, tt := range tests {
		fn, ok := Builtins[bltin]

		assert.True(t, ok, "Failed to find %s() function", bltin)

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		if len(tt.scrn) > 0 {
			mt.StrVal = &tt.scrn
		}
		env := object.NewTermEnvironment(mt)

		if tt.lnum != 0 {
			env.Set(token.LINENUM, &object.IntDbl{Value: int32(tt.lnum)})
		}
		res := fn.Fn(env, fn, tt.inp...)

		compareObjects(tt.cmd, res, tt.exp, t)
	}
}

func TestAbs(t *testing.T) {
	tests := []test{
		{cmd: "10 ABS(1)", lnum: 10, inp: []object.Object{&object.Integer{Value: 1}}, exp: &object.Integer{Value: 1}},
		{cmd: "20 ABS(-20)", inp: []object.Object{&object.Integer{Value: -20}}, exp: &object.Integer{Value: 20}},
		{cmd: "30 ABS(30.14)", inp: []object.Object{&object.Fixed{Value: decimal.New(3014, -2)}}, exp: &object.Fixed{Value: decimal.New(3014, -2)}},
		{cmd: `40 ABS(-40.25)`, inp: []object.Object{&object.Fixed{Value: decimal.New(4025, -2)}}, exp: &object.Fixed{Value: decimal.New(4025, -2)}},
		{cmd: `50 ABS(5.05E+4)`, inp: []object.Object{&object.FloatSgl{Value: float32(50500)}}, exp: &object.FloatSgl{Value: float32(50500)}},
		{cmd: `60 ABS(-6.05E+4)`, inp: []object.Object{&object.FloatSgl{Value: float32(-60500)}}, exp: &object.FloatSgl{Value: float32(60500)}},
		{cmd: `70 ABS(7.05D+4)`, inp: []object.Object{&object.FloatDbl{Value: float64(70500)}}, exp: &object.FloatDbl{Value: float64(70500)}},
		{cmd: `80 ABS(-8.05D+4)`, inp: []object.Object{&object.FloatDbl{Value: float64(-80500)}}, exp: &object.FloatDbl{Value: float64(80500)}},
		{cmd: `90 ABS( "Foo" )`, lnum: 90, inp: []object.Object{&object.String{Value: "Foo"}}, exp: &object.Error{Message: syntaxErr + " in 90"}},
		{cmd: `100 ABS( "Foo", "Bar" )`, lnum: 100, inp: []object.Object{&object.String{Value: "Foo"}, &object.String{Value: "Bar"}}, exp: &object.Error{Message: syntaxErr + " in 100"}},
		{cmd: `110 ABS(-32769)`, inp: []object.Object{&object.IntDbl{Value: -32769}}, exp: &object.IntDbl{Value: 32769}},
		{cmd: `120 X% = -40 : ABS(X%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: -40}}}, exp: &object.Integer{Value: 40}},
	}

	runTests(t, "ABS", tests)
}

func TestAsc(t *testing.T) {
	var mt mocks.MockTerm
	mocks.InitMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	bval := bstrEncode(2, env, &object.Integer{Value: 2251})
	tests := []test{
		{cmd: `10 ASC("Alpha")`, lnum: 10, inp: []object.Object{&object.String{Value: "Alpha"}}, exp: 65},
		{cmd: `20 ASC("")`, lnum: 20, inp: []object.Object{&object.String{Value: ""}}, exp: &object.Error{Message: illegalFuncCallErr + " in 20"}},
		{cmd: `30 ASC("FRED", "Joe")`, lnum: 30, inp: []object.Object{&object.String{Value: "FRED"}, &object.String{Value: "Joe"}}, exp: &object.Error{Message: syntaxErr + " in 30"}},
		{cmd: `40 A$ = "Alpha" : ASC(A$)`, lnum: 40, inp: []object.Object{&object.TypedVar{TypeID: "$", Value: &object.String{Value: "Alpha"}}}, exp: 65},
		{cmd: `50 A$ = MKI$(2251) : ASC(A$)`, lnum: 50, inp: []object.Object{&object.TypedVar{TypeID: "$", Value: bval}}, exp: 8},
		{cmd: `60 ASC(3)`, lnum: 60, inp: []object.Object{&object.Integer{Value: 3}}, exp: &object.Error{Message: syntaxErr + " in 60"}},
	}

	runTests(t, "ASC", tests)
}

func TestAtn(t *testing.T) {
	tests := []test{
		{cmd: `10 ATN(3)`, inp: []object.Object{&object.Integer{Value: 3}}, exp: &object.FloatDbl{Value: 1.2490457723982544}},
		{cmd: `20 ATN(3.335)`, inp: []object.Object{&object.Fixed{Value: decimal.New(3335, -3)}}, exp: &object.FloatDbl{Value: 1.2794770838980052}},
		{cmd: `30 ATN(3.335E+0)`, inp: []object.Object{&object.FloatSgl{Value: 3.335}}, exp: &object.FloatDbl{Value: 1.2794770870448673}},
		{cmd: `40 ATN(3.335D+0)`, inp: []object.Object{&object.FloatDbl{Value: 3.335}}, exp: &object.FloatDbl{Value: 1.2794770838980052}},
		{cmd: `50 ATN(3, 33)`, lnum: 50, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 33}}, exp: &object.Error{Message: syntaxErr + " in 50"}},
		{cmd: `60 ATN("Fred")`, lnum: 60, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 60"}},
		{cmd: `70 ATN(32769)`, inp: []object.Object{&object.IntDbl{Value: 32769}}, exp: &object.FloatDbl{Value: 1.5707658101480753}},
		{cmd: `80 X% = 3 : ATN(X%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 3}}}, exp: &object.FloatDbl{Value: 1.2490457723982544}},
	}

	runTests(t, "ATN", tests)
}

func TestCdbl(t *testing.T) {

	tests := []test{
		{cmd: `10 CDBL(32767)`, inp: []object.Object{&object.Integer{Value: 32767}}, exp: &object.IntDbl{Value: 32767}},
		{cmd: `20 CDBL(3.335)`, inp: []object.Object{&object.Fixed{Value: decimal.New(3335, -3)}}, exp: &object.FloatDbl{Value: float64(3.335)}},
		{cmd: `30 CDBL(7.3350E+1)`, inp: []object.Object{&object.FloatSgl{Value: 73.35}}, exp: &object.FloatDbl{Value: float64(73.3499984741211)}},
		{cmd: `40 CDBL(3.1234D+2)`, inp: []object.Object{&object.FloatDbl{Value: 312.34}}, exp: &object.FloatDbl{Value: float64(312.34)}},
		{cmd: `50 CDBL(3, 33)`, lnum: 50, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 33}}, exp: &object.Error{Message: syntaxErr + " in 50"}},
		{cmd: `60 CDBL("Fred")`, inp: []object.Object{&object.String{Value: "Fred"}}, lnum: 60, exp: &object.Error{Message: typeMismatchErr + " in 60"}},
		{cmd: `70 CDBL(32769)`, inp: []object.Object{&object.IntDbl{Value: 32769}}, exp: &object.IntDbl{Value: 32769}},
		{cmd: `80 X% = 500 : exp: CDBL(X%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 500}}}, exp: &object.IntDbl{Value: 500}},
	}

	runTests(t, "CDBL", tests)
}

func TestChr(t *testing.T) {
	tests := []test{
		{cmd: `10 CHR$(3, 33)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 33}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 CHR$("Fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 CHR$(320)`, lnum: 30, inp: []object.Object{&object.Integer{Value: 320}}, exp: &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{cmd: `40 CHR$(-32)`, lnum: 40, inp: []object.Object{&object.Integer{Value: -32}}, exp: &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{cmd: `50 CHR$(41)`, inp: []object.Object{&object.Integer{Value: 41}}, exp: &object.String{Value: ")"}},
	}

	runTests(t, "CHR$", tests)
}

func TestCint(t *testing.T) {

	tests := []test{
		{cmd: `10 CINT(46)`, inp: []object.Object{&object.Integer{Value: 46}}, exp: &object.Integer{Value: 46}},
		{cmd: `70 CINT(3, 33)`, lnum: 70, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 33}}, exp: &object.Error{Message: syntaxErr + " in 70"}},
		{cmd: `80 CINT("Fred")`, lnum: 80, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 80"}},
		{cmd: `100 CINT(-32769)`, lnum: 100, inp: []object.Object{&object.IntDbl{Value: -32769}}, exp: &object.Error{Message: overflowErr + " in 100"}},
	}

	runTests(t, "CINT", tests)
}

func TestCos(t *testing.T) {
	//rc10, _ := decimal.NewFromString("-.4321779")
	//rc20, _ := decimal.NewFromString("-.981355")

	//rc50, _ := decimal.NewFromString("-.4594971")
	//rc60, _ := decimal.NewFromString("-.2459203")

	//rc100, _ := decimal.NewFromString("-.579265")

	tests := []test{
		{cmd: `10 COS(46)`, inp: []object.Object{&object.Integer{Value: 46}}, exp: &object.FloatDbl{Value: -0.4321779448847783}},
		{cmd: `20 COS(3.335)`, inp: []object.Object{&object.Fixed{Value: decimal.New(3335, -3)}}, exp: &object.FloatDbl{Value: -0.9813550281508611}},
		{cmd: `50 COS(7.3350E+1)`, inp: []object.Object{&object.FloatSgl{Value: 73.35}}, exp: &object.FloatDbl{Value: -0.45949708599024536}},
		{cmd: `60 COS(3.1234D+2)`, inp: []object.Object{&object.FloatDbl{Value: 312.34}}, exp: &object.FloatDbl{Value: -0.2459202961599014}},
		{cmd: `70 COS(3, 33)`, lnum: 70, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 33}}, exp: &object.Error{Message: syntaxErr + " in 70"}},
		{cmd: `80 COS("Fred")`, lnum: 80, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 80"}},
		{cmd: `100 COS(-32769)`, inp: []object.Object{&object.IntDbl{Value: -32769}}, exp: &object.FloatDbl{Value: -0.5792650135068360}},
		{cmd: `110 X% = 46 : COS(X%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 46}}}, exp: &object.FloatDbl{Value: -0.4321779448847783}},
	}

	runTests(t, "COS", tests)
}

func TestCsng(t *testing.T) {
	rc20, _ := decimal.NewFromString("3.335")
	rc50 := &object.FloatSgl{Value: float32(73.350)}

	rc100 := &object.IntDbl{Value: -32769}

	tests := []test{
		{cmd: `10 CSNG(46)`, inp: []object.Object{&object.Integer{Value: 46}}, exp: &object.Integer{Value: 46}},
		{cmd: `20 CSNG(3.335)`, inp: []object.Object{&object.Fixed{Value: decimal.New(3335, -3)}}, exp: &object.Fixed{Value: rc20}},
		{cmd: `50 CSNG(7.3350E+1)`, inp: []object.Object{&object.FloatSgl{Value: 73.35}}, exp: rc50},
		{cmd: `60 CSNG(3.1234D+2)`, inp: []object.Object{&object.FloatDbl{Value: 312.34}}, exp: &object.FloatSgl{Value: 312.34}},
		{cmd: `70 CSNG(3, 33)`, lnum: 70, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 33}}, exp: &object.Error{Message: syntaxErr + " in 70"}},
		{cmd: `80 CSNG("Fred")`, inp: []object.Object{&object.String{Value: "Fred"}}, lnum: 80, exp: &object.Error{Message: typeMismatchErr + " in 80"}},
		{cmd: `100 CSNG(-32769)`, inp: []object.Object{&object.IntDbl{Value: -32769}}, exp: rc100},
		{cmd: `110 X% = 46 : CSNG(X%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 46}}}, exp: &object.Integer{Value: 46}},
	}

	runTests(t, "CSNG", tests)
}

func TestCvd(t *testing.T) {
	// a simple way to get my parameter

	var mt mocks.MockTerm
	fn, _ := Builtins["MKD$"]
	mocks.InitMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	res := fn.Fn(env, fn, &object.Integer{Value: -12})

	tests := []test{
		{cmd: `10 CVD("ABCD", "EFGH")`, lnum: 10, inp: []object.Object{&object.String{Value: "ABCD"}, &object.String{Value: "EFGH"}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `30 CVD(123)`, lnum: 30, inp: []object.Object{&object.Integer{Value: 123}}, exp: &object.Error{Message: typeMismatchErr + " in 30"}},
		{cmd: `40 CVD("........")`, inp: []object.Object{&object.String{Value: "........"}}, exp: &object.FloatDbl{Value: 3327647950551526912}},
		{cmd: `50 A$ = MKD$(-12) : CVD(A$)`, inp: []object.Object{res}, exp: &object.Integer{Value: -12}},
	}

	runTests(t, "CVD", tests)
}

func TestCvi(t *testing.T) {
	tests := []test{
		{cmd: `10 CVI("ABCD", "EFGH")`, lnum: 10, inp: []object.Object{&object.String{Value: "ABCD"}, &object.String{Value: "EFGH"}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `30 CVI(123)`, lnum: 30, inp: []object.Object{&object.Integer{Value: 123}}, exp: &object.Error{Message: typeMismatchErr + " in 30"}},
		{cmd: `60 CVI("..")`, inp: []object.Object{&object.String{Value: ".."}}, exp: &object.Integer{Value: 11822}},
	}

	runTests(t, "CVI", tests)
}

func TestCvs(t *testing.T) {
	tests := []test{
		{cmd: `10 CVS("ABCD", "EFGH")`, lnum: 10, inp: []object.Object{&object.String{Value: "ABCD"}, &object.String{Value: "EFGH"}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `30 CVS(123)`, lnum: 30, inp: []object.Object{&object.Integer{Value: 123}}, exp: &object.Error{Message: typeMismatchErr + " in 30"}},
		//{cmd: `40 Y$ = MKS$(35) : CVS(Y$)`, exp: &object.IntDbl{Value: 35}},
		{cmd: `50 CVS("..")`, inp: []object.Object{&object.String{Value: ".."}}, exp: &object.Integer{Value: 11822}},
	}

	runTests(t, "CVS", tests)
}

func TestExp(t *testing.T) {
	tests := []test{
		{cmd: `10 EXP(2, 3)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 3}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 EXP("ABCDEF")`, lnum: 20, inp: []object.Object{&object.String{Value: "ABCDEF"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 EXP(123)`, lnum: 30, inp: []object.Object{&object.Integer{Value: 123}}, exp: &object.Error{Message: overflowErr + " in 30"}},
		{cmd: `40 Y% = 35 : EXP(Y%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 35}}}, exp: &object.FloatSgl{Value: 1586013445029888}},
		{cmd: `50 Y = 38999 : EXP(Y)`, lnum: 50, inp: []object.Object{&object.TypedVar{TypeID: "#", Value: &object.IntDbl{Value: 38999}}}, exp: &object.Error{Message: overflowErr + " in 50"}},
		{cmd: `60 Y = 3.8999 : EXP(Y)`, inp: []object.Object{&object.Fixed{Value: decimal.New(38999, -4)}}, exp: &object.FloatSgl{Value: 49.397511}},
		{cmd: `70 Y = 3.8999E+00 : EXP(Y)`, inp: []object.Object{&object.FloatSgl{Value: 3.8999}}, exp: &object.FloatSgl{Value: 49.397507}},
		{cmd: `80 Y = 3.8999D+00 : EXP(Y)`, inp: []object.Object{&object.FloatDbl{Value: 3.8999}}, exp: &object.FloatSgl{Value: 49.397511}},
		{cmd: `90 Y$ = "ABCD" : EXP(Y$)`, lnum: 50, inp: []object.Object{&object.TypedVar{TypeID: "$", Value: &object.String{Value: "ABCD"}}}, exp: &object.Error{Message: typeMismatchErr + " in 50"}},
	}

	runTests(t, "EXP", tests)
}

func TestFix(t *testing.T) {
	tests := []test{
		{cmd: `10 FIX(2, 3)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 3}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 FIX("ABCDEF")`, lnum: 20, inp: []object.Object{&object.String{Value: "ABCDEF"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 FIX(4294967296)`, lnum: 30, inp: []object.Object{&object.FloatDbl{Value: 4294967296}}, exp: &object.Error{Message: overflowErr + " in 30"}},
		{cmd: `60 Y = 3.8999 : FIX(Y)`, inp: []object.Object{&object.Fixed{Value: decimal.New(38999, -4)}}, exp: &object.Integer{Value: 3}},
		{cmd: `90 Y = 65999 : FIX(Y)`, inp: []object.Object{&object.IntDbl{Value: 65999}}, exp: &object.IntDbl{Value: 65999}},
	}

	runTests(t, "FIX", tests)
}

func TestHex(t *testing.T) {
	tests := []test{
		{cmd: `10 HEX$(2, 3)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 3}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 HEX$("ABCDEF")`, lnum: 20, inp: []object.Object{&object.String{Value: "ABCDEF"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 HEX$(65999)`, lnum: 30, inp: []object.Object{&object.IntDbl{Value: 65999}}, exp: &object.Error{Message: overflowErr + " in 30"}},
		{cmd: `40 Y% = 35 : HEX$(y%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 35}}}, exp: &object.String{Value: "23"}},
	}

	runTests(t, "HEX$", tests)
}

func TestInputStr(t *testing.T) {
	tests := []struct {
		tt   test
		keys string
	}{
		{tt: test{cmd: `10 INPUT$(1, 2)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}}, keys: ""},
		{tt: test{cmd: `20 INPUT$("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}}, keys: ""},
		{tt: test{cmd: `30 INPUT$(0)`, lnum: 30, inp: []object.Object{&object.Integer{Value: 0}}, exp: &object.Error{Message: illegalFuncCallErr + " in 30"}}, keys: ""},
		{tt: test{cmd: `40 INPUT$(2)`, inp: []object.Object{&object.Integer{Value: 2}}, exp: &object.String{Value: "AB"}}, keys: "AB"},
	}

	for _, tt := range tests {
		fn, ok := Builtins["INPUT$"]

		assert.True(t, ok, "Failed to find %s() function", "INPUT$")

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		mt.StrVal = &tt.keys
		env := object.NewTermEnvironment(mt)
		if tt.tt.lnum != 0 {
			env.Set(token.LINENUM, &object.IntDbl{Value: int32(tt.tt.lnum)})
		}
		res := fn.Fn(env, fn, tt.tt.inp...)

		compareObjects(tt.tt.cmd, res, tt.tt.exp, t)
	}
}

func TestInstr(t *testing.T) {
	tests := []test{
		{cmd: `10 INSTR("Fred")`, lnum: 10, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 INSTR( 4, "fred", "George", "Sam")`, lnum: 20, inp: []object.Object{
			&object.Integer{Value: 4}, &object.String{Value: "fred"}, &object.String{Value: "George"}, &object.String{Value: "Sam"},
		}, exp: &object.Error{Message: syntaxErr + " in 20"}},
		{cmd: `30 INSTR("George", "org")`, inp: []object.Object{
			&object.Integer{Value: 2}, &object.String{Value: "George"}, &object.String{Value: "org"},
		}, exp: &object.Integer{Value: 3}},
		{cmd: `40 INSTR(2, "George", "get")`, inp: []object.Object{
			&object.Integer{Value: 2}, &object.String{Value: "George"}, &object.String{Value: "get"},
		}, exp: &object.Integer{Value: 0}},
		{cmd: `50 X# = 6 : INSTR(2, "George", X#)`, lnum: 50, inp: []object.Object{
			&object.Integer{Value: 2}, &object.String{Value: "George"}, &object.TypedVar{TypeID: "#", Value: &object.IntDbl{Value: 6}},
		}, exp: &object.Error{Message: syntaxErr + " in 50"}},
		{cmd: `60 INSTR(0, "George", "ge")`, lnum: 60, inp: []object.Object{
			&object.Integer{Value: 0}, &object.String{Value: "George"}, &object.String{Value: "Ge"},
		}, exp: &object.Error{Message: illegalArgErr + " in 60"}},
		{cmd: `70 INSTR(390, "George", "ge")`, lnum: 70, inp: []object.Object{
			&object.Integer{Value: 390}, &object.String{Value: "George"}, &object.String{Value: "Ge"},
		}, exp: &object.Error{Message: illegalFuncCallErr + " in 70"}},
		{cmd: `80 INSTR(10, "George", "ge")`, inp: []object.Object{
			&object.Integer{Value: 10}, &object.String{Value: "George"}, &object.String{Value: "ge"},
		}, exp: &object.Integer{Value: 0}},
	}

	runTests(t, "INSTR", tests)
}

func TestInt(t *testing.T) {
	tests := []test{
		{cmd: `10 INT(2, 3)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 3}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 INT("ABCDEF")`, lnum: 20, inp: []object.Object{&object.String{Value: "ABCDEF"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 INT(429496729)`, lnum: 30, inp: []object.Object{&object.FloatDbl{Value: 4294967296}}, exp: &object.Error{Message: overflowErr + " in 30"}},
		{cmd: `40 Y% = 35 : INT(Y%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 35}}}, exp: &object.Integer{Value: 35}},
		{cmd: `90 Y = 65999 : INT(Y)`, inp: []object.Object{&object.Fixed{Value: decimal.New(65999, 0)}}, exp: &object.IntDbl{Value: 65999}},
	}

	runTests(t, "INT", tests)
}

func TestLeft(t *testing.T) {
	// a simple way to get my parameter

	var mt mocks.MockTerm
	fn, _ := Builtins["MKS$"]
	mocks.InitMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	res := fn.Fn(env, fn, &object.IntDbl{Value: 65999})

	tests := []test{
		{cmd: `10 LEFT$("Fred")`, lnum: 10, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 LEFT$( 3, "fred")`, lnum: 20, inp: []object.Object{&object.Integer{Value: 3}, &object.String{Value: "fred"}}, exp: &object.Error{Message: syntaxErr + " in 20"}},
		{cmd: `30 LEFT$("George", 3)`, lnum: 30, inp: []object.Object{&object.String{Value: "George"}, &object.Integer{Value: 3}}, exp: &object.String{Value: "Geo"}},
		//{`40 LEFT$("George", 0)`, &object.String{Value: ""}},
		{cmd: `50 LEFT$("George", 300)`, lnum: 50, inp: []object.Object{&object.String{Value: "George"}, &object.Integer{Value: 300}}, exp: &object.Error{Message: illegalFuncCallErr + " in 50"}},
		{cmd: `60 X$ = MKS$(65999) : LEFT$( X$, 2)`, inp: []object.Object{res, &object.Integer{Value: 2}}, exp: &object.BStr{Value: []byte{0xcf, 0x01}}},
	}
	runTests(t, "LEFT$", tests)
}

func TestLen(t *testing.T) {
	tests := []test{
		{cmd: `30 LEN("hello world")`, inp: []object.Object{&object.String{Value: "hello world"}}, exp: 11},
		{cmd: `40 LEN(1)`, lnum: 40, inp: []object.Object{&object.Integer{Value: 1}}, exp: &object.Error{Message: typeMismatchErr + " in 40"}},
		{cmd: `50 LEN("one", "two")`, lnum: 50, inp: []object.Object{&object.String{Value: "one"}, &object.String{Value: "two"}}, exp: &object.Error{Message: syntaxErr + " in 50"}},
	}
	runTests(t, "LEN", tests)
}

func TestLog(t *testing.T) {
	tests := []test{
		{cmd: `10 LOG(2, 3)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 3}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 LOG("ABCDEF")`, lnum: 20, inp: []object.Object{&object.String{Value: "ABCDEF"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 LOG(-2)`, lnum: 30, inp: []object.Object{&object.Integer{Value: -2}}, exp: &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{cmd: `40 Y% = 35 : LOG(Y%)`, inp: []object.Object{&object.TypedVar{TypeID: "%", Value: &object.Integer{Value: 35}}}, exp: &object.FloatSgl{Value: 3.5553482}},
	}

	runTests(t, "LOG", tests)
}

func TestLPOS(t *testing.T) {
	// TODO: will need more once I implement printing
	tests := []test{
		{cmd: `10 LPOS(2, 3)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 3}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 LPOS("A")`, inp: []object.Object{&object.String{Value: "A"}}, exp: &object.Integer{Value: 0}},
	}

	runTests(t, "LPOS", tests)
}

func TestMID(t *testing.T) {
	// a simple way to get my parameter

	var mt mocks.MockTerm
	fn := Builtins["MKD$"]
	mocks.InitMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	res := fn.Fn(env, fn, &object.IntDbl{Value: 35456778})

	tests := []test{
		{cmd: `10 A$ = "Georgia" : MID$(A$)`, lnum: 10, inp: []object.Object{&object.TypedVar{TypeID: "$", Value: &object.String{Value: "Georgia"}}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 A$ = "Georgia" : MID$(A$,4)`, inp: []object.Object{
			&object.String{Value: "Georgia"}, &object.Integer{Value: 4},
		}, exp: &object.String{Value: "rgia"}},
		{cmd: `30 A$ = "Georgia" : MID$(A$,4,2)`, inp: []object.Object{
			&object.TypedVar{TypeID: "$", Value: &object.String{Value: "Georgia"}}, &object.Integer{Value: 4}, &object.Integer{Value: 2},
		}, exp: &object.String{Value: "rg"}},
		{cmd: `40 A$ = "Georgia" : MID$(A$,4,300)`, lnum: 40, inp: []object.Object{
			&object.TypedVar{TypeID: "$", Value: &object.String{Value: "Georgia"}}, &object.Integer{Value: 4}, &object.Integer{Value: 300},
		}, exp: &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{cmd: `50 A$ = "Georgia" : MID$(A$,"4",3)`, lnum: 50, inp: []object.Object{
			&object.TypedVar{TypeID: "$", Value: &object.String{Value: "Georgia"}}, &object.String{Value: "4"}, &object.Integer{Value: 3},
		}, exp: &object.Error{Message: syntaxErr + " in 50"}},
		{cmd: `60 A$ = MKD$(35456778) : MID$(A$,2,2)`, inp: []object.Object{
			res, &object.Integer{Value: 2}, &object.Integer{Value: 2},
		}, exp: &object.BStr{Value: []byte{0x07, 0x1d}}},
	}

	runTests(t, "MID$", tests)
}

func TestMKD(t *testing.T) {
	tests := []test{
		{cmd: `10 MKD$("..")`, lnum: 10, inp: []object.Object{&object.String{Value: ".."}}, exp: &object.Error{Message: typeMismatchErr + " in 10"}},
		{cmd: `20 MKD$(1, 2)`, lnum: 20, inp: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 20"}},
		{cmd: `30 MKD$(35)`, inp: []object.Object{&object.Integer{Value: 35}}, exp: &object.BStr{Value: []byte{0x23, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
	}

	runTests(t, "MKD$", tests)
}

func TestMKI(t *testing.T) {
	tests := []test{
		{cmd: `10 MKI$("..")`, lnum: 10, inp: []object.Object{&object.String{Value: ".."}}, exp: &object.Error{Message: typeMismatchErr + " in 10"}},
		{cmd: `20 MKI$(1, 2)`, lnum: 20, inp: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 20"}},
		{cmd: `60 MKI$(-32769)`, lnum: 60, inp: []object.Object{&object.IntDbl{Value: -32769}}, exp: &object.Error{Message: overflowErr + " in 60"}},
		{cmd: `70 MKI$(32768)`, lnum: 70, inp: []object.Object{&object.IntDbl{Value: 32768}}, exp: &object.Error{Message: overflowErr + " in 70"}},
	}

	runTests(t, "MKI$", tests)
}

func TestMKS(t *testing.T) {
	tests := []test{
		{cmd: `10 MKS$("..")`, lnum: 10, inp: []object.Object{&object.String{Value: ".."}}, exp: &object.Error{Message: typeMismatchErr + " in 10"}},
		{cmd: `20 MKS$(1, 2)`, lnum: 20, inp: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 20"}},
		{cmd: `30 MKS$(35)`, inp: []object.Object{&object.Integer{Value: 35}}, exp: &object.BStr{Value: []byte{0x23, 0x00, 0x00, 0x00}}},
	}

	runTests(t, "MKS$", tests)
}

func TestOct(t *testing.T) {
	tests := []test{
		{cmd: `10 OCT$(2, 3)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 OCT$("ABCDEF")`, lnum: 20, inp: []object.Object{&object.String{Value: "ABCDEF"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 OCT$(65999)`, lnum: 30, inp: []object.Object{&object.IntDbl{Value: 65999}}, exp: &object.Error{Message: overflowErr + " in 30"}},
		{cmd: `40 Y% = 35 : OCT$(Y%)`, inp: []object.Object{&object.Integer{Value: 35}}, exp: &object.String{Value: "43"}},
	}

	runTests(t, "OCT$", tests)
}

func TestRight(t *testing.T) {
	tests := []test{
		{cmd: `10 RIGHT$("Fred")`, lnum: 10, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 RIGHT$( 3, "fred")`, lnum: 20, inp: []object.Object{&object.Integer{Value: 3}, &object.String{Value: "fred"}}, exp: &object.Error{Message: syntaxErr + " in 20"}},
		{cmd: `30 RIGHT$("George", 3)`, inp: []object.Object{&object.String{Value: "George"}, &object.Integer{Value: 3}}, exp: &object.String{Value: "rge"}},
		{cmd: `40 RIGHT$("George", 0)`, inp: []object.Object{&object.String{Value: "George"}, &object.Integer{Value: 0}}, exp: &object.String{Value: ""}},
		{cmd: `50 RIGHT$("George", 300)`, lnum: 50, inp: []object.Object{&object.String{Value: "George"}, &object.Integer{Value: 300}}, exp: &object.Error{Message: illegalFuncCallErr + " in 50"}},
		{cmd: `60 X$ = MKS$(65999) : RIGHT$( X$, 2)`, inp: []object.Object{&object.BStr{Value: []byte{0x03, 0x02, 0x01, 0x00}}, &object.Integer{Value: 2}}, exp: &object.BStr{Value: []byte{0x01, 0x00}}},
		{cmd: `70 RIGHT$("George", 10)`, inp: []object.Object{&object.String{Value: "George"}, &object.Integer{Value: 10}}, exp: &object.String{Value: "George"}},
	}
	runTests(t, "RIGHT$", tests)
}

func TestRnd(t *testing.T) {
	tests := []test{
		{cmd: `10 RND(5, 5)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 5}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 RND("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "Fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 RND(0)`, inp: []object.Object{&object.Integer{Value: 0}}, exp: &object.FloatSgl{Value: 0.615608156}},
	}
	runTests(t, "RND", tests)
}

func TestScreen(t *testing.T) {
	tests := []test{
		{cmd: `10 SCREEN(5)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}}, scrn: "", exp: &object.Error{Message: illegalFuncCallErr + " in 10"}},
		{cmd: `20 SCREEN(5, "fred")`, lnum: 20, inp: []object.Object{&object.Integer{Value: 5}, &object.String{Value: "fred"}}, scrn: "", exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 SCREEN(1,1)`, inp: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 1}}, scrn: "470", exp: &object.Integer{Value: 52}},
		{cmd: `40 SCREEN(1,2)`, inp: []object.Object{&object.Integer{Value: 1}, &object.Integer{Value: 2}}, scrn: "470", exp: &object.Integer{Value: 55}},
	}

	runTests(t, "SCREEN", tests)
}

func TestSgn(t *testing.T) {
	tests := []test{
		{cmd: `10 SGN(5, 2)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 SGN("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 SGN(5)`, inp: []object.Object{&object.Integer{Value: 5}}, exp: &object.Integer{Value: 1}},
		{cmd: `40 SGN(0)`, inp: []object.Object{&object.Integer{Value: 0}}, exp: &object.Integer{Value: 0}},
		{cmd: `50 SGN(-2)`, inp: []object.Object{&object.Integer{Value: -2}}, exp: &object.Integer{Value: -1}},
	}

	runTests(t, "SGN", tests)
}

func TestSin(t *testing.T) {
	tests := []test{
		{cmd: `10 SIN(5, 2)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 SIN("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 SIN(5)`, inp: []object.Object{&object.Integer{Value: 5}}, exp: &object.FloatSgl{Value: -0.9589243}},
		{cmd: `40 SIN(0.5)`, inp: []object.Object{&object.FloatSgl{Value: 0.5}}, exp: &object.FloatSgl{Value: 0.479425550}},
	}

	runTests(t, "SIN", tests)
}

func TestSpaces(t *testing.T) {
	tests := []test{
		{cmd: `10 SPACE$(5, 2)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 SPACE$("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 SPACE$(-1)`, lnum: 30, inp: []object.Object{&object.Integer{Value: -1}}, exp: &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{cmd: `40 SPACE$(256)`, lnum: 40, inp: []object.Object{&object.Integer{Value: 256}}, exp: &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{cmd: `50 SPACE$(2)`, inp: []object.Object{&object.Integer{Value: 2}}, exp: &object.String{Value: "  "}},
		{cmd: `60 SPACE$(0)`, inp: []object.Object{&object.Integer{Value: 0}}, exp: &object.String{Value: ""}},
	}

	runTests(t, "SPACE$", tests)
}

func TestSqr(t *testing.T) {
	tests := []test{
		{cmd: `10 SQR(5, 2)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 SQR("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 SQR(25)`, inp: []object.Object{&object.Integer{Value: 25}}, exp: &object.FloatSgl{Value: 5}},
		{cmd: `40 SQR(0.5)`, inp: []object.Object{&object.FloatSgl{Value: 0.5}}, exp: &object.FloatSgl{Value: 0.707106769}},
	}

	runTests(t, "SQR", tests)
}

func TestStrs(t *testing.T) {
	tests := []test{
		{cmd: `10 STR$(5, 2)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 STR$("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 STR$(-1)`, inp: []object.Object{&object.Integer{Value: -1}}, exp: &object.String{Value: "-1"}},
		{cmd: `40 STR$(256)`, inp: []object.Object{&object.Integer{Value: 256}}, exp: &object.String{Value: "256"}},
		{cmd: `50 STR$(2)`, inp: []object.Object{&object.Integer{Value: 2}}, exp: &object.String{Value: "2"}},
		{cmd: `60 STR$(1.5)`, inp: []object.Object{&object.FloatSgl{Value: 1.5}}, exp: &object.String{Value: "1.5"}},
	}

	runTests(t, "STR$", tests)
}

func TestStrings(t *testing.T) {
	tests := []test{
		{cmd: `10 STRING$(5)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 STRING$("fred", 52)`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}, &object.Integer{Value: 52}}, exp: &object.Error{Message: typeMismatchErr + " in 20"}},
		{cmd: `30 STRING$(-1, 52)`, lnum: 30, inp: []object.Object{&object.Integer{Value: -1}, &object.Integer{Value: 52}}, exp: &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{cmd: `40 STRING$(3, 256)`, lnum: 40, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 256}}, exp: &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{cmd: `50 STRING$(2, 50)`, inp: []object.Object{&object.Integer{Value: 2}, &object.Integer{Value: 50}}, exp: &object.String{Value: "22"}},
		{cmd: `60 STRING$(3, "34")`, inp: []object.Object{&object.Integer{Value: 3}, &object.String{Value: "34"}}, exp: &object.String{Value: "333"}},
		{cmd: `70 STRING$(3, 0)`, inp: []object.Object{&object.Integer{Value: 3}, &object.Integer{Value: 0}}, exp: &object.BStr{Value: []byte{0x00, 0x00, 0x00}}},
	}

	runTests(t, "STRING$", tests)
}

func TestTab(t *testing.T) {
	tests := []struct {
		inp []object.Object
		exp object.Object
		col int
	}{
		{inp: []object.Object{&object.Integer{Value: 5}}, col: 5},
		{inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 0}}, exp: &object.Error{Code: 2, Message: "Syntax error"}},
		{inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Code: 13, Message: "Type mismatch"}},
	}

	for _, tt := range tests {
		fn, ok := Builtins["TAB"]

		assert.True(t, ok, "Failed to find TAB() function")

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		res := fn.Fn(env, fn, tt.inp...)

		assert.EqualValuesf(t, tt.exp, res, "call to TAB(%s) returned %T", tt.inp[0].Inspect(), res)
		assert.EqualValuesf(t, tt.col, *mt.Col, "expected to be @col %d but at %d", tt.col, *mt.Col)
	}

}

func TestTan(t *testing.T) {
	tests := []test{
		{cmd: `10 TAN(5, 2)`, lnum: 10, inp: []object.Object{&object.Integer{Value: 5}, &object.Integer{Value: 2}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 TAN("fred")`, lnum: 20, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.Error{Message: typeMismatchErr}},
		{cmd: `30 TAN(5)`, inp: []object.Object{&object.Integer{Value: 5}}, exp: &object.FloatSgl{Value: -3.380515099}},
		{cmd: `40 TAN(0.5)`, inp: []object.Object{&object.FloatSgl{Value: 0.5}}, exp: &object.FloatSgl{Value: 0.546302497}},
	}

	runTests(t, "TAN", tests)
}

func TestVal(t *testing.T) {
	tests := []test{
		{cmd: `10 VAL("5", "2")`, lnum: 10, inp: []object.Object{&object.String{Value: "5"}, &object.String{Value: "2"}}, exp: &object.Error{Message: syntaxErr + " in 10"}},
		{cmd: `20 VAL("fred")`, inp: []object.Object{&object.String{Value: "fred"}}, exp: &object.FloatSgl{Value: 0}},
		{cmd: `30 VAL("5")`, inp: []object.Object{&object.String{Value: "5"}}, exp: &object.FloatSgl{Value: 5}},
		{cmd: `40 VAL("0.5")`, inp: []object.Object{&object.String{Value: "0.5"}}, exp: &object.FloatSgl{Value: 0.5}},
		{cmd: `50 VAL(5)`, lnum: 50, inp: []object.Object{&object.Integer{Value: 5}}, exp: &object.Error{Message: typeMismatchErr + " in 50"}},
	}
	runTests(t, "VAL", tests)
}

func Test_BstrEncode(t *testing.T) {
	tests := []struct {
		cmd  string
		inp  object.Object
		size int
		lnum int
		exp  object.Object
	}{
		{cmd: "string-fail", inp: &object.String{Value: "Fred"}, size: 2, lnum: 10, exp: &object.Error{Message: typeMismatchErr + " in 10"}},
		{cmd: "intdbl(40000) overflow", inp: &object.IntDbl{Value: 40000}, size: 2, lnum: 20, exp: &object.Error{Message: overflowErr + " in 20"}},
		{cmd: "int(2)", inp: &object.Integer{Value: 10}, size: 2, exp: &object.BStr{Value: []byte{0x0a, 0x00}}},
		{cmd: "intdbl(4)", inp: &object.IntDbl{Value: 459}, size: 4, exp: &object.BStr{Value: []byte{0xcb, 0x01, 0x0, 0x00}}},
		{cmd: "Fixed(8)", inp: &object.Fixed{Value: decimal.New(1594897133, -4)}, size: 8, exp: &object.BStr{Value: []byte{2, 111, 2, 0, 0, 0, 0, 0}}},
		{cmd: "FloatSgl(4)", inp: &object.FloatSgl{Value: 3.14159}, size: 4, exp: &object.BStr{Value: []byte{3, 0, 0, 0}}},
		{cmd: "FloatDbl(4)", inp: &object.FloatDbl{Value: 35.14159}, size: 4, exp: &object.BStr{Value: []byte{35, 0, 0, 0}}},
	}

	for _, tt := range tests {

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		if tt.lnum != 0 {
			env.Set(token.LINENUM, &object.IntDbl{Value: int32(tt.lnum)})
		}
		res := bstrEncode(tt.size, env, tt.inp)

		compareObjects(tt.cmd, res, tt.exp, t)
	}
}

func Test_FixType(t *testing.T) {
	tests := []struct {
		cmd  string
		inp  interface{}
		lnum int
		exp  object.Object
	}{
		{cmd: "FixType(int16)", inp: int16(3000), exp: &object.Integer{Value: 3000}},
		{cmd: "FixType(int32)", inp: int32(32769), exp: &object.IntDbl{Value: 32769}},
		{cmd: "FixType(int)", inp: int(3000), exp: &object.Integer{Value: 3000}},
		{cmd: "FixType(float32)", inp: float32(3000), exp: &object.Integer{Value: 3000}},
		{cmd: "FixType(float64(3000))", inp: float64(3000), exp: &object.Integer{Value: 3000}},
		{cmd: "FixType(float64(3276912345677))", inp: float64(3276912345677), exp: &object.FloatDbl{Value: 3276912345677}},
		{cmd: `FixType("Fred")`, inp: "Fred", lnum: 10, exp: &object.Error{Message: typeMismatchErr + " in 10"}},
	}

	for _, tt := range tests {

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		if tt.lnum != 0 {
			env.Set(token.LINENUM, &object.IntDbl{Value: int32(tt.lnum)})
		}
		res := FixType(env, tt.inp)

		compareObjects(tt.cmd, res, tt.exp, t)
	}
}

/*
func Test_TryFixed(t *testing.T) {
	tests := []struct {
		cmd string
		inp interface{}
		exp object.Object
	}{
		{cmd: "123.45", inp: 123.45, exp: &object.Fixed{Value: decimal.New(12345, -2)}},
		{cmd: "BIG Float", inp: 41234560009191.4589765421, exp: nil},
		{cmd: "41.45812341234", inp: 41.45812341234, exp: nil},
	}

	for _, tt := range tests {
		res := tryFixed(tt.inp)

		if (res == nil) && (tt.exp == nil) {
			continue
		}

		compareObjects(tt.cmd, res, tt.exp, t)
	}
}*/
