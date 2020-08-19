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
	CLS    = "CLS"
	CLEAR  = "CLEAR"
	COMMON = "COMMON"
	DEF    = "DEF"
	DIM    = "DIM"
	ELSE   = "ELSE"
	END    = "END"
	FALSE  = "FALSE"
	GOSUB  = "GOSUB"
	GOTO   = "GOTO"
	IF     = "IF"
	LET    = "LET"
	LIST   = "LIST"
	LOCATE = "LOCATE"
	MOD    = "MOD"
	PRINT  = "PRINT"
	REM    = "REM"
	RETURN = "RETURN"
	THEN   = "THEN"
	TRUE   = "TRUE"
	USING  = "USING"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"clear":  CLEAR,
	"cls":    CLS,
	"def":    DEF,
	"dim":    DIM,
	"else":   ELSE,
	"end":    END,
	"false":  FALSE,
	"gosub":  GOSUB,
	"goto":   GOTO,
	"if":     IF,
	"let":    LET,
	"list":   LIST,
	"locate": LOCATE,
	"mod":    MOD,
	"print":  PRINT,
	"rem":    REM,
	"return": RETURN,
	"then":   THEN,
	"true":   TRUE,
	"using":  USING,
}

func LookupIdent(ident string) TokenType {

	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
