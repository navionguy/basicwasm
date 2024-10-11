package parser

import (
	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/token"
)

// parse a comma separated series of expressions
func (p *Parser) parseCommaSeparatedExpressions() ([]ast.Expression, []ast.TrashStatement) {
	var exp []ast.Expression
	var trash []ast.TrashStatement
	done := false
	for ; !done; p.nextToken() {
		next := p.parseNextExpression()
		exp = append(exp, next)

		// if there is a trailing comma, there is likely more params
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}

		done = p.chkEndOfStatement()

		// series can't end with a comma
		if done && p.curTokenIs(token.COMMA) {
			p.parseTrash(&trash)
		}
	}

	return exp, trash
}

// parse the next expression or add a nil parameter
func (p *Parser) parseNextExpression() ast.Expression {
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
