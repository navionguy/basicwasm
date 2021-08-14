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
	AUTO    = "AUTO"
	BEEP    = "BEEP"
	CHAIN   = "CHAIN"
	CLEAR   = "CLEAR"
	CLS     = "CLS"
	COLOR   = "COLOR"
	COMMON  = "COMMON"
	CSRLIN  = "CSRLIN"
	DATA    = "DATA"
	DEF     = "DEF"
	DIM     = "DIM"
	ELSE    = "ELSE"
	END     = "END"
	FALSE   = "FALSE"
	FILES   = "FILES"
	GOSUB   = "GOSUB"
	GOTO    = "GOTO"
	IF      = "IF"
	LET     = "LET"
	LIST    = "LIST"
	LOCATE  = "LOCATE"
	MOD     = "MOD"
	NEW     = "NEW"
	PRINT   = "PRINT"
	READ    = "READ"
	REM     = "REM"
	RESTORE = "RESTORE"
	RETURN  = "RETURN"
	RUN     = "RUN"
	THEN    = "THEN"
	TRON    = "TRON"
	TROFF   = "TROFF"
	TRUE    = "TRUE"
	USING   = "USING"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"auto":    AUTO,
	"beep":    BEEP,
	"chain":   CHAIN,
	"clear":   CLEAR,
	"cls":     CLS,
	"color":   COLOR,
	"common":  COMMON,
	"csrlin":  CSRLIN,
	"data":    DATA,
	"def":     DEF,
	"dim":     DIM,
	"else":    ELSE,
	"end":     END,
	"false":   FALSE,
	"files":   FILES,
	"gosub":   GOSUB,
	"goto":    GOTO,
	"if":      IF,
	"let":     LET,
	"list":    LIST,
	"locate":  LOCATE,
	"mod":     MOD,
	"new":     NEW,
	"print":   PRINT,
	"read":    READ,
	"rem":     REM,
	"restore": RESTORE,
	"return":  RETURN,
	"run":     RUN,
	"then":    THEN,
	"tron":    TRON,
	"troff":   TROFF,
	"true":    TRUE,
	"using":   USING,
}

// LookupIdent returns a TokenType object
func LookupIdent(ident string) TokenType {

	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}

var builtins = map[string]bool{
	"ABS":  true,
	"ASC":  true,
	"ATN":  true,
	"CDBL": true,
	"CHR$": true,
	"CINT": true,
	"COS":  true,
	"CSNG": true,
}
