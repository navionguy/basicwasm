package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/settings"
	"github.com/navionguy/basicwasm/token"
	"github.com/stretchr/testify/assert"
)

func TestAutoCommand(t *testing.T) {
	tests := []struct {
		inp    string
		params []ast.Expression
	}{
		{inp: "AUTO 10, 10, 10", params: []ast.Expression{
			&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "10"}, Value: 10},
			&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "10"}, Value: 10},
		}},
		{inp: "AUTO"},
		{inp: "AUTO 20", params: []ast.Expression{&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "20"}, Value: 20}}},
		{inp: "AUTO , 20", params: []ast.Expression{nil, &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "20"}, Value: 20}}},
		{inp: "AUTO ., 20", params: []ast.Expression{&ast.Identifier{Value: "."}, &ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "20"}, Value: 20}}},
		{inp: "AUTO .", params: []ast.Expression{&ast.Identifier{Value: "."}}},
	}

	for _, tt := range tests {
		mt := mocks.MockTerm{}
		env := object.NewTermEnvironment(mt)
		cl := ParseInput(tt.inp, env)

		assert.EqualValues(t, 1, cl.LineLength())

		stmt := cl.NextStatement()

		if stmt.TokenLiteral() != token.AUTO {
			t.Fatal("TestAutoCommand didn't get an Auto command")
		}

		atc := stmt.(*ast.AutoCommand)

		assert.NotNil(t, atc, "couldn't extract AutoCommand object")

		assert.EqualValues(t, len(tt.params), len(atc.Params), tt.inp)

		for i, p := range atc.Params {
			assert.EqualValuesf(t, tt.params[i], p, "param %d didn't match expected", i)
		}
	}
}

func Test_BeepStatement(t *testing.T) {
	tests := []struct {
		inp   string
		trash bool
	}{
		{inp: "BEEP"},
		{inp: "BEEP BEEP", trash: true},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)
		assert.EqualValues(t, 1, sl.LineLength())

		stmt := sl.NextStatement()
		assert.Equalf(t, token.BEEP, stmt.TokenLiteral(), "TestBeepStatement didn't get an Beep Statement")

		atc := stmt.(*ast.BeepStatement)
		assert.NotNil(t, atc, "TestBeepStatement didn't get an Beep Statement")

		if tt.trash {
			assert.True(t, atc.HasTrash(), tt.inp)
		}
	}

}

func Test_BuiltinExpression(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: `ABS(5)`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		if sl.LineLength() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := sl.NextStatement()

		exp, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok, "Test_BuiltinExpression didn't get ExpressionStatement")

		assert.Equal(t, " = ABS(5)", exp.String(), "unexpected Builtin")
	}
}

func Test_ChainStatement(t *testing.T) {
	tests := []struct {
		cmd   string // command to parse
		exp   string
		trash bool // I expect to have trash
	}{
		{cmd: `CHAIN`, exp: ` CHAIN`, trash: true},
		{cmd: `CHAIN MERGE`, exp: `CHAIN MERGE`, trash: true},
		{cmd: `CHAIN "MENU.BAS"`, exp: `CHAIN "MENU.BAS"`},
		{cmd: `CHAIN "MENU2.BAS", PRINT`, exp: `CHAIN "MENU2.BAS", PRINT`},
		{cmd: `CHAIN "MENU.BAS", 10`, exp: `CHAIN "MENU.BAS", 10`},
		{cmd: `CHAIN "MENU.BAS",, all`, exp: `CHAIN "MENU.BAS",, ALL`},
		{cmd: `CHAIN "MENU2.BAS",, all OPEN`, exp: `CHAIN "MENU2.BAS",, ALL OPEN`, trash: true},
		{cmd: `CHAIN "C:\MENU\HCAL.BAS", 100,all,delete 100-1000`, exp: `CHAIN "C:\MENU\HCAL.BAS", 100, ALL, DELETE 100 - 1000`},
		{cmd: `CHAIN "C:\MENU\HCAL.BAS", 100,all,delete 100-1000 PRINT`, exp: `CHAIN "C:\MENU\HCAL.BAS", 100, ALL, DELETE 100 - 1000 PRINT`, trash: true},
		{cmd: `CHAIN MERGE "C:\MENU\HIWORLD.BAS"`, exp: `CHAIN MERGE "C:\MENU\HIWORLD.BAS"`},
		{cmd: `CHAIN "C:\MENU\START.BAS", 100,fred`, exp: `CHAIN "C:\MENU\START.BAS", 100 fred`, trash: true},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.cmd, env)

		if sl.LineLength() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := sl.NextStatement()
		assert.Equal(t, tt.exp, stmt.String(), "chain failed")

		chain, ok := stmt.(*ast.ChainStatement)
		assert.True(t, ok, "didn't get a chain statement")
		if tt.trash {
			assert.True(t, chain.HasTrash(), "expected to get trash")
		}
	}
}

func Test_ChrS(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: `X$ = CHR$(20)`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		if sl.LineLength() != 1 {
			t.Fatalf("program.Statements does not contain single command")
		}

		stmt := sl.NextStatement()
		_, ok := stmt.(*ast.LetStatement)

		assert.True(t, ok, "fail")

	}
}

func Test_ChDir(t *testing.T) {
	tests := []struct {
		inp string
		exp []ast.Expression
	}{
		{inp: `CHDIR`},
		{inp: `CHDIR "D:\"`, exp: []ast.Expression{&ast.StringLiteral{Value: `D:\`}}},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		if sl.LineLength() != 1 {
			t.Fatalf("program.Statements does not contain single command")
		}

		stmt := sl.NextStatement()
		cdstmt, ok := stmt.(*ast.ChDirStatement)

		assert.True(t, ok, "stmt not *ast.ChiDirStatement. got=%T", stmt)

		assert.Equal(t, "CHDIR", cdstmt.TokenLiteral(), "CHDIR token literal wrong")

		assert.Equal(t, len(tt.exp), len(cdstmt.Path), "CHDIR unexpected path")
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: "CLOSE #1"},
		{inp: "CLOSE 12"},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		stmt := sl.NextStatement()

		assert.Equal(t, tt.inp, stmt.String(), "Close failed to parse")
	}
}

func TestCls(t *testing.T) {
	tests := []struct {
		inp   string
		param int
	}{
		{"CLS", -1},
		{"CLS 0", 0},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		if sl.LineLength() != 1 {
			t.Fatalf("program.Statements does not contain single command")
		}

		stmt := sl.NextStatement()
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

func Test_ColorStatement(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: "COLOR 1,2,3"},
		{inp: "COLOR ,,3"},
		{inp: "COLOR"},
		{inp: "COLOR 1,2,3,4"},
		{inp: "COLOR 1,2,"},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		if sl.LineLength() != 1 {
			t.Fatalf("program.Statements does not contain single command")
		}

		stmt := sl.NextStatement()
		colorStmt, ok := stmt.(*ast.ColorStatement)
		if !ok {
			t.Fatalf("stmt not *ast.ClsStatement. got=%T", stmt)
		}
		if colorStmt.TokenLiteral() != "COLOR" {
			t.Fatalf("colorStmt.TokenLiteral not 'COLOR', got %q", colorStmt.TokenLiteral())
		}
	}
}

func Test_Commands(t *testing.T) {
	tests := []struct {
		inp string
		tk  string
		lst string
	}{
		{inp: "CLEAR", tk: token.CLEAR, lst: "CLEAR "},
		{inp: "CLEAR 32767", tk: token.CLEAR, lst: "CLEAR 32767"},
		{inp: "CLEAR 2,32767,32767", tk: token.CLEAR, lst: "CLEAR 2,32767,32767"},
		{inp: "CLEAR ,32767", tk: token.CLEAR, lst: "CLEAR ,32767"},
		{inp: "FILES", tk: token.FILES, lst: "FILES"},
		{inp: `FILES "C:\MENU"`, tk: token.FILES, lst: `FILES "C:\MENU"`},
		{inp: `FILES "C:\MENU", "AndSuch"`, tk: token.FILES, lst: `FILES "C:\MENU", "AndSuch"`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		if sl.LineLength() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := sl.NextStatement()

		if stmt.TokenLiteral() != tt.tk {
			t.Fatalf("Test_Commands(%s) didn't get a %s command", tt.inp, tt.tk)
		}

		lst := stmt.String()
		if tt.lst != "" {
			assert.Equal(t, tt.lst, lst, "Test_Commands(%s) expected %s, got %s", tt.inp, tt.lst, lst)
		}
	}
}

func Test_CommonStatement(t *testing.T) {
	tests := []struct {
		inp string
		cnt int
	}{
		{inp: "COMMON A()", cnt: 1},
		{inp: "COMMON A(), B[] : REM", cnt: 2},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)
		stmt := sl.NextStatement()

		cmn, ok := stmt.(*ast.CommonStatement)

		if !ok {
			t.Fatalf("Test_CommonStatement didn't return correct object")
		}

		assert.Equal(t, tt.cnt, len(cmn.Vars))
	}

}

// the "CONT" command means continue running the program
func Test_ContCommand(t *testing.T) {
	env := object.NewTermEnvironment(mocks.MockTerm{})
	sl := ParseInput("CONT", env)
	assert.EqualValues(t, 1, sl.LineLength())
}

func Test_Csrlin(t *testing.T) {
	env := object.NewTermEnvironment(mocks.MockTerm{})
	sl := ParseInput("PRINT CSRLIN", env)
	assert.EqualValues(t, 1, sl.LineLength())
}

func Test_DataStatement(t *testing.T) {
	tkInt := token.Token{Type: token.INT, Literal: "INT"}
	tkFixed := token.Token{Type: token.FIXED, Literal: "123.45"}
	tkString := token.Token{Type: token.STRING, Literal: "STRING"}
	tkFloatS := token.Token{Type: token.FLOAT, Literal: "3.14159E+0"}
	tkFloatD := token.Token{Type: token.FLOAT, Literal: "3.14159D+0"}
	tkDblInt := token.Token{Type: token.INTD, Literal: "99999#"}

	tests := []struct {
		inp     string           // source line
		stmtNum int              // # of statements expected
		lineNum uint16           // line number
		cnt     int              // number of expressions expected
		exp     []ast.Expression // expected values
	}{
		{`10 DATA "Fred", George Foreman`, 2, 10, 2, []ast.Expression{
			&ast.StringLiteral{Token: tkString, Value: "Fred"},
			&ast.StringLiteral{Token: tkString, Value: "George Foreman"},
		}},
		{`20 DATA 123, 123.45, "Fred", 99999#`, 2, 20, 4, []ast.Expression{
			&ast.IntegerLiteral{Token: tkInt, Value: 123},
			&ast.FixedLiteral{Token: tkFixed, Value: tkFixed},
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
		sl := ast.NewSourceLine(tt.inp, tt.lineNum)
		l := lexer.New(tt.inp)
		p := New(l)
		p.ParseSourceLine(sl)
		FinishParseSourceLine(sl)

		assert.EqualValues(t, tt.stmtNum, sl.LineLength())

		stmt := sl.NextStatement()
		lm, ok := stmt.(*ast.LineNumStmt)

		assert.True(t, ok)
		assert.EqualValues(t, tt.lineNum, lm.Value)

		stmt = sl.NextStatement()
		dstmt, ok := stmt.(*ast.DataStatement)
		assert.True(t, ok)

		assert.Equal(t, tt.cnt, len(dstmt.Consts))

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
		inp     string
		exp     string
		stmtNum int
		lineNum uint16
		numIDs  int8
		dims    []dimensions
	}{
		{`10 DIM A(20)`, `DIM A(20)`, 2, 10, 1, []dimensions{{"A()", []int8{20}}}},
		{`20 DIM A[20, 10]`, `DIM A[20,10]`, 2, 20, 1, []dimensions{{"A[]", []int8{20, 10}}}},
		{`30 DIM A[20, 30],B[15,5]`, `DIM A[20,30], B[15,5]`, 2, 30, 2, []dimensions{{"A[]", []int8{20, 30}}, {"B[]", []int8{15, 5}}}},
		{`40 DIM A(20)`, `DIM A(20)`, 2, 40, 1, []dimensions{{"A()", []int8{20}}}},
		{`50 DIM A(20) : REM A Comment`, `DIM A(20)`, 3, 50, 1, []dimensions{{"A()", []int8{20}}}},
	}

	for _, tt := range tests {
		sl := ast.NewSourceLine(tt.inp, tt.lineNum)
		l := lexer.New(tt.inp)
		p := New(l)
		p.ParseSourceLine(sl)

		assert.EqualValues(t, tt.stmtNum, sl.LineLength())
		stmt := sl.NextStatement()

		lm, ok := stmt.(*ast.LineNumStmt)
		assert.True(t, ok)
		assert.EqualValues(t, tt.lineNum, lm.Value)

		stmt = sl.NextStatement()
		dstmt, ok := stmt.(*ast.DimStatement)
		assert.True(t, ok)

		assert.Equal(t, tt.exp, dstmt.String(), "TestDimStatement got %s, expected %s", dstmt.String(), tt.exp)

		if int8(len(dstmt.Vars)) != tt.numIDs {
			t.Fatalf("expected %d dimensioned variables, got %d on %s", tt.numIDs, len(dstmt.Vars), tt.inp)
		}

		for dNum, d := range tt.dims {
			if dstmt.Vars[dNum].Token.Literal != d.id {
				t.Fatalf("got literal %s, expected %s on line %s", dstmt.Vars[dNum].Token.Literal, d.id, tt.inp)
			}

			for dnum, dim := range d.dims {
				indExp, ok := dstmt.Vars[dNum].Index[dnum].Index.(*ast.IntegerLiteral)

				if !ok {
					t.Fatalf("dimension %d for %s is not an index", dnum, tt.inp)
				}

				if int8(indExp.Value) != dim {
					t.Fatalf("expeced dimension %d, got %d, on %s", dim, indExp.Value, tt.inp)
				}
			}
		}
	}
}

func Test_ErrorStatement(t *testing.T) {
	tests := []struct {
		inp string
		exp int
	}{
		{inp: `10 ERROR 31`, exp: 31},
	}

	for _, tt := range tests {
		sl := ast.NewSourceLine(tt.inp, 10)
		l := lexer.New(tt.inp)
		p := New(l)
		p.ParseSourceLine(sl)

		sl.NextStatement()
		val := sl.NextStatement()

		ers, ok := val.(*ast.ErrorStatement)
		assert.True(t, ok, "Failed to get an ErrorStatement")

		err, ok := ers.ErrNum.(*ast.IntegerLiteral)

		assert.True(t, ok, "ErrorStatement didn't have IntegerLiteral")
		assert.EqualValues(t, tt.exp, err.Value)
	}
}

func Test_KeyStatement(t *testing.T) {
	tests := []struct {
		inp  string
		lnum uint16
		parm ast.Expression
	}{
		{inp: `10 KEY`, lnum: 10},
		{inp: `20 KEY OFF`, lnum: 20, parm: &ast.OffExpression{}},
		{inp: `30 KEY 1,"FILES"`, lnum: 30, parm: &ast.IntegerLiteral{Token: token.Token{Type: "INT", Literal: "1"}, Value: 1}},
		{inp: `40 KEY ON`, lnum: 40, parm: &ast.OnExpression{}},
		{inp: `50 KEY LIST`, lnum: 50, parm: &ast.ListExpression{Token: token.Token{Type: "LIST", Literal: "LIST"}}},
		{inp: `60 KEY 1, CHR$(03)+CHR$(25)`, lnum: 60, parm: &ast.IntegerLiteral{Token: token.Token{Type: "INT", Literal: "1"}, Value: 1}},
	}

	for _, tt := range tests {
		sl := ast.NewSourceLine(tt.inp, tt.lnum)
		l := lexer.New(sl.String())
		p := New(l)
		p.ParseSourceLine(sl)
		sl.NextStatement()
		k := sl.NextStatement()

		assert.NotNil(t, k, "failed to get a KeyStatement")
		key, ok := k.(*ast.KeyStatement)

		assert.True(t, ok, "statement was not a KeyStatement")
		assert.NotNil(t, key, "*KeyStatement was nil")
		assert.Equal(t, tt.parm, key.Param)
	}
}

func Test_LetStatementImplied(t *testing.T) {
	tests := []struct {
		inp  string
		lnum uint16
		exp  []string
	}{
		{inp: `10 X = 5: Y = 20`, lnum: 10, exp: []string{` X = 5`, ` Y = 20`}},
		{inp: `20 CALIBRATE PORT 10`, lnum: 20, exp: []string{`CALIBRATE PORT 10`}},
	}

	for _, tt := range tests {
		sl := ast.NewSourceLine(tt.inp, tt.lnum)
		l := lexer.New(sl.String())
		p := New(l)
		p.ParseSourceLine(sl)

		assert.EqualValues(t, len(tt.exp)+1, sl.LineLength())

		sl.NextStatement()
		for _, e := range tt.exp {
			assert.Equal(t, e, sl.NextStatement().String())
		}
	}
}

func Test_LetStatement(t *testing.T) {
	inp := `10 let x = 5: let y$ = "test": let foobar% = 838383 : LET BANG! = 46.8 : LET POUND# = 7654321.1234`
	//input := `10 LET 4 = 5` ToDo support this

	sl := ast.NewSourceLine(inp, 10)
	l := lexer.New(sl.String())
	p := New(l)
	p.ParseSourceLine(sl)

	assert.EqualValues(t, 6, sl.LineLength())

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

	for _, tt := range tests {
		stmt := sl.NextStatement()

		_, ok := stmt.(*ast.LineNumStmt)
		if ok {
			continue
		}

		assert.True(t, strings.EqualFold("LET", stmt.TokenLiteral()))

		letStmt, ok := stmt.(*ast.LetStatement)
		assert.True(t, ok)
		assert.True(t, strings.EqualFold(tt.expectedIdentifier, letStmt.Name.String()))
		assert.True(t, strings.EqualFold(tt.expectedIdentifier, letStmt.Name.TokenLiteral()))
	}
}

func TestLetWithTypes(t *testing.T) {
	type result struct {
		expectedToken      string
		expectedIdentifier string
	}

	type results []result

	tests := []struct {
		inp     string
		lnum    uint16
		results results
	}{
		{inp: `10 LET A$ = "a test string"`, lnum: 10, results: results{
			{token.LINENUM, "10"},
			{token.LET, "A$"},
		},
		},
		{inp: `20 LET B% = "a test string"`, lnum: 20, results: results{
			{token.LINENUM, "10"},
			{token.LET, "B%"},
		},
		},
		{inp: `30 LET C! = "a test string"`, lnum: 30, results: results{
			{token.LINENUM, "10"},
			{token.LET, "C!"},
		},
		},
		{inp: `40 LET D# = "a test string"`, lnum: 40, results: results{
			{token.LINENUM, "10"},
			{token.LET, "D#"},
		},
		},
	}

	for _, tt := range tests {
		sl := ast.NewSourceLine(tt.inp, tt.lnum)
		l := lexer.New(sl.String())
		p := New(l)
		p.ParseSourceLine(sl)
		stmt := sl.NextStatement()

		for _, ttt := range tt.results {
			_, ok := stmt.(*ast.LineNumStmt)
			if ok {
				continue
			}

			assert.True(t, strings.EqualFold("LET", stmt.TokenLiteral()))

			letStmt, ok := stmt.(*ast.LetStatement)
			assert.True(t, ok)
			assert.True(t, strings.EqualFold(ttt.expectedIdentifier, letStmt.Name.String()))
			assert.True(t, strings.EqualFold(ttt.expectedIdentifier, letStmt.Name.TokenLiteral()))
		}
	}
}

func TestLineNumbers(t *testing.T) {
	tk := token.Token{Type: token.AUTO, Literal: "AUTO"}

	env := object.NewTermEnvironment(mocks.MockTerm{})
	env.SaveSetting(settings.Auto, &ast.AutoCommand{Token: tk, Params: []ast.Expression{&ast.IntegerLiteral{Value: 30}, &ast.IntegerLiteral{Value: 10}}})

	ParseInput("10 REM", env)
	ParseInput("20 REM", env)
	ParseInput("30 REM", env)

	tests := []struct {
		tok  string
		lnum uint16
	}{
		{tok: token.LINENUM, lnum: 10},
		{tok: token.LINENUM, lnum: 20},
		{tok: token.LINENUM, lnum: 30},
	}

	for _, tt := range tests {
		stmt := env.NextStatement()
		if stmt != nil {
			if !testLineNumber(t, stmt, tt.lnum) {
				return
			}
		}
	}

}

func testLineNumber(t *testing.T, s ast.Statement, line uint16) bool {
	lineStmt, ok := s.(*ast.LineNumStmt)
	if !ok {
		t.Errorf("s not *ast.LineNumStmt. got=%T", s)
		return false
	}
	if lineStmt.Value != uint16(line) {
		t.Errorf("lineStmt.Value not '%d'. got=%d", line, lineStmt.Value)
		return false
	}
	return true
}

func Test_LineNumberWithTrash(t *testing.T) {
	var mocklex mocks.MockLexer
	mocklex.AddToken(token.Token{Type: token.LINENUM, Literal: "10PRINT"})
	mocklex.AddToken(token.Token{Type: token.EOF})
	p := New(&mocklex)
	p.parseLineNumber()
}

func Test_LiteralsWithTrash(t *testing.T) {
	tests := []struct {
		tokens   []token.Token
		hasTrash bool
	}{
		{tokens: []token.Token{{Type: token.FLOAT, Literal: "3.14159E+0"},
			{Type: token.EOF}}},
		{tokens: []token.Token{{Type: token.FLOAT, Literal: "3.14159EFG+0"},
			{Type: token.EOF}}, hasTrash: true},
		{tokens: []token.Token{{Type: token.FLOAT, Literal: "1.09432D-06"},
			{Type: token.EOF}}},
		{tokens: []token.Token{{Type: token.FLOAT, Literal: "1.09432DBA-06"},
			{Type: token.EOF}}, hasTrash: true},
	}

	for _, tt := range tests {
		var mocklex mocks.MockLexer
		for _, tok := range tt.tokens {
			mocklex.AddToken(tok)
		}
		p := New(&mocklex)
		exp := p.parseExpression(LOWEST)
		assert.NotZero(t, len(tt.tokens))

		switch v := exp.(type) {
		case *ast.FloatSingleLiteral:
		case *ast.FloatDoubleLiteral:
			assert.Equal(t, tt.hasTrash, v.HasTrash())
		default:
			assert.Fail(t, "Got unsupported type in test")
		}
	}
}

func Test_LoadCommand(t *testing.T) {
	tests := []struct {
		inp      string // command to parse
		keepOpen bool   // flag should be set
		trash    bool
	}{
		{inp: `LOAD "HEWORLD.BAS"`},
		{inp: `LOAD "HIWORLD.BAS",R`, keepOpen: true},
		{inp: `LOAD "HERWORLD.BAS",F`, trash: true},
		{inp: `LOAD "HERWORLD.BAS" F`, trash: true},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		fmt.Println(tt.inp)
		sl := ParseInput(tt.inp, env)

		stmt := sl.NextStatement()
		cmd, ok := stmt.(*ast.LoadCommand)

		if !ok {
			t.Fatalf("(%s) parse didn't return LoadCommand, got %T instead", tt.inp, stmt)
		}

		assert.Equal(t, tt.keepOpen, cmd.KeepOpen, "KeepOpen incorrect")

		if tt.trash {
			assert.True(t, cmd.HasTrash(), "trash not found")
		} else {
			assert.False(t, cmd.HasTrash(), "unexpected trash found")
		}
	}
}

func Test_LocateStatement(t *testing.T) {
	tests := []struct {
		inp string           // statement to test
		exp []ast.Expression // array of parameter expressions expected
		err bool             // true if I expect a parse error
	}{
		{inp: `LOCATE 5,5 : PRINT "Hello"`, exp: []ast.Expression{&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
			&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5}}},
		{inp: `LOCATE`},
		{inp: `LOCATE $`, err: true},
		{inp: `LOCATE 1,2`, exp: []ast.Expression{&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1},
			&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2}}},
		{inp: `LOCATE 1,,2`, exp: []ast.Expression{&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1}, nil,
			&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2}}},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		stmt := sl.NextStatement()
		lct, ok := stmt.(*ast.LocateStatement)

		if !ok {
			t.Fatalf("Test_LocateStatement didn't return Locate object")
		}

		if !tt.err {
			for i, res := range lct.Parms {
				if res != nil {
					if tt.exp[i] != nil {
						assert.Equal(t, tt.exp[i], res, "parseLocateStatement param %d mismatch", i)
					} else {
						t.Fatalf("Test_LocateStatement got a param it didn't expect")
					}
				}
			}
		} else {
			//assert.True(t, lct.H)
		}
	}
}

func TestIdentifierExpression(t *testing.T) {
	env := object.NewTermEnvironment(mocks.MockTerm{})
	cmd := ParseInput("10 foobar", env)
	assert.Nil(t, cmd)

	// since the line
	n, src := env.NextStatement()
	assert.Nil(t, n)

	n = env.NextStatement()
	assert.NotNil(t, n)

	stmt, ok := n.(*ast.ExpressionStatement)

	assert.True(t, ok)
	ident, ok := stmt.Expression.(*ast.Identifier)

	assert.True(t, ok)
	assert.Equal(t, "FOOBAR", ident)
}

func TestNewCommand(t *testing.T) {
	inp := "new"
	env := object.NewTermEnvironment(mocks.MockTerm{})
	sl := ParseInput(inp, env)
	assert.EqualValues(t, 1, sl.LineLength(), "NewCommand didn't create one command")
}

func TestNextCommand(t *testing.T) {
	tests := []struct {
		inp string
		err bool
	}{
		{inp: `30 NEXT`},
		{inp: `40 NEXT X`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := object.NewTermEnvironment(mocks.MockTerm{})
		p.ParseProgram(env)
	}
}

func TestParseTrash(t *testing.T) {
	tests := []struct {
		inp string
		exp string
		len int
	}{
		{inp: "Buddy", exp: "EOF", len: 1},
		{inp: "Buddy:REM", exp: ":", len: 1},
		{inp: "Buddy,Neighbor:REM", exp: ":", len: 3},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		p.nextToken()
		var Trash []ast.TrashStatement
		p.parseTrash(&Trash)

		assert.Equal(t, tt.exp, p.peekToken.Literal, "Trash stopped for wrong reason")
	}
}

func TestOnStatement(t *testing.T) {
	// Currently support
	// ON ERROR GOTO
	// ON exp GOTO
	// ON exp GOSUB
	tests := []struct {
		inp string
		exp string
		jmp int
		tpe int
	}{
		{inp: "10 ON ERROR GOTO 100", exp: "ON ERROR GOTO 100", jmp: 100},
		{inp: "10 ON ERROR GOTO END", exp: "ON ERROR GOTO END"},
		{inp: "10 ON ERROR", exp: "ON ERROR"},
		{inp: "10 ON ERROR GOSUB 100", exp: "ON ERROR GOSUB 100"},
		{inp: "10 ON ERROR GOTO 10000000000000000000", exp: "ON ERROR GOTO", jmp: 0},
		{inp: "10 ON X GOTO 100, 200, 300", exp: "ON X GOTO 100, 200, 300"},
		{inp: "10 ON X GOSUB 100, 200, 300", exp: "ON X GOSUB 100, 200, 300"},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		n := env.NextStatement()
		n = env.NextStatement()
		switch stmt := n.(type) {
		case *ast.OnErrorGoto:
			assert.EqualValues(t, tt.exp, stmt.String(), "ON ERROR parse fail")
			assert.EqualValues(t, tt.jmp, stmt.Jump, "got the wrong line")
		case *ast.OnGoStatement:
			assert.EqualValues(t, tt.exp, stmt.String(), "ON ERROR parse fail")
		}
	}
}

func TestOpenStatement(t *testing.T) {
	tests := []struct {
		inp string
		exp string
	}{
		// brief syntax
		{inp: `10 open "O",1,"test.out",128`,
			exp: `open "O", 1, "test.out",128`},
		{inp: `20 open "O", #2, "test.out",128`,
			exp: `open "O", #2, "test.out",128`},
		{inp: `30 open "O" #3, "test.out",128`,
			exp: `open "O" # 3, "test.out", 128`},

		// verbose syntax
		{inp: `40 open "test.out" FOR OUTPUT ACCESS WRITE SHARED AS #1 LEN = 128`,
			exp: `open "test.out" FOR OUTPUT ACCESS WRITE SHARED AS #1 LEN = 128`},
		{inp: `50 open "test2.out" FOR OUTPUT ACCESS WRITE SHARED AS #2 LEN = 128 FOR`,
			exp: `open "test2.out" FOR OUTPUT ACCESS WRITE SHARED AS #2 LEN = 128 FOR`},
		// this next one would eval to a syntax error
		{inp: `60 open "test3.out" FOR OUTPUT ACCESS WRITE LOCK READ AS #3 LEN = 128`,
			exp: `open "test3.out" FOR OUTPUT ACCESS WRITE LOCK READ AS # 3 LEN = 128`},
		// error case
		{inp: `60 open 3, "test3.out" FOR OUTPUT ACCESS WRITE LOCK READ AS #3 LEN = 128`,
			exp: `open 3, "test3.out" FOR OUTPUT ACCESS WRITE LOCK READ AS # 3 LEN = 128`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		// go get the second statement in program

		n := env.NextStatement()
		n = env.NextStatement()

		assert.Equal(t, tt.exp, n.String())
	}
}

func TestPaletteStatement(t *testing.T) {
	tests := []struct {
		inp string
		err bool
	}{
		{inp: `10 PALETTE 3,2`},
		{inp: `20 PALETTE USING PAL(3)`},
		{inp: `30 PALETTE t,x`},
		{inp: `40 PALETTE t x`, err: true},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		n := env.NextStatement()
		n = env.NextStatement()
		stmt, ok := n.(*ast.PaletteStatement)

		assert.True(t, ok, "Didn't get a palette statement")

		if tt.err {
			assert.True(t, stmt.HasTrash(), "failed to catch the error")
		}
	}
}

func Test_ReadStatement(t *testing.T) {
	tkAs := token.Token{Type: token.IDENT, Literal: "A$"}
	tkBs := token.Token{Type: token.IDENT, Literal: "B$"}

	tests := []struct {
		inp     string
		stmtNum int              // expected count of statments
		lineNum uint16           // line number
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		n := env.NextStatement()
		n = env.NextStatement()

		lm, ok := n.(*ast.LineNumStmt)

		assert.True(t, ok)
		assert.Equal(t, tt.lineNum, lm.Value)

		n = env.NextStatement()
		rstmt, ok := n.(*ast.ReadStatement)

		assert.True(t, ok)
		assert.Equal(t, tt.vars, len(rstmt.Vars))
	}
}

func Test_RemStatement(t *testing.T) {
	tests := []struct {
		inp string
		res string
	}{
		{inp: "10 REM A code comment", res: "REM A code comment"},
		{inp: "20 REM", res: "REM "},
		{inp: "30 ' Alternate form remark", res: "' Alternate form remark"},
		{inp: "40 'Once a remark : GOTO 20", res: "' Once a remark : GOTO 20"},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		n := env.NextStatement()
		n = env.NextStatement()

		assert.Equal(t, tt.res, n.String())
	}
}

func TestRestore(t *testing.T) {
	rsTk := token.Token{Type: token.RESTORE, Literal: "RESTORE"}

	tests := []struct {
		inp string
		exp interface{}
	}{
		{inp: `10 RESTORE`, exp: &ast.RestoreStatement{Token: rsTk, Line: -1}},
		{inp: `20 RESTORE 300`, exp: &ast.RestoreStatement{Token: rsTk, Line: 300}},
		{inp: `30 RESTORE X`},
		{inp: `40 RESTORE : END`, exp: &ast.RestoreStatement{Token: rsTk, Line: -1}},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		n := env.NextStatement()
		n = env.NextStatement()

		if tt.exp != nil {
			compareStatements(tt.inp, n, tt.exp, t)
		} else {
			res, ok := n.(*ast.RestoreStatement)
			assert.True(t, ok)
			assert.True(t, res.HasTrash())
		}
	}
}

func Test_ResumeStatement(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: `100 RESUME`},
		{inp: `100 RESUME NEXT`},
		{inp: `100 RESUME 0`},
		{inp: `100 RESUME 100`},
		{inp: `100 RESUME FRED`},
	}

	for _, tt := range tests {
		l := lexer.New(tt.inp)
		p := New(l)
		env := object.NewTermEnvironment(mocks.MockTerm{})
		p.ParseProgram(env)
	}
}

func Test_ScreenStatement(t *testing.T) {
	tests := []struct {
		inp   string
		exp   []ast.Expression
		trash bool
	}{
		{inp: "10 SCREEN 0,1", exp: []ast.Expression{&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "0"}, Value: 0},
			&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "1"}, Value: 1}}},
		{inp: "20 SCREEN 2,,3", exp: []ast.Expression{&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "2"}, Value: 2},
			nil,
			&ast.IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "3"}, Value: 3}}},
		{inp: "30 SCREEN", trash: true},
		{inp: "30 SCREEN 1,2,3,4,5", trash: true},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		cd := env.NextStatement()
		assert.NotNil(t, cd)

		cd = env.NextStatement()
		assert.NotNil(t, cd)

		scrn := cd.(*ast.ScreenStatement)
		assert.NotNil(t, scrn)

		for i, exp := range tt.exp {
			assert.Equal(t, tt.exp[i], exp)
		}

		assert.Equal(t, tt.trash, scrn.HasTrash())
	}
}

func Test_StopStatement(t *testing.T) {

	input := `10 STOP : REM a test`
	env := object.NewTermEnvironment(mocks.MockTerm{})
	ParseInput(input, env)
	stmt := env.NextStatement()
	assert.NotNil(t, stmt)

	stmt = env.NextStatement()
	assert.NotNil(t, stmt)
	stop, ok := stmt.(*ast.StopStatement)

	assert.True(t, ok)
	assert.Equal(t, token.STOP, stop)
}

func Test_StringLiteralExpression(t *testing.T) {
	tests := []struct {
		inp  string
		outp string
		exp  ast.Statement
	}{
		{inp: `10 "Hello World!"`, outp: `Hello World!`, exp: &ast.ExpressionStatement{Expression: &ast.StringLiteral{Value: `Hello World!`}}},
		//{inp: `20 CALIBRATE PORT 10`},
	}

	for _, tt := range tests {
		input := tt.inp
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(input, env)
		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		step, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok)

		literal, ok := step.Expression.(*ast.StringLiteral)
		assert.True(t, ok)

		assert.Equal(t, tt.outp, literal.Value)
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		assert.EqualValues(t, 1, sl.LineLength(), "program.Cmd does not contain single command")

		stmt := sl.NextStatement()
		assert.Equal(t, tt.tok, stmt.TokenLiteral(), "TestTronTroffCommands didn't get an %s command", tt.inp)
	}
}

func Test_UsingExpression(t *testing.T) {
	tst := []struct {
		inp string
	}{
		//{inp: `10 PRINT USING`},
		//{inp: `10 PRINT USING "###.##"`},
		//{inp: `10 PRINT USING "###.##";`},
		{inp: `10 PRINT USING "###.##"; X; Y;`},
	}

	for _, tt := range tst {
		l := lexer.New(tt.inp)
		p := New(l)
		env := object.NewTermEnvironment(mocks.MockTerm{})
		p.ParseProgram(env)
	}
}

func Test_UsingRunTime(t *testing.T) {
	tst := []struct {
		inp  string
		exp  string
		data float32
		fmt  string
	}{
		{inp: `#####`, exp: `%5.f`, data: 328, fmt: "  328"},
		{inp: `###.##`, exp: `%6.2f`, data: 123.456, fmt: "123.46"},
		{inp: `###.##`, exp: `%6.2f`, data: -123.456, fmt: "-123.46"},
		{inp: `+###.##`, exp: `%+6.2f`, data: 123.456, fmt: "+123.46"},
		{inp: `+###.##`, exp: `%+6.2f`, data: -123.456, fmt: "-123.46"},
	}

	for _, tt := range tst {
		l := lexer.New(tt.inp)
		p := New(l)
		//env := object.NewTermEnvironment(mocks.MockTerm{})
		rc := p.ParseUsingRunTime()

		assert.EqualValuesf(t, tt.exp, rc, "%s", tt.inp)
		res := fmt.Sprintf(rc, tt.data)
		assert.EqualValuesf(t, tt.fmt, res, "%s", tt.inp)
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	dblTok := token.Token{Type: token.INTD, Literal: "65999"}
	fltTok := token.Token{Type: token.FLOAT, Literal: "4294967295"}

	tests := []struct {
		inp   string
		stmts int
		lit   interface{}
	}{
		{`10 5`, 2, &ast.IntegerLiteral{Value: 5, Token: token.Token{Type: token.INT, Literal: "5"}}},
		{`20 65999#`, 2, &ast.DblIntegerLiteral{Value: 65999, Token: dblTok}},
		{`30 4294967295`, 2, &ast.FloatSingleLiteral{Token: fltTok, Value: 4294967295}},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		exp, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok)

		compareStatements(tt.inp, tt.lit, exp, t)
	}
}

func TestHexOctalConstants(t *testing.T) {
	tests := []struct {
		inp   string
		lit   interface{}
		trash bool
	}{
		{inp: "10 &HF76F", lit: &ast.HexConstant{Value: "F76F"}},
		{inp: "20 &HF7F6F", lit: &ast.HexConstant{Value: "F7F6F"}},
		{inp: "30 &767", lit: &ast.OctalConstant{Value: "767"}},
		{inp: "40 &O767", lit: &ast.OctalConstant{Value: "767"}},
		{inp: "50 &F767", lit: nil, trash: true},
		{inp: "60 &O F767", lit: nil},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		exp, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok)

		if !tt.trash {
			compareStatements(tt.inp, tt.lit, exp, t)
		} else {
			oct, ok := exp.Expression.(*ast.OctalConstant)
			assert.True(t, ok, tt.inp)
			assert.True(t, oct.HasTrash(), tt.inp)
		}
	}
}

type parseFunc func(*Parser) ast.Expression

// TODO: This code is crap
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
		}, " 53a"},
		{"62.4d5", token.FIXED, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, "62.4d5"},
		{"53", token.INT, func(p *Parser) ast.Expression {
			return p.parseIntegerLiteral()
		}, "53"},
		{"62.45", token.FIXED, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, "62.45"},
		{"62.", token.INT, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, "62."},
		{"62.45.37", token.INT, func(p *Parser) ast.Expression {
			return p.parseFixedPointLiteral()
		}, "62.45.37"},
		{"624537", token.INT, func(p *Parser) ast.Expression {
			return p.parseIntegerLiteral()
		}, "624537"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)

		// this is where I cheat
		p.curToken.Type = tt.tok
		p.curToken.Literal = tt.input

		res := tt.fn(p)

		if (tt.res == "") && (strings.Compare(res.TokenLiteral(), tt.input) != 0) {
			t.Errorf("Parse succeeded when it should have failed, %s", tt.input)
		}

		if tt.res != "" {
			assert.EqualValues(t, tt.res, res.String())
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.input, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		exp, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok)

		pre, ok := exp.Expression.(*ast.PrefixExpression)
		assert.True(t, ok)

		assert.Equal(t, tt.operator, pre.Operator)
		assert.Equal(t, tt.integerValue, pre.Right)
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
		lineNum    uint16
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.input, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		exp, ok := stmt.(*ast.ExpressionStatement)
		assert.True(t, ok)

		exp2, ok := exp.Expression.(*ast.InfixExpression)
		assert.True(t, ok)

		assert.Equal(t, tt.leftValue, exp2.Left)
		assert.Equal(t, tt.operator, exp2.Operator)
		assert.Equal(t, tt.rightValue, exp2.Right)
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
		{"20 X = ((5 < 4) <> (3 > 4))", "20  X = ((5 < 4) <> (3 > 4))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		env := object.NewTermEnvironment(mocks.MockTerm{})
		p.ParseProgram(env)
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
		{`80 LET A[0] = 5 : LET A(1) = 2`, "A[]", "", 1, []string{"0"}},
		{`90 LET A$[0] = "Hello"`, "A$[]", "$", 1, []string{"0"}},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.input, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		ls, ok := stmt.(*ast.LetStatement)
		assert.True(t, ok)

		assert.Equal(t, tt.literal, ls.Name.Value)
		assert.Equal(t, tt.typ, ls.Name.Type)
		assert.Equal(t, tt.indCount, ls.Name.Index)

		for i, dim := range tt.indVal {
			assert.Equal(t, dim, ls.Name.Index[i].Index.String())
		}
	}
}

func TestIfStatement(t *testing.T) {
	tests := []struct {
		inp  string // input source
		cons string // consequence
		alt  string // alternative
		op   string // comparison operand
		exp  string // expected string output
	}{
		{inp: "10 IF X < Y THEN GOTO 300", cons: "GOTO", alt: "nil", op: "<", exp: "IF X < Y THEN GOTO 300"},
		{inp: "20 IF (X < Y) GOTO 300", cons: "GOTO", alt: "nil", op: "<", exp: "IF (X < Y) THEN GOTO 300"},
		{inp: "30 IF X > Y THEN 300 ELSE 400", cons: "GOTO", alt: "GOTO", op: ">", exp: "IF X > Y THEN 300 ELSE 400"},
		{inp: "40 IF X >= Y THEN END", cons: "END", alt: "nil", op: ">=", exp: "IF X >= Y THEN END"},
		{inp: "50 IF X < Y THEN 300 ELSE END", cons: "GOTO", alt: "END", op: "<", exp: "IF X < Y THEN 300 ELSE END"},
		{inp: "60 IF X < Y, THEN 300 ELSE END", cons: "GOTO", alt: "END", op: "<", exp: "IF X < Y THEN 300 ELSE END"},
		{inp: "70 IF X = Y, THEN 300 ELSE END", cons: "GOTO", alt: "END", op: "=", exp: "IF X = Y THEN 300 ELSE END"},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		ifstmt, ok := stmt.(*ast.IfStatement)
		assert.True(t, ok)
		assert.Equal(t, tt.exp, ifstmt.String())

		gexp, ok := ifstmt.Condition.(*ast.GroupedExpression)
		assert.True(t, ok)

		iexp, ok := gexp.Exp.(*ast.InfixExpression)
		assert.True(t, ok)

		assert.True(t, testInfixExpression(t, iexp, "X", tt.op, "Y"))
		assert.True(t, testInfixExpression(t, ifstmt.Condition, "X", tt.op, "Y"))
		assert.True(t, testIfConsequence(t, tt.cons, ifstmt.Consequence))
		assert.True(t, testIfAlternative(t, tt.alt, ifstmt.Alternative))
	}
}

func Test_ParseInkeyExpression(t *testing.T) {
	tests := []struct {
		inp string
		exp int
	}{
		{inp: `10 X$ = INKEY$ : END`, exp: 3},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.Nil(t, stmt)
	}

}

func Test_ParseInput(t *testing.T) {
	tests := []struct {
		inp string
		len int
	}{
		{inp: `10 PRINT "Hello World!"`, len: 0},
		{inp: `PRINT 10`, len: 1},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})

		sl := ParseInput(tt.inp, env)

		assert.EqualValues(t, tt.len, sl.LineLength())
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.input, env)

		assert.EqualValues(t, tt.expStmts, sl.LineLength())

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		gotoStmt, ok := stmt.(*ast.GotoStatement)

		assert.True(t, ok, "stmt not *ast.GotoStatement")
		assert.Equalf(t, "GOTO", gotoStmt.TokenLiteral(), "returnStmt.TokenLiteral not 'GOTO', got %q", gotoStmt.TokenLiteral())
		assert.Equalf(t, 1, len(gotoStmt.JmpTo), "Goto didn't have 1 JmpTo, it had %d", len(gotoStmt.JmpTo))
		assert.Equalf(t, tt.expectedValue, gotoStmt.JmpTo[0].Literal, "expected linenum %s, got %s", tt.expectedValue, gotoStmt.JmpTo[0].Literal)
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.input, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		gosubStmt, ok := stmt.(*ast.GosubStatement)

		assert.True(t, ok, "stmt not *ast.GosubStatement")
		assert.Equalf(t, "GOSUB", gosubStmt.TokenLiteral(), "returnStmt.TokenLiteral not 'GOSUB', got %q", gosubStmt.TokenLiteral())
		assert.Equalf(t, 1, len(gosubStmt.Gosub), "GOSUB had more than one destination")
		assert.Equalf(t, tt.expectedValue, gosubStmt.Gosub[0].Literal, "expected linenum %s, got %s", tt.expectedValue, gosubStmt.Gosub[0].Literal)
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.input, env)

		assert.EqualValues(t, tt.expStmts, sl.LineLength())

		sl.NextStatement()         // line number
		stmt := sl.NextStatement() // return statement
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		assert.True(t, ok)
		assert.Equal(t, "RETURN", returnStmt.TokenLiteral())
		assert.EqualValues(t, tt.expectedValue, returnStmt.ReturnTo)
	}
}

func Test_RunCommand(t *testing.T) {
	tests := []struct {
		inp   string
		start int
		file  string
		err   bool // I expect parsing to faile
	}{
		{inp: "RUN"},
		{inp: "RUN 20", start: 20},
		{inp: `RUN "TESTFILE.BAS"`, file: `"TESTFILE.BAS"`},
		{inp: `RUN "TESTFILE.BAS",r`, file: `"TESTFILE.BAS"`},
		{inp: `RUN "TESTFILE.BAS",k`, file: `"TESTFILE.BAS"`, err: true},
		{inp: `RUN "TESTFILE.BAS",-`, file: `"TESTFILE.BAS"`, err: true},
	}

	fmt.Println("TestRunCommand Parsing")
	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		if sl.LineLength() != 1 {
			t.Fatal("program.Cmd does not contain single command")
		}

		stmt := sl.NextStatement()

		if stmt.TokenLiteral() != token.RUN {
			t.Fatal("TestRunCommand didn't get an Run command")
		}

		atc := stmt.(*ast.RunCommand)

		if atc == nil {
			t.Fatal("TestRunCommand couldn't extract RunCommand object")
		}

		if tt.err {
			assert.True(t, atc.HasTrash(), "didn't find trash")
		}

		if atc.StartLine != tt.start {
			t.Fatalf("TestRunCommand got start = %d, expected %d", atc.StartLine, tt.start)
		}

		if atc.LoadFile != nil {
			assert.Equalf(t, tt.file, atc.LoadFile.String(), "TestRun(%s) expected %s, got %s", tt.inp, tt.file, atc.LoadFile.String())
		}
	}
}

func TestCheckForFuncCall(t *testing.T) {
	tst := []struct {
		inp string
		exp bool
	}{
		{inp: "LEN", exp: true},
		{inp: "FNA", exp: true},
		{inp: "MUFIN", exp: false},
	}

	for _, tt := range tst {
		l := lexer.New(tt.inp)
		p := New(l)
		p.nextToken() // skip the starting EOL
		p.checkForFuncCall()
	}
}

func TestDefFN(t *testing.T) {
	tests := []struct {
		inp string
		exp string
	}{
		{inp: `DEF FNINC(X) = X + 1`, exp: `DEF FNINC(X) = X + 1`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)
		assert.NotNil(t, sl)
		assert.NotZero(t, sl.LineLength())

		lst := sl.NextStatement().String()
		assert.Equal(t, tt.exp, lst, "Unexpected DEF string")
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input string
		err   bool
	}{
		{input: "10 DEF FNID(x) = x : PRINT FNID(5)"},
		{input: "20 DEF FNMUL(x,y) = x*y : PRINT FNMUL(2,3)"},
		{input: "30 DEF FNSKIP(x)= (x + 2): PRINT FNSKIP(3)"},
		{input: "40 DEF FN(z) = z + 2", err: true},
		{input: "50 DEF AFUNC(t) = t * 5", err: true},
		{input: "60 DEF FNMUL(x,y)", err: true},
		{input: "70 DEF FNMUL  = 5", err: true},
		{input: "80 DEF FNMUL(x,y)", err: true},
		{input: "90 DEF FNMUL(x,y = x * y", err: true},
		{input: "100 DEF FNMUL() = x * y"},
		{input: "110 X$ = MKD$(65999)"},
		{input: "120 MKD$(65999)", err: true},
	}
	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.input, env)

		assert.NotZero(t, sl.LineLength())

		if tt.err {
			sl.NextStatement()
			stmt := sl.NextStatement()
			exp, ok := stmt.(*ast.ExpressionStatement)

			if ok {
				// expression statement
				//assert.True(t, (exp.HasTrash() || exp.Expression.HasTrash()), "didn't get expression trash")
				if !exp.HasTrash() && !exp.Expression.HasTrash() {
					assert.False(t, exp.HasTrash(), "bogus test")
				}
			} else {
				// try for a let statement
				let, ok := stmt.(*ast.LetStatement)

				assert.True(t, ok, "got a wierd statement")

				assert.True(t, let.HasTrash() || let.Name.HasTrash(), "failed to get let trash")
			}
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.input, env)

		assert.EqualValues(t, tt.expStmts, sl.LineLength())

		sl.NextStatement()
		stmt := sl.NextStatement()
		endStmt, ok := stmt.(*ast.EndStatement)
		assert.True(t, ok)
		assert.Equal(t, "END", endStmt.TokenLiteral())
	}
}

func Test_FilesCommand(t *testing.T) {
	tests := []struct {
		input string
	}{
		{`FILES`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.input, env)

		// ToDo: finish writing this test

	}
}

func Test_FixedLiteral(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: `10 X = 12.5`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		// ToDo: finish writing this test
	}
}

func Test_ForStatement(t *testing.T) {
	tests := []struct {
		inp string
	}{
		{inp: `10 FOR I = 1 to 10 STEP 2`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.inp, env)

		// ToDo: finish writing this test

	}
}

func Test_PrintStatements(t *testing.T) {
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
		{`70 PRINT TAB(20);"Hello"`, 2},
		{`80 PRINT " ";USING Z$;Z;:PRINT " ";C$(Z+Y);  'comment`, 4},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		ParseInput(tt.input, env)

		stmt := env.NextStatement()
		assert.NotNil(t, stmt)

		stmt = env.NextStatement()
		assert.NotNil(t, stmt)

		//ToDo: expect results
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
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		stmt := sl.NextStatement()

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
			t.Fatalf("bad value from %s, got %s, wanted %s", inp, gotFixed.Value.Literal, wantVal.Value.Literal)
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

func Test_ViewStatement(t *testing.T) {
	tests := []struct {
		inp string
		exp int // expected size of vw.Parms
	}{
		{inp: `VIEW PRINT 3 TO 23`, exp: 3},
		{inp: `VIEW SCREEN (5,5) - (120,150)`},
	}

	for _, tt := range tests {
		env := object.NewTermEnvironment(mocks.MockTerm{})
		sl := ParseInput(tt.inp, env)

		cmd := sl.NextStatement()
		assert.Equal(t, tt.inp, cmd.String())
	}
}
