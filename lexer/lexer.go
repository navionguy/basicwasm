package lexer

import (
	"github.com/navionguy/basicwasm/token"
)

//Lexer a lexical analyzer instance
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	newLine      bool // flag that I'm at the start of a line
}

//New create a new lexer object
func New(input string) *Lexer {
	l := &Lexer{
		input:    input,
		newLine:  true,
		position: -1,
	}
	return l
}

//NextToken scans for the next token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	// detect startup
	if -1 == l.position {
		l.position++
		l.readChar()
		return newToken(token.EOL, '\n')
	}

	l.skipWhitespace()

	switch l.ch {
	// Type tokens
	case '\n':
		tok = newToken(token.EOL, l.ch)
	case '$':
		tok = newToken(token.TYPE_STR, l.ch)
	case '%':
		tok = newToken(token.TYPE_INT, l.ch)
	case '#':
		tok = newToken(token.TYPE_DBL, l.ch)
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case ':':
		tok = newToken(token.COLON, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
	case ']':
		tok = newToken(token.RBRACKET, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '&':
		tok = newToken(token.AMPERSAND, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.TYPE_SGL, l.ch)
		}
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.LTE, Literal: literal}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.GTE, Literal: literal}
		} else {
			tok = newToken(token.GT, l.ch)
		}
	case '\\':
		tok = newToken(token.BSLASH, l.ch)
	case '"':
		literal := l.readString()
		tok = token.Token{Type: token.STRING, Literal: literal}
	case 0:
		tok.Literal = token.EOF
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type, tok.Literal = l.readNumber()
			if l.newLine && (tok.Type == token.INT) {
				tok.Type = token.LINENUM
				l.newLine = false
			}
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if (l.ch == '"') || (l.ch == 0) {
			break
		}
	}

	return l.input[position:l.position]
}

// reads a numeric value and
// reads a string of digits
func (l *Lexer) readNumber() (token.TokenType, string) {
	var tt token.TokenType
	tt = token.INT
	position := l.position

	err := false
	for !err {
		switch l.ch {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			l.readChar()
		case '.':
			err, tt = l.chgType(tt, token.INT, token.FIXED)
			l.readChar()
		case 'e', 'E', 'd', 'D':
			err, tt = l.chgType(tt, token.FIXED, token.FLOAT)
			if err {
				err, tt = l.chgType(tt, token.INT, token.FLOAT)
			}
			l.readChar()
			if (l.ch == '-') || (l.ch == '+') {
				l.readChar()
			}
		default:
			err = true
		}
	}

	return tt, l.input[position:l.position]
}

func (l *Lexer) chgType(curTok token.TokenType, ifTok token.TokenType, newTok token.TokenType) (bool, token.TokenType) {
	if curTok == ifTok {
		return false, newTok
	}
	return true, curTok
}

func (l *Lexer) readHexConstant() int {
	return 0
}

func (l *Lexer) readOctalConstant() int {
	return 0
}

//peekChar - take a look at, but don't consume the next character
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}

	return l.input[l.readPosition]
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' {
		if l.ch == '\n' {
			l.newLine = true
			return
		}
		l.readChar()
	}
}
