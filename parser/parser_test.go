// parser/parser_test.go
package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/token"
)

func TestAutoCommand(t *testing.T) {
	tests := []struct {
		inp   string
		start int
		step  int
		curr  bool
	}{
		{"AUTO", -1, 10, false},
		{"AUTO 20", 20, 10, false},
		{"AUTO , 20", -1, 20, false},
		{"AUTO ., 20", -1, 20, true},
		{"AUTO .", -1, 10, true},
	}

	fmt.Println("TestAutoCommand Parsing")
	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseCmd(env)
		program := env.Program

		checkParserErrors(t, p)

		itr := program.CmdLineIter()

		if itr.Len() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := itr.Value()

		if stmt.TokenLiteral() != token.AUTO {
			t.Fatal("TestAutoCommand didn't get an Auto command")
		}

		atc := stmt.(*ast.AutoCommand)

		if atc == nil {
			t.Fatal("TestAutoCommand couldn't extract AutoCommand object")
		}

		if atc.Start != tt.start {
			t.Fatalf("TestAutoCommand got start = %d, expected %d", atc.Start, tt.start)
		}

		if atc.Increment != tt.step {
			t.Fatalf("TestAutoCommand got increment = %d, expected %d", atc.Increment, tt.step)
		}

		if atc.Curr != tt.curr {
			t.Fatalf("TestAutoCommand got curr = %t, expected %t", atc.Curr, tt.curr)
		}
	}
}

func Test_BeepStatement(t *testing.T) {
	l := lexer.New("BEEP")
	p := New(l)
	env := &object.Environment{}
	p.ParseCmd(env)
	program := env.Program

	checkParserErrors(t, p)

	itr := program.CmdLineIter()

	if itr.Len() != 1 {
		t.Fatal("program.Cmd does not contain single command")
	}

	stmt := itr.Value()

	if stmt.TokenLiteral() != token.BEEP {
		t.Fatal("TestBeepStatement didn't get an Beep Statement")
	}

	atc := stmt.(*ast.BeepStatement)

	if atc == nil {
		t.Fatal("TestBeepStatement couldn't extract BeepStatement object")
	}

}

func Test_ChainStatement(t *testing.T) {
	tests := []struct {
		file string
	}{
		{file: `c:\menu\HCAL.BAS`},
	}

	for _, tt := range tests {
		cmd := fmt.Sprintf(`CHAIN "%s"`, tt.file)
		l := lexer.New(cmd)
		p := New(l)
		env := &object.Environment{}
		p.ParseCmd(env)
		program := env.Program

		checkParserErrors(t, p)

		itr := program.CmdLineIter()

		if itr.Len() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := itr.Value()

		if stmt.TokenLiteral() != token.BEEP {
			t.Fatal("TestChainStatement didn't get an Chain Statement")
		}

		atc := stmt.(*ast.ChainStatement)

		if atc == nil {
			t.Fatal("TestChainStatement couldn't extract ChainStatement object")
		}
	}
}

func Test_FilesCommand(t *testing.T) {
	tests := []struct {
		path string
	}{
		{`"C:\PROG\"`},
		{},
	}

	for _, tt := range tests {
		inp := "FILES"
		if len(tt.path) > 0 {
			inp = inp + " " + tt.path
		}

		l := lexer.New(inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseCmd(env)
		prog := env.Program

		checkParserErrors(t, p)

		itr := prog.CmdLineIter()

		if itr.Len() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := itr.Value()

		if stmt.TokenLiteral() != token.FILES {
			t.Fatal("Test_FilesCommand didn't get a FILES command")
		}

		fls := stmt.(*ast.FilesCommand)

		if fls == nil {
			t.Fatal("Test_FilesCommand couldn't extract FilesCommand object")
		}

		if (len(tt.path) > 0) || (len(fls.Path) > 0) {
			if strings.Compare(fls.Path, tt.path) == 0 {
				t.Fatalf("FILES cmd expected path %s, got %s\n", tt.path, fls.Path)
			}
		}
	}
}

func Test_LetStatementImplied(t *testing.T) {
	input := `10 x = 5: y = 20`

	l := lexer.New(input)
	p := New(l)
	fmt.Println("Test_LetStatementImplied Parsing")
	env := &object.Environment{}
	p.ParseProgram(env)
	program := env.Program

	checkParserErrors(t, p)

	if program == nil {
		t.Fatal("ParseProgram() returned nil")
	}

	if program.StatementIter().Len() != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", program.StatementIter().Len())
	}

	tests := []struct {
		expectedToken      string
		expectedIdentifier string
	}{
		{token.LINENUM, "10"},
		{"", "X"},
		{"", "Y"},
	}

	itr := program.StatementIter()
	for _, tt := range tests {
		stmt := itr.Value()
		itr.Next()

		_, ok := stmt.(*ast.LineNumStmt)
		if !ok {
			if !testLetStatement("", t, stmt, tt.expectedIdentifier) {
				return
			}
		}
	}
}

func Test_LetStatement(t *testing.T) {
	input := `10 let x = 5: let y$ = "test": let foobar% = 838383 : LET BANG! = 46.8 : LET POUND# = 7654321.1234`

	l := lexer.New(input)
	p := New(l)
	fmt.Println("Test_LetStatement Parsing")
	env := &object.Environment{}
	p.ParseProgram(env)
	program := env.Program

	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if program.StatementIter().Len() != 6 {
		t.Fatalf("program.Statements does not contain 4 statements. got=%d", program.StatementIter().Len())
	}

	tests := []struct {
		expectedToken      string
		expectedIdentifier string
	}{
		{token.LINENUM, "10"},
		{token.LET, "X"},
		{token.LET, "Y$"},
		{token.LET, "FOOBAR%"},
		{token.LET, "BANG!"},
		{token.LET, "POUND#"},
	}

	itr := program.StatementIter()
	for _, tt := range tests {
		stmt := itr.Value()
		itr.Next()

		_, ok := stmt.(*ast.LineNumStmt)
		if !ok {
			if !testLetStatement("LET", t, stmt, tt.expectedIdentifier) {
				return
			}
		}
	}
}

func testLetStatement(texp string, t *testing.T, s ast.Statement, name string) bool {
	if s.TokenLiteral() != texp {
		t.Errorf("s.TokenLiteral not 'let'. got=%q", s.TokenLiteral())
		return false
	}
	letStmt, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}
	if letStmt.Name.String() != strings.ToUpper(name) {
		t.Errorf("letStmt.Name.Value not '%s'. got=%s", strings.ToUpper(name), letStmt.Name.String())
		return false
	}
	if letStmt.Name.TokenLiteral() != name {
		t.Errorf("letStmt.Name.TokenLiteral() not '%s'. got=%s", name, letStmt.Name.TokenLiteral())
		return false
	}
	return true
}

func TestLetWithTypes(t *testing.T) {
	type result struct {
		expectedToken      string
		expectedIdentifier string
	}

	type results []result

	tests := []struct {
		input   string
		results results
	}{
		{input: `10 LET A$ = "a test string"`, results: results{
			{token.LINENUM, "10"},
			{token.LET, "A$"},
		},
		},
		{input: `20 LET B% = "a test string"`, results: results{
			{token.LINENUM, "10"},
			{token.LET, "B%"},
		},
		},
		{input: `30 LET C! = "a test string"`, results: results{
			{token.LINENUM, "10"},
			{token.LET, "C!"},
		},
		},
		{input: `40 LET D# = "a test string"`, results: results{
			{token.LINENUM, "10"},
			{token.LET, "D#"},
		},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program

		checkParserErrors(t, p)

		if program == nil {
			t.Fatalf("ParseProgram() returned nil")
		}

		itr := program.StatementIter()
		for _, ttt := range tt.results {
			stmt := itr.Value()
			itr.Next()

			_, ok := stmt.(*ast.LineNumStmt)
			if !ok {
				if !testLetStatement("LET", t, stmt, ttt.expectedIdentifier) {
					return
				}
			}
		}

	}
}

func TestLineNumbers(t *testing.T) {
	input := "10\n20\n30*"
	tk := token.Token{Type: token.AUTO, Literal: "AUTO"}

	l := lexer.New(input)
	p := New(l)
	fmt.Println("TestLineNumbers Parsing")
	env := &object.Environment{}
	env.SetAuto(&ast.AutoCommand{Token: tk, Start: 30, Increment: 10})
	p.ParseProgram(env)
	program := env.Program

	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if program.StatementIter().Len() != 3 {
		t.Fatalf("program.Statements does not contain 3 statements. got=%d", program.StatementIter().Len())
	}

	tests := []struct {
		expectedToken string
		expectedValue int32
	}{
		{token.LINENUM, 10},
		{token.LINENUM, 20},
		{token.LINENUM, 30},
	}

	itr := program.StatementIter()
	for _, tt := range tests {
		stmt := itr.Value()
		itr.Next()
		if !testLineNumber(t, stmt, tt.expectedValue) {
			return
		}
	}

}

func testLineNumber(t *testing.T, s ast.Statement, line int32) bool {
	lineStmt, ok := s.(*ast.LineNumStmt)
	if !ok {
		t.Errorf("s not *ast.LineNumStmt. got=%T", s)
		return false
	}
	if lineStmt.Value != line {
		t.Errorf("lineStmt.Value not '%d'. got=%d", line, lineStmt.Value)
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}
	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestDataStatement(t *testing.T) {
	tkInt := token.Token{Type: token.INT, Literal: "INT"}
	tkFixed := token.Token{Type: token.FIXED, Literal: "123.45"}
	tkString := token.Token{Type: token.STRING, Literal: "STRING"}
	tkFloatS := token.Token{Type: token.FLOAT, Literal: "3.14159E+0"}
	tkFloatD := token.Token{Type: token.FLOAT, Literal: "3.14159D+0"}
	tkDblInt := token.Token{Type: token.INTD, Literal: "INTD"}

	fixed, _ := decimal.NewFromString("123.45")

	tests := []struct {
		inp     string           // source line
		stmtNum int              // # of statements expected
		lineNum int32            // line number
		cnt     int              // number of expressions expected
		exp     []ast.Expression // expected values
	}{
		{`10 DATA "Fred", George Foreman`, 2, 10, 2, []ast.Expression{
			&ast.StringLiteral{Token: tkString, Value: "Fred"},
			&ast.StringLiteral{Token: tkString, Value: "George Foreman"},
		}},
		{`20 DATA 123, 123.45, "Fred", 99999`, 2, 20, 4, []ast.Expression{
			&ast.IntegerLiteral{Token: tkInt, Value: 123},
			&ast.FixedLiteral{Token: tkFixed, Value: fixed},
			&ast.StringLiteral{Token: tkString, Value: "Fred"},
			&ast.DblIntegerLiteral{Token: tkDblInt, Value: 99999},
		},
		},
		{`30 DATA "Fred", George : PRINT`, 3, 30, 2, []ast.Expression{
			&ast.StringLiteral{Token: tkString, Value: "Fred"},
			&ast.StringLiteral{Token: tkString, Value: "George"},
		}},
		{`40 DATA 3.14159E+0, 3.14159D+0`, 2, 40, 2, []ast.Expression{
			&ast.FloatSingleLiteral{Token: tkFloatS, Value: 3.14159},
			&ast.FloatDoubleLiteral{Token: tkFloatD, Value: 3.14159},
		},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		iter := program.StatementIter()
		if iter.Len() != tt.stmtNum {
			t.Fatalf("expected %d statements, got %d", tt.stmtNum, iter.Len())
		}
		stmt := iter.Value()

		lm, ok := stmt.(*ast.LineNumStmt)

		if !ok {
			t.Fatalf("no line number, expected %d", tt.lineNum)
		}

		if lm.Value != tt.lineNum {
			t.Fatalf("expected line %d, got %d", tt.lineNum, lm.Value)
		}

		iter.Next()
		stmt = iter.Value()

		dstmt, ok := stmt.(*ast.DataStatement)

		if !ok {
			t.Fatalf("unexpected this is")
		}

		if len(dstmt.Consts) != tt.cnt {
			t.Fatalf("expected %d constants, got %d!", tt.cnt, len(dstmt.Consts))
		}

		for i, want := range tt.exp {
			compareStatements(tt.inp, dstmt.Consts[i], want, t)
		}
	}
}

func TestDimStatement(t *testing.T) {
	type dimensions struct {
		id   string
		dims []int8
	}
	tests := []struct {
		input   string
		stmtNum int
		lineNum int32
		numIDs  int8
		dims    []dimensions
	}{
		{`10 DIM A[20]`, 2, 10, 1, []dimensions{{"A[]", []int8{20}}}},
		{`20 DIM A[20, 10]`, 2, 20, 1, []dimensions{{"A[]", []int8{20, 10}}}},
		{`30 DIM A[20, 30],B[15,5]`, 2, 30, 2, []dimensions{{"A[]", []int8{20, 30}}, {"B[]", []int8{15, 5}}}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		iter := program.StatementIter()
		if iter.Len() != tt.stmtNum {
			t.Fatalf("expected %d statements, got %d", tt.stmtNum, iter.Len())
		}
		stmt := iter.Value()

		lm, ok := stmt.(*ast.LineNumStmt)

		if !ok {
			t.Fatalf("no line number, expected %d", tt.lineNum)
		}

		if lm.Value != tt.lineNum {
			t.Fatalf("expected line %d, got %d", tt.lineNum, lm.Value)
		}

		iter.Next()
		stmt = iter.Value()

		dstmt, ok := stmt.(*ast.DimStatement)

		if !ok {
			t.Fatalf("unexpected this is")
		}

		dstmt.String()

		if int8(len(dstmt.Vars)) != tt.numIDs {
			t.Fatalf("expected %d dimensioned variables, got %d on %s", tt.numIDs, len(dstmt.Vars), tt.input)
		}

		for dNum, d := range tt.dims {
			if dstmt.Vars[dNum].Value != d.id {
				t.Fatalf("got id %s, expected %s on line %s", dstmt.Vars[dNum].Value, d.id, tt.input)
			}

			for dnum, dim := range d.dims {
				indExp, ok := dstmt.Vars[dNum].Index[dnum].Index.(*ast.IntegerLiteral)

				if !ok {
					t.Fatalf("dimension %d for %s is not an index", dnum, tt.input)
				}

				if int8(indExp.Value) != dim {
					t.Fatalf("expeced dimension %d, got %d, on %s", dim, indExp.Value, tt.input)
				}
			}
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := "10 foobar"
	l := lexer.New(input)
	p := New(l)
	env := &object.Environment{}
	p.ParseProgram(env)
	program := env.Program
	checkParserErrors(t, p)
	if program.StatementIter().Len() != 2 {
		t.Fatalf("program has not enough statements. got=%d", program.StatementIter().Len())
	}

	iter := program.StatementIter()
	iter.Next()
	step := iter.Value()
	stmt, ok := step.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[1] is not ast.ExpressionStatement. got=%T", step)
	}
	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}
	if ident.Value != "FOOBAR" {
		t.Errorf("ident.Value not %s. got=%s", "FOOBAR", ident.Value)
	}
	if ident.TokenLiteral() != "FOOBAR" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "FOOBAR", ident.TokenLiteral())
	}
}

func TestReadStatement(t *testing.T) {
	tkAs := token.Token{Type: token.IDENT, Literal: "A$"}
	tkBs := token.Token{Type: token.IDENT, Literal: "B$"}

	tests := []struct {
		inp     string
		stmtNum int              // expected count of statments
		lineNum int32            // line number
		vars    int              // number of expressions expected
		exp     []ast.Expression // expected values
	}{
		{`10 READ A$`, 2, 10, 1, []ast.Expression{
			&ast.Identifier{Token: tkAs, Value: "A$", Type: "$"},
		}},
		{`20 READ A$, B$`, 2, 20, 2, []ast.Expression{
			&ast.Identifier{Token: tkAs, Value: "A$", Type: "$"},
			&ast.Identifier{Token: tkBs, Value: "B$", Type: "$"},
		}},
		{`30 READ A$, B$ : PRINT "Hello"`, 3, 30, 2, []ast.Expression{
			&ast.Identifier{Token: tkAs, Value: "A$", Type: "$"},
			&ast.Identifier{Token: tkBs, Value: "B$", Type: "$"},
		}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		iter := program.StatementIter()
		if iter.Len() != tt.stmtNum {
			t.Fatalf("expected %d statements, got %d", tt.stmtNum, iter.Len())
		}
		stmt := iter.Value()

		lm, ok := stmt.(*ast.LineNumStmt)

		if !ok {
			t.Fatalf("no line number, expected %d", tt.lineNum)
		}

		if lm.Value != tt.lineNum {
			t.Fatalf("expected line %d, got %d", tt.lineNum, lm.Value)
		}

		iter.Next()
		stmt = iter.Value()

		rstmt, ok := stmt.(*ast.ReadStatement)

		if !ok {
			t.Fatalf("unexpected this is")
		}

		if len(rstmt.Vars) != tt.vars {
			t.Fatalf("expected %d contants, got %d!", tt.vars, len(rstmt.Vars))
		}

		/*for i, want := range tt.exp {
			compareStatements(tt.inp, rstmt.[i], want, t)
		}*/
	}
}

func TestRemStatement(t *testing.T) {
	tests := []struct {
		inp string
		res string
	}{
		{inp: "10 REM A code comment", res: "REM A code comment"},
		{inp: "20 REM", res: "REM"},
		{inp: "30 ' Alternate form remark", res: "' Alternate form remark"},
		{inp: "40 ' Once a remark : GOTO 20", res: "' Once a remark : GOTO 20"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)

		itr := program.StatementIter()
		itr.Next()
		stmt := itr.Value()

		if strings.Compare(stmt.String(), tt.res) != 0 {
			t.Fatalf("REM stmt expected %s, got %s", tt.res, stmt.String())
		}
	}
}

func TestRestore(t *testing.T) {
	rsTk := token.Token{Type: token.RESTORE, Literal: "RESTORE"}

	tests := []struct {
		inp string
		exp interface{}
	}{
		{`10 RESTORE`, &ast.RestoreStatement{Token: rsTk, Line: -1}},
		{`20 RESTORE 300`, &ast.RestoreStatement{Token: rsTk, Line: 300}},
		{`30 RESTORE X`, nil},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		if tt.exp != nil {
			checkParserErrors(t, p)
		} else {
			if len(p.errors) == 0 {
				t.Fatalf("%s parsed but should have failed!", tt.inp)
			}
		}

		itr := program.StatementIter()
		itr.Next()
		stmt := itr.Value()

		if tt.exp != nil {
			compareStatements(tt.inp, stmt, tt.exp, t)
		}
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `10 "hello world"`
	l := lexer.New(input)
	p := New(l)
	env := &object.Environment{}
	p.ParseProgram(env)
	program := env.Program
	checkParserErrors(t, p)
	iter := program.StatementIter()

	iter.Next()
	step := iter.Value()
	stmt, ok := step.(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", step)
	}
	literal, ok := stmt.Expression.(*ast.StringLiteral)

	if !ok {
		t.Fatalf("program.Statements[1] is not an ast.StringLiteral.  got=%T", step)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestTronTroffCommands(t *testing.T) {
	tests := []struct {
		inp string
		tok string
	}{
		{"TRON", token.TRON},
		{"TROFF", token.TROFF},
	}

	fmt.Println("TestTronTroffCommands Parsing")
	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseCmd(env)
		program := env.Program

		checkParserErrors(t, p)

		itr := program.CmdLineIter()

		if itr.Len() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := itr.Value()

		if stmt.TokenLiteral() != tt.tok {
			t.Fatalf("TestTronTroffCommands didn't get an %s command", tt.inp)
		}
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	intTok := token.Token{Type: token.INT, Literal: "5"}
	dblTok := token.Token{Type: token.INTD, Literal: "65999"}
	fltTok := token.Token{Type: token.FLOAT, Literal: "4294967295"}

	tests := []struct {
		inp   string
		stmts int
		lit   interface{}
	}{
		{`10 5`, 2, &ast.IntegerLiteral{Value: 5, Token: intTok}},
		{`20 65999#`, 2, &ast.DblIntegerLiteral{Value: 65999, Token: dblTok}},
		{`30 4294967295`, 2, &ast.FloatSingleLiteral{Token: fltTok, Value: 4294967295}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		if program.StatementIter().Len() != tt.stmts {
			t.Fatalf("program has not enough statements. got=%d", program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		step := iter.Value()
		stmt, ok := step.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[1] is not ast.ExpressionStatement. got=%T", step)
		}

		compareStatements(tt.inp, tt.lit, stmt, t)
	}
}

func TestHexOctalConstants(t *testing.T) {
	tests := []struct {
		inp string
		lit interface{}
	}{
		{"10 &HF76F", &ast.HexConstant{Value: "F76F"}},
		{"20 &HF7F6F", &ast.HexConstant{Value: "F7F6F"}},
		{"30 &767", &ast.OctalConstant{Value: "767"}},
		{"30 &O767", &ast.OctalConstant{Value: "767"}},
		{"40 &F767", nil},
		{"30 &O F767", nil},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program

		if tt.lit == nil {
			if len(p.errors) != 1 {
				t.Fatalf("%s passed and it shouldn't", tt.inp)
			}
		} else {
			checkParserErrors(t, p)
			if program.StatementIter().Len() != 2 {
				t.Fatalf("program has not enough statements. got=%d", program.StatementIter().Len())
			}

			iter := program.StatementIter()
			iter.Next()
			step := iter.Value()
			stmt, ok := step.(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("program.Statements[1] is not ast.ExpressionStatement. got=%T", step)
			}

			compareStatements(tt.inp, tt.lit, stmt, t)
		}
	}
}

type parseFunc func(*Parser) ast.Expression

func TestNumericConversion(t *testing.T) {
	tests := []struct {
		input string
		tok   token.TokenType
		fn    parseFunc
		res   string
	}{
		{"235.988E-7", token.FLOAT, func(p *Parser) ast.Expression {
			return p.parseFloatingPointLiteral()
		}, "235.988E-7"},
		{"235.988D-7", token.FLOAT, func(p *Parser) ast.Expression {
			return p.parseFloatingPointLiteral()
		}, "235.988D-7"},
		{"53a", token.INT, func(p *Parser) ast.Expression {
			return p.parseIntegerLiteral()
		}, ""},
		{"62.4d5", token.FIXED, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, ""},
		{"53", token.INT, func(p *Parser) ast.Expression {
			return p.parseIntegerLiteral()
		}, "53"},
		{"62.45", token.FIXED, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, "62.45"},
		{"62.", token.INT, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, "62"},
		{"62.45.37", token.INT, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, ""},
		{"624537", token.INT, func(p *Parser) ast.Expression {
			return p.parseIntegerLiteral()
		}, "624537!"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		// this is where I cheat
		p.curToken.Type = tt.tok
		p.curToken.Literal = tt.input

		res := tt.fn(p)

		if (tt.res == "") && (res != nil) {
			t.Errorf("Parse succeeded when it should have failed, %s", tt.input)
		}

		if (tt.res == "") && (len(p.errors) == 0) {
			t.Errorf("Parse failed to report error")
			break
		}

		if tt.res != "" {
			//fmt.Printf("got %T", res)
			if tt.res != res.String() {
				t.Errorf("expected %s, got %s", tt.res, res.String())
			}
		}
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input        string
		operator     string
		integerValue int16
	}{
		{"10 -15", "-", 15},
	}
	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		if program.StatementIter().Len() != 2 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 2, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		step := iter.Value()
		stmt, ok := step.(*ast.ExpressionStatement)

		if !ok {
			t.Fatalf("program.Statements[1] is not ast.ExpressionStatement. got=%T", step)
		}
		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T = %s", stmt.Expression, stmt.String())
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
		if !testIntegerLiteral(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int16) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}
	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}
	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value, integ.TokenLiteral())
		return false
	}
	return true
}

func TestParsingInfixExpressions(t *testing.T) {

	infixTests := []struct {
		input      string
		leftValue  int16
		operator   string
		rightValue int16
		lineNum    int32
	}{
		{"10 5 + 5", 5, "+", 5, 10},
		{"20 5 - 5", 5, "-", 5, 20},
		{"30 5 * 5", 5, "*", 5, 30},
		{"40 5 / 5", 5, "/", 5, 40},
		{"50 5 > 5", 5, ">", 5, 50},
		{"60 5 < 5", 5, "<", 5, 60},
		{"80 5 <> 5", 5, "<>", 5, 80},
	}
	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		if program.StatementIter().Len() != 2 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 2, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		step := iter.Value()
		stmt, ok := step.(*ast.LineNumStmt)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.LineNumStmt. got=%T", step)
		}
		if stmt.Value != tt.lineNum {
			t.Fatalf("wrong line number, expected %d, got %d\n", tt.lineNum, stmt.Value)
		}

		iter.Next()
		step = iter.Value()
		stmt2, ok := step.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[1] is not ast.ExpressionStatement. got=%T, line %d", step, tt.lineNum)
		}
		exp, ok := stmt2.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("exp is not ast.InfixExpression. got=%T", stmt2.Expression)
		}
		if !testIntegerLiteral(t, exp.Left, tt.leftValue) {
			fmt.Println("exiting at first testIntegerLiteral")
			return
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s", tt.operator, exp.Operator)
		}
		if !testIntegerLiteral(t, exp.Right, tt.rightValue) {
			fmt.Println("exiting at second testIntegerLiteral")
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"10 -a * b", "10 -A * B"},
		{"10 a + b + c", "10 A + B + C"},
		{"10 a + b - c", "10 A + B - C"},
		{"10 a * b * c", "10 A * B * C"},
		{"10 a * b / c", "10 A * B / C"},
		{"10 a + b / c", "10 A + B / C"},
		{"10 a + b * c + d / e - f", "10 A + B * C + D / E - F"},
		{"10 5 > 4 = 3 < 4", "10 5 > 4 = 3 < 4"},
		{"20 ((5 < 4) <> (3 > 4))", "20 ((5 < 4) <> (3 > 4))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}
func TestParsingIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		literal  string
		typ      string
		indCount int
		indVal   []string
	}{
		{"10 LET simpleArray[x] = 5", "SIMPLEARRAY[]", "", 1, []string{"X"}},
		{"20 LET myArray[0,1] = 5", "MYARRAY[]", "", 2, []string{"0", "1"}},
		{"30 impliedArray[4,3] = 5", "IMPLIEDARRAY[]", "", 2, []string{"4", "3"}},
		{`40 str$ = "Hello"`, "STR$", "$", 0, nil},
		{`50 num% = 46`, "NUM%", "%", 0, nil},
		{`60 sng! = 3.14E+0`, "SNG!", "!", 0, nil},
		{`70 dbl# = 3.14159E+0`, "DBL#", "#", 0, nil},
		{`80 LET A[0] = 5 : LET A[1] = 2`, "A[]", "", 1, []string{"0"}},
		{`90 LET A$[0] = "Hello"`, "A$[]", "$", 1, []string{"0"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)
		iter := program.StatementIter()
		if iter.Len() < 2 {
			t.Fatalf("got %d expressions, wanted %d", iter.Len(), 2)
		}
		iter.Next()
		stmt := iter.Value().(*ast.LetStatement)
		if stmt == nil {
			t.Fatalf("got %T, was expecting *ast.LetStatement", iter.Value())
		}
		if stmt.Name.Value != tt.literal {
			t.Fatalf("got name value %s was expecting %s", stmt.Name.Value, tt.literal)
		}
		if stmt.Name.Type != tt.typ {
			t.Fatalf("got type %s, was expecting %s", stmt.Name.Type, tt.typ)
		}
		if len(stmt.Name.Index) != tt.indCount {
			t.Fatalf("got %d indicies, expected %d", len(stmt.Name.Index), tt.indCount)
		}

		for i, dim := range tt.indVal {
			if stmt.Name.Index[i].Index.String() != dim {
				t.Fatalf("index %d, got expression %s, expected %s", i, stmt.Name.Index[i].Index.String(), dim)
			}
		}
	}
}

func TestIfExpression(t *testing.T) {
	tests := []struct {
		input string
		cons  string
		alt   string
		op    string
	}{
		{"10 IF X < Y THEN 300", "GOTO", "nil", "<"},
		{"20 IF (X < Y) GOTO 300", "GOTO", "nil", "<"},
		{"30 IF X > Y THEN 300 ELSE 400", "GOTO", "GOTO", ">"},
		{"40 IF X >= Y THEN END", "END", "nil", ">="},
		{"50 IF X < Y THEN 300 ELSE END", "GOTO", "END", "<"},
		{"60 IF X < Y, THEN 300 ELSE END", "GOTO", "END", "<"},
		{"70 IF X = Y, THEN 300 ELSE END", "GOTO", "END", "="},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)

		if program.StatementIter().Len() != 2 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n", 2, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		stmt := iter.Value()

		stmt1, ok := stmt.(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[1] is not ast.ExpressionStatement. got=%T", stmt)
		}

		exp, ok := stmt1.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T", stmt1.Expression)
		}

		gexp, ok := exp.Condition.(*ast.GroupedExpression)
		if ok {
			iexp, ok := gexp.Exp.(*ast.InfixExpression)

			if ok {
				if !testInfixExpression(t, iexp, "X", tt.op, "Y") {
					return
				}
			}
		} else {
			if !testInfixExpression(t, exp.Condition, "X", tt.op, "Y") {
				return
			}
		}

		if !testIfConsequence(t, tt.cons, exp.Consequence) {
			return
		}

		if !testIfAlternative(t, tt.alt, exp.Alternative) {
			return
		}
	}
}

func TestGotoStatements(t *testing.T) {
	tests := []struct {
		input         string
		expStmts      int
		expectedValue string
	}{
		{"10 GOTO 100", 2, "100"},
		{"20 GOTO 100 : GOTO 200", 3, "100"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)

		if program.StatementIter().Len() != tt.expStmts {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", tt.expStmts, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		stmt := iter.Value()
		gotoStmt, ok := stmt.(*ast.GotoStatement)
		if !ok {
			t.Fatalf("stmt not *ast.GotoStatement. got=%T", stmt)
		}
		if gotoStmt.TokenLiteral() != "GOTO" {
			t.Fatalf("returnStmt.TokenLiteral not 'GOTO', got %q", gotoStmt.TokenLiteral())
		}
		if gotoStmt.Goto != tt.expectedValue {
			t.Fatalf("expected linenum %s, got %s", tt.expectedValue, gotoStmt.Goto)
		}
	}
}

func TestGosubStatements(t *testing.T) {
	tests := []struct {
		input         string
		expStmts      int
		expectedValue string
	}{
		{"10 GOSUB 100", 2, "100"},
		{"20 GOSUB 100 : GOSUB 200", 3, "100"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)

		if program.StatementIter().Len() != tt.expStmts {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", tt.expStmts, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		stmt := iter.Value()
		gosubStmt, ok := stmt.(*ast.GosubStatement)
		if !ok {
			t.Fatalf("stmt not *ast.GosubStatement. got=%T", stmt)
		}
		if gosubStmt.TokenLiteral() != "GOSUB" {
			t.Fatalf("returnStmt.TokenLiteral not 'GOSUB', got %q", gosubStmt.TokenLiteral())
		}
		if gosubStmt.Gosub != tt.expectedValue {
			t.Fatalf("expected linenum %s, got %s", tt.expectedValue, gosubStmt.Gosub)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expStmts      int
		expectedValue interface{}
	}{
		{"10 return 5", 2, "5"},
		{"20 return", 2, ""},
		{"30 return : return", 3, ""},
		{"40 return 10: return", 3, "10"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)

		if program.StatementIter().Len() != tt.expStmts {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", tt.expStmts, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		stmt := iter.Value()
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Fatalf("stmt not *ast.ReturnStatement. got=%T", stmt)
		}
		if returnStmt.TokenLiteral() != "RETURN" {
			t.Fatalf("returnStmt.TokenLiteral not 'RETURN', got %q", returnStmt.TokenLiteral())
		}
		if returnStmt.ReturnTo != tt.expectedValue {
			t.Fatalf("got return to %T, expected %T", returnStmt.ReturnTo, tt.expectedValue)
			return
		}
	}
}

func TestRunCommand(t *testing.T) {
	tests := []struct {
		inp   string
		start int
		file  string
	}{
		{"RUN", 0, ""},
		{"RUN 20", 20, ""},
		{"RUN \"TESTFILE.BAS\"", 0, "TESTFILE.BAS"},
	}

	fmt.Println("TestRunCommand Parsing")
	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseCmd(env)
		program := env.Program

		checkParserErrors(t, p)

		itr := program.CmdLineIter()

		if itr.Len() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := itr.Value()

		if stmt.TokenLiteral() != token.RUN {
			t.Fatal("TestRunCommand didn't get an Run command")
		}

		atc := stmt.(*ast.RunCommand)

		if atc == nil {
			t.Fatal("TestRunCommand couldn't extract AutoCommand object")
		}

		if atc.StartLine != tt.start {
			t.Fatalf("TestRunCommand got start = %d, expected %d", atc.StartLine, tt.start)
		}

		if atc.LoadFile != tt.file {
			t.Fatalf("TestRunCommand got LoadFile = %s, expected %s", atc.LoadFile, tt.file)
		}
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		errCount int
	}{
		{"10 DEF FNID(x) = x : FNID(5)", 0},
		{"20 DEF FNMUL(x,y) = x*y : FNMUL(2,3)", 0},
		{"30 DEF FNSKIP(x)= (x + 2): FNSKIP(3)", 0},
		{"40 DEF FN(z) = z + 2", 1},
		{"50 DEF AFUNC(t) = t * 5", 1},
		{"60 DEF FNMUL(x,y)", 1},
		{"70 DEF FNMUL  = 5", 1},
		{"80 DEF FNMUL(x,y)", 1},
		{"90 DEF FNMUL(x,y = x * y", 1},
		{"100 DEF FNMUL() = x * y", 0},
		{"110 MKD$(65999)", 0},
	}
	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program

		if len(p.errors) != tt.errCount {
			t.Fatalf("expected %d errors, got %d instead on test %d", tt.errCount, len(p.errors), i)
		}

		if program.StatementIter().Len() == 0 {
			t.Fatalf("parser failed to produce statements")
		}
	}
}

func TestEndStatements(t *testing.T) {
	tests := []struct {
		input    string
		expStmts int
	}{
		{"10 END", 2},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program
		checkParserErrors(t, p)

		if program.StatementIter().Len() != tt.expStmts {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", tt.expStmts, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		stmt := iter.Value()
		endStmt, ok := stmt.(*ast.EndStatement)
		if !ok {
			t.Fatalf("stmt not *ast.EndStatement. got=%T", stmt)
		}
		if endStmt.TokenLiteral() != "END" {
			t.Fatalf("endStmt.TokenLiteral not 'END', got %q", endStmt.TokenLiteral())
		}
	}
}

func TestCls(t *testing.T) {
	tests := []struct {
		input string
		param int
	}{
		{"CLS", -1},
		{"CLS 0", 0},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		//l.NextToken()
		p := New(l)
		env := &object.Environment{}
		p.ParseCmd(env)
		program := env.Program
		checkParserErrors(t, p)

		if program.CmdLineIter().Len() != 1 {
			t.Fatalf("program.Statements does not contain single command")
		}

		iter := program.CmdLineIter()
		stmt := iter.Value()
		clsStmt, ok := stmt.(*ast.ClsStatement)
		if !ok {
			t.Fatalf("stmt not *ast.ClsStatement. got=%T", stmt)
		}
		if clsStmt.TokenLiteral() != "CLS" {
			t.Fatalf("clsStmt.TokenLiteral not 'CLS', got %q", clsStmt.TokenLiteral())
		}
		if tt.param != clsStmt.Param {
			t.Fatalf("cls param expected %d, got %d", tt.param, clsStmt.Param)
		}
	}
}

func TestFilesCommand(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`20 FILES`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		//program := env.Program

	}
}

func TestPrintStatements(t *testing.T) {
	tests := []struct {
		input    string
		expStmts int
	}{
		{`5 PRINT X * Y`, 2},
		{`7 PRINT (X * Y)`, 2},
		{`10 PRINT "Hello World!`, 2},
		{`20 PRINT "This is ";"a test"`, 2},
		{`30 PRINT "Another test " "program."`, 2},
		{`40 PRINT "Test of tab","due to comma"`, 2},
		{`50 PRINT "Test of a run on";`, 2},
		{`60 PRINT " sentence"`, 2},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := &object.Environment{}
		p.ParseProgram(env)
		program := env.Program

		if program.StatementIter().Len() != tt.expStmts {
			t.Fatalf("program.Statements does not contain %d statements. got=%d", tt.expStmts, program.StatementIter().Len())
		}

		iter := program.StatementIter()
		iter.Next()
		stmt := iter.Value()

		fmt.Printf("stmt[1] = %T\n", stmt)
	}
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{},
	operator string, right interface{}) bool {

	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	//	et := exp.(type)
	//	fmt.Printf("expecting a %T\n", et)
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int16(v))
	case string:
		return testIdentifier(t, exp, v)
	case nil:
		return exp == nil
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}

	return true
}

func testIfConsequence(t *testing.T, exp string, stmt ast.Statement) bool {

	return testIfResult(t, "Consequence", exp, stmt)
}

func testIfAlternative(t *testing.T, exp string, stmt ast.Statement) bool {
	// the one result that is not shared with Consequence
	if exp == "nil" {
		if nil == stmt {
			return true
		}
		t.Errorf("exp.Alternative.Statements was not %s. got=%+v", exp, stmt)
		return false
	}
	return testIfResult(t, "Alternative", exp, stmt)
}

func testIfResult(t *testing.T, rt string, exp string, stmt ast.Statement) bool {
	var ok bool
	switch exp {
	case "GOTO":
		_, ok = stmt.(*ast.GotoStatement)
	case "END":
		_, ok = stmt.(*ast.EndStatement)
	}

	if !ok {
		t.Errorf("exp.%s.Statements was not %s. got=%+v", rt, exp, stmt)
		return false
	}

	return true
}

func TestListStatement(t *testing.T) {
	tests := []struct {
		inp string
		res *ast.ListStatement
	}{
		{"LIST", &ast.ListStatement{
			Token:  token.Token{Type: token.LIST, Literal: "LIST"},
			Start:  "",
			Lrange: "",
			Stop:   "",
		}},
		{"LIST 50", &ast.ListStatement{
			Token:  token.Token{Type: token.LIST, Literal: "LIST"},
			Start:  "50",
			Lrange: "",
			Stop:   "",
		}},
		{"LIST 50-", &ast.ListStatement{
			Token:  token.Token{Type: token.LIST, Literal: "LIST"},
			Start:  "50",
			Lrange: "-",
			Stop:   "",
		}},
		{"LIST 50-100", &ast.ListStatement{
			Token:  token.Token{Type: token.LIST, Literal: "LIST"},
			Start:  "50",
			Lrange: "-",
			Stop:   "100",
		}},
		{"LIST -100", &ast.ListStatement{
			Token:  token.Token{Type: token.LIST, Literal: "LIST"},
			Start:  "",
			Lrange: "-",
			Stop:   "100",
		}},
		{"LIST -", &ast.ListStatement{ // this is actually valid, same as "LIST"
			Token:  token.Token{Type: token.LIST, Literal: "LIST"},
			Start:  "",
			Lrange: "-",
			Stop:   "",
		}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := &object.Environment{}
		p.ParseCmd(env)
		program := env.Program

		itr := program.CmdLineIter()
		stmt := itr.Value()

		if strings.Compare(stmt.TokenLiteral(), tt.res.TokenLiteral()) != 0 {
			t.Fatalf("Parse(%s), expected Literal %s, got %s", tt.inp, tt.res.TokenLiteral(), stmt.TokenLiteral())
		}

		cmd, ok := stmt.(*ast.ListStatement)

		if !ok {
			t.Fatalf("stmt failed to conver to ListStatement")
		}

		if strings.Compare(tt.res.Start, cmd.Start) != 0 {
			t.Fatalf("Parse(%s), expected Start = %s, got %s", tt.inp, tt.res.Start, cmd.Start)
		}

		if strings.Compare(tt.res.Lrange, cmd.Lrange) != 0 {
			t.Fatalf("Parse(%s), expected Lrange = %s, got %s", tt.inp, tt.res.Lrange, cmd.Lrange)
		}

		if strings.Compare(tt.res.Stop, cmd.Stop) != 0 {
			t.Fatalf("Parse(%s), expected Stop = %s, got %s", tt.inp, tt.res.Stop, cmd.Stop)
		}
	}
}

func compareStatements(inp string, got interface{}, want interface{}, t *testing.T) {
	switch wantVal := want.(type) {
	case *ast.IntegerLiteral:
		gotInt, ok := got.(*ast.IntegerLiteral)

		if !ok {
			t.Fatalf("got incorrect statement from %s, got %T, wanted %T", inp, got, want)
		}

		if gotInt.Value != wantVal.Value {
			t.Fatalf("bad value from %s, got %d, wanted %d", inp, gotInt.Value, wantVal.Value)
		}

		if gotInt.Token.Literal != wantVal.Token.Literal {
			t.Fatalf("unexpected token from %s, got %s, expected %s", inp, gotInt.Token.Literal, wantVal.Token.Literal)
		}
	case *ast.FixedLiteral:
		gotFixed, ok := got.(*ast.FixedLiteral)

		if !ok {
			t.Fatalf("got incorrect statement from %s, got %T, wanted %T", inp, got, want)
		}

		if gotFixed.Value != wantVal.Value {
			t.Fatalf("bad value from %s, got %d, wanted %d", inp, gotFixed.Value, wantVal.Value)
		}

		if gotFixed.Token.Literal != wantVal.Token.Literal {
			t.Fatalf("unexpected token from %s, got %s, expected %s", inp, gotFixed.Token.Literal, wantVal.Token.Literal)
		}
	case *ast.FloatSingleLiteral:
		gotFloat, ok := got.(*ast.FloatSingleLiteral)

		if !ok {
			t.Fatalf("got incorrect statement from %s, got %T, wanted %T", inp, got, want)
		}

		if gotFloat.Value != wantVal.Value {
			t.Fatalf("bad value from %s, got %f, wanted %f", inp, gotFloat.Value, wantVal.Value)
		}

		if gotFloat.Token.Literal != wantVal.Token.Literal {
			t.Fatalf("unexpected token from %s, got %s, expected %s", inp, gotFloat.Token.Literal, wantVal.Token.Literal)
		}
	case *ast.FloatDoubleLiteral:
		gotFloat, ok := got.(*ast.FloatDoubleLiteral)

		if !ok {
			t.Fatalf("got incorrect statement from %s, got %T, wanted %T", inp, got, want)
		}

		if gotFloat.Value != wantVal.Value {
			t.Fatalf("bad value from %s, got %f, wanted %f", inp, gotFloat.Value, wantVal.Value)
		}

		if gotFloat.Token.Literal != wantVal.Token.Literal {
			t.Fatalf("unexpected token from %s, got %s, expected %s", inp, gotFloat.Token.Literal, wantVal.Token.Literal)
		}
	case *ast.DblIntegerLiteral:
		gotInt, ok := got.(*ast.DblIntegerLiteral)

		if !ok {
			t.Fatalf("got incorrect statement from %s, got %T, wanted %T", inp, got, want)
		}

		if gotInt.Value != wantVal.Value {
			t.Fatalf("bad value from %s, got %d, wanted %d", inp, gotInt.Value, wantVal.Value)
		}

		if gotInt.Token.Literal != wantVal.Token.Literal {
			t.Fatalf("unexpected token from %s, got %s, expected %s", inp, gotInt.Token.Literal, wantVal.Token.Literal)
		}
	case *ast.StringLiteral:
		gotString, ok := got.(*ast.StringLiteral)

		if !ok {
			t.Fatalf("got incorrect statement from %s, got %T, wanted %T", inp, got, want)
		}

		if gotString.Value != wantVal.Value {
			t.Fatalf("bad value from %s, got %s, wanted %s", inp, gotString.Value, wantVal.Value)
		}
	case *ast.RestoreStatement:
		gotRestore, ok := got.(*ast.RestoreStatement)

		if !ok {
			t.Fatalf("got incorrect statement from %s, got %T, wanted %T", inp, got, want)
		}

		if gotRestore.Line != wantVal.Line {
			t.Fatalf("bad value from %s, got %d, wanted %d", inp, gotRestore.Line, wantVal.Line)
		}
	}
}
