package evaluator

import (
	"fmt"

	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"

	"testing"
)

type mockTerm struct {
	row    *int
	col    *int
	strVal *string
	sawCls *bool
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
}

func (mt mockTerm) Locate(int, int) {

}

func (mt mockTerm) GetCursor() (int, int) {
	return *mt.row, *mt.col
}

func (mt mockTerm) Read(col, row, len int) string {
	return *mt.strVal
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
		cmd := env.Program.CmdLineIter().Value()

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		Eval(cmd, env.Program.CmdLineIter(), env)

		if !*mt.sawCls {
			t.Errorf("No call to Cls() seen")
		}
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
			"identifier not found: FOOBAR at 10",
		},
		{
			`10 "Hello" - "World"`,
			"unknown operator: STRING - STRING at 10",
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
	// 3.306000E+04
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
	// 2.350300E+04
	// 2.349700E+04
	// 7.050000E+04
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
	// 7.050000E+04
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
		{`10 LET Y[0] = 5 : PRINT Y[0]`},
		{`15 LET Y[0] = 4 : PRINT Y[5]`},
		{`20 LET Y[0] = 5 : LET Y[1] = 1: PRINT Y[0]`},
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
	// ERROR: identifier not found: Y$[] at 70
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
	// Undefined line number at 10
	// identifier not found: FNBANG at 20
	// 5
	// index out of range at 40
	// Subscript out of range at 50
	// type mis-match at 60
	// unsupport negative on STRING at 70
	// type mis-match at 80
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
		cmd := env.Program.CmdLineIter().Value()

		Eval(cmd, prog.CmdLineIter(), env)
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
		cmd := env.Program.CmdLineIter().Value()

		Eval(cmd, prog.CmdLineIter(), env)
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
		cmd := env.Program.CmdLineIter().Value()

		Eval(cmd, prog.CmdLineIter(), env)
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
		res string
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
		cmd := env.Program.CmdLineIter().Value()

		Eval(cmd, prog.CmdLineIter(), env)
	}

	// Output:
	// 10 REM This is a test program
	// 20 PRINT "Hello World!"
	// 30 PRINT "And Goodbye Cruel World." : REM A trailing comment
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`10 LEN("")`, 0},
		{`20 LEN("four")`, 4},
		{`30 LEN("hello world")`, 11},
		{`40 LEN(1)`, "argument to `len` not supported, got INTEGER at 40"},
		{`50 LEN("one", "two")`, "wrong number of arguments. got=2, want=1 at 50"},
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
