package evaluator

import (
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/object"
)

func compareObjects(inp string, evald object.Object, want interface{}, t *testing.T) {
	if evald == nil {
		t.Fatalf("got nil return value!")
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
		{`90 ABS( "Foo" )`, &object.Error{Message: syntaxErr + " in 90"}},
		{`100 ABS( "Foo", "Bar" )`, &object.Error{Message: syntaxErr + " in 100"}},
		{`110 ABS(-32769)`, &object.IntDbl{Value: 32769}},
		{`120 X% = -40 : ABS(X%)`, &object.Integer{Value: 40}},
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
		{`20 ASC("")`, &object.Error{Message: illegalFuncCallErr + " in 20"}},
		{`30 ASC("FRED", "Joe")`, &object.Error{Message: syntaxErr + " in 30"}},
		{`40 A$ = "Alpha" : ASC(A$)`, 65},
		{`50 A$ = MKI$(2251) : ASC(A$)`, 8},
		{`60 ASC(3)`, &object.Error{Message: syntaxErr + " in 60"}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestAtn(t *testing.T) {
	// build the decimal
	rc10 := decimal.New(1249046, -6)
	rc20 := decimal.New(1279477, -6)
	rc70 := decimal.New(1570766, -6)

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 ATN(3)`, &object.Fixed{Value: rc10}},
		{`20 ATN(3.335)`, &object.Fixed{Value: rc20}},
		{`30 ATN(3.335E+0)`, &object.Fixed{Value: rc20}},
		{`40 ATN(3.335D+0)`, &object.Fixed{Value: rc20}},
		{`50 ATN(3, 33)`, &object.Error{Message: syntaxErr + " in 50"}},
		{`60 ATN("Fred")`, &object.Error{Message: typeMismatchErr + " in 60"}},
		{`70 ATN(32769)`, &object.Fixed{Value: rc70}},
		{`80 X% = 3 : ATN(X%)`, &object.Fixed{Value: rc10}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestCdbl(t *testing.T) {

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 CDBL(32767)`, &object.IntDbl{Value: 32767}},
		{`20 CDBL(3.335)`, &object.FloatDbl{Value: float64(3.335)}},
		{`30 CDBL(7.3350E+1)`, &object.FloatDbl{Value: float64(73.3499984741211)}},
		{`40 CDBL(3.1234D+2)`, &object.FloatDbl{Value: float64(312.34)}},
		{`50 CDBL(3, 33)`, &object.Error{Message: syntaxErr + " in 50"}},
		{`60 CDBL("Fred")`, &object.Error{Message: typeMismatchErr + " in 60"}},
		{`70 CDBL(32769)`, &object.IntDbl{Value: 32769}},
		{`80 X% = 500 : CDBL(X%)`, &object.IntDbl{Value: 500}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestChr(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 CHR$(3, 33)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 CHR$("Fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 CHR$(320)`, &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{`40 CHR$(-32)`, &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{`50 CHR$(41)`, &object.String{Value: ")"}},
		{`60 CHR$(41#)`, &object.String{Value: ")"}},
		{`70 CHR$(41.3)`, &object.String{Value: ")"}},
		{`80 CHR$(40.5)`, &object.String{Value: ")"}},
		{`90 CHR$(4.13E+01)`, &object.String{Value: ")"}},
		{`100 CHR$(4.05E+01)`, &object.String{Value: ")"}},
		{`110 CHR$(4.13D+01)`, &object.String{Value: ")"}},
		{`120 CHR$(4.05D+01)`, &object.String{Value: ")"}},
		{`130 X% = 41 : CHR$(X%)`, &object.String{Value: ")"}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}
func TestCint(t *testing.T) {

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 CINT(46)`, &object.Integer{Value: 46}},
		{`20 CINT(3.335)`, &object.Integer{Value: 3}},
		{`30 CINT(3.678335)`, &object.Integer{Value: 4}},
		{`40 CINT(-3.678335)`, &object.Integer{Value: -4}},
		{`50 CINT(7.3350E+1)`, &object.Integer{Value: 73}},
		{`60 CINT(3.1234D+2)`, &object.Integer{Value: 312}},
		{`70 CINT(3, 33)`, &object.Error{Message: syntaxErr + " in 70"}},
		{`80 CINT("Fred")`, &object.Error{Message: typeMismatchErr + " in 80"}},
		{`90 CINT(32768)`, &object.Error{Message: overflowErr + " in 90"}},
		{`100 CINT(-32769)`, &object.Error{Message: overflowErr + " in 100"}},
		{`110 X! = -3.678335 : CINT(X!)`, &object.Integer{Value: -4}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestCos(t *testing.T) {
	rc10, _ := decimal.NewFromString("-.4321779")
	rc20, _ := decimal.NewFromString("-.981355")
	rc30, _ := decimal.NewFromString("-.8593790")

	rc50, _ := decimal.NewFromString("-.4594971")
	rc60, _ := decimal.NewFromString("-.2459203")

	rc90, _ := decimal.NewFromString("0.3729378")
	rc100, _ := decimal.NewFromString("-.579265")

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 COS(46)`, &object.Fixed{Value: rc10}},
		{`20 COS(3.335)`, &object.Fixed{Value: rc20}},
		{`30 COS(3.678335)`, &object.Fixed{Value: rc30}},
		{`40 COS(-3.678335)`, &object.Fixed{Value: rc30}},
		{`50 COS(7.3350E+1)`, &object.Fixed{Value: rc50}},
		{`60 COS(3.1234D+2)`, &object.Fixed{Value: rc60}},
		{`70 COS(3, 33)`, &object.Error{Message: syntaxErr + " in 70"}},
		{`80 COS("Fred")`, &object.Error{Message: typeMismatchErr + " in 80"}},
		{`90 COS(32768)`, &object.Fixed{Value: rc90}},
		{`100 COS(-32769)`, &object.Fixed{Value: rc100}},
		{`110 X% = 46 : COS(X%)`, &object.Fixed{Value: rc10}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestCsng(t *testing.T) {
	rc20, _ := decimal.NewFromString("3.335")
	rc30, _ := decimal.NewFromString("3.678335")
	rc40, _ := decimal.NewFromString("-3.678335")
	rc50 := &object.FloatSgl{Value: float32(73.350)}
	rc60 := &object.FloatSgl{Value: float32(312.34)}

	rc90 := &object.IntDbl{Value: 32768}
	rc100 := &object.IntDbl{Value: -32769}

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 CSNG(46)`, &object.Integer{Value: 46}},
		{`20 CSNG(3.335)`, &object.Fixed{Value: rc20}},
		{`30 CSNG(3.678335)`, &object.Fixed{Value: rc30}},
		{`40 CSNG(-3.678335)`, &object.Fixed{Value: rc40}},
		{`50 CSNG(7.3350E+1)`, rc50},
		{`60 CSNG(3.1234D+2)`, rc60},
		{`70 CSNG(3, 33)`, &object.Error{Message: syntaxErr + " in 70"}},
		{`80 CSNG("Fred")`, &object.Error{Message: typeMismatchErr + " in 80"}},
		{`90 CSNG(32768)`, rc90},
		{`100 CSNG(-32769)`, rc100},
		{`110 X% = 46 : CSNG(X%)`, &object.Integer{Value: 46}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestCvd(t *testing.T) {

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 CVD("ABCD", "EFGH")`, &object.Error{Message: syntaxErr + " in 10"}},
		{`30 CVD(123)`, &object.Error{Message: typeMismatchErr + " in 30"}},
		{`40 CVD("........")`, &object.FloatDbl{Value: 3327647950551526912}},
		{`50 A$ = MKD$(-12) : CVD(A$)`, &object.Integer{Value: -12}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestCvi(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 CVI("ABCD", "EFGH")`, &object.Error{Message: syntaxErr + " in 10"}},
		{`30 CVI(123)`, &object.Error{Message: typeMismatchErr + " in 30"}},
		{`40 Y$ = MKI$(35) : CVI(Y$)`, &object.Integer{Value: 35}},
		{`50 Y$ = MKS$(65999) : CVI(Y$)`, &object.Integer{Value: 463}},
		{`60 CVI("..")`, &object.Integer{Value: 11822}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestCvs(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 CVS("ABCD", "EFGH")`, &object.Error{Message: syntaxErr + " in 10"}},
		{`30 CVS(123)`, &object.Error{Message: typeMismatchErr + " in 30"}},
		{`40 Y$ = MKS$(35) : CVS(Y$)`, &object.IntDbl{Value: 35}},
		{`50 CVS("..")`, &object.IntDbl{Value: 11822}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestExp(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 EXP(2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 EXP("ABCDEF")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 EXP(123)`, &object.Error{Message: overflowErr + " in 30"}},
		{`40 Y% = 35 : EXP(Y%)`, &object.FloatSgl{Value: 1586013445029888}},
		{`50 Y = 38999 : EXP(Y)`, &object.Error{Message: overflowErr + " in 50"}},
		{`60 Y = 3.8999 : EXP(Y)`, &object.FloatSgl{Value: 49.397511}},
		{`70 Y = 3.8999E+00 : EXP(Y)`, &object.FloatSgl{Value: 49.397507}},
		{`80 Y = 3.8999D+00 : EXP(Y)`, &object.FloatSgl{Value: 49.397511}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestFix(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 FIX(2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 FIX("ABCDEF")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 FIX(4294967296)`, &object.Error{Message: overflowErr + " in 30"}},
		{`40 Y% = 35 : FIX(Y%)`, &object.Integer{Value: 35}},
		{`50 Y = -35 : FIX(Y)`, &object.Integer{Value: -35}},
		{`60 Y = 3.8999 : FIX(Y)`, &object.Integer{Value: 3}},
		{`70 Y = 3.8999E+00 : FIX(Y)`, &object.Integer{Value: 3}},
		{`80 Y = 3.8999D+00 : FIX(Y)`, &object.Integer{Value: 3}},
		{`90 Y = 65999 : FIX(Y)`, &object.IntDbl{Value: 65999}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestHex(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 HEX$(2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 HEX$("ABCDEF")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 HEX$(65999)`, &object.Error{Message: overflowErr + " in 30"}},
		{`40 Y% = 35 : HEX$(Y%)`, &object.String{Value: "23"}},
		{`50 Y = -35 : HEX$(Y)`, &object.String{Value: "FFDD"}},
		{`60 Y = 3.8999 : HEX$(Y)`, &object.String{Value: "4"}},
		{`70 Y = 3.8999E+00 : HEX$(Y)`, &object.String{Value: "4"}},
		{`80 Y = 3.8999D+00 : HEX$(Y)`, &object.String{Value: "4"}},
		{`90 Y = 65500 : HEX$(Y)`, &object.String{Value: "FFDC"}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestInputStr(t *testing.T) {
	tests := []struct {
		inp  string
		keys string
		exp  interface{}
	}{
		{`10 INPUT$(1, 2)`, "", &object.Error{Message: syntaxErr + " in 10"}},
		{`20 INPUT$("fred")`, "", &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 INPUT$(0)`, "", &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{`40 INPUT$(2)`, "AB", &object.String{Value: "AB"}},
		{`50 INPUT$(2#)`, "AB", &object.String{Value: "AB"}},
		{`60 INPUT$(1.56)`, "AB", &object.String{Value: "AB"}},
		{`70 INPUT$(1.56E+00)`, "AB", &object.String{Value: "AB"}},
		{`80 INPUT$(1.56D+00)`, "AB", &object.String{Value: "AB"}},
		{`90 Y# = 2 : INPUT$(Y#)`, "AB", &object.String{Value: "AB"}},
	}
	for _, tt := range tests {
		evald := testEvalWithTerm(tt.inp, tt.keys)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestInstr(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 INSTR("Fred")`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 INSTR( 4, "fred", "George", "Sam")`, &object.Error{Message: syntaxErr + " in 20"}},
		{`30 INSTR("George", "org")`, &object.Integer{Value: 3}},
		{`40 INSTR(2, "George", "Ge")`, &object.Integer{Value: 0}},
		{`50 G# = 6 : INSTR(2, "George", G#)`, &object.Error{Message: syntaxErr + " in 50"}},
		{`60 INSTR(0, "George", "ge")`, &object.Error{Message: illegalArgErr + " in 60"}},
		{`70 INSTR(390, "George", "ge")`, &object.Error{Message: illegalFuncCallErr + " in 70"}},
		{`80 INSTR(10, "George", "ge")`, &object.Integer{Value: 0}},
		{`90 X = 2 : INSTR(X, "George", "ge")`, &object.Integer{Value: 5}},
		{`100 INSTR("George", "ff")`, &object.Integer{Value: 0}},
		{`90 X$ = "2" : INSTR(X$, "George", "ge")`, &object.Error{Message: illegalArgErr + " in 90"}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 INT(2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 INT("ABCDEF")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 INT(4294967296)`, &object.Error{Message: overflowErr + " in 30"}},
		{`40 Y% = 35 : INT(Y%)`, &object.Integer{Value: 35}},
		{`50 Y = -35 : INT(Y)`, &object.Integer{Value: -35}},
		{`60 Y = 3.8999 : INT(Y)`, &object.Integer{Value: 3}},
		{`70 Y = 3.8999E+00 : INT(Y)`, &object.Integer{Value: 3}},
		{`80 Y = 3.8999D+00 : INT(Y)`, &object.Integer{Value: 3}},
		{`90 Y = 65999 : INT(Y)`, &object.IntDbl{Value: 65999}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestLeft(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 LEFT$("Fred")`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 LEFT$( 3, "fred")`, &object.Error{Message: syntaxErr + " in 20"}},
		{`30 LEFT$("George", 3)`, &object.String{Value: "Geo"}},
		{`40 LEFT$("George", 0)`, &object.String{Value: ""}},
		{`50 LEFT$("George", 300)`, &object.Error{Message: illegalFuncCallErr + " in 50"}},
		{`60 X$ = MKS$(65999) : LEFT$( X$, 2)`, &object.BStr{Value: []byte{0xcf, 0x01}}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestLen(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`5 Y$ = "Test" : LEN(Y$)`, 4},
		{`10 LEN("")`, 0},
		{`20 LEN("four")`, 4},
		{`30 LEN("hello world")`, 11},
		{`40 LEN(1)`, &object.Error{Message: typeMismatchErr + " in 40"}},
		{`50 LEN("one", "two")`, &object.Error{Message: syntaxErr + " in 50"}},
		{`70 LEN("four" / "five")`, &object.Error{Message: "unknown operator: STRING / STRING in 70"}},
		{`80 LEN(MKI$(1279))`, &object.Integer{Value: 2}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestLog(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 LOG(2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 LOG("ABCDEF")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 LOG(-2)`, &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{`40 Y% = 35 : LOG(Y%)`, &object.FloatSgl{Value: 3.5553482}},
		{`50 Y = 38999 : LOG(Y)`, &object.FloatSgl{Value: 10.571291}},
		{`60 Y = 3.8999 : LOG(Y)`, &object.FloatSgl{Value: 1.360951}},
		{`70 Y = 3.8999E+00 : LOG(Y)`, &object.FloatSgl{Value: 1.360951}},
		{`80 Y = 3.8999D+00 : LOG(Y)`, &object.FloatSgl{Value: 1.360951}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestLPOS(t *testing.T) {
	// TODO: will need more once I implement printing
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 LPOS(2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 LPOS("A")`, &object.Integer{Value: 0}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestMID(t *testing.T) {
	// TODO: will need more once I implement printing
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 A$ = "Georgia" : MID$(A$)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 A$ = "Georgia" : MID$(A$,4)`, &object.String{Value: "rgia"}},
		{`30 A$ = "Georgia" : MID$(A$,4,2)`, &object.String{Value: "rg"}},
		{`40 A$ = "Georgia" : MID$(A$,4,300)`, &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{`50 A$ = "Georgia" : MID$(A$,"4",3)`, &object.Error{Message: syntaxErr + " in 50"}},
		{`60 A$ = MKD$(35456778) : MID$(A$,2,2)`, &object.BStr{Value: []byte{0x07, 0x1d}}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestMKD(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 MKD$("..")`, &object.Error{Message: typeMismatchErr + " in 10"}},
		{`20 MKD$(1, 2)`, &object.Error{Message: syntaxErr + " in 20"}},
		{`30 MKD$(35)`, &object.BStr{Value: []byte{0x23, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestMKI(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 MKI$("..")`, &object.Error{Message: typeMismatchErr + " in 10"}},
		{`20 MKI$(1, 2)`, &object.Error{Message: syntaxErr + " in 20"}},
		{`30 MKI$(35)`, &object.BStr{Value: []byte{0x23, 0x00}}},
		{`40 MKI$(-35)`, &object.BStr{Value: []byte{0xdd, 0xff}}},
		{`50 MKI$(35.3)`, &object.BStr{Value: []byte{0x23, 0x00}}},
		{`60 MKI$(-32769)`, &object.Error{Message: overflowErr + " in 60"}},
		{`70 MKI$(32768)`, &object.Error{Message: overflowErr + " in 70"}},
		{`80 MKI$(3.53E+01)`, &object.BStr{Value: []byte{0x23, 0x00}}},
		{`90 MKI$(3.53D+01)`, &object.BStr{Value: []byte{0x23, 0x00}}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestMKS(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 MKS$("..")`, &object.Error{Message: typeMismatchErr + " in 10"}},
		{`20 MKS$(1, 2)`, &object.Error{Message: syntaxErr + " in 20"}},
		{`30 MKS$(35)`, &object.BStr{Value: []byte{0x23, 0x00, 0x00, 0x00}}},
		{`40 MKS$(-35)`, &object.BStr{Value: []byte{0xdd, 0xff, 0xff, 0xff}}},
		{`50 MKS$(35.3)`, &object.BStr{Value: []byte{0x23, 0x00, 0x00, 0x00}}},
		{`60 MKS$(3.53E+01)`, &object.BStr{Value: []byte{0x23, 0x00, 0x00, 0x00}}},
		{`70 MKS$(3.53D+01)`, &object.BStr{Value: []byte{0x23, 0x00, 0x00, 0x00}}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestOct(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 OCT$(2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 OCT$("ABCDEF")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 OCT$(65999)`, &object.Error{Message: overflowErr + " in 30"}},
		{`40 Y% = 35 : OCT$(Y%)`, &object.String{Value: "43"}},
		{`50 Y = -35 : OCT$(Y)`, &object.String{Value: "177735"}},
		{`60 Y = 3.8999 : OCT$(Y)`, &object.String{Value: "4"}},
		{`70 Y = 3.8999E+00 : OCT$(Y)`, &object.String{Value: "4"}},
		{`80 Y = 3.8999D+00 : OCT$(Y)`, &object.String{Value: "4"}},
		{`90 Y = 65500 : OCT$(Y)`, &object.String{Value: "177734"}},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestRight(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 RIGHT$("Fred")`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 RIGHT$( 3, "fred")`, &object.Error{Message: syntaxErr + " in 20"}},
		{`30 RIGHT$("George", 3)`, &object.String{Value: "rge"}},
		{`40 RIGHT$("George", 0)`, &object.String{Value: ""}},
		{`50 RIGHT$("George", 300)`, &object.Error{Message: illegalFuncCallErr + " in 50"}},
		{`60 X$ = MKS$(65999) : RIGHT$( X$, 2)`, &object.BStr{Value: []byte{0x01, 0x00}}},
		{`70 RIGHT$("George", 10)`, &object.String{Value: "George"}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestRnd(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 RND(5, 5)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 RND("fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 RND(0)`, &object.FloatSgl{Value: 0.615608156}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestScreen(t *testing.T) {
	tests := []struct {
		inp string
		msg string
		exp interface{}
	}{
		{`10 SCREEN(5)`, "", &object.Error{Message: illegalFuncCallErr + " in 10"}},
		{`20 SCREEN(5, "fred")`, "", &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 SCREEN(1,1)`, "470", &object.Integer{Value: 52}},
		{`40 SCREEN(1,2)`, "470", &object.Integer{Value: 55}},
	}
	for _, tt := range tests {
		evald := testEvalWithTerm(tt.inp, tt.msg)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestSgn(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 SGN(5, 2)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 SGN("fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 SGN(5)`, &object.Integer{Value: 1}},
		{`40 SGN(0)`, &object.Integer{Value: 0}},
		{`50 SGN(-2)`, &object.Integer{Value: -1}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestSin(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 SIN(5, 2)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 SIN("fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 SIN(5)`, &object.FloatSgl{Value: -0.9589243}},
		{`40 SIN(0.5)`, &object.FloatSgl{Value: 0.479425550}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestSpaces(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 SPACE$(5, 2)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 SPACE$("fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 SPACE$(-1)`, &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{`40 SPACE$(256)`, &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{`50 SPACE$(2)`, &object.String{Value: "  "}},
		{`60 SPACE$(0)`, &object.String{Value: ""}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestSqr(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 SQR(5, 2)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 SQR("fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 SQR(25)`, &object.FloatSgl{Value: 5}},
		{`40 SQR(0.5)`, &object.FloatSgl{Value: 0.707106769}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestStrs(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 STR$(5, 2)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 STR$("fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 STR$(-1)`, &object.String{Value: "-1"}},
		{`40 STR$(256)`, &object.String{Value: "256"}},
		{`50 STR$(2)`, &object.String{Value: "2"}},
		{`60 STR$(1.5)`, &object.String{Value: "1.5"}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 STRING$(5, 2, 3)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 STRING$("fred", 52)`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 STRING$(-1, 52)`, &object.Error{Message: illegalFuncCallErr + " in 30"}},
		{`40 STRING$(3, 256)`, &object.Error{Message: illegalFuncCallErr + " in 40"}},
		{`50 STRING$(2, 50)`, &object.String{Value: "22"}},
		{`60 STRING$(3, "34")`, &object.String{Value: "333"}},
		{`70 STRING$(3, 0)`, &object.BStr{Value: []byte{0x00, 0x00, 0x00}}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestTan(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 TAN(5, 2)`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 TAN("fred")`, &object.Error{Message: typeMismatchErr + " in 20"}},
		{`30 TAN(5)`, &object.FloatSgl{Value: -3.380515099}},
		{`40 TAN(0.5)`, &object.FloatSgl{Value: 0.546302497}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}

func TestVal(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 VAL("5", "2")`, &object.Error{Message: syntaxErr + " in 10"}},
		{`20 VAL("fred")`, &object.FloatSgl{Value: 0}},
		{`30 VAL("5")`, &object.FloatSgl{Value: 5}},
		{`40 VAL("0.5")`, &object.FloatSgl{Value: 0.5}},
		{`50 VAL(5)`, &object.Error{Message: typeMismatchErr + " in 50"}},
	}
	for _, tt := range tests {
		evald := testEval(tt.inp)

		compareObjects(tt.inp, evald, tt.exp, t)
	}
}
