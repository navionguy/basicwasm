package mocks

import "github.com/navionguy/basicwasm/token"

type MockLexer struct {
	tokens []token.Token
}

// add a token to the array
func (ml *MockLexer) AddToken(token token.Token) {
	ml.tokens = append(ml.tokens, token)
}

// return the next token
func (ml *MockLexer) NextToken() token.Token {
	rc := ml.tokens[0]
	ml.tokens = ml.tokens[1:]
	return rc
}

func (ml *MockLexer) PassOn()  {}
func (ml *MockLexer) PassOff() {}
