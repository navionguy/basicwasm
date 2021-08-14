package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/navionguy/basicwasm/ast"
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
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token
	curLine   int
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
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.INTD, p.parseIntDoubleLiteral)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.DEF, p.parseFunctionLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatingPointLiteral)
	p.registerPrefix(token.FIXED, p.parseFixedPointLiteral)
	p.registerPrefix(token.AMPERSAND, p.parseHexOctalConstant)

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

	if env.Program == nil {
		env.Program = &ast.Program{}
		env.Program.New() // make sure to initialize the new program
	}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			env.Program.AddStatement(stmt)
		}
		p.nextToken()
	}

	env.Program.Parsed()
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

	if env.Program == nil {
		env.Program = &ast.Program{}
		env.Program.New() // make sure to initialize the new program
	}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			env.Program.AddCmdStmt(stmt)
		}
		p.nextToken()
	}

	env.Program.CmdParsed()
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
	case token.CLEAR:
		return p.parseClearCommand()
	case token.CLS:
		return p.parseClsStatement()
	case token.COLOR:
		return p.parseColorStatement()
	case token.COMMON:
		return p.parseCommonStatement()
	case token.DATA:
		return p.parseDataStatement()
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
	case token.LET:
		return p.parseLetStatement()
	case token.LINENUM:
		return p.parseLineNumber()
	case token.LIST:
		return p.parseListStatement()
	case token.LOCATE:
		return p.parseLocateStatement()
	case token.GOTO:
		return p.parseGotoStatement()
	case token.GOSUB:
		return p.parseGosubStatement()
	case token.IF:
		return p.parseExpressionStatement()
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
	case token.TROFF:
		return p.parseTroffCommand()
	case token.TRON:
		return p.parseTronCommand()
	case token.PRINT:
		return p.parsePrintStatement()
	case token.DIM:
		return p.parseDimStatement()
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

	if !p.peekTokenIs(token.STRING) {
		p.errors = append(p.errors, "Syntax Error")
		return nil
	}

	p.nextToken()
	chain.File = p.curToken.Literal

	return &chain
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

func (p *Parser) parseColorStatement() *ast.ColorStatement {
	defer untrace(trace("parseColorStatement"))
	stmt := ast.ColorStatement{Token: p.curToken}

	if p.chkEndOfStatement() {
		p.errors = append(p.errors, "Syntax error")
		return &stmt
	}

	for i := 0; (i < 3) && !p.chkEndOfStatement(); i++ {
		p.nextToken()
		if !p.curTokenIs(token.COMMA) {
			stmt.Parms[i] = p.parseExpression(LOWEST)
			p.nextToken()
		}
	}

	return &stmt
}

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

func (p *Parser) parseFilesCommand() *ast.FilesCommand {
	defer untrace(trace("parseFilesCommand"))
	cd := &ast.FilesCommand{Token: p.curToken}

	if p.peekTokenIs(token.STRING) {
		p.nextToken()
		cd.Path = p.curToken.Literal
	}

	return cd
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))
	lit := &ast.IntegerLiteral{Token: p.curToken}
	value, err := strconv.Atoi(p.curToken.Literal)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer line %d", p.curToken.Literal, p.curLine)
		p.errors = append(p.errors, msg)
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
			msg := fmt.Sprintf("could not parse %q as octal prefix line %d", p.curToken.Literal, p.curLine)
			p.errors = append(p.errors, msg)
			return nil
		}
		lit.Token = token.Token{Type: token.OCTAL, Literal: "&O"}
		p.nextToken()
	}

	if !p.curTokenIs(token.INT) {
		msg := fmt.Sprintf("could not parse %q as octal value line %d", p.curToken.Literal, p.curLine)
		p.errors = append(p.errors, msg)
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

func (p *Parser) chkEndOfStatement() bool {
	return p.peekTokenIs(token.COLON) || p.peekTokenIs(token.LINENUM) || p.peekTokenIs(token.EOF) || p.peekTokenIs(token.EOL)
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer untrace(trace("parseLetStatement"))

	p.curToken.Literal = strings.ToUpper(p.curToken.Literal)
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
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

func (p *Parser) parseLocateStatement() *ast.LocateStatement {
	defer untrace(trace("parseLocateStatement"))
	stmt := ast.LocateStatement{Token: p.curToken}

	for i := 0; (i < len(stmt.Parms)) && (!p.chkEndOfStatement()); i++ {
		p.nextToken()
		if !p.curTokenIs(token.COMMA) {
			stmt.Parms[i] = p.parseExpression(LOWEST)
			p.nextToken()
		}
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

// gosub - uncondition transfer to subroutine
func (p *Parser) parseGosubStatement() *ast.GosubStatement {
	defer untrace(trace("parseGosubStatement"))
	stmt := ast.GosubStatement{Token: p.curToken, Gosub: ""}

	if !p.expectPeek(token.INT) {
		return nil
	}

	stmt.Gosub = p.curToken.Literal

	if p.peekTokenIs(token.COLON) { // if a colon follows consume it
		p.nextToken()
	}

	return &stmt
}

// read constant data from DATA statements
func (p *Parser) parseReadStatement() *ast.ReadStatement {
	defer untrace(trace("parseReadStatement"))
	stmt := &ast.ReadStatement{Token: p.curToken}

	for !p.peekTokenIs(token.LINENUM) && !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.EOL) && !p.peekTokenIs(token.COLON) {
		p.nextToken()

		if p.curTokenIs(token.IDENT) {
			id := p.parseIdentifier()

			if id != nil {
				stmt.Vars = append(stmt.Vars, id)
			}

			if p.peekTokenIs(token.COMMA) {
				p.nextToken()
			}
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
	cmd := &ast.RunCommand{Token: p.curToken}

	// check for line number to start at
	if p.peekTokenIs(token.INT) {
		p.nextToken()
		cmd.StartLine, _ = strconv.Atoi(p.curToken.Literal)
	} else if p.peekTokenIs(token.STRING) { // check for file to load
		p.nextToken()
		cmd.LoadFile = p.curToken.Literal
	}

	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if p.peekTokenIs(token.IDENT) {

		}
	}

	return cmd
}

func (p *Parser) parseEndStatement() *ast.EndStatement {
	defer untrace(trace("parseEndStatement"))
	stmt := &ast.EndStatement{Token: p.curToken}

	return stmt
}

func (p *Parser) parseIdentifier() ast.Expression {
	defer untrace(trace("parseIdentifier"))
	exp := p.innerParseIdentifier()

	return exp
}

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

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
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

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

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

	if !p.expectPeek(token.RPAREN) {
		return nil
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

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}
