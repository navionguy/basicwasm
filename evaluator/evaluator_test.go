package evaluator

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/fileserv"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"
	"github.com/navionguy/basicwasm/settings"
	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"

	"testing"
)

func initMockTerm(mt *mocks.MockTerm) {
	mt.Row = new(int)
	*mt.Row = 0

	mt.Col = new(int)
	*mt.Col = 0

	mt.StrVal = new(string)
	*mt.StrVal = ""

	mt.SawCls = new(bool)
	*mt.SawCls = false

	mt.ExpMsg = &mocks.Expector{}
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

// TODO FULLY test applyFunction()
func TestApplyFuncion(t *testing.T) {
	fn := &object.Integer{}
	args := []object.Object{}
	var mt mocks.MockTerm
	env := object.NewTermEnvironment(mt)

	rc := applyFunction(fn, args, nil, env)

	_, ok := rc.(*object.Error)

	assert.Truef(t, ok, "failed to get error, instead got object %T", rc)
}

func TestAutoCommand(t *testing.T) {
	tests := []struct {
		inp  string
		strt ast.DblIntegerLiteral
		step ast.DblIntegerLiteral
		err  bool
	}{
		//{inp: "AUTO 10,10,10", err: true},
		{inp: "AUTO", strt: ast.DblIntegerLiteral{Value: 10}, step: ast.DblIntegerLiteral{Value: 10}},
		{inp: "AUTO 500", strt: ast.DblIntegerLiteral{Value: 500}, step: ast.DblIntegerLiteral{Value: 10}},
		{inp: "AUTO 500, 50", strt: ast.DblIntegerLiteral{Value: 500}, step: ast.DblIntegerLiteral{Value: 50}},
		{inp: "AUTO , 20", strt: ast.DblIntegerLiteral{Value: 10}, step: ast.DblIntegerLiteral{Value: 20}},
		{inp: "AUTO ., 20", strt: ast.DblIntegerLiteral{Value: 10}, step: ast.DblIntegerLiteral{Value: 20}},
		{inp: "AUTO 10, 20, 50", err: true},
		{inp: "AUTO 10.5, 20", strt: ast.DblIntegerLiteral{Value: 11}, step: ast.DblIntegerLiteral{Value: 20}},
	}

	for _, tt := range tests {

		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		rc := Eval(&ast.Program{}, env.CmdLineIter(), env)

		if tt.err {
			assert.NotNilf(t, rc, "%s failed to produce error", tt.inp)
			continue
		}
		aut := env.GetSetting(settings.Auto)

		assert.NotNilf(t, aut, "%s failed to save an auto setting", tt.inp)

		ato, ok := aut.(*ast.AutoCommand)

		assert.True(t, ok, "saved auto command wrong type")

		assert.EqualValues(t, &tt.strt, ato.Params[0].(*ast.DblIntegerLiteral))
		assert.EqualValues(t, &tt.step, ato.Params[1].(*ast.DblIntegerLiteral))

	}
}

func Test_BeepStatement(t *testing.T) {
	l := lexer.New("BEEP")
	p := parser.New(l)
	var mt mocks.MockTerm
	initMockTerm(&mt)
	chk := false
	mt.SawBeep = &chk
	env := object.NewTermEnvironment(mt)

	p.ParseCmd(env)

	if len(p.Errors()) > 0 {
		for _, er := range p.Errors() {
			fmt.Println(er)
		}
		return
	}

	Eval(&ast.Program{}, env.CmdLineIter(), env)

	assert.True(t, chk, "Test_BeepStatement term.beep() not called!")
}

func Test_BreakSignal(t *testing.T) {
	tests := []struct {
		inp string
		exp string
	}{
		{inp: `10 PRINT "Hello World!"`, exp: "Hello World!"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		mt.ExpMsg = &mocks.Expector{Exp: []string{tt.exp}}
		flag := true
		mt.SawBreak = &flag
		env := object.NewTermEnvironment(mt)

		p.ParseProgram(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		env.SetRun(true)
		rc := Eval(&ast.Program{}, env.StatementIter(), env)

		_, ok := rc.(*object.HaltSignal)

		assert.True(t, ok, "failed to get a HaltSignal")
	}
}

func Test_ChainStatementCommandLine(t *testing.T) {
	tests := []struct {
		stmt string
		rs   int // response code  eg '200'
		send string
		exp  string
	}{
		{stmt: `CHAIN "start.bas"`, rs: 404, send: ``},
		{stmt: `CHAIN "start.bas"`, rs: 200, send: `10 PRINT "Hello World!"`, exp: "Hello World!"},
		{stmt: `CHAIN 5`, rs: 200, send: `10 PRINT`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.stmt)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		mt.ExpMsg = &mocks.Expector{Exp: []string{tt.exp}}
		env := object.NewTermEnvironment(mt)
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(tt.rs)
			res.Write([]byte(tt.send))
		}))
		defer ts.Close()
		env.SaveSetting(object.SERVER_URL, &ast.StringLiteral{Value: ts.URL})

		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		Eval(&ast.Program{}, env.CmdLineIter(), env)

		if len(tt.exp) != 0 {
			assert.False(t, mt.ExpMsg.Failed, "cmdline CHAIN file no output!")
		}
	}
}

func Test_ChainStatementRunning(t *testing.T) {
	tests := []struct {
		stmt string
		rs   int // response code  eg '200'
		send string
		exp  string
	}{
		{stmt: `10 CHAIN "start.bas"`, rs: 200, send: `10 PRINT "Hello World!"`, exp: "Hello World!"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.stmt)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		mt.ExpMsg = &mocks.Expector{Exp: []string{tt.exp}}
		env := object.NewTermEnvironment(mt)
		ts := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(tt.rs)
			res.Write([]byte(tt.send))
		}))
		defer ts.Close()
		env.SaveSetting(object.SERVER_URL, &ast.StringLiteral{Value: ts.URL})

		p.ParseProgram(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		env.SetRun(true)
		Eval(&ast.Program{}, env.StatementIter(), env)

		assert.False(t, mt.ExpMsg.Failed, "running CHAIN file no output!")
	}
}

func Test_ChrS(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: `10 X$ = CHR$(45)`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		//mt.ExpMsg = &mocks.Expector{Exp: []string{tt.exp}}
		env := object.NewTermEnvironment(mt)

		p.ParseProgram(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		env.SetRun(true)
		Eval(&ast.Program{}, env.StatementIter(), env)
	}
}

func Test_ChDirStatement(t *testing.T) {
	tests := []struct {
		path  ast.Expression
		exp   string // what the CWD should end up being
		rpath string // request path the server should see
		rc    int    // return code for mock server
	}{
		{path: &ast.StringLiteral{Value: `D:\`}, rpath: `/driveD`, exp: `D:\`, rc: 200},
		{path: &ast.StringLiteral{Value: `D:\`}, rpath: `/driveD`, exp: `C:\`, rc: 404},
		{path: &ast.StringLiteral{Value: `PROG`}, rpath: `/driveC/PROG`, exp: `C:\PROG\`, rc: 200},
		{path: &ast.StringLiteral{Value: `PROG`}, rpath: `/driveC/PROG`, exp: `C:\`, rc: 404},
		{path: &ast.IntegerLiteral{Value: 6}, exp: `C:\`},
		{path: &ast.StringLiteral{Value: `\prog`}, rpath: `/driveC/prog`, exp: `C:\prog\`, rc: 200},
	}

	for _, tt := range tests {
		cd := ast.ChDirStatement{}
		if tt.path != nil {
			cd.Path = append(cd.Path, tt.path)
		}
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		ts := mocks.NewMockServer(tt.rc, []byte("Body text"))
		defer ts.Close()

		env.SaveSetting(object.SERVER_URL, &ast.StringLiteral{Value: ts.ExpURL})

		Eval(&cd, env.CmdLineIter(), env)

		// make sure current working directory is what I expect
		cwd := fileserv.GetCWD(env)
		assert.Equal(t, tt.exp, cwd)

		// make sure request looked correct
		if len(tt.rpath) > 0 {
			assert.EqualValues(t, tt.rpath, ts.Requests[0], "Server request badly formed!")
		}
	}
}

func Test_ColorMode(t *testing.T) {
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	scr := evalColorMode(env)

	assert.Equal(t, ast.ScrnModeMDA, scr.Settings[ast.ScrnMode])

	scr.Settings[ast.ScrnMode] = ast.ScrnModeCGA
	env.SaveSetting(settings.Screen, scr)
	scr = evalColorMode(env)

	assert.Equal(t, ast.ScrnModeCGA, scr.Settings[ast.ScrnMode])
}

func Test_ColorPalette(t *testing.T) {
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	scr := &ast.ScreenStatement{}
	scr.InitValue()
	plt := evalColorPalette(scr, env)

	assert.Equal(t, object.XCyan, plt.BasePalette[object.GWCyan], "BasePalette[GWCyan] came back as %d", plt.BasePalette[object.GWCyan])
	plt.BasePalette[object.GWCyan] = object.XBlack
	env.SaveSetting(settings.Palette, plt)
	plt = evalColorPalette(scr, env)

	assert.Equal(t, object.XBlack, plt.BasePalette[object.GWCyan])
}

func Test_ColorStatement(t *testing.T) {
	tests := []struct {
		inp  string
		mode int
		exp  string
	}{
		{inp: "COLOR 15", mode: 0, exp: "\x1b[97m"},
		{inp: "COLOR 7", mode: 0, exp: "\x1b[37m"},
		{inp: "COLOR 1", mode: 0, exp: "\x1b[1m"},
		{inp: "COLOR 1", mode: 3},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		mt.ExpMsg = &mocks.Expector{}

		if len(tt.exp) > 0 {
			mt.ExpMsg.Exp = append(mt.ExpMsg.Exp, tt.exp)
		}
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		// set the screen to desired mode
		scrn := ast.ScreenStatement{}
		scrn.Settings[ast.ScrnMode] = tt.mode
		env.SaveSetting(settings.Screen, &scrn)
		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		Eval(&ast.Program{}, env.CmdLineIter(), env)

		if len(tt.exp) > 0 {
			assert.False(t, mt.ExpMsg.Failed, "COLOR got unexpected output")
		}
	}
}

//
func Test_ClearCommand(t *testing.T) {
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	cmd := ast.ClearCommand{}

	Eval(&cmd, env.CmdLineIter(), env)
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
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			return
		}

		Eval(&ast.Program{}, env.CmdLineIter(), env)

		if !*mt.SawCls {
			t.Errorf("No call to Cls() seen")
		}
	}
}

func Test_CommonStatement(t *testing.T) {
	tests := []struct {
		inp string
		chk string
		exp string
	}{
		//{inp: "10 COMMON A$, B$"},
		{inp: `10 COMMON A$
		20 A$ = "Test"
		30 A$ = "Foo"`, chk: "A$", exp: "Foo"},
	}

	for _, tt := range tests {
		rc := testEval(tt.inp, tt.chk)

		assert.Equal(t, tt.exp, rc.Inspect(), "Test_Common got %s", rc.Inspect())
	}
}

func Test_ContCommand_Errors(t *testing.T) {
	tests := []struct {
		inp    string
		setRun bool
	}{
		{inp: "CONT", setRun: true},
		{inp: "CONT"},
	}

	for _, tt := range tests {

		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		if tt.setRun {
			env.SetRun(true)
		}
		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			t.Fatalf("%s command failed!", tt.inp)
			return
		}

		Eval(&ast.Program{}, env.CmdLineIter(), env)
	}
}

//func Test_ContCommand_Start(t *testing.T) {
func ExampleContCommand() {
	// create my test program
	inp := `10 PRINT "Hello!" : X = 5: STOP : PRINT "Goodbye!"`

	l := lexer.New(inp)
	p := parser.New(l)
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	env.SetRun(true)
	Eval(&ast.Program{}, env.StatementIter(), env)
	env.SetRun(false)

	// now try to continue
	l = lexer.New("CONT")
	p = parser.New(l)
	p.ParseCmd(env)

	Eval(&ast.Program{}, env.CmdLineIter(), env)

	// Output:
	// Hello!
	// Goodbye!
}

func Test_CsrLinExpression(t *testing.T) {
	// create my test program
	inp := `10 X = CSRLIN`

	l := lexer.New(inp)
	p := parser.New(l)
	var mt mocks.MockTerm
	initMockTerm(&mt)
	row := 5
	mt.Row = &row
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	rc := Eval(&ast.Program{}, env.StatementIter(), env)

	assert.Nil(t, rc, "CSRLIN unexpectedly returned %s", fmt.Sprintf("%T", rc))
	csrlin := env.Get("X")

	newRow, ok := csrlin.(*object.Integer)

	assert.True(t, ok, "CSRLIN did not return an integer!")
	assert.Equal(t, row+1, int(newRow.Value), "CSRLIN returned %d, expected %d", newRow.Value, row+1)
}

func Test_ErrorStatement(t *testing.T) {
	tests := []struct {
		inp string
		err int
	}{
		{inp: `10 ERROR 200`, err: 200},
		{inp: `10 ERROR @`, err: berrors.IllegalFuncCallErr},
		{inp: `10 ERROR 300`, err: berrors.Syntax},
		{inp: `10 ERROR "300"`, err: berrors.Syntax},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		row := 5
		mt.Row = &row
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)

		rc := Eval(&ast.Program{}, env.StatementIter(), env)

		assert.NotNil(t, rc, "ErrorStatement didn't produce an error")

		err, ok := rc.(*object.Error)
		assert.True(t, ok, "ErrorStatement didn't return an error")
		assert.Equalf(t, tt.err, err.Code, "ErrorStatement expected %d, got %d", tt.err, err.Code)
	}
}

func Test_EvalExpressionNode(t *testing.T) {
	tests := []struct {
		nd  ast.Node
		inp string
		err bool
	}{
		{inp: `10 REM A comment`, err: true},
		{inp: `20 REM A comment`, nd: &ast.IntegerLiteral{Value: 5}},
		{inp: `20 REM A comment`, nd: &ast.ClsStatement{}, err: true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		row := 5
		mt.Row = &row
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)
		iter := env.StatementIter()

		rc := evalExpressionNode(tt.nd, iter, env)

		if tt.err {
			_, ok := rc.(*object.Error)
			assert.True(t, ok, "evalExpressionNode for %s should have thrown an error", tt.inp)
		}
	}
}

func Test_EvalExpressionNodeTyped(t *testing.T) {
	tests := []struct {
		inp  ast.Node
		want object.Object
		err  bool
		dsc  string
	}{
		{inp: &ast.StringLiteral{Value: "A test!"}, want: &object.String{}, err: false, dsc: "string to get string"},
		{inp: &ast.StringLiteral{Value: "A test!"}, want: &object.Integer{}, err: true, dsc: "string to get integer"},
		{inp: &ast.IntegerLiteral{Value: 12}, want: &object.Integer{}, err: false, dsc: "integer to get integer"},
	}

	for _, tt := range tests {

		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		rc := evalExpressionNodeTyped(tt.inp, nil, env, tt.want)

		if !tt.err {
			assert.NotNil(t, rc, tt.dsc)
		} else {
			assert.Nil(t, rc, tt.dsc)
		}
	}
}

func Test_FilesCommand(t *testing.T) {
	tests := []struct {
		param  string // parameter to the files command
		expURL string // the URL I expect him to ask for
		cwd    string
		send   string
		exp    string
		rs     int
		err    bool
	}{
		{param: "", expURL: `/driveC`, cwd: `C:\`, send: "10 PRINT \"Main Menu\"\n", exp: "10 PRINT \"Main Menu\"\n", rs: 404, err: false},
		{param: "", expURL: `/driveC`, cwd: `C:\`, send: "10 PRINT \"Main Menu\"\n", exp: "10 PRINT \"Main Menu\"\n", rs: 200, err: false},
		{param: "HamCalc", expURL: `/driveC/HamCalc`, cwd: `C:\`, send: "10 PRINT \"Main Menu\"\n", exp: "10 PRINT \"Main Menu\"\n", rs: 200, err: false},
		{param: "", expURL: `/driveC`, cwd: `C:\`, send: `[{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":true}]`, exp: `[{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":true}]`, rs: 200, err: false},
	}

	for _, tt := range tests {
		cmd := "FILES"

		if len(tt.param) > 0 {
			cmd = cmd + ` "` + tt.param + `"`
		}

		l := lexer.New(cmd)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		ts := mocks.NewMockServer(tt.rs, []byte(tt.send))
		defer ts.Close()

		env.SaveSetting(object.SERVER_URL, &ast.StringLiteral{Value: ts.ExpURL})

		if len(tt.cwd) > 0 {
			env.SaveSetting(object.WORK_DRIVE, &ast.StringLiteral{Value: tt.cwd})
		}

		p.ParseCmd(env)

		if len(p.Errors()) > 0 {
			for _, er := range p.Errors() {
				fmt.Println(er)
			}
			t.Fatal("FILES command failed!")
			return
		}

		Eval(&ast.Program{}, env.CmdLineIter(), env)

		// did he request the URL I expected
		if len(tt.expURL) > 0 {
			assert.EqualValues(t, tt.expURL, ts.Requests[0], "FILES asked for wrong URL")
		}
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
		var mt mocks.MockTerm
		initMockTerm(&mt)
		var rec string
		mt.SawStr = &rec
		env := object.NewTermEnvironment(mt)
		env.SaveSetting(object.WORK_DRIVE, &ast.StringLiteral{Value: `C:\`})

		catchNotDir(tt.path, errors.New(tt.send), env)
		assert.Equal(t, tt.exp, rec, "Test_CatchNotDir got unexpected return")
	}
}

func Test_ForStatement(t *testing.T) {
	// along the way we also test the NEXT statement

	tests := []struct {
		inp string
		err bool
	}{
		{inp: `10 FOR I = `, err: true},
		{inp: `10 FOR I = 5 TO 2 : PRINT I : NEXT I`},
		{inp: `10 FOR I = 1 TO 2 STEP 0.5 : PRINT I : NEXT I`},
		{inp: `10 FOR I = 1 TO 3 : PRINT I : NEXT I`},
		{inp: `10 FOR I = 1 TO 3 : PRINT I : NEXT J`, err: true},
		{inp: `10 FOR I = 1 TO 4 STEP 2 : PRINT I : NEXT I`},
		{inp: `10 FOR I = 1 TO 4 STEP 0 : PRINT I : NEXT I`},
		{inp: `10 FOR I = 5 TO -3 STEP -1 : PRINT I : NEXT I`},
		{inp: `10 FOR I = 7 TO 7 : PRINT I : NEXT I`},
		{inp: `10 PRINT I : NEXT I`, err: true},
	}

	for _, tt := range tests {

		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		p.ParseProgram(env)
		itr := env.StatementIter()
		res := Eval(&ast.Program{}, itr, env)

		if tt.err {
			_, ok := res.(*object.Error)

			assert.Truef(t, ok, "%s expected error, didn't get one", tt.inp)
		}
	}
}

func Test_ForSkip(t *testing.T) {
	tests := []struct {
		inp string
		id  string
		err bool
	}{
		{inp: `10 FOR I = 1 TO 10 : PRINT I : NEXT I : END`, id: "I"},
		{inp: `20 FOR I = 1 TO 10 : PRINT I
		30 FOR J = 1 TO 5 : PRINT J : NEXT J
		40 NEXT I`, id: "I"},
		{inp: `30 FOR I = 1 TO 10 : PRINT I : END`, err: true},
		{inp: `40 FOR I = 1 TO 10 : FOR J = 2 TO 3 : PRINT I`, err: true},
		{inp: `50 FOR I = 1 TO 10 : PRINT I : NEXT J : END`, id: "I"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		p.ParseProgram(env)
		itr := env.StatementIter()
		itr.Next()
		itr.Next()
		four := &ast.ForStatment{Init: &ast.LetStatement{Name: &ast.Identifier{Value: tt.id}}}
		evalForSkipLoop(four, itr, env)

		if !tt.err {
			nxt := itr.Value()

			_, ok := nxt.(*ast.NextStatement)

			if !ok {
				t.Fatalf("%s failed to find an expected NEXT", tt.inp)
			}
		}
	}
}

func Test_GosubStatement(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: "10 GOSUB"},
		{inp: "10 GOSUB 30"},
		{inp: "10 GOSUB 30\n20 END\n30 STOP"},
	}

	for _, tt := range tests {

		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		p.ParseProgram(env)
		itr := env.StatementIter()
		Eval(&ast.Program{}, itr, env)
	}
}

func Test_EvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int16
	}{
		{"10 X = -5", -5},
		{"50 X=5 + 5", 10},
		{"70 X=5 < 10", 1},
		{"80 x=5 > 10", 0},
		{"110 x=10 > 1", 1},
		{"120 x=10 < 1", 0},
		{"130 x=10 / 2", 5},
		{"160 X=10 \\ 2", 5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)

		assert.Zero(t, len(p.Errors()), "Parse(%s) failed with errors!")

		// need to execute run command
		env.SetRun(true)
		rc := Eval(&ast.Program{}, env.StatementIter(), env)

		assert.Nil(t, rc, "eval of %s returned a %T", tt.input, rc)

		val := env.Get("X")

		x, ok := val.(*object.Integer)
		assert.True(t, ok, "eval %s didn't set X!", tt.input)
		assert.Equal(t, tt.expected, int16(x.Value), "eval %s expected %d, got %d", tt.input, tt.expected, x.Value)
	}
}

func TestDblInetegerExpression(t *testing.T) {
	tests := []struct {
		inp string
		exp int32
	}{
		{"10 x=99999", 99999},
		{"20 x=-99999", -99999},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)

		assert.Zero(t, len(p.Errors()), "Parse(%s) failed with errors!")

		// need to execute run command
		env.SetRun(true)
		rc := Eval(&ast.Program{}, env.StatementIter(), env)

		assert.Nil(t, rc, "eval %s returned a %T", tt.inp, rc)

		val := env.Get("X")

		x, ok := val.(*object.IntDbl)
		assert.True(t, ok, "eval %s didn't set X!", tt.inp)
		assert.Equal(t, tt.exp, int32(x.Value), "eval %s expected %d, got %d", tt.inp, tt.exp, x.Value)
	}
}

func TestDim_Statements(t *testing.T) {
	tests := []struct {
		inp string
		chk string
		exp int
	}{
		{inp: `10 DIM A[20] : A[11] = 6 : PRINT A[11]`, chk: "A[11]", exp: 6},
		/*{`20 DIM B[10,10]`},
		{`30 DIM A[9,10], B[14,15] : B[5,6] = 12 : PRINT B[5,6]`},*/
	}

	for _, tt := range tests {
		rc := testEval(tt.inp, tt.chk)

		assert.NotNil(t, rc, "TestDim_Statements failed to get value")
	}

	// want
	// 4
}

func testEval(input string, vbl string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	if len(p.Errors()) > 0 {
		return nil
	}

	// need to execute run command
	env.SetRun(true)
	rc := Eval(&ast.Program{}, env.StatementIter(), env)

	if rc != nil {
		return rc
	}

	return env.Get(vbl)
}

func testEvalEnv(input string, vbl string, env *object.Environment) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	p.ParseProgram(env)

	if len(p.Errors()) > 0 {
		return nil
	}

	// need to execute run command
	env.SetRun(true)
	rc := Eval(&ast.Program{}, env.StatementIter(), env)

	if rc != nil {
		return rc
	}

	return env.Get(vbl)
}

func testEvalWithClient(input string, file string, err *error) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	mc := &mocks.MockClient{Contents: file}
	if err != nil {
		mc.Err = *err
	}
	env.SetClient(mc)

	p.ParseCmd(env)

	if len(p.Errors()) > 0 {
		return nil
	}

	return Eval(&ast.Program{}, env.CmdLineIter(), env)
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

func Test_IfExpression(t *testing.T) {
	tests := []struct {
		inp string
		exp int
	}{
		{inp: "10 IF 5 < 6 THEN 30\n20 x=5\n30 x=6", exp: 6},
		//{inp: "10 IF 5 < 6 GOTO 30\n20 x=5\n30 x=7", exp: 7},
		{inp: "10 IF 5 < 6 THEN END\n20 x=5", exp: 0},
		//{"10 IF 5 > 6 THEN 20 ELSE END\n20 5", &object.HaltSignal{}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)

		assert.Zero(t, len(p.Errors()), "Parse(%s) failed with errors!")

		// need to execute run command
		env.SetRun(true)
		rc := Eval(&ast.Program{}, env.StatementIter(), env)

		assert.Nil(t, rc, "eval %s returned a %T", tt.inp, rc)

		x := env.Get("X")

		val, ok := x.(*object.Integer)

		assert.True(t, ok, "eval of %s failed to set X", tt.inp)
		assert.Equal(t, tt.exp, int(val.Value), "eval of %s, expected %d, got %d", tt.inp, tt.exp, val.Value)
	}
}

func Test_EndStatement(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"10 END\n20 5\n30 6"},
	}

	for _, tt := range tests {
		rc := testEval(tt.input, "A")

		assert.NotNil(t, rc, "End statement returned a nil!")
		ec, ok := rc.(*object.Integer)
		assert.True(t, ok, "End statement didn't return an integer")
		assert.Equal(t, 0, int(ec.Value))
	}
}

func Test_KeyStatement(t *testing.T) {
	tests := []struct {
		inp string
		len int
		exp string
		err bool
	}{
		{inp: `10 KEY OFF`},
		{inp: `10 KEY ON`},
		{inp: `10 KEY LIST`, exp: "F1 \r\nF2 \r\nF3 \r\nF4 \r\nF5 \r\nF6 \r\nF7 \r\nF8 \r\nF9 \r\nF10 \r\n"},
		{inp: `10 KEY 4,"FILES"`, len: 1},
		{inp: `10 KEY 4,"FILES" : KEY LIST`, len: 1, exp: "F1 \r\nF2 \r\nF3 \r\nF4 FILES\r\nF5 \r\nF6 \r\nF7 \r\nF8 \r\nF9 \r\nF10 \r\n"},
		{inp: `10 KEY 1`, err: true},
		{inp: `20 KEY 25,"FILES"`, err: true},
		{inp: `20 KEY "25","FILES"`, err: true},
		{inp: `20 KEY 2,30`, err: true},
		{inp: `60 KEY 15, CHR$(03)+CHR$(25)`},
		{inp: `60 KEY 15, 34`, err: true},
	}

	for _, tt := range tests {
		mt := mocks.MockTerm{}
		mt.ExpMsg = &mocks.Expector{Exp: []string{tt.exp}}
		env := object.NewTermEnvironment(mt)
		err := testEvalEnv(tt.inp, "Key", env)

		ks := env.GetSetting(settings.KeyMacs)

		if !tt.err {
			assert.NotNil(t, ks, "Key statement didn't create setting")

			kset := ks.(*ast.KeySettings)
			assert.NotNil(t, kset, "Key settings is incorrect type")
			assert.EqualValues(t, tt.len, len(kset.Keys), "Key settings count is wrong")

			if len(tt.exp) > 0 {
				assert.Falsef(t, mt.ExpMsg.Failed, "KEY LIST didn't return expected value < %s", tt.inp)
			}
		} else {
			assert.NotNil(t, err, "expected KEY to return an error and he didn't")
			eval := err.(*object.Error)
			assert.NotNilf(t, eval, "expected KEY to retrun error but got %T", err)
		}
	}
}

func Test_LetStatements(t *testing.T) {
	tests := []struct {
		inp string
		chk string
		exp int16
	}{
		{inp: "10 LET a = 5", chk: "a", exp: 5},
		{inp: "20 LET a = 5 * 5", chk: "a", exp: 25},
		{inp: "30 LET a = 5: let b = a:", chk: "b", exp: 5},
		{inp: "40 LET a = 5: let b = a: let c = a + b + 5", chk: "c", exp: 15},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.inp, tt.chk), tt.exp)
	}
}

func Test_LoadCommand(t *testing.T) {
	tests := []struct {
		src  string // source code of the file to run
		cmd  string // the load command to
		fail bool   // should not get a file
		emsg string // an error I want the httpClient to return
	}{
		{src: `10 PRINT "Hello!"`, cmd: `LOAD "HELLO.BAS"`},
		{src: `10 PRINT "Goodbye!"`, cmd: `LOAD 5`, fail: true},
		{src: `10 PRINT "And I Ran!"`, cmd: `LOAD "HELLO.BAS",R`},
		{src: `10 PRINT "And I don't run"`, cmd: `LOAD "HELLO.BAS",R`, emsg: "File not found"},
	}

	for _, tt := range tests {
		//rc := testEvalWithClient(tt.cmd, tt.src)
		var emsg *error

		if len(tt.emsg) != 0 {
			err := errors.New(tt.emsg)
			emsg = &err
		}
		rc := testEvalWithClient(tt.cmd, tt.src, emsg)

		fmt.Printf("%s got %T\n", tt.src, rc)

		if tt.fail && (rc == nil) {
			t.Fatalf("%s should have errored, but didn't", tt.cmd)
		}
	}
}

func Test_LoadCommandWithLiveServer(t *testing.T) {
	tests := []struct {
		cmd string // the load command to
	}{
		{cmd: `LOAD "HCALC.txt"`},
	}

	for _, tt := range tests {
		p := parser.New(lexer.New(tt.cmd))

		errs := p.Errors()
		for _, e := range errs {
			t.Fatalf(e)
		}
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseCmd(env)

		Eval(&ast.Program{}, env.CmdLineIter(), env)

	}

}

func Test_LocateStatement(t *testing.T) {
	tests := []struct {
		inp string
		err bool
	}{
		{inp: `LOCATE`, err: true},
		{inp: `LOCATE ,,1`},
		{inp: `LOCATE X`, err: true},
		{inp: `LOCATE 4,5`},
		{inp: `LOCATE 4,X`, err: true},
		{inp: `LOCATE 5,6,1`, err: true},
		{inp: `LOCATE 5`}, // just change the row
		{inp: `LOCATE 1,2,3,4,5,6`, err: true},
	}

	for _, tt := range tests {
		p := parser.New(lexer.New(tt.inp))

		errs := p.Errors()
		for _, e := range errs {
			t.Fatalf(e)
		}
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseCmd(env)

		res := Eval(&ast.Program{}, env.CmdLineIter(), env)

		// check for any error coming back
		if res != nil {
			err := res.(*object.Error)
			if (err == nil) && tt.err {
				t.Fatalf("%s got no error but expected one", tt.inp)
			}

			if (err != nil) && !tt.err {
				t.Fatalf("%s got an unexpected error", tt.inp)
			}
		}
	}
}

func Test_NewCommand(t *testing.T) {
	l := lexer.New(`10 PRINT "Hello!"`)
	p := parser.New(l)
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)
	code := env.StatementIter()
	cmd := ast.NewCommand{}

	rc := Eval(&cmd, code, env)

	_, ok := rc.(*object.HaltSignal)

	assert.True(t, ok, "New command failed to send halt!")
}

func Test_OnErrorGotoStatement(t *testing.T) {
	tests := []struct {
		inp string
		jmp int
		err bool
		msg string
	}{
		{inp: `10 ON ERROR GOTO 100
		100 END`, jmp: 100},
		{inp: `10 ON ERROR GOTO 0`, jmp: 0},
		{inp: `10 ON ERROR GOTO 100`, err: true, msg: "Undefined line number in 10"},
		{inp: `10 ON ERROR GOTO -5`, err: true, msg: "Syntax error in 10"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)
		code := env.StatementIter()
		env.SetRun(true)
		rc := Eval(&ast.Program{}, code, env)
		env.SetRun(false)

		if rc != nil {
			// should be the expected error
			err, ok := rc.(*object.Error)

			assert.True(t, ok, "didn't get the expected error")
			assert.EqualValues(t, tt.msg, err.Message, "didn't get the expected error")
		} else if tt.jmp > 0 {
			// check the setting
			set := env.GetSetting(settings.OnError)
			oeg, ok := set.(*ast.OnErrorGoto)

			assert.True(t, ok, "failed to get OnErrorGoto node")
			assert.EqualValues(t, tt.jmp, oeg.Jump)
		} else {
			set := env.GetSetting(settings.OnError)

			assert.Nil(t, set, "OnError failed to clear")
		}
	}
}

func Test_OnGoStatement(t *testing.T) {
	tests := []struct {
		inp string
		jmp int
		err bool
		msg string
	}{
		{inp: `10 X = 1 : ON X GOTO 100, 200
		100 END
		200 END`, jmp: 100},
		{inp: `10 X = 2 : ON X GOSUB 100, 200
		100 END
		200 END`, jmp: 200},
		{inp: `10 X = 0 : ON X GOSUB 100, 200
		20 END
		100 END
		200 END`, jmp: 20},
		{inp: `10 X = 3 : ON X GOSUB 100, 200
		20 END
		100 END
		200 END`, jmp: 20},
		{inp: `10 X = 2 : ON X JUMP 100, 200
		100 END
		200 END`, err: true, msg: "Syntax error in 10"},
		{inp: `10 X = 2 : ON X JUMP 100, 200`, err: true, msg: "Undefined line number in 10"},
		{inp: `10 ON GOTO 100, 200`, err: true, msg: "Syntax error in 10"},
		{inp: `10 ON ! GOTO 100, 200`, err: true, msg: "Syntax error in 10"},
		{inp: `10 ON ERROR GOTO 100`, err: true, msg: "Undefined line number in 10"},
		{inp: `10 ON ERROR GOTO -5`, err: true, msg: "Syntax error in 10"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)
		code := env.StatementIter()
		env.SetRun(true)
		rc := Eval(&ast.Program{}, code, env)
		env.SetRun(false)

		if rc != nil {
			// should be the expected error
			err, ok := rc.(*object.Error)

			assert.True(t, ok, "didn't get the expected error")
			assert.EqualValues(t, tt.msg, err.Message, "didn't get the expected error")
		} else {
			// check the setting
			assert.Equal(t, tt.jmp, code.CurLine(), "Jumped to wrong line")
		}
	}
}

func Test_PrintStatement(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: `10 LET Y[0] = 5`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		p.ParseProgram(env)

		errs := p.Errors()

		assert.Zero(t, len(errs), "parser threw some errors")

		code := env.StatementIter()
		env.SetRun(true)
		rc := Eval(&ast.Program{}, code, env)
		env.SetRun(false)

		assert.Nil(t, nil, "got %T back", rc)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `10 A$ = "Hello World!"`
	rc := testEval(input, "A$")

	assert.Equal(t, rc.Inspect(), "Hello World!", "TestStringLiteral got %s", rc.Inspect())
}

func TestStringConcatenation(t *testing.T) {
	input := `10 A$ = "Hello" + " " + "World!"`
	evaluated := testEval(input, "A$")

	assert.Equal(t, evaluated.Inspect(), "Hello World!", "TestStringConcatenation got %s", evaluated.Inspect())
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
		evaluated := testEval(tt.input, "")

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
	input := "10 DEF FNSKIP(x)= x + 2"

	rc := testEval(input, "FNSKIP")

	fn, ok := rc.(*object.Function)

	assert.Truef(t, ok, "object is not Function. got=%T (%+v)", rc, rc)
	assert.Equal(t, 1, len(fn.Parameters))
	assert.Equal(t, "x", fn.Parameters[0].String())
	assert.Equal(t, "x = X + 2", fn.Body.String())
}

func TestFunctionExecution(t *testing.T) {
	tests := []struct {
		inp string
		res object.Object
		vbl string
	}{
		{inp: `10 DEF FNMUL(x,y)= x * y : Y = FNMUL(2,5)`, res: &object.Integer{Value: 10}, vbl: "Y"},
		{inp: `10 DEF FNSKIP(x)= x + 2 : Y = FNSKIP(1)`, res: &object.Integer{Value: 3}, vbl: "Y"},
		{inp: `10 DEF FNSKIP(x)= x + 2 : Y = FNSKIP(1)`, res: &object.Function{}, vbl: "FNSKIP"},
	}

	for _, tt := range tests {
		rc := testEval(tt.inp, tt.vbl)
		assert.IsTypef(t, tt.res, rc, "%s return %T", tt.inp, rc)
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
		var mt mocks.MockTerm
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
		{"10 DEF FNID(x) = x : y = FNID(5)", 5},
		//{"20 DEF FNMUL(x,y) = x*y : y = FNMUL(2,3)", 6},
		//{"30 DEF FNSKIP(x)= (x + 2): y = FNSKIP(3)", 5},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input, "y"), tt.expected)
	}
}

func TestHexOctalConstants(t *testing.T) {
	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 X = &H7F`, int16(127)},
		{`20 &HG7F`, "Syntax error in 20"},
		{`30 &H7FFFFF`, "Overflow in 30"},
		{`40 X = &O7`, int16(7)},
		{`50 X = &O77`, int16(63)},
		{`60 x = &O77777`, int16(32767)},
		{`70 &O777777`, "Overflow in 70"},
		{`80 x = &77777`, int16(32767)},
		{`90 &O78777`, "Syntax error in 90"},
	}

	for _, tt := range tests {
		evald := testEval(tt.inp, "X")
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

func Test_ReadStatement(t *testing.T) {
	fixedInt, _ := decimal.NewFromString("999.99")

	tests := []struct {
		inp string
		chk string
		exp object.Object
	}{
		{inp: `10 DATA "Fred", "George" : READ A$`, chk: `A$`, exp: &object.String{Value: "Fred"}},
		{inp: `20 DATA 123 : READ A`, chk: `A`, exp: &object.Integer{Value: 123}},
		{inp: `30 DATA 99999 : READ A`, chk: `A`, exp: &object.IntDbl{Value: 99999}},
		{inp: `40 DATA 999.99 : READ A`, chk: `A`, exp: &object.Fixed{Value: fixedInt}},
		{inp: `50 DATA 2.35123412341234E+4 : READ A`, chk: `A`, exp: &object.FloatSgl{Value: 23512.341796875}},
		{inp: `60 DATA 2.35123412341234D+4 : READ A`, chk: `A`, exp: &object.FloatDbl{Value: 23512.3412341234}},
		{inp: `70 DATA -2.35123412341234D+4 : READ A`, chk: `A`, exp: &object.FloatDbl{Value: -23512.3412341234}},
		{inp: `80 DATA "Fred" : READ A$ : READ B$`, chk: `A$`, exp: &object.Error{Message: "Out of DATA in 80"}},
		{inp: `90 DATA 3,4,5 : READ 3+5`, chk: ``, exp: &object.Error{Message: "Syntax error in 90"}},
		{inp: `100 DATA 3,4,5 : READ`, exp: &object.Error{Message: "Syntax error in 100"}},
	}

	for _, tt := range tests {
		res := testEval(tt.inp, tt.chk)

		if tt.exp == nil {
			assert.Nil(t, res, "got an object when I didn't expect one!")
		} else {
			compareObjects(tt.inp, res, tt.exp, t)
		}
	}
}

func Test_RestoreStatement(t *testing.T) {

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 DATA "Fred", "George" : RESTORE`, nil},
		{`20 DATA "Fred", "George" : RESTORE 20`, nil},
		{`30 DATA "Fred", "George" : RESTORE 5`, &object.Error{Message: "Undefined line number in 30"}},
	}

	for _, tt := range tests {
		res := testEval(tt.inp, "")

		if tt.exp != nil {
			compareObjects("Restore", res, tt.exp, t)
		}
	}
}

func Test_ResumeStatement(t *testing.T) {
	tests := []struct {
		inp string
		err bool
	}{
		{inp: `10 ON ERROR GOTO 100
		20 ERROR 17 : PRINT "Test"
		30 PRINT "I'm back!"
		40 END
		100 PRINT ERR, ERL
		110 RESUME NEXT`},
		{inp: `10 ON ERROR GOTO 100
		20 ERROR 17 : REM syntax error
		30 PRINT "I'm back!"
		40 END
		100 PRINT ERR, ERL
		110 RESUME 40`},
		{inp: `10 ON ERROR GOTO 100
		20 ERROR 17 : REM syntax error
		30 PRINT "I'm back!"
		40 END
		100 PRINT ERR, ERL
		110 RESUME 400`, err: true},
		{inp: `10 ON ERROR GOTO 100
		20 ERROR 17 : REM syntax error
		30 PRINT "I'm back!"
		40 END
		100 PRINT ERR, ERL
		110 RESUME FRED`, err: true},
		{inp: `10 ON ERROR GOTO 100
		20 ERROR 17 : REM syntax error
		30 PRINT "I'm back!"
		40 END
		100 PRINT ERR, ERL
		110 RESUME NEXT FRED`, err: true},
		// last test blows through the switch, just 4 fun
		{inp: `10 ON ERROR GOTO 100
		20 ERROR 17 : REM syntax error
		30 PRINT "I'm back!"
		40 END
		100 PRINT ERR, ERL
		110 RESUME PRINT`, err: true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		p.ParseProgram(env)

		env.SetRun(true)
		rc := Eval(&ast.Program{}, env.StatementIter(), env)
		env.SetRun(false)

		if tt.err {
			assert.NotNil(t, rc, "%s expected error, didn't get one", tt.inp)
		} else {
			assert.Nil(t, rc, "got an unexpected error %T", rc)
		}
	}
}

func Test_ReturnStatement(t *testing.T) {
	tests := []struct {
		src string
		err int
	}{
		{src: `10 RETURN`, err: berrors.ReturnWoGosub},
	}

	for _, tt := range tests {
		res := testEval(tt.src, "")

		if tt.err != 0 {
			var mt mocks.MockTerm
			initMockTerm(&mt)
			env := object.NewTermEnvironment(mt)

			exp := object.StdError(env, tt.err)
			exp.Message = exp.Message + " in 10"
			assert.Equal(t, exp, res)
		}
	}
}

func ExampleReturnStatement() {
	prg := `10 GOSUB 100
	20 PRINT "I'm back!"
	30 END
	100 PRINT "Subroutine"
	110 RETURN`

	l := lexer.New(prg)
	p := parser.New(l)
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)

	p.ParseProgram(env)

	env.SetRun(true)
	Eval(&ast.Program{}, env.StatementIter(), env)
	env.SetRun(false)

	// Output:
	// Subroutine
	// I'm back!
}

func Test_RunParameters(t *testing.T) {
	tests := []struct {
		src  string // source code of the file to run
		strt int    // line # to start on
		url  string // give an incorrect url so I can get an error
		exp  object.Object
	}{
		{src: `10 PRINT "Hello!"`},
		{src: `10 PRINT "Goodbye!"`, strt: 10},
		{src: `10 PRINT "Fail!"`, strt: 10, url: "http://localhost:8000/driveC/noprog.txt"},
		{src: `10 PRINT "Not found."`, strt: 20, exp: &object.Error{}},
	}

	for _, tt := range tests {
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		mc := &mocks.MockClient{Contents: tt.src, Url: tt.url}
		env.SetClient(mc)
		cmd := ast.RunCommand{LoadFile: &ast.StringLiteral{Value: "HELLO.BAS"}, StartLine: tt.strt}
		code := env.CmdLineIter()
		rc := evalRunCommand(&cmd, code, env)

		if (rc != nil) && (tt.exp == nil) {
			t.Fatalf("eval of %s returned a non-nil result %T", tt.src, rc)
		}

		if tt.exp != nil {
			rct := fmt.Sprintf("%T", rc)
			expt := fmt.Sprintf("%T", tt.exp)
			if rct != expt {
				t.Fatalf("(%s) expected object of type %T, got result type %T", tt.src, tt.exp, rc)
			}
		}
	}
}

func Test_ScreenStatement(t *testing.T) {
	tests := []struct {
		inp   string
		exp   [4]int
		err   bool
		ecode int
	}{
		{inp: "SCREEN 0,1", exp: [4]int{0, 1, -1, -1}},
		{inp: "SCREEN 0,1 : SCREEN ,2", exp: [4]int{0, 2, -1, -1}},
		{inp: "SCREEN 3", err: true, ecode: berrors.IllegalFuncCallErr},
	}

	for _, tt := range tests {
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		rc := Eval(&ast.Program{}, env.CmdLineIter(), env)

		if !tt.err {
			set := env.GetSetting(settings.Screen)
			scrn := set.(*ast.ScreenStatement)

			assert.NotNil(t, scrn, "Screen settings failed to save!")

			for i := range tt.exp {
				// -1 means it should be nil
				if tt.exp[i] != -1 {
					assert.Equal(t, scrn.Settings[i], tt.exp[i], "Line %s expected %d but got %d", tt.inp, tt.exp[i], scrn.Settings[i])
				} else {
					assert.Zero(t, scrn.Settings[i], "Line %s, setting %d unexpected was %d", tt.inp, i, scrn.Settings[i])
				}
			}
		} else {
			err := rc.(*object.Error)

			assert.NotNil(t, err, "%s failed to return an error", tt.inp)

			assert.Equal(t, err.Code, tt.ecode, "%s didn't return syntax error, gave %s instead", tt.inp, err.Message)
		}
	}
}

func ExampleStopStatement() {
	tests := []struct {
		inp string
		cmd bool
	}{
		{inp: `10 PRINT "Hello!" : STOP : PRINT "Goodbye!"`},
		{inp: `STOP`, cmd: true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		if !tt.cmd {
			p.ParseProgram(env)

			env.SetRun(true)
			Eval(&ast.Program{}, env.StatementIter(), env)
			env.SetRun(false)

			// now try to continue
			l = lexer.New("CONT")
			p = parser.New(l)
			p.ParseCmd(env)

			Eval(&ast.Program{}, env.CmdLineIter(), env)
		} else {
			l = lexer.New(tt.inp)
			p = parser.New(l)
			p.ParseCmd(env)

			Eval(&ast.Program{}, env.CmdLineIter(), env)

		}
	}

	// Output:
	// Hello!
	// Goodbye!
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
		var mt mocks.MockTerm
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

		Eval(&ast.Program{}, env.CmdLineIter(), env)

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
		{`250 PRINT LEN("Hello")`},
		{`260 PRINT 10`},
	}

	for _, tt := range tests {
		testEval(tt.input, "")
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
	// 5
	// 10
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
		{`80 LET Y = 10 / 0 : PRINT Y`},
	}

	for _, tt := range tests {
		testEval(tt.input, "")
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
		{`250 LET Y = 10.25 / 0 : PRINT Y`},
	}

	for _, tt := range tests {
		testEval(tt.input, "")
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
		{`170 LET Y = 2.35E+4 / 0 : PRINT Y`},
	}

	for _, tt := range tests {
		testEval(tt.input, "")
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
		{`220 LET X = 2.35D+4 / 0 : PRINT`},
	}

	for _, tt := range tests {
		testEval(tt.input, "")
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
		testEval(tt.input, "")
	}

	// Output:
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
		testEval(tt.input, "")
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
		/*{`5 REM A comment to get started.`},
		{`10 GOTO 200`},
		{`20 LET X = FNBANG(32)`},*/
		{`30 LET Y = 1.5 : LET X[Y] = 5 : PRINT X[Y]`},
		{`40 LET Y[11] = 5`},
		{`50 LET Y[1] = 5 : LET Y[11] = 4`},
		{`60 LET Y% = 5 : LET Y% = 3.5`},
		{`70 LET A$ = -"A negative msg"`},
		{`80 LET A = 5 + HELLO`},
	}

	for _, tt := range tests {
		testEval(tt.input, "")
	}

	// Output:
	// 5
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
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(&ast.Program{}, env.CmdLineIter(), env)
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
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(&ast.Program{}, env.CmdLineIter(), env)
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
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(&ast.Program{}, env.CmdLineIter(), env)
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
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(&ast.Program{}, env.CmdLineIter(), env)
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
	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	p.ParseProgram(env)

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		p.ParseCmd(env)

		Eval(&ast.Program{}, env.CmdLineIter(), env)
	}

	// Output:
	// Hello World!
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`10 X = LEN("")`, 0},
		/*{`20 LEN("four")`, 4},
		{`30 LEN("hello world")`, 11},
		{`40 LEN(1)`, "Type mismatch in 40"},
		{`50 LEN("one", "two")`, "Syntax error in 50"},
		{`70 LEN("four" / "five")`, &object.Error{}},*/
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input, "")
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int16(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v) test %s", evaluated, evaluated, tt.input)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q test %s", expected, errObj.Message, tt.input)
			}
		}
	}
}

func Test_BuiltinFunctionMissing(t *testing.T) {
	bltin := ast.BuiltinExpression{Token: token.Token{Type: token.BUILTIN, Literal: "FooBar"}}

	var mt mocks.MockTerm
	initMockTerm(&mt)
	env := object.NewTermEnvironment(mt)
	env.AddCmdStmt(&ast.ExpressionStatement{Expression: &bltin})
	code := env.CmdLineIter()

	rc := Eval(&bltin, code, env)

	if err, ok := rc.(*object.Error); ok {
		assert.Equal(t, "Syntax error", err.Message, "Builtin Foobar, didn't get an")
	}
}

func Test_UsingStatement(t *testing.T) {
	tests := []struct {
		inp string
		err object.Object
		exp []string
	}{
		{inp: `PRINT USING "###.##"; 23.45`, err: nil, exp: []string{" 23.45"}},
		{inp: `PRINT "Totals:"; USING "###.##"; 23.45`, err: nil, exp: []string{"Totals:", " 23.45"}},
		{inp: `PRINT USING "##.##"; X`, err: nil, exp: []string{" 0.00"}},
		{inp: `PRINT USING "##.##"; X#`, err: nil, exp: []string{" 0.00"}},
		{inp: `X=2.134E1 : PRINT USING "##.##"; X`, err: nil, exp: []string{"21.34"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		if len(tt.exp) != 0 {
			exp := &mocks.Expector{}
			exp.Exp = append(exp.Exp, tt.exp...)
			mt.ExpMsg = exp
		}
		env := object.NewTermEnvironment(mt)

		p.ParseCmd(env)

		rc := Eval(&ast.Program{}, env.CmdLineIter(), env)

		if tt.err == nil {
			assert.Nilf(t, rc, "%s returned %T", tt.inp, rc)
		} else {
			assert.Equalf(t, tt.err, rc, "%s returned %T", tt.inp, rc)
		}

		if len(tt.exp) != 0 {
			assert.Falsef(t, mt.ExpMsg.Failed, "%s didn't get %s", tt.inp, tt.exp)
		}
	}
}

func Test_ViewPrintStatement(t *testing.T) {
	tests := []struct {
		inp string
		exp string
		err bool
	}{
		{inp: `VIEW PRINT 3 TO 19`, exp: "\x1b[3;19r"},
		{inp: `VIEW PRINT -33 TO 19`, err: true},
		{inp: `VIEW PRINT`, exp: "\x1b[1;24r"},
		{inp: `VIEW PRINT 3 TO`, err: true},
		{inp: `VIEW PRINT 3 4 19`, err: true},
		{inp: `VIEW PRINT FOR TO 19`, err: true},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := parser.New(l)
		var mt mocks.MockTerm
		initMockTerm(&mt)
		expr := mocks.Expector{Exp: []string{tt.exp}}
		mt.ExpMsg = &expr
		env := object.NewTermEnvironment(mt)

		p.ParseCmd(env)

		rc := Eval(&ast.Program{}, env.CmdLineIter(), env)

		assert.Equal(t, tt.err, (rc != nil))

		if tt.err && rc != nil {
			if rc.Type() != object.ERROR_OBJ {

			}
		}

		if len(tt.exp) > 0 {
			assert.False(t, expr.Failed, "term didn't get the word")
		}
	}

}
