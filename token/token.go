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
	HASHTAG   = "#"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	HEX   = "&H"
	OCTAL = "&O"

	// Keywords
	ACCESS  = "ACCESS"
	ALL     = "ALL"
	APPEND  = "APPEND"
	AS      = "AS"
	AUTO    = "AUTO"
	BEEP    = "BEEP"
	BUILTIN = "BUILTIN"
	CHAIN   = "CHAIN"
	CHDIR   = "CHDIR"
	CLEAR   = "CLEAR"
	CLOSE   = "CLOSE"
	CLS     = "CLS"
	COLOR   = "COLOR"
	COMMON  = "COMMON"
	CONT    = "CONT"
	CSRLIN  = "CSRLIN"
	DATA    = "DATA"
	DEF     = "DEF"
	DELETE  = "DELETE"
	DIM     = "DIM"
	ELSE    = "ELSE"
	END     = "END"
	ERROR   = "ERROR"
	FALSE   = "FALSE"
	FILES   = "FILES"
	FOR     = "FOR"
	GOSUB   = "GOSUB"
	GOTO    = "GOTO"
	IF      = "IF"
	INKEY   = "INKEY$"
	INPUT   = "INPUT"
	KEY     = "KEY"
	LEN     = "LEN"
	LET     = "LET"
	LIST    = "LIST"
	LOAD    = "LOAD"
	LOCATE  = "LOCATE"
	LOCK    = "LOCK"
	MERGE   = "MERGE"
	MOD     = "MOD"
	NEW     = "NEW"
	NEXT    = "NEXT"
	OFF     = "OFF"
	ON      = "ON"
	OPEN    = "OPEN"
	OUTPUT  = "OUTPUT"
	PALETTE = "PALETTE"
	PRINT   = "PRINT"
	RANDOM  = "RANDOM"
	READ    = "READ"
	REM     = "REM"
	RESTORE = "RESTORE"
	RESUME  = "RESUME"
	RETURN  = "RETURN"
	RUN     = "RUN"
	SCREEN  = "SCREEN"
	SHARED  = "SHARED"
	STOP    = "STOP"
	THEN    = "THEN"
	TO      = "TO"
	TRON    = "TRON"
	TROFF   = "TROFF"
	TRUE    = "TRUE"
	USING   = "USING"
	VIEW    = "VIEW"
	WRITE   = "WRITE"
)

type Token struct {
	Type    TokenType
	Literal string
}

var keywords = map[string]TokenType{
	"access":  ACCESS,
	"all":     ALL,
	"append":  APPEND,
	"auto":    AUTO,
	"as":      AS,
	"beep":    BEEP,
	"builtin": BUILTIN,
	"chain":   CHAIN,
	"chdir":   CHDIR,
	"clear":   CLEAR,
	"close":   CLOSE,
	"cls":     CLS,
	"color":   COLOR,
	"common":  COMMON,
	"cont":    CONT,
	"csrlin":  CSRLIN,
	"data":    DATA,
	"def":     DEF,
	"delete":  DELETE,
	"dim":     DIM,
	"else":    ELSE,
	"end":     END,
	"error":   ERROR,
	"false":   FALSE,
	"files":   FILES,
	"for":     FOR,
	"gosub":   GOSUB,
	"goto":    GOTO,
	"if":      IF,
	"inkey$":  INKEY,
	"input":   INPUT,
	"key":     KEY,
	//"len":     LEN,
	"let":     LET,
	"list":    LIST,
	"load":    LOAD,
	"locate":  LOCATE,
	"lock":    LOCK,
	"merge":   MERGE,
	"mod":     MOD,
	"new":     NEW,
	"next":    NEXT,
	"off":     OFF,
	"on":      ON,
	"open":    OPEN,
	"output":  OUTPUT,
	"palette": PALETTE,
	"print":   PRINT,
	"random":  RANDOM,
	"read":    READ,
	"rem":     REM,
	"restore": RESTORE,
	"resume":  RESUME,
	"return":  RETURN,
	"run":     RUN,
	"screen":  SCREEN,
	"shared":  SHARED,
	"stop":    STOP,
	"then":    THEN,
	"to":      TO,
	"tron":    TRON,
	"troff":   TROFF,
	"true":    TRUE,
	"using":   USING,
	"view":    VIEW,
	"write":   WRITE,
}

// LookupIdent returns a TokenType object
func LookupIdent(ident string) TokenType {

	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
