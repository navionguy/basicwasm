package token

import "strings"

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	EOL     = "EOL"

	// Identifiers + literals
	BSTR    = "BSTR"  //string of bytes
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

	HEX   = "&H"
	OCTAL = "&O"

	// Keywords
	AUTO   = "AUTO"
	CLS    = "CLS"
	CLEAR  = "CLEAR"
	COMMON = "COMMON"
	DATA   = "DATA"
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
	READ   = "READ"
	REM    = "REM"
	RETURN = "RETURN"
	RUN    = "RUN"
	THEN   = "THEN"
	TRON   = "TRON"
	TROFF  = "TROFF"
	TRUE   = "TRUE"
	USING  = "USING"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"auto":   AUTO,
	"clear":  CLEAR,
	"cls":    CLS,
	"data":   DATA,
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
	"read":   READ,
	"rem":    REM,
	"return": RETURN,
	"run":    RUN,
	"then":   THEN,
	"tron":   TRON,
	"troff":  TROFF,
	"true":   TRUE,
	"using":  USING,
}

func LookupIdent(ident string) TokenType {

	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
