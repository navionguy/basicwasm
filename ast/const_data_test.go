package ast

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ConstData(t *testing.T) {
	/*tests := []struct {

	}*/

	cd := ConstData{}
	cd.statementNode()
	assert.EqualValues(t, "DATA", cd.TokenLiteral())
	assert.EqualValues(t, "DATA", cd.String())
}
