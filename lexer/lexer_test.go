package lexer

import (
	"testing"

	"github.com/navionguy/basicwasm/token"
)

func TestNextToken(t *testing.T) {

	input := `10 REM let five = 5
	20 let ten = 10
	30 let add = fn(x, y) 	
	40 let result = add(five, ten)
	50 !-/*5 	5 < 10 >	if ( 5 < 10 ) 		return true	 else 		return false	
	60 10 == 10	9 != 10	9 <= 10	10>=9	let result = "Hello there!" $%:[]+&<>
	`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.EOL, "\n"},
		{token.INT, "10"},
		{token.REM, "REM"},
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="}, //5
		{token.INT, "5"},
		{token.EOL, "\n"},
		{token.INT, "20"},
		{token.LET, "let"},
		{token.IDENT, "ten"}, // 10
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.EOL, "\n"},
		{token.INT, "30"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.IDENT, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.EOL, "\n"},
		{token.INT, "40"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.EOL, "\n"},
		{token.INT, "50"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.ELSE, "else"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.EOL, "\n"},
		{token.INT, "60"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.INT, "9"},
		{token.NOT_EQ, "!="},
		{token.INT, "10"},
		{token.INT, "9"},
		{token.LTE, "<="},
		{token.INT, "10"},
		{token.INT, "10"},
		{token.GTE, ">="},
		{token.INT, "9"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.STRING, "Hello there!"},
		{token.TYPE_STR, "$"},
		{token.TYPE_INT, "%"},
		{token.COLON, ":"},
		{token.LBRACKET, "["},
		{token.RBRACKET, "]"},
		{token.PLUS, "+"},
		{token.AMPERSAND, "&"},
		{token.NOT_EQ, "<>"},
		{token.EOL, "\n"},
		{token.EOF, "EOF"},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestReadNumber(t *testing.T) {
	tests := []struct {
		input string
		tok   token.TokenType
	}{
		{"43.8", token.FIXED},
		{"235.988E-7", token.FLOAT},
		{"235.988D-12", token.FLOAT},
	}

	for _, tt := range tests {
		l := New(tt.input)
		l.position = 0
		l.readChar()

		tk, val := l.readNumber()

		if tk != tt.tok {
			t.Fatalf("got back a %s, expected a %s", tk, tt.tok)
		}

		if val != tt.input {
			t.Fatalf("expected to get back %s, got back %s instead", tt.input, val)
		}
	}
}

func TestLineNumbers(t *testing.T) {
	input := `
	10
	20
	30`

	l := New(input)

	tt := l.NextToken()
	tt = l.NextToken()
	tt = l.NextToken()

	println(tt.Literal)
}

func TestStatements(t *testing.T) {
	type result struct {
		expectedType    token.TokenType
		expectedLiteral string
	}

	tests := []struct {
		input   string
		results []result
	}{
		{`10 DIM A[10]`, []result{
			{token.EOL, "\n"},
			{token.INT, "10"},
			{token.DIM, "DIM"},
			{token.IDENT, "A"},
			{token.LBRACKET, "["},
			{token.INT, "10"},
			{token.RBRACKET, "]"},
		}},
	}

	for _, tt := range tests {
		l := New(tt.input)

		for _, res := range tt.results {
			nt := l.NextToken()

			if nt.Type != res.expectedType {
				t.Errorf("expected tok %s, got %s", res.expectedType, nt.Type)
			}

			if nt.Literal != res.expectedLiteral {
				t.Errorf("expected literal %s, got %s", res.expectedLiteral, nt.Literal)
			}
		}

		for nt := l.NextToken(); nt.Type != token.EOF; nt = l.NextToken() {
			t.Errorf("extra token %s - %s", nt.Type, nt.Literal)
		}
	}
}
