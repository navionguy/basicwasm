package token

import "strings"

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	EOL     = "EOL"

	// Identifiers + literals
	IDENT   = "IDENT" // add, foobar, x, y, ...
	LINENUM = "####"  // 10, 15, 20, ...
	INT     = "INT"   // -32768 to 32767
	INTD    = "INTD"
	STRING  = "STRING" // "A string literal"
	FLOAT   = "FLOAT"  // 2.539999E+01
	FIXED   = "FIXED"  // 42.14

	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	BSLASH   = "\\"
	MOD      = "MOD"

	LT = "<"
	GT = ">"

	EQ     = "="
	NOT_EQ = "<>"
	GTE    = ">="
	LTE    = "<="

	// Type designators
	TYPE_STR = "$"
	TYPE_INT = "%"
	TYPE_SGL = "!"
	TYPE_DBL = "#"

	// Delimiters
	PERIOD    = "."
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	AMPERSAND = "&"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	CLEAR    = "CLEAR"
	COMMON   = "COMMON"
	DIM      = "DIM"
	DEFINE   = "DEFINE"
	FUNCTION = "DEF"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	THEN     = "THEN"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	GOTO     = "GOTO"
	GOSUB    = "GOSUB"
	LOCATE   = "LOCATE"
	PRINT    = "PRINT"
	USING    = "USING"
	REM      = "REM"
	END      = "END"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"def":    FUNCTION,
	"dim":    DIM,
	"let":    LET,
	"true":   TRUE,
	"false":  FALSE,
	"if":     IF,
	"then":   THEN,
	"else":   ELSE,
	"return": RETURN,
	"gosub":  GOSUB,
	"goto":   GOTO,
	"end":    END,
	"locate": LOCATE,
	"clear":  CLEAR,
	"print":  PRINT,
	"using":  USING,
	"mod":    MOD,
}

func LookupIdent(ident string) TokenType {

	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
