package object

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AddSrcLine(t *testing.T) {
	env := newEnvironment()
	env.ClearProgramMemory()
	env.AddSourceLine("10 REM A comment", 10)
	env.AddSourceLine(`20 PRINT "Hello World!`, 20)
	env.AddSourceLine("30 END", 30)
	assert.EqualValues(t, 3, env.GetSrcLineCount())
}

// ClearProgramMemory should delete all source lines from the tree
func Test_ClearProgramMemory(t *testing.T) {
	env := newEnvironment()
	env.source.AddSrcLine("10 REM A comment", 10)
	assert.EqualValues(t, 1, env.GetSrcLineCount())

	env.ClearProgramMemory()
	assert.EqualValues(t, 0, env.GetSrcLineCount())
}

// GetSrcLineCount returns the total number of lines in the source tree
func Test_GetSrcLineCount(t *testing.T) {
	env := newEnvironment()
	env.ClearProgramMemory()
	env.AddSourceLine("10 REM A comment", 10)
	env.AddSourceLine(`20 PRINT "Hello World!`, 20)
	env.AddSourceLine("30 END", 30)
	assert.EqualValues(t, 3, env.GetSrcLineCount())
}

// NextStatement should return the next statement to execute
func Test_NextStatement(t *testing.T) {
	env := newEnvironment()
	env.ClearProgramMemory()
	env.AddSourceLine("10 REM A comment", 10)
	stmt := env.NextStatement()

	assert.EqualValues(t, "10", stmt.String())

}
