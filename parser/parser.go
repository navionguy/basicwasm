package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/token"
)

const (
	_ int = iota
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
	curLine   int16

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
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatingPointLiteral)
	p.registerPrefix(token.FIXED, p.parseFixedPointLiteral)

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

// ParseProgram time to get busy
func (p *Parser) ParseProgram() *ast.Program {
	defer untrace(trace("ParseProgram"))

	program := &ast.Program{}
	//program.Statements = []ast.Statement{}
	program.New()

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.AddStatement(stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	defer untrace(trace("parseStatement"))
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.EOL:
		// EOF means that was the last line
		if p.peekTokenIs(token.EOF) {
			stmt := &ast.EndStatement{}
			return stmt
		}

		/* Newline signals a line number should follow
		if !p.expectPeek(token.LINENUM) {
			p.errors = append(p.errors, fmt.Sprintf("missing line number after %d", p.curLine))
			return nil
		}*/
		return nil
	case token.LINENUM:
		return p.parseLineNumber()
	case token.GOTO:
		return p.parseGotoStatement()
	case token.GOSUB:
		return p.parseGosubStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.PRINT:
		return p.parsePrintStatement()
	case token.END:
		return p.parseEndStatement()
	case token.DIM:
		return p.parseDimStatement()
	case token.CLS:
		return p.parseClsStatement()
	default:
		if strings.ContainsAny(p.peekToken.Literal, "=[$%!#") {
			return p.parseImpliedLetStatement(p.curToken.Literal)
		}
		return p.parseExpressionStatement()
	}
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

	if (value > 32767) || (value < -32768) {
		return p.parseDoubleIntegerLiteral(value)
	}

	lit.Value = int16(value)
	return lit
}

func (p *Parser) parseDoubleIntegerLiteral(value int) ast.Expression {
	defer untrace(trace("parseDoubleIntegerLiteral"))
	lit := &ast.DblIntegerLiteral{Token: p.curToken}
	lit.Value = int32(value)
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
	lit := &ast.FLoatDoubleLiteral{Token: p.curToken}
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
	stmt := &ast.LineNumStmt{Token: p.curToken}
	stmt.Token.Literal = p.curToken.Literal // rebrand it a line number

	var err error
	tv, err := strconv.Atoi(p.curToken.Literal)

	if err != nil {
		p.generalError("Invalid line number")
	}
	stmt.Value = int16(tv)
	p.curLine = int16(tv)

	return stmt
}

func (p *Parser) parseClsStatement() *ast.ClsStatement {
	defer untrace(trace("parseClsStatement"))
	stmt := &ast.ClsStatement{Token: p.curToken, Param: -1}

	if p.peekTokenIs(token.INT) {
		p.nextToken()
		stmt.Param, _ = strconv.Atoi(p.curToken.Literal)
	}

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
	return p.peekTokenIs(token.COLON) || p.peekTokenIs(token.LINENUM) || p.peekTokenIs(token.EOF)
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

// goto - uncondition transfer to line
func (p *Parser) parseGotoStatement() *ast.GotoStatement {
	stmt := &ast.GotoStatement{Token: p.curToken, Goto: ""}

	if !p.expectPeek(token.INT) {
		return nil
	}

	stmt.Goto = p.curToken.Literal

	if p.peekTokenIs(token.COLON) { // if a colon follows consume it
		p.nextToken()
	}

	return stmt
}

// gosub - uncondition transfer to subroutine
func (p *Parser) parseGosubStatement() *ast.GosubStatement {
	stmt := &ast.GosubStatement{Token: p.curToken, Gosub: ""}

	if !p.expectPeek(token.INT) {
		return nil
	}

	stmt.Gosub = p.curToken.Literal

	if p.peekTokenIs(token.COLON) { // if a colon follows consume it
		p.nextToken()
	}

	return stmt
}

// returns are much simpler in gwbasic
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
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

func (p *Parser) parseEndStatement() *ast.EndStatement {
	stmt := &ast.EndStatement{Token: p.curToken}

	return stmt
}

func (p *Parser) parseDimStatement() *ast.DimStatement {
	defer untrace(trace("parseDimStatement"))
	exp := &ast.DimStatement{Token: p.curToken, Vars: []*ast.Identifier{}}

	for !p.peekTokenIs(token.EOF) && !p.peekTokenIs(token.EOL) && !p.peekTokenIs(token.COLON) {
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

func (p *Parser) parseIdentifier() ast.Expression {
	defer untrace(trace("parseIdentifier"))
	return p.innerParseIdentifier()
}

func (p *Parser) innerParseIdentifier() *ast.Identifier {
	defer untrace(trace("innerParseIdentifier"))
	exp := &ast.Identifier{Token: p.curToken, Value: strings.ToUpper(p.curToken.Literal)}

	switch p.peekToken.Literal {
	case "$", "%", "!", "#":
		p.nextToken()
		exp.Token.Literal = exp.Token.Literal + p.curToken.Literal
		exp.Value = exp.Value + p.curToken.Literal
		exp.Type = p.curToken.Literal

	case "[":
		p.nextToken()
		exp.Token.Literal = exp.Token.Literal + "[]"
		exp.Value = exp.Value + "[]"
		exp.Array = true
		exp.Index = make([]*ast.IndexExpression, 0)

		for p.curToken.Literal != "]" {
			exp.Index = append(exp.Index, p.innerParseIndexExpression(exp))
			p.nextToken()
		}
	}

	// type might also be an array
	if p.peekTokenIs("[") && strings.ContainsAny(exp.Token.Literal, "$%!#") {
		exp.Token.Literal = exp.Token.Literal + "[]"
		exp.Value = exp.Value + "[]"
		exp.Array = true
		exp2 := p.innerParseIdentifier()
		exp.Index = exp2.Index
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
	p.nextToken()

	exp := p.parseExpression(LOWEST)

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

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	for !p.curTokenIs(token.COLON) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
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

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	defer untrace(trace("parseCallExpression"))
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
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

	if !p.expectPeek(end) {
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
