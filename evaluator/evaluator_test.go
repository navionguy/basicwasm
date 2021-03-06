package evaluator

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"
	"github.com/stretchr/testify/assert"

	"testing"
)

type mockTerm struct {
	row     *int
	col     *int
	strVal  *string
	sawStr  *string
	sawCls  *bool
	sawBeep *bool
}

func initMockTerm(mt *mockTerm) {
	mt.row = new(int)
	*mt.row = 0

	mt.col = new(int)
	*mt.col = 0

	mt.strVal = new(string)
	*mt.strVal = ""

	mt.sawCls = new(bool)
	*mt.sawCls = false
}

func (mt mockTerm) Cls() {
	*mt.sawCls = true
}

func (mt mockTerm) Print(msg string) {
	fmt.Print(msg)
}

func (mt mockTerm) Println(msg string) {
	fmt.Println(msg)
	if mt.sawStr != nil {
		*mt.sawStr = *mt.sawStr + msg
	}
}

func (mt mockTerm) SoundBell() {
	fmt.Print("\x07")
	*mt.sawBeep = true
}

func (mt mockTerm) Locate(int, int) {
}

func (mt mockTerm) GetCursor() (int, int) {
	return *mt.row, *mt.col
}

func (mt mockTerm) Read(col, row, len int) string {
	// make sure your test is correct
	trim := (row-1)*80 + (col - 1)

	tstr := *mt.strVal

	newstr := tstr[trim : trim+len]

	return newstr
}

func (mt mockTerm) ReadKeys(count int) []byte {
	if mt.strVal == nil {
		return nil
	}

	bt := []byte(*mt.strVal)

	if count >= len(bt) {
		mt.strVal = nil
		return bt
	}

	v := (*mt.strVal)[:count]
	mt.strVal = &v

	return bt[:count]
}

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

func TestAutoCommand(t *testing.T) {
	tests := []struct {
		inp  string
		strt int
		step int
		curr bool
	}{
		{inp: "AUTO", strt: 10, step: 10, curr: false},
		{inp: "AUTO 500", strt: 500, step: 10, curr: false},
		{inp: "AUTO 500, 50", strt: 500, step: 50, curr: false},
		{inp: "AUTO ., 50", strt: 10, step: 50, curr: true},
	}

	for _, tt := range tests {

		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(env.Program, env.Program.CmdLineIter(), env)

		aut := env.GetAuto()

		if aut == nil {
			t.Fatalf("TestAutoCommand environment doesn't have an auto struct")
		}

		if tt.strt != aut.Start {
			t.Fatalf("TestAutoCommand expected start = %d, got %d", tt.strt, aut.Start)
		}

		if tt.step != aut.Increment {
			t.Fatalf("TestAutoCommand expected step = %d, got %d", tt.step, aut.Increment)
		}

		if tt.curr != aut.Curr {
			t.Fatalf("TestAutoCommand expected curr = %t, got %t", tt.curr, aut.Curr)
		}
	}
}

func Test_BeepStatement(t *testing.T) {
	l := lexer.New("BEEP")
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	chk := false
	mt.sawBeep = &chk
	env := object.NewTermEnvironment(mt)

	p.ParseCmd(env)

	if len(p.Errors()) > 0 {
		for _, er := range p.Errors() {
			fmt.Println(er)
		}
		return
	}

	Eval(env.Program, env.Program.CmdLineIter(), env)

	assert.True(t, chk, "Test_BeepStatement term.beep() not called!")
}

func Test_ChainStatement(t *testing.T) {
	tests := []struct {
		stmt string
		rs   int // response code  eg '200'
		send string
	}{
		{stmt: `CHAIN "start.bas"`, rs: 404, send: ``},
		{stmt: `CHAIN "start.bas"`, rs: 200, send: `10 PRINT "Hello World!"`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.stmt)
		p := parser.New(l)
		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(tt.rs)
			res.Write([]byte(tt.send))
		}))
		defer ts.Close()
		url := object.String{Value: ts.URL}
		/*		if len(tt.url) > 0 {
				url = object.String{Value: tt.url}
			}*/
		env.Set(object.SERVER_URL, &url)

		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		Eval(env.Program, env.Program.CmdLineIter(), env)
	}
}

func TestClsStatement(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"Cls"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		Eval(env.Program, env.Program.CmdLineIter(), env)

		if !*mt.sawCls {
			t.Errorf("No call to Cls() seen")
		}
	}
}

func Test_FilesCommand(t *testing.T) {
	tests := []struct {
		param string
		cwd   string
		send  string
		exp   string
		rs    int
		err   bool
	}{
		{param: "", cwd: `C:\`, send: "10 PRINT \"Main Menu\"\n", exp: "10 PRINT \"Main Menu\"\n", rs: 404, err: false},
		{param: "", cwd: `C:\`, send: "10 PRINT \"Main Menu\"\n", exp: "10 PRINT \"Main Menu\"\n", rs: 200, err: false},
		{param: "", cwd: `C:\`, send: `[{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":true}]`, exp: `[{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":true}]`, rs: 200, err: false},
	}

	for _, tt := range tests {
		cmd := "FILES"

		if len(tt.param) > 0 {
			cmd = cmd + tt.param
		}

		l := lexer.New(cmd)
		p := parser.New(l)
		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(tt.rs)
			res.Write([]byte(tt.send))
		}))
		defer ts.Close()

		url := object.String{Value: ts.URL}
		env.Set(object.SERVER_URL, &url)

		if len(tt.cwd) > 0 {
			drv := object.String{Value: tt.cwd}
			env.Set(object.WORK_DRIVE, &drv)
		}

		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			t.Fatal("FILES command failed!")
			return
		}

		Eval(env.Program, env.Program.CmdLineIter(), env)

	}
}

func Test_CatchNotDir(t *testing.T) {
	tests := []struct {
		path string
		send string
		exp  string
	}{
		{path: "file.ext", send: "NotDir", exp: `C:\file.ext`},
		{path: "file.ext", send: "File not found", exp: `File not found`},
	}

	for _, tt := range tests {
		var mt mockTerm
		initMockTerm(&mt)
		var rec string
		mt.sawStr = &rec
		env := object.NewTermEnvironment(mt)
		env.Set(object.WORK_DRIVE, &object.String{Value: `C:\`})

		catchNotDir(tt.path, errors.New(tt.send), env)
		assert.Equal(t, tt.exp, rec, "Test_CatchNotDir got unexpected return")
	}
}

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int16
	}{
		{"10 -5", -5},
		{"20 -10", -10},
		{"30 5", 5},
		{"40 10", 10},
		{"50 5 + 5", 10},
		{"60 5 + 5 + 5 + 5 -10", 10},
		{"70 5 < 10", 1},
		{"80 5 > 10", 0},
		{"110 10 > 1", 1},
		{"120 10 < 1", 0},
		{"130 10 / 2", 5},
		{"160 10 \\ 2", 5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestDblInetegerExpression(t *testing.T) {
	tests := []struct {
		inp string
		exp int32
	}{
		{"10 99999", 99999},
		{"20 -99999", -99999},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)
		testIntDblObject(t, evald, tt.exp)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	program := env.Program

	if len(p.Errors()) > 0 {
		return nil
	}

	return Eval(program, program.StatementIter(), env)
}

func testEvalWithTerm(input string, keys string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	mt.strVal = &keys
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	program := env.Program

	if len(p.Errors()) > 0 {
		return nil
	}

	return Eval(program, program.StatementIter(), env)
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

func testIntDblObject(t *testing.T, obj object.Object, expected int32) bool {
	result, ok := obj.(*object.IntDbl)
	if !ok {
		t.Errorf("object is not IntDbl. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int16
	}{
		{"10 IF 5 < 6 THEN 30\n20 5\n30 6", 6},
		{"10 IF 5 < 6 GOTO 30\n20 5\n30 6", 6},
		{"10 IF 5 < 6 THEN END\n20 5", 0},
		{"10 IF 5 > 6 THEN 20 ELSE END\n20 5", 0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEndStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int16
	}{
		{"10 END\n20 5\n30 6", 0},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int16
	}{
		{"10 LET a = 5: a", 5},
		{"20 LET a = 5 * 5: a", 25},
		{"30 LET a = 5: let b = a: b", 5},
		{"40 LET a = 5: let b = a: let c = a + b + 5: c", 15},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestDimStatements(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`10 DIM A[20] : A[11] = 6 : PRINT A[11]`},
		{`20 DIM B[10,10]`},
		{`30 DIM A[9,10], B[14,15] : B[5,6] = 12 : PRINT B[5,6]`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}

	// want
	// 4
}

func TestStringLiteral(t *testing.T) {
	input := `10 "Hello World!"`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `10 "Hello" + " " + "World!"`
	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}
func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"10 foobar",
			"Syntax error in 10",
		},
		{
			`10 "Hello" - "World"`,
			"unknown operator: STRING - STRING in 10",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}
func TestFunctionObject(t *testing.T) {
	input := "10 DEF FNSKIP(x)= (x + 2)"

	evaluated := testEval(input)

	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v", fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(X + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestInvalidFunctionName(t *testing.T) {
	tests := []struct {
		input    string
		expError string
	}{
		{"10 DEF ID(x)", "function names must be in the form FNname"},
		{"20 DEF NFID(x)", "function names must be in the form FNname"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)

		if len(p.Errors()) != 1 {
			t.Errorf("expected 1 error, got %d", len(p.Errors()))
		}

		if tt.expError != p.Errors()[0] {
			t.Errorf("expected error %s, got %s", tt.expError, p.Errors()[0])
		}
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int16
	}{
		{"10 DEF FNID(x) = x : FNID(5)", 5},
		{"20 DEF FNMUL(x,y) = x*y : FNMUL(2,3)", 6},
		{"30 DEF FNSKIP(x)= (x + 2): FNSKIP(3)", 5},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestHexOctalConstants(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 &H7F`, int16(127)},
		{`20 &HG7F`, "Syntax error in 20"},
		{`30 &H7FFFFF`, "Overflow in 30"},
		{`40 &O7`, int16(7)},
		{`50 &O77`, int16(63)},
		{`60 &O77777`, int16(32767)},
		{`70 &O777777`, "Overflow in 70"},
		{`80 &77777`, int16(32767)},
		{`90 &O78777`, "Syntax error in 90"},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp)
		switch expected := tt.exp.(type) {
		case int16:
			testIntegerObject(t, evald, expected)
		case string:
			errObj, ok := evald.(*object.Error)
			if !ok {
				t.Errorf("unexepected result, go %t (%+v)", evald, evald)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message!  expected %q, got %q", expected, errObj.Message)
			}
		}
	}
}

func TestReadStatement(t *testing.T) {
	fixedInt, _ := decimal.NewFromString("999.99")

	tests := []struct {
		inp string
		exp object.Object
	}{
		{`10 DATA "Fred", "George" : READ A$`, &object.String{Value: "Fred"}},
		{`20 DATA 123 : READ A`, &object.Integer{Value: 123}},
		{`30 DATA 99999 : READ A`, &object.IntDbl{Value: 99999}},
		{`40 DATA 999.99 : READ A`, &object.Fixed{Value: fixedInt}},
		{`50 DATA 2.35123412341234E+4 : READ A`, &object.FloatSgl{Value: 23512.341796875}},
		{`60 DATA 2.35123412341234D+4 : READ A`, &object.FloatDbl{Value: 23512.3412341234}},
		{`70 DATA -2.35123412341234D+4 : READ A`, &object.FloatDbl{Value: -23512.3412341234}},
		{`80 DATA "Fred" : READ A$ : READ B$`, &object.Error{Message: "Out of data in 80"}},
	}

	for _, tt := range tests {
		res := testEval(tt.inp)

		compareObjects(tt.inp, res, tt.exp, t)
	}
}

func TestRestoreStatement(t *testing.T) {

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 DATA "Fred", "George" : RESTORE`, nil},
		{`20 DATA "Fred", "George" : RESTORE 20`, nil},
		{`30 DATA "Fred", "George" : RESTORE 5`, &object.Error{Message: "Undefined line number in 30"}},
	}

	for _, tt := range tests {
		res := testEval(tt.inp)

		if (res != nil) || (tt.exp != nil) {
			compareObjects(tt.inp, res, tt.exp, t)
		}
	}
}

func TestTronTroffCommands(t *testing.T) {
	tests := []struct {
		inp string
		trc bool
	}{
		{"TRON", true},
		{"TRON : TROFF", false},
	}

	for _, tt := range tests {
		var mt mockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		Eval(env.Program, env.Program.CmdLineIter(), env)

		if env.GetTrace() != tt.trc {
			t.Errorf("TestTronTroffCommands trace expected %t, got %t", tt.trc, env.GetTrace())
		}
	}
}

func ExamplePrint() {
	tests := []struct {
		input string
	}{
		{`5 IF 5 = 5 THEN PRINT "Same"`},
		{`10 IF "A" = "A" THEN PRINT "Same"`},
		{`10 PRINT "Hello World!`},
		{`20 PRINT "This is ";"a test"`},
		{`30 PRINT "Another test " "program."`},
		{`40 PRINT "Test of tab","due to comma"`},
		{`50 PRINT "Test of a run on";`},
		{`60 PRINT " sentence"`},
		{`70 LET X = 45.12 : PRINT X`},
		{`80 LET Y = 45.12 + 12 : PRINT Y`},
		{`90 LET Y = 2 * 45.12 : PRINT Y`},
		{`90 LET Y = 45.12 / 2 : PRINT Y`},
		{`100 LET Y = 45.12 < 53.6 : PRINT Y`},
		{`110 LET Y = 45.12 - 12.6 : PRINT Y`},
		{`120 LET Y = 45.12 < 23.6 : PRINT Y`},
		{`130 LET Y = 45.12 <= 53.6 : PRINT Y`},
		{`140 LET Y = 45.12 <= 23.6 : PRINT Y`},
		{`150 LET Y = 45.12 > 53.6 : PRINT Y`},
		{`160 LET Y = 45.12 > 23.6 : PRINT Y`},
		{`170 LET Y = 45.12 >= 53.6 : PRINT Y`},
		{`180 LET Y = 45.12 >= 23.6 : PRINT Y`},
		{`190 LET Y = 45.12 <> 53.6 : PRINT Y`},
		{`200 LET Y = 45.12 <> 45.12 : PRINT Y`},
		{`210 LET Y = 45.12 * 3.4 : PRINT Y`},
		{`220 LET Y = 45.12 / 3.4 : PRINT Y`},
		{`230 LET Y = 235.988E+2 + 1.354E+1 : PRINT Y`},
		{`240 X = 5 : Y = 3.2 : PRINT X * Y`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}
	// Output:
	// Same
	// Same
	// Hello World!
	// This is a test
	// Another test program.
	// Test of tab	due to comma
	// Test of a run on sentence
	// 45.12
	// 57.12
	// 90.24
	// 22.56
	// 1
	// 32.52
	// 0
	// 1
	// 0
	// 0
	// 1
	// 0
	// 1
	// 1
	// 0
	// 153.408
	// 13.27059
	// 2.361234E+04
	// 16
}

func ExampleT_int() {
	tests := []struct {
		input string
	}{
		{`10 LET X = 32760 + 300 : PRINT X`},
		{`20 LET Y = 32767 / 3 : PRINT Y`},
		{`30 LET Y = 11 MOD 3 : PRINT Y`},
		{`40 LET Y = 10 <> 10 : PRINT Y`},
		{`50 LET Y = 10 <> 3 : PRINT Y`},
		{`60 LET Y = 10 = 10 : PRINT Y`},
		{`70 LET Y = 10 = 3 : PRINT Y`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}
	// Output:
	// 33060
	// 1.092233E+04
	// 2
	// 0
	// 1
	// 1
	// 0
}

func ExampleT_fixed() {
	tests := []struct {
		input string
	}{
		{`10 LET X = 45.12 : PRINT X`},
		{`20 LET Y = 45.12 + 12 : PRINT Y`},
		{`30 LET Y = 2 * 45.12 : PRINT Y`},
		{`40 LET Y = 45.12 / 2 : PRINT Y`},
		{`50 LET Y = 45.12 < 53.6 : PRINT Y`},
		{`60 LET Y = 45.12 - 12.6 : PRINT Y`},
		{`70 LET Y = 45.12 < 23.6 : PRINT Y`},
		{`80 LET Y = 45.12 <= 53.6 : PRINT Y`},
		{`90 LET Y = 45.12 <= 23.6 : PRINT Y`},
		{`100 LET Y = 45.12 > 53.6 : PRINT Y`},
		{`110 LET Y = 45.12 > 23.6 : PRINT Y`},
		{`120 LET Y = 45.12 >= 53.6 : PRINT Y`},
		{`130 LET Y = 45.12 >= 23.6 : PRINT Y`},
		{`140 LET Y = 45.12 <> 53.6 : PRINT Y`},
		{`150 LET Y = 45.12 <> 45.12 : PRINT Y`},
		{`160 LET Y = 45.12 * 3.4 : PRINT Y`},
		{`170 LET Y = 45.12 / 3.4 : PRINT Y`},
		{`180 LET Y = 235.988E+2 + 1.354E+1 : PRINT Y`},
		{`190 LET Y = 235.988E+2 = 235.988E+2 : PRINT Y`},
		{`200 LET Y = 235.988E+2 = 1.354E+1 : PRINT Y`},
		{`210 LET Y = 45.12 = 45.12 : PRINT Y`},
		{`220 LET Y = 45.12 = 12 : PRINT Y`},
		{`230 LET Y = 45 >= 12 : PRINT Y`},
		{`240 LET Y = 45 <= 12 : PRINT Y`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}
	// Output:
	// 45.12
	// 57.12
	// 90.24
	// 22.56
	// 1
	// 32.52
	// 0
	// 1
	// 0
	// 0
	// 1
	// 0
	// 1
	// 1
	// 0
	// 153.408
	// 13.27059
	// 2.361234E+04
	// 1
	// 0
	// 1
	// 0
	// 1
	// 0
}

func ExampleT_float() {
	tests := []struct {
		input string
	}{
		{`10 LET Y = 235.988E+2 + 1.354E+1 : PRINT Y`},
		{`20 LET Y = 2.35E+4 + 3.14: PRINT Y`},
		{`30 LET Y = 2.35E+4 + 3: PRINT Y`},
		{`40 LET Y = 2.35E+4 - 3: PRINT Y`},
		{`50 LET Y = 3 * 2.35E+4: PRINT Y`},
		{`60 LET Y = 45123.62 / 2.35E+4: PRINT Y`},
		{`70 LET Y = 2.35E+4 < 53.6 : PRINT Y`},
		{`80 LET Y = 2.35E+4 < 23.6 : PRINT Y`},
		{`90 LET Y = 2.35E+4 <= 53.6 : PRINT Y`},
		{`100 LET Y = 2.35E+4 <= 23.6 : PRINT Y`},
		{`110 LET Y = 2.35E+4 > 53.6 : PRINT Y`},
		{`120 LET Y = 2.35E+4 > 23.6 : PRINT Y`},
		{`130 LET Y = 2.35E+4 >= 53.6 : PRINT Y`},
		{`140 LET Y = 2.35E+4 >= 23.6 : PRINT Y`},
		{`150 LET Y = 2.35E+4 <> 53.6 : PRINT Y`},
		{`160 LET Y = 2.35E+4 <> 45.12 : PRINT Y`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}
	// Output:
	// 2.361234E+04
	// 2.350314E+04
	// 23503
	// 23497
	// 70500
	// 1.920154E+00
	// 0
	// 0
	// 0
	// 0
	// 1
	// 1
	// 1
	// 1
	// 1
	// 1
}

func ExampleT_floatDbl() {
	tests := []struct {
		input string
	}{
		{`10 LET Y = 235.988D+12 + 1.354D+4 : PRINT Y`},
		{`20 LET Y = -2.35D+4 + 314: PRINT Y`},
		{`30 LET Y = 2.35D+4 + 3.14159: PRINT Y`},
		{`40 LET Y = 2.35D+4 - 3.1415E+3: PRINT Y`},
		{`50 LET Y = 3 * 2.35D+4: PRINT Y`},
		{`60 LET Y = 123.45 / 2.35D+4: PRINT Y`},
		{`70 LET Y = 2.35E+4 < 4.56D+4 : PRINT Y`},
		{`80 LET Y = 2.35D+4 < 23.6 : PRINT Y`},
		{`90 LET Y = 2.35D+4 <= 53.6 : PRINT Y`},
		{`100 LET Y = 2.35D+4 <= 23.6 : PRINT Y`},
		{`110 LET Y = 2.35D+4 > 53.6 : PRINT Y`},
		{`120 LET Y = 2.35D+4 > 23.6 : PRINT Y`},
		{`130 LET Y = 2.35D+4 >= 53.6 : PRINT Y`},
		{`140 LET Y = 2.35D+4 >= 23.6 : PRINT Y`},
		{`150 LET Y = 2.35D+4 <> 53.6 : PRINT Y`},
		{`160 LET Y = 2.35D+4 <> 45.12 : PRINT Y`},
		{`170 LET Y = 2.35D+4 = 2.35D+4 : PRINT Y`},
		{`180 LET Y = 2.35D+4 = 2.35 : PRINT Y`},
		{`190 LET X = -2.35123412341234D+4 : PRINT X`},
		{`200 LET X = -2.35123412341234E+4 : PRINT X`},
		{`210 LET X = -2.351 : PRINT X`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}

	// Output:
	// 2.359880E+14
	// -23186
	// 2.350314E+04
	// 2.035850E+04
	// 70500
	// 5.253191E-03
	// 1
	// 0
	// 0
	// 0
	// 1
	// 1
	// 1
	// 1
	// 1
	// 1
	// 1
	// 0
	// -2.351234E+04
	// -2.351234E+04
	// -2.351
}

func ExampleT_array() {
	tests := []struct {
		input string
	}{
		{`10 LET Y[0] = 5 : PRINT Y(0)`},
		{`15 LET Y[0] = 4 : PRINT Y[5]`},
		{`20 LET Y(0) = 5 : LET Y[1] = 1: PRINT Y[0]`},
		{`30 LET Y[0] = 5 : LET Y[1] = 1: PRINT Y[1]`},
		{`40 LET Y$[0] = "Hello" : PRINT Y$[0]`},
		{`50 LET Y$[0] = "Hello" : Y$[0] = "Goodbye" : PRINT Y$[0]`},
		{`60 LET Y$[0] = "Hello" : PRINT Y$[5]`},
		{`70 LET Y$ = "HELLO" : PRINT Y$[0]`},
		{`80 LET Y# = 5 : PRINT Y#`},
		{`90 LET Y#[0] = 5 : PRINT Y#[0]`},
		{`100 LET Y#[0] = 5 : PRINT Y#[1]`},
		{`110 LET Y%[0] = 5 : LET Y%[1] = 3 : PRINT Y%[0]`},
		{`120 LET Y![0] = 5 : LET Y![1] = 3 : PRINT Y![0]`},
		{`130 DIM A[20] : LET A[11] = 6 : PRINT A[11]`},
		{`140 DIM M[10,10] : LET M[4,5] = 13 : PRINT M[4,5] : PRINT M[5,4]`},
		{`150 DIM A[9,10], B[5,6] : LET B[4,5] = 12 : PRINT B[4,5]`},
		{`160 DIM Y[12.5] : LET Y[1.5] = 5 : PRINT Y[1.5]`},
		{`170 LET Y[4] = 31 : PRINT Y[3.6E+00]`},
		{`170 LET Y[4] = 31 : PRINT Y[3.6D+00]`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}

	// Output:
	// 5
	// 0
	// 5
	// 1
	// Hello
	// Goodbye
	//
	// ERROR: Syntax error in 70
	// 5
	// 5
	// 0.000000E+00
	// 5
	// 5
	// 6
	// 13
	// 0
	// 12
	// 5
	// 31
	// 31
}

func ExampleT_strings() {
	tests := []struct {
		input string
	}{
		{`10 LET Y$ = "Hello" : PRINT Y$`},
		{`20 LET Y$ = "Hello" : Y$ = "Goodbye" : PRINT Y$`},
		{`10 LET Y$ = "Hello" + " Goodbye" : PRINT Y$`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}

	// Output:
	// Hello
	// Goodbye
	// Hello Goodbye
}

func ExampleT_errors() {
	tests := []struct {
		input string
	}{
		{`5 REM A comment to get started.`},
		{`10 GOTO 200`},
		{`20 LET X = FNBANG(32)`},
		{`30 LET Y = 1.5 : LET X[Y] = 5 : PRINT X[Y]`},
		{`40 LET Y[11] = 5`},
		{`50 LET Y[0] = 5 : LET Y[11] = 4`},
		{`60 LET Y% = 5 : LET Y% = 3.5`},
		{`70 LET A$ = -"A negative msg"`},
		{`80 LET A = 5 + HELLO`},
	}

	for _, tt := range tests {
		testEval(tt.input)
	}

	// Output:
	// Undefined line number in 10
	// Undefined user function in 20
	// 5
	// index out of range in 40
	// Subscript out of range in 50
	// type mis-match in 60
	// unsupport negative on STRING in 70
	// type mis-match in 80
}

func ExampleT_list() {
	src := `
	10 rem This is a test program
	20 print "Hello World!"
	30 PRINT "And Goodbye Cruel World." : REM A trailing comment
	40 REM The end of the test program
	50 PRINT A$
	60 END`

	tests := []struct {
		inp string
		res string
	}{
		{inp: "LIST"},
	}

	l := lexer.New(src)
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	prog := env.Program

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(env.Program, prog.CmdLineIter(), env)
	}

	// Output:
	// 10 REM This is a test program
	// 20 PRINT "Hello World!"
	// 30 PRINT "And Goodbye Cruel World." : REM A trailing comment
	// 40 REM The end of the test program
	// 50 PRINT A$
	// 60 END
}

func ExampleT_list2() {
	src := `
	10 rem This is a test program
	20 print "Hello World!"
	30 PRINT "And Goodbye Cruel World." : REM A trailing comment
	40 REM The end of the test program
	50 PRINT A$
	60 END`

	tests := []struct {
		inp string
		res string
	}{
		{inp: "LIST 20-"},
	}

	l := lexer.New(src)
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	prog := env.Program

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(env.Program, prog.CmdLineIter(), env)
	}

	// Output:
	// 20 PRINT "Hello World!"
	// 30 PRINT "And Goodbye Cruel World." : REM A trailing comment
	// 40 REM The end of the test program
	// 50 PRINT A$
	// 60 END
}

func ExampleT_list3() {
	src := `
	10 rem This is a test program
	20 print "Hello World!"
	30 PRINT "And Goodbye Cruel World." : REM A trailing comment
	40 REM The end of the test program
	50 PRINT A$
	60 END`

	tests := []struct {
		inp string
		res string
	}{
		{inp: "LIST 20"},
	}

	l := lexer.New(src)
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	prog := env.Program

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(env.Program, prog.CmdLineIter(), env)
	}

	// Output:
	// 20 PRINT "Hello World!"
}

func ExampleT_list4() {
	src := `
	10 rem This is a test program
	20 print "Hello World!"
	30 PRINT "And Goodbye Cruel World." : REM A trailing comment
	40 REM The end of the test program
	50 PRINT A$
	60 END`

	tests := []struct {
		inp string
	}{
		{inp: "LIST -30"},
	}

	l := lexer.New(src)
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	prog := env.Program

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(env.Program, prog.CmdLineIter(), env)
	}

	// Output:
	// 10 REM This is a test program
	// 20 PRINT "Hello World!"
	// 30 PRINT "And Goodbye Cruel World." : REM A trailing comment
}

func ExampleT_Run() {
	src := `
	10 REM This is a test program
	20 PRINT "Hello World!"`

	tests := []struct {
		inp string
	}{
		{"RUN"},
	}

	l := lexer.New(src)
	p := parser.New(l)
	var mt mockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	prog := env.Program

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(env.Program, prog.CmdLineIter(), env)
	}

	// Output:
	// Hello World!
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`10 LEN("")`, 0},
		{`20 LEN("four")`, 4},
		{`30 LEN("hello world")`, 11},
		{`40 LEN(1)`, "Type mismatch in 40"},
		{`50 LEN("one", "two")`, "Syntax error in 50"},
		{`70 LEN("four" / "five")`, &object.Error{}},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int16(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}
