package parser

import (
	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/token"
)

// parse a comma seperated series of expressions
func (p *Parser) parseCommaSeparatedExpressions() []ast.Expression {
	var exp []ast.Expression
	done := false
	for ; !done; p.nextToken() {
		exp = append(exp, p.parseNextExpression(exp))

		// if there is a trailing comma, there is likely more params
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}

		done = p.chkEndOfStatement()

		// series can't end with a comma
		if done && p.curTokenIs(token.COMMA) {
			p.reportError(berrors.Syntax)
		}
	}

	return exp
}

// parse the next expression or add a nil parameter
func (p *Parser) parseNextExpression(exp []ast.Expression) ast.Expression {
	// if it is a comma, user is skipping a parameter
	if p.curTokenIs(token.COMMA) {
		return nil
	}

	// parse the expression to calculate the parameter
	return p.parseExpression(LOWEST)
}

// parser can't make sense of the input
// just soak up all the tokens until the next statement
func (p *Parser) parseTrash(Trash *[]ast.TrashStatement) {

	for {
		if !p.atEndOfStatement() {
			*Trash = append(*Trash, ast.TrashStatement{Token: token.Token{Type: p.curToken.Type, Literal: p.curToken.Literal}})
		}

		if p.chkEndOfStatement() {
			return
		}
		p.nextToken()
	}
}
