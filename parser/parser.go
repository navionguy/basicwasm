package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/builtins"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/token"
)

const (
	_ int = iota
	// LOWEST defines the bottom of the priority stack
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.GTE:      LESSGREATER,
	token.LTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.BSLASH:   PRODUCT,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MOD:      PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

// Parser an instance
type Parser struct {
	l      *lexer.Lexer // the lexer feeding me tokens
	errors []string     // array of error messages, TODO: stop parsing after first error

	curToken  token.Token
	peekToken token.Token
	curLine   int  // current line number being parsed
	cmdInput  bool // are we parsing from the terminal?
	env       *object.Environment

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New create and return a Parser instance
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:       l,
		curLine: 0,
		errors:  []string{},
	}

	// create map parsers for prefix elements
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.AMPERSAND, p.parseHexOctalConstant)
	p.registerPrefix(token.CSRLIN, p.parseCsrLinVar)
	p.registerPrefix(token.DEF, p.parseFunctionLiteral)
	p.registerPrefix(token.EOF, p.parseEOFExpression)
	p.registerPrefix(token.FLOAT, p.parseFloatingPointLiteral)
	p.registerPrefix(token.FIXED, p.parseFixedPointLiteral)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.INTD, p.parseIntDoubleLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.STRING, p.parseStringLiteral)

	// and infix elements
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.BSLASH, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.RPAREN, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

// Errors returns list of errors seen while parsing
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	// If I see EOL followed by INT, that is actually a line number
	if p.curTokenIs(token.EOL) && p.peekTokenIs(token.INT) {
		p.peekToken.Type = token.LINENUM
	}
}

// ParseProgram time to get busy and build the Abstract Syntax Tree
// The program object holds the code and he lives in the environment
func (p *Parser) ParseProgram(env *object.Environment) {
	defer untrace(trace("ParseProgram"))
	p.env = env

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			env.AddStatement(stmt)
		}
		p.nextToken()
	}

	env.Parsed()
}

// ParseCmd is used to parse out a command entered directly
//
func (p *Parser) ParseCmd(env *object.Environment) {
	defer untrace(trace("ParseCmd"))

	if p.peekTokenIs(token.LINENUM) {
		p.ParseProgram(env)
		return
	}

	p.env = env

	p.cmdInput = true
	for (!p.curTokenIs(token.EOF)) && (len(p.errors) == 0) {
		stmt := p.parseStatement()
		if stmt != nil {
			env.AddCmdStmt(stmt)
		}
		p.nextToken()
	}

	env.CmdParsed()
	return
}

func (p *Parser) parseStatement() ast.Statement {
	defer untrace(trace("parseStatement"))
	switch p.curToken.Type {
	case token.AUTO:
		return p.parseAutoCommand()
	case token.BEEP:
		return p.parseBeepStatement()
	case token.CHAIN:
		return p.parseChainStatement()
	case token.CHDIR:
		return p.parseChDirStatement()
	case token.CLEAR:
		return p.parseClearCommand()
	case token.CLS:
		return p.parseClsStatement()
	case token.COLOR:
		return p.parseColorStatement()
	case token.COMMON:
		return p.parseCommonStatement()
	case token.CONT:
		return p.parseContCommand()
	case token.DATA:
		return p.parseDataStatement()
	case token.DIM:
		return p.parseDimStatement()
	case token.END:
		return p.parseEndStatement()
	case token.EOL:
		// EOF means that was the last line
		if p.peekTokenIs(token.EOF) {
			stmt := &ast.EndStatement{}
			return stmt
		}
		return nil
	case token.FILES:
		return p.parseFilesCommand()
	case token.FOR:
		return p.parseForStatement()
	case token.GOSUB:
		return p.parseGosubStatement()
	case token.GOTO:
		return p.parseGotoStatement()
	case token.IF:
		return p.parseExpressionStatement()
	case token.KEY:
		return p.parseKeyStatement()
	case token.LET:
		return p.parseLetStatement()
	case token.LINENUM:
		return p.parseLineNumber()
	case token.LIST:
		return p.parseListStatement()
	case token.LOCATE:
		return p.parseLocateStatement()
	case token.LOAD:
		return p.parseLoadCommand()
	case token.NEW:
		return p.parseNewCommand()
	case token.NEXT:
		return p.parseNextStatement()
	case token.PALETTE:
		return p.parsePaletteStatement()
	case token.READ:
		return p.parseReadStatement()
	case token.REM:
		return p.parseRemStatement()
	case token.RESTORE:
		return p.parseRestoreStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.RUN:
		return p.parseRunCommand()
	case token.SCREEN:
		return p.parseScreenCommand()
	case token.STOP:
		return p.parseStopStatement()
	case token.TROFF:
		return p.parseTroffCommand()
	case token.TRON:
		return p.parseTronCommand()
	case token.PRINT:
		return p.parsePrintStatement()
	case token.VIEW:
		return p.parseViewStatement()
	default:
		if strings.ContainsAny(p.peekToken.Literal, "=[($%!#") {
			stmt := p.parseImpliedLetStatement(p.curToken.Literal)

			if !p.checkForFuncCall() {
				return stmt
			}
			// yikes!  It is actually a function call
			// recover the full name

			stmt.Value = p.parseCallExpression(stmt.Name)
			return stmt
			//p.curToken = token.Token{Type: token.IDENT, Literal: stmt.Name.Value}
		}
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseAutoCommand() *ast.AutoCommand {
	auto := &ast.AutoCommand{Token: p.curToken, Start: -1, Increment: 10, Curr: false}

	// check for the starting line number
	if p.peekTokenIs(token.INT) {
		p.nextToken()
		auto.Start, _ = strconv.Atoi(p.curToken.Literal)
	}

	// check for '.' to start with the current line number
	if p.peekTokenIs(token.PERIOD) {
		p.nextToken()
		auto.Curr = true
	}

	// did he specify an increment value?
	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(token.INT) {
			p.nextToken()
			auto.Increment, _ = strconv.Atoi(p.curToken.Literal)
		}
	}
	p.nextToken()

	return auto
}

// he has no params, he just, well, beeps
func (p *Parser) parseBeepStatement() *ast.BeepStatement {
	beep := ast.BeepStatement{Token: p.curToken}

	return &beep
}

// a questionable name for parsing a function definition
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	for !p.curTokenIs(token.COLON) && !p.curTokenIs(token.EOF) && !p.curTokenIs(token.EOL) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}
	//p.nextToken()

	return block
}

// parseChainStatement()
func (p *Parser) parseChainStatement() *ast.ChainStatement {
	chain := ast.ChainStatement{Token: p.curToken}

	p.nextToken()

	// check for merge option
	if p.curTokenIs(token.MERGE) {
		chain.Merge = true
		p.nextToken()
	}

	return p.parseChainPath(&chain)
}

// parse out the path for file to chain in
func (p *Parser) parseChainPath(chain *ast.ChainStatement) *ast.ChainStatement {
	chain.Path = p.parseExpression(LOWEST)

	if !p.peekTokenIs(token.COMMA) {
		return chain
	}

	return p.parseChainParams(chain)
}

// parse any parameters to the chain statement
func (p *Parser) parseChainParams(chain *ast.ChainStatement) *ast.ChainStatement {
	for i := 0; p.peekTokenIs(token.COMMA); i++ {
		p.nextToken()
		p.parseChainParameter(i, chain)
	}
	return chain
}

// parse the next chain statement parameter
func (p *Parser) parseChainParameter(param int, chain *ast.ChainStatement) {
	p.nextToken()

	switch param {
	case 0: // start at linenum
		chain.Line = p.parseExpression(LOWEST)
	case 1: // ALL preserve variables
		if !p.curTokenIs(token.ALL) {
			p.reportError(berrors.Syntax)
		}
		chain.All = true
	case 2: // DELETE range of lines
		chain.Delete = true
		p.nextToken()
		chain.Range = p.parseExpression(LOWEST)
	}
	return
}

// ChDir should have one expression that evaluates to a string
func (p *Parser) parseChDirStatement() *ast.ChDirStatement {
	defer untrace(trace("parseChDirStatement"))
	cd := ast.ChDirStatement{Token: p.curToken}

	if p.chkEndOfStatement() {
		return &cd // no path supplied, evaluator will display error
	}

	p.nextToken()
	cd.Path = append(cd.Path, p.parseExpression(LOWEST))
	p.nextToken()

	return &cd
}

func (p *Parser) parseClearCommand() *ast.ClearCommand {
	defer untrace(trace("parseClearStatement"))
	clr := ast.ClearCommand{Token: p.curToken}

	if p.chkEndOfStatement() {
		return &clr // no parameters to the clear
	}

	for i := 0; (i < 3) && !p.chkEndOfStatement(); i++ {
		p.nextToken()
		if !p.curTokenIs(token.COMMA) {
			clr.Exp[i] = p.parseExpression(LOWEST)
			p.nextToken()
		}
	}

	return &clr
}

func (p *Parser) parseClsStatement() *ast.ClsStatement {
	defer untrace(trace("parseClsStatement"))
	stmt := ast.ClsStatement{Token: p.curToken, Param: -1}

	if p.peekTokenIs(token.INT) {
		p.nextToken()
		stmt.Param, _ = strconv.Atoi(p.curToken.Literal)
	}

	p.nextToken()

	return &stmt
}

// parse th color statement
func (p *Parser) parseColorStatement() *ast.ColorStatement {
	defer untrace(trace("parseColorStatement"))
	stmt := ast.ColorStatement{Token: p.curToken}

	if p.chkEndOfStatement() {
		return &stmt
	}
	p.nextToken()

	exp := p.parseCommaSperatedExpressions()

	if len(exp) > 3 {
		p.reportError(berrors.Syntax)
		return &stmt
	}

	stmt.Parms = exp
	return &stmt
}

// pars
func (p *Parser) parseCommonStatement() *ast.CommonStatement {
	defer untrace(trace("parseCommonStatement"))
	stmt := ast.CommonStatement{Token: p.curToken}

	for !p.chkEndOfStatement() {
		p.nextToken()
		stmt.Vars = append(stmt.Vars, p.innerParseIdentifier())

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return &stmt
}

func (p *Parser) parseContCommand() *ast.ContCommand {
	defer untrace(trace("parseContCommand"))
	cmd := ast.ContCommand{Token: token.Token{Type: token.CONT, Literal: "CONT"}}

	return &cmd
}

func (p *Parser) parseCsrLinVar() ast.Expression {
	defer untrace(trace("parseCsrLinVar()"))
	csr := ast.Csrlin{Token: p.curToken}

	return &csr
}

func (p *Parser) parseDataStatement() *ast.DataStatement {
	defer untrace(trace("parseDataStatement"))
	stmt := &ast.DataStatement{Token: p.curToken}

	p.l.PassOn()
	elem := ""
	for !p.peekTokenIs(token.LINENUM) && !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.EOL) && !p.peekTokenIs(token.COLON) {
		p.nextToken()

		if p.curTokenIs(token.COMMA) || p.curTokenIs(token.COLON) {
			stmt.Consts = append(stmt.Consts, p.parseDataElement(strings.Trim(elem, " ")))
			elem = ""
		} else {
			elem = elem + p.curToken.Literal
		}
	}

	stmt.Consts = append(stmt.Consts, p.parseDataElement(strings.Trim(elem, " ")))

	p.l.PassOff()
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseDataElement(elem string) ast.Expression {
	l := lexer.New(elem)
	ip := New(l)
	ip.nextToken()
	stmt := ip.parseStatement()

	if len(ip.errors) != 0 {
		p.errors = append(p.errors, ip.errors...)
		return nil
	}

	ln, ok := stmt.(*ast.LineNumStmt)

	if ok {
		// line number is actually a int or double int
		if (ln.Value < 32767) && (ln.Value > -32768) {
			tk := token.Token{Type: token.INT, Literal: "INT"}
			return &ast.IntegerLiteral{Token: tk, Value: int16(ln.Value)}
		}

		tk := token.Token{Type: token.INTD, Literal: "INTD"}
		return &ast.DblIntegerLiteral{Token: tk, Value: ln.Value}
	}

	exp, ok := stmt.(*ast.ExpressionStatement)

	if !ok {
		return nil
	}

	if exp.Token.Type == token.IDENT {
		// if it parsed to an identifier, it's actually a string
		// and ignore the parser flipping it to all caps
		tk := token.Token{Type: token.STRING, Literal: "STRING"}
		return &ast.StringLiteral{Token: tk, Value: elem}
	}

	return exp.Expression
}

func (p *Parser) parseDimStatement() *ast.DimStatement {
	defer untrace(trace("parseDimStatement"))
	exp := &ast.DimStatement{Token: p.curToken, Vars: []*ast.Identifier{}}

	for !p.chkEndOfStatement() {
		p.nextToken()
		exp.Vars = append(exp.Vars, p.innerParseIdentifier())

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return exp
}

func (p *Parser) parseEndStatement() *ast.EndStatement {
	defer untrace(trace("parseEndStatement"))
	stmt := &ast.EndStatement{Token: p.curToken}

	return stmt
}

// parse in a FILES command
func (p *Parser) parseFilesCommand() *ast.FilesCommand {
	defer untrace(trace("parseFilesCommand"))
	cd := &ast.FilesCommand{Token: p.curToken}

	if p.peekTokenIs(token.STRING) {
		p.nextToken()
		cd.Path = p.curToken.Literal
	}

	return cd
}

// parse the begining of a FOR loop
func (p *Parser) parseForStatement() *ast.ForStatment {
	defer untrace(trace("parseForStatement"))
	four := ast.ForStatment{Token: p.curToken}
	p.nextToken()

	// get the starting assignment
	if p.curTokenIs(token.IDENT) {
		four.Init = p.parseImpliedLetStatement(p.curToken.Literal)
		p.nextToken()
	}

	// get the termination value, if it is there
	if p.curTokenIs(token.TO) && (strings.EqualFold(p.curToken.Literal, "to")) {
		p.nextToken()
		// read the final expression
		four.Final = append(four.Final, p.parseExpression(LOWEST))
	}

	// their may be a step size specified
	if p.peekTokenIs(token.IDENT) && (strings.EqualFold(p.peekToken.Literal, "step")) {
		p.nextToken()
		p.nextToken()
		four.Step = append(four.Step, p.parseExpression(LOWEST))
	}

	return &four
}

// parse an Integer Literal
func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.Atoi(p.curToken.Literal)
	if err != nil {
		p.reportError(berrors.Syntax)
		return nil
	}

	if (value > math.MaxInt16) || (value < math.MinInt16) {
		return p.buildDoubleIIntegerLiteral(value)
	}

	lit.Value = int16(value)
	return lit
}

func (p *Parser) parseHexOctalConstant() ast.Expression {
	defer untrace(trace("parseHexOctalConstant"))

	if p.peekTokenIs(token.IDENT) && (strings.Compare(p.peekToken.Literal[0:1], "H") == 0) {
		return p.parseHexConstant()
	}

	if !p.peekTokenIs(token.INT) && !p.peekTokenIs(token.IDENT) {
		return nil
	}

	return p.parseOctalConstant()
}

func (p *Parser) parseOctalConstant() ast.Expression {
	tk := token.Token{Type: token.OCTAL, Literal: "&"}
	lit := &ast.OctalConstant{Token: tk}
	p.nextToken()

	if p.curTokenIs(token.IDENT) {
		if strings.Compare(p.curToken.Literal, "O") != 0 {
			p.reportError(berrors.Syntax)
			return nil
		}
		lit.Token = token.Token{Type: token.OCTAL, Literal: "&O"}
		p.nextToken()
	}

	if !p.curTokenIs(token.INT) {
		p.reportError(berrors.TypeMismatch)
		return nil
	}

	lit.Value = p.curToken.Literal

	return lit
}

func (p *Parser) parseHexConstant() ast.Expression {
	tk := token.Token{Type: token.HEX, Literal: "&H"}
	lit := &ast.HexConstant{Token: tk}

	// there may be hex A-F stuck to the H
	p.nextToken()

	val := ""
	if len(p.curToken.Literal) > 1 {
		val += p.curToken.Literal[1:]
	}

	// scoop up all letters and numbers
	// we'll figure out if they are valid later
	for p.peekTokenIs(token.IDENT) || p.peekTokenIs(token.INT) {
		p.nextToken()
		val += p.curToken.Literal
	}

	lit.Value = val

	return lit
}

func (p *Parser) parseIntDoubleLiteral() ast.Expression {
	defer untrace(trace("parseIntDoubleLiteral"))

	value, err := strconv.Atoi(strings.TrimRight(p.curToken.Literal, "#"))

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer in line %d", p.curToken.Literal, p.curLine)
		p.errors = append(p.errors, msg)
		return nil
	}

	return p.buildDoubleIIntegerLiteral(value)
}

func (p *Parser) buildDoubleIIntegerLiteral(value int) ast.Expression {
	defer untrace(trace("buildDoubleIIntegerLiteral"))

	// make sure it will fit in in32
	if (value >= math.MinInt32) && (value <= math.MaxInt32) {
		tk := token.Token{Type: token.INTD, Literal: strconv.Itoa(value)}
		lit := &ast.DblIntegerLiteral{Token: tk}
		lit.Value = int32(value)
		return lit
	}

	// Interestingly, basic willingly lost precision in this case
	fv := float32(value)
	tk := token.Token{Type: token.FLOAT, Literal: strconv.Itoa(value)}
	lit := &ast.FloatSingleLiteral{Token: tk, Value: fv}

	return lit
}

func (p *Parser) parseFixedPointLiteral() ast.Expression {
	defer untrace(trace("parseFixedPointLiteral"))

	val, err := decimal.NewFromString(p.curToken.Literal)

	if err != nil {
		msg := fmt.Sprintf("numeric %s invalid at line %d", p.curToken.Literal, p.curLine)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit := &ast.FixedLiteral{Token: p.curToken, Value: val}

	return lit
}

func (p *Parser) parseFloatingPointLiteral() ast.Expression {
	defer untrace(trace("parseFloatingPointLiteral"))
	lit := &ast.FloatSingleLiteral{Token: p.curToken}
	src := p.curToken.Literal
	if strings.ContainsAny(src, "dD") {
		src = strings.Replace(src, "d", "E", 1)
		src = strings.Replace(src, "D", "E", 1)
		return p.parseDoubleFloatingPointLiteral(src)
	}
	value, err := strconv.ParseFloat(p.curToken.Literal, 32)
	if err != nil {
		msg := fmt.Sprintf("could not parse %s as float at line %d", p.curToken.Literal, p.curLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = float32(value)
	return lit
}

func (p *Parser) parseDoubleFloatingPointLiteral(newTokLit string) ast.Expression {
	defer untrace(trace("parseDoubleFloatingPointLiteral"))
	lit := &ast.FloatDoubleLiteral{Token: p.curToken}
	value, err := strconv.ParseFloat(newTokLit, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %s as float at line %d", p.curToken.Literal, p.curLine)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = float64(value)
	return lit

}

func (p *Parser) parseStringLiteral() ast.Expression {
	defer untrace(trace("parseStringLiteral"))
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseLineNumber() *ast.LineNumStmt {
	defer untrace(trace("parseLineNumber"))

	// rebrand it a line number
	stmt := &ast.LineNumStmt{Token: p.curToken}
	stmt.Token.Literal = p.curToken.Literal

	var err error
	tv, err := strconv.Atoi(p.curToken.Literal)

	if err != nil {
		p.generalError("Invalid line number")
	}
	stmt.Value = int32(tv)
	p.curLine = tv

	// little detour here, if I see linenum*EOL AND auto is on
	// user has decided *not* to overwrite an existing line
	// I should return nil
	//
	// if I see linenum*statement AND auto is on
	// user has decided to overwrite an existing line
	// I should consume and ignore the asterisk

	if p.peekTokenIs(token.ASTERISK) && (p.env.GetAuto() != nil) {
		p.nextToken()

		if p.peekTokenIs(token.EOL) || p.peekTokenIs(token.EOF) {
			return stmt
		}

		// consume the asterisk
		p.nextToken()
	}

	return stmt
}

// user wants to list part or all of the program
func (p *Parser) parseListStatement() *ast.ListStatement {
	defer untrace(trace("parseListStatement"))
	stmt := &ast.ListStatement{Token: p.curToken, Start: "", Lrange: "", Stop: ""}

	if !p.peekTokenIs(token.INT) && !p.peekTokenIs(token.MINUS) {
		p.nextToken()
		return stmt
	}

	if p.peekTokenIs(token.INT) {
		p.nextToken()
		stmt.Start = p.curToken.Literal
	}

	if !p.peekTokenIs(token.MINUS) {
		p.nextToken()
		return stmt
	}

	p.nextToken()
	stmt.Lrange = p.curToken.Literal

	if !p.peekTokenIs(token.INT) {
		p.nextToken()
		return stmt
	}

	p.nextToken()
	stmt.Stop = p.curToken.Literal

	p.nextToken()
	return stmt
}

func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	defer untrace(trace("parsePrintStatement"))
	stmt := &ast.PrintStatement{Token: p.curToken}

	for !p.chkEndOfStatement() {
		p.nextToken()
		stmt.Items = append(stmt.Items, p.parseExpression(LOWEST))

		if p.peekTokenIs(token.COMMA) || p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
			stmt.Seperators = append(stmt.Seperators, p.curToken.Literal)
		} else {
			stmt.Seperators = append(stmt.Seperators, " ")
		}
	}

	p.nextToken()
	return stmt
}

// returns true if current token is the end of the statement
func (p *Parser) chkOnEndOfStatement() bool {
	return p.curTokenIs(token.COLON) || p.curTokenIs(token.LINENUM) || p.curTokenIs(token.EOF) || p.curTokenIs(token.EOL)
}

// returns true if the next token would put us at the end of a statement
func (p *Parser) chkEndOfStatement() bool {
	return p.peekTokenIs(token.COLON) || p.peekTokenIs(token.LINENUM) || p.peekTokenIs(token.EOF) || p.peekTokenIs(token.EOL)
}

// gosub - uncondition transfer to subroutine
func (p *Parser) parseGosubStatement() *ast.GosubStatement {
	defer untrace(trace("parseGosubStatement"))
	stmt := ast.GosubStatement{Token: p.curToken, Gosub: 0}

	if !p.expectPeek(token.INT) {
		return &stmt
	}

	stmt.Gosub, _ = strconv.Atoi(p.curToken.Literal)

	if p.peekTokenIs(token.COLON) { // if a colon follows consume it
		p.nextToken()
	}

	return &stmt
}

// goto - uncondition transfer to line
func (p *Parser) parseGotoStatement() *ast.GotoStatement {
	defer untrace(trace("parseGotoStatement"))
	stmt := ast.GotoStatement{Token: p.curToken, Goto: ""}

	if !p.expectPeek(token.INT) {
		return nil
	}

	stmt.Goto = p.curToken.Literal

	if p.peekTokenIs(token.COLON) { // if a colon follows consume it
		p.nextToken()
	}

	return &stmt
}

// Key statement can come in many forms
func (p *Parser) parseKeyStatement() *ast.KeyStatement {
	defer untrace(trace("parseKeyStatement"))

	stmt := &ast.KeyStatement{Token: p.curToken}

	// if there is a parameter, save it
	if !p.chkEndOfStatement() {
		p.nextToken()
		stmt.Param = p.curToken
	}

	// load up any data items
	for ; p.peekTokenIs(token.COMMA); p.nextToken() {
		p.nextToken()
		if !p.chkEndOfStatement() {
			p.nextToken()
			item := p.parseExpression(LOWEST)
			stmt.Data = append(stmt.Data, item)
		}
	}

	return stmt
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer untrace(trace("parseLetStatement"))

	p.curToken.Literal = strings.ToUpper(p.curToken.Literal)
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return stmt
	}
	stmt.Name = p.innerParseIdentifier()

	return p.finishParseLetStatment(stmt)
}

func (p *Parser) parseImpliedLetStatement(id string) *ast.LetStatement {
	defer untrace(trace("parseImpliedLetStatement"))
	tk := token.Token{
		Type:    token.LookupIdent("let"),
		Literal: "",
	}
	stmt := &ast.LetStatement{Token: tk}
	stmt.Name = p.innerParseIdentifier()

	if p.checkForFuncCall() {
		// whoops, it's a function identifier
		return stmt
	}

	return p.finishParseLetStatment(stmt)
}

func (p *Parser) finishParseLetStatment(stmt *ast.LetStatement) *ast.LetStatement {

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

// build array of parameter expressions
func (p *Parser) parseLocateStatement() *ast.LocateStatement {
	defer untrace(trace("parseLocateStatement"))
	stmt := ast.LocateStatement{Token: p.curToken}

	for !p.chkEndOfStatement() {
		p.nextToken()
		if p.curTokenIs(token.COMMA) && p.peekTokenIs(token.COMMA) {
			stmt.Parms = append(stmt.Parms, nil)
			p.nextToken()
			continue
		}

		if p.curTokenIs(token.COMMA) && !p.peekTokenIs(token.COMMA) {
			p.nextToken()
			stmt.Parms = append(stmt.Parms, p.parseExpression(LOWEST))
			continue
		}

		stmt.Parms = append(stmt.Parms, p.parseExpression(LOWEST))
	}
	p.nextToken()

	return &stmt
}

func (p *Parser) parseLoadCommand() *ast.LoadCommand {
	defer untrace(trace("parseLoadCommand"))
	stmt := ast.LoadCommand{Token: p.curToken}

	p.nextToken()
	// parse the Expression to get the filename, hopefully
	stmt.Path = p.parseExpression(LOWEST)

	// if there isn't a comma, we are done
	if p.peekToken.Literal != token.COMMA {
		return &stmt
	}

	return p.parseLoadCommandRunOption(&stmt)
}

// parseLoadCommandRunOption, only called if param is present
func (p *Parser) parseLoadCommandRunOption(cmd *ast.LoadCommand) *ast.LoadCommand {
	p.nextToken()
	p.nextToken()

	if strings.ToUpper(p.curToken.Literal) == "R" {
		cmd.KeppOpen = true
		return cmd
	}

	// anything other than an 'R' or 'r' is a syntax error
	p.reportError(berrors.Syntax)
	return nil
}

// parseNewCommand, a very simple thing to do
func (p *Parser) parseNewCommand() *ast.NewCommand {
	defer untrace(trace("parseNewCommand"))
	cmd := ast.NewCommand{Token: p.curToken}

	return &cmd
}

// parse the NEXT statement
func (p *Parser) parseNextStatement() *ast.NextStatement {
	defer untrace(trace("parseNextStatement"))
	nxt := ast.NextStatement{Token: p.curToken}

	if !p.chkEndOfStatement() {
		p.nextToken()
		if p.curTokenIs(token.IDENT) {
			nxt.Id = *p.innerParseIdentifier()
		} else {
			p.reportError(berrors.Syntax)
		}
	}

	return &nxt
}

// adjust the screen color palette as directed
func (p *Parser) parsePaletteStatement() *ast.PaletteStatement {
	defer untrace(trace("parsePaletteStatement"))
	stmt := &ast.PaletteStatement{Token: p.curToken}
	p.nextToken()

	// check to see if this is a USING version
	if p.curTokenIs(token.USING) {
		return p.parsePaletteUsingStatement(stmt)
	}

	// go handle a single attribute
	return p.parsePaletteSingleStatement(stmt)
}

// parse PALETTE single attribute
func (p *Parser) parsePaletteSingleStatement(stmt *ast.PaletteStatement) *ast.PaletteStatement {
	stmt.Attrib = p.parseIdentifier()
	p.nextToken()

	if !p.curTokenIs(token.COMMA) {
		p.reportError(berrors.Syntax)
		return nil
	}
	p.nextToken()
	stmt.Color = p.parseIdentifier()
	return stmt
}

// parse a PALETTE USING statement
func (p *Parser) parsePaletteUsingStatement(stmt *ast.PaletteStatement) *ast.PaletteStatement {
	// move to the variable to use
	p.nextToken()

	stmt.Color = p.parseIdentifier()
	return stmt
}

// read constant data from DATA statements
func (p *Parser) parseReadStatement() *ast.ReadStatement {
	defer untrace(trace("parseReadStatement"))
	stmt := &ast.ReadStatement{Token: p.curToken}

	for !p.peekTokenIs(token.LINENUM) && !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.EOL) && !p.peekTokenIs(token.COLON) {
		p.nextToken()

		// parse in the variable expression
		exp := p.parseExpression(LOWEST)

		// add him to the list of variables
		if exp != nil {
			stmt.Vars = append(stmt.Vars, exp)
		}

		// check if there is more coming
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

// not a hard one to parse
func (p *Parser) parseRemStatement() *ast.RemStatement {
	defer untrace(trace("parseRemStatement"))
	stmt := &ast.RemStatement{Token: p.curToken}
	//stmt := &ast.RemStatement{Token: p.curToken, Comment: strings.ToUpper(p.curToken.Literal) + " "}

	p.l.PassOn()
	for !p.peekTokenIs(token.LINENUM) && !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.EOL) {
		p.nextToken()
		stmt.Comment += p.curToken.Literal
	}
	p.l.PassOff()
	return stmt
}

// RESTORE resets to read from the beginning of const DATA
// it can optionally take a line number to restore to
func (p *Parser) parseRestoreStatement() *ast.RestoreStatement {
	defer untrace(trace("parseRestoreStatement"))
	stmt := &ast.RestoreStatement{Token: p.curToken, Line: -1}

	if !p.peekTokenIs(token.LINENUM) && !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.EOL) && !p.peekTokenIs(token.COLON) {
		p.nextToken()
		targ, err := strconv.Atoi(p.curToken.Literal)

		if err != nil {
			msg := fmt.Sprintf("undefined line number %s", p.curToken.Literal)
			p.errors = append(p.errors, msg)
			return nil
		}

		stmt.Line = targ
	}

	if p.peekTokenIs(token.COLON) {
		p.nextToken()
	}

	return stmt
}

// returns are much simpler in gwbasic
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	defer untrace(trace("parseReturnStatement"))
	stmt := &ast.ReturnStatement{Token: p.curToken, ReturnTo: ""}

	if p.peekToken.Literal == token.EOF {
		return stmt
	}

	if p.peekTokenIs(token.COLON) { // if a colon follows consume it
		p.nextToken()
		return stmt
	}

	p.nextToken()

	stmt.ReturnTo = p.curToken.Literal

	if p.peekTokenIs(token.COLON) { // if a colon follows consume it
		p.nextToken()
	}

	return stmt
}

// Run commands come in two forms
// RUN [line number][,r]
// RUN filename[,r]
func (p *Parser) parseRunCommand() *ast.RunCommand {
	cmd := ast.RunCommand{Token: p.curToken}

	// check for line number to start at
	if p.peekTokenIs(token.INT) {
		p.nextToken()
		cmd.StartLine, _ = strconv.Atoi(p.curToken.Literal)
	} else if p.peekTokenIs(token.STRING) { // check for file to load
		p.nextToken()
		cmd.LoadFile = p.parseExpression(LOWEST)
	}

	if !p.peekTokenIs(token.COMMA) {
		return &cmd
	}

	return p.parseRunKeepOpen(cmd)
}

// does it look like files should stay open?
func (p *Parser) parseRunKeepOpen(cmd ast.RunCommand) *ast.RunCommand {
	p.nextToken()
	if !p.peekTokenIs(token.IDENT) {
		p.reportError(berrors.Syntax)
		return &cmd
	}

	return p.parseRunValidFlag(cmd)
}

// the only valid flag is "R"
func (p *Parser) parseRunValidFlag(cmd ast.RunCommand) *ast.RunCommand {
	p.nextToken()

	if strings.ToUpper(p.curToken.Literal) != "R" {
		p.reportError(berrors.Syntax)
		return &cmd
	}

	cmd.KeepOpen = true

	return &cmd
}

// ScreenStatement allows user to configure screen mode for
// different display adapters.  MDA,CGA,EGA and such
func (p *Parser) parseScreenCommand() *ast.ScreenStatement {
	// create my object
	stmt := ast.ScreenStatement{Token: p.curToken}

	// need to move past the SCREEN token, unless their are no params
	if p.chkEndOfStatement() {
		p.reportError(berrors.MissingOp)
		return &stmt
	}
	p.nextToken()

	return p.parseScreenCommandParameters(stmt)
}

// parse the parameters for the SCREEN statement
func (p *Parser) parseScreenCommandParameters(stmt ast.ScreenStatement) *ast.ScreenStatement {
	// load up all the parameters, max 4

	stmt.Params = p.parseCommaSperatedExpressions()

	// check for too many params
	if len(stmt.Params) > 4 {
		p.reportError(berrors.Syntax)
	}

	return &stmt
}

// not much to do, just return a StopStatement object
func (p *Parser) parseStopStatement() *ast.StopStatement {
	defer untrace(trace("parseStopStatement"))
	stmt := ast.StopStatement{Token: p.curToken}

	return &stmt
}

// start parsing an Identifier
func (p *Parser) parseIdentifier() ast.Expression {
	defer untrace(trace("parseIdentifier"))
	exp := p.innerParseIdentifier()

	return exp
}

// innerParseIdentifier is called from many other statements consume identifiers
func (p *Parser) innerParseIdentifier() *ast.Identifier {
	defer untrace(trace("innerParseIdentifier"))
	exp := &ast.Identifier{Token: p.curToken, Value: strings.ToUpper(p.curToken.Literal)}

	if strings.ContainsAny(p.peekToken.Literal, "$%!#") {
		p.nextToken()
		exp.Token.Literal = exp.Token.Literal + p.curToken.Literal
		exp.Value = exp.Value + p.curToken.Literal
		exp.Type = p.curToken.Literal
	}

	if p.checkForFuncCall() {
		return exp
	}

	if strings.ContainsAny(p.peekToken.Literal, "[(") {
		p.nextToken()
		if p.curTokenIs(")") {
			return exp
		}
		exp.Token.Literal = exp.Token.Literal + p.curToken.Literal
		exp.Value = exp.Value + "["
		exp.Array = true
		exp.Index = make([]*ast.IndexExpression, 0)

		for (p.curToken.Literal != "]") && (p.curToken.Literal != ")") {
			if (p.peekToken.Literal != "]") && (p.peekToken.Literal != ")") {
				exp.Index = append(exp.Index, p.innerParseIndexExpression(exp))
			}
			p.nextToken()
		}
		exp.Token.Literal = exp.Token.Literal + p.curToken.Literal
		exp.Value = exp.Value + "]"
	}

	return exp
}

// true if curToken has same type as t
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// true if peekToken is param type
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// true if peekToken matches and advances it to curtoken, error otherwise
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be '%s', got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) generalError(msg string) {
	p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	exp := &ast.GroupedExpression{Token: p.curToken}

	p.nextToken()

	exp.Exp = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	defer untrace(trace("parseIfExpression"))
	expression := &ast.IfExpression{Token: p.curToken}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)
	p.nextToken()

	if p.curTokenIs(token.COMMA) { // ignore optional comma
		p.nextToken()
	}

	// can be 'IF exp THEN'
	// or 'IF exp GOTO'
	// in fact, it can be 'IF exp LINE#'
	// and don't forget 'IF exp THEN END'
	if p.curTokenIs(token.THEN) {
		p.nextToken()
	}

	expression.Consequence = p.parseIfOption()

	// if there is no ELSE we are done
	if !p.peekTokenIs(token.ELSE) {
		return expression
	}

	p.nextToken()
	p.nextToken()

	expression.Alternative = p.parseIfOption()

	return expression
}

// parseIfOption handles the special cases
func (p *Parser) parseIfOption() ast.Statement {
	var exp ast.Statement

	switch p.curToken.Type {
	case token.INT:
		exp = &ast.GotoStatement{Token: token.Token{Type: token.LookupIdent(token.GOTO), Literal: ""}, Goto: p.curToken.Literal}

	case token.END:
		exp = &ast.EndStatement{Token: token.Token{Type: token.LookupIdent(token.END), Literal: "END"}}

	default:
		exp = p.parseStatement()
	}

	return exp
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// enforce the basic rule that the id starts with "FN"

	if len(p.curToken.Literal) < 3 {
		p.generalError("function names must be in the form FNname")
		return nil
	}

	if "FN" != strings.ToUpper(p.curToken.Literal[0:2]) {
		p.generalError("function names must be in the form FNname")
		return nil
	}

	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.EQ) {
		return nil
	}
	p.nextToken()

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// parse an unexpected EOF (or really, end of input)
// the error message should come from the eval layer
func (p *Parser) parseEOFExpression() ast.Expression {
	stmt := &ast.EOFExpression{}

	return stmt
}

func (p *Parser) checkForFuncCall() bool {
	// user defined functions must start with FN
	if (len(p.curToken.Literal) > 2) && (p.curToken.Literal[0:2] == "FN") {
		return true
	}

	// not user defined, check for builtin
	_, ok := builtins.Builtins[p.curToken.Literal]

	if ok {
		return true
	}

	return false
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	defer untrace(trace("parseCallExpression"))
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	p.nextToken()
	exp.Arguments = p.parseCallArguments()
	return exp
}

//
func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.curTokenIs(token.LPAREN) {
		p.nextToken() // skip past the left brace
	}

	if p.curTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	//if !p.expectPeek(token.RPAREN) {
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		//return nil
	}

	return args
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseIndexExpression"))
	return p.innerParseIndexExpression(left)
}

func (p *Parser) innerParseIndexExpression(left ast.Expression) *ast.IndexExpression {
	defer untrace(trace("innerParseIndexExpression"))
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}
	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)
	return exp

}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	defer untrace(trace("parseExpressionList"))
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.peekTokenIs(end) {
		return nil
	}

	return list
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()
	for !p.peekTokenIs(token.COLON) && !p.peekTokenIs((token.RBRACKET)) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

// Trace Off, no parameters so nothing much to do
func (p *Parser) parseTroffCommand() *ast.TroffCommand {
	stmt := &ast.TroffCommand{Token: p.curToken}
	p.nextToken()

	return stmt
}

// Trace On, no parameters so nothing much to do
func (p *Parser) parseTronCommand() *ast.TronCommand {
	stmt := &ast.TronCommand{Token: p.curToken}
	p.nextToken()

	return stmt
}

// Precedence of the peekToken
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// Precedence of current token
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// reportError builds an error message and adds to the array
func (p *Parser) reportError(err int) {
	msg := berrors.TextForError(err)

	if !p.cmdInput && (p.curLine != 0) {
		msg = fmt.Sprintf("%s in %d", msg, p.curLine)
	}

	p.errors = append(p.errors, msg)
}

// parse VIEW, also catches VIEW PRINT
func (p *Parser) parseViewStatement() ast.Statement {
	untrace(trace("parseViewStatement"))
	vw := ast.ViewStatement{Token: p.curToken}

	// if the first param is PRINT, this affects text window
	if p.peekTokenIs(token.PRINT) {
		return p.parseViewPrintStatement()
	}

	for !p.chkEndOfStatement() {
		p.nextToken()
		switch p.curToken.Type {
		case token.COMMA:
			vw.Parms = append(vw.Parms, &ast.Identifier{Value: ","})
		case token.LPAREN:
			vw.Parms = append(vw.Parms, &ast.Identifier{Value: "("})
		case token.MINUS:
			vw.Parms = append(vw.Parms, &ast.Identifier{Value: " - "})
		case token.RPAREN:
			vw.Parms = append(vw.Parms, &ast.Identifier{Value: ")"})
		case token.SCREEN:
			vw.Parms = append(vw.Parms, &ast.ScreenStatement{Token: token.Token{Type: token.SCREEN, Literal: "SCREEN"}})
		default:
			vw.Parms = append(vw.Parms, p.parseExpression(LOWEST))
		}
	}

	return &vw
}

// View Print changes the boundaries of the text window
func (p *Parser) parseViewPrintStatement() ast.Statement {
	untrace(trace("parseViewPrintStatement"))

	// create a ViewPrintStatement, build the literal
	vp := ast.ViewPrintStatement{Token: p.curToken}
	p.nextToken()
	vp.Token.Literal += " " + p.curToken.Literal // winds up "VIEW PRINT"

	// loop up parameters until the end of the statement
	for !p.chkEndOfStatement() {
		p.nextToken()
		switch p.curToken.Type {
		case token.TO:
			vp.Parms = append(vp.Parms, &ast.ToStatement{Token: token.Token{Type: token.TO, Literal: "TO"}})
		default:
			vp.Parms = append(vp.Parms, p.parseExpression(LOWEST))
		}
	}

	return &vp
}
