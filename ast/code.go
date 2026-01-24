package ast

// Code allows iterating over the code lines subject to control transfer
type Code struct {
	srcCode  *SourceTree // a BTree holding all of the source code
	currLine *SourceLine // current line executing
	err      error
}

// Make my code a node
func (c *Code) TokenLiteral() string { return "" }
func (c *Code) String() string       { return "SourceCode" }

// The implementation for Code type
// Initialize a new Code structure to support the environment space
func InitCode() *Code {
	c := &Code{
		srcCode: initSourceTree(), // creates an empty BTree to hold all the source code
	}

	return c
}

// AddSrcLine adds a line to the source tree
func (c *Code) AddSrcLine(src string, lnum uint16) {
	sl := NewSourceLine(src, lnum)

	c.srcCode.addSourceLine(sl)
}

// GetReturnPoint is used for GOSUB, ON GOSUB and FOR loops
func (c *Code) GetReturnPoint() RetPoint {
	return RetPoint{Line: c.currLine.lineNum, Stmt: uint8(c.currLine.itr + 1)}
}

// GetSrcLineCount returns the number lines in the tree
func (c *Code) GetSrcLineCount() uint16 {
	lc := c.srcCode.tree.Len()

	return uint16(lc)
}

// Goto does a moves to the beginning of the requested line.
// Returns the line number requested, zero if it was not found.
func (c *Code) Goto(l uint16) uint16 {
	t := c.srcCode.JumpToLine(l)
	if t != 0 {
		c.currLine = c.srcCode.NextLine()
	} else {
		c.currLine = nil
	}

	return t
}

// Check if a source line is already in the tree
// return true if it does
// Only know use is when auto line number is active for the console.
// If the line already exists, we append an '*' to the line number
// to warn the user they are about to overwrite existing code.
func (c *Code) LineExists(l uint16) bool {
	if c.Goto(l) == 0 {
		return false
	}
	return true
}

// LineParsed returns true if it has been parsed
// false if it has note.
func (c *Code) LineParsed() bool {
	return len(c.currLine.statements) > 0
}

// Fetch the next statement to be evaluated.
// If currLine is all evaluated, this will try
// to advance to the next line in the tree.
// If no more lines, returns nil
func (c *Code) NextStmt() Statement {
	if c.currLine == nil {
		c.currLine = c.srcCode.FirstLine()
		stmt := c.currLine.NextStatement()
		return stmt
	}

	// extract the next statement from the current line
	stmt := c.currLine.NextStatement()

	if stmt == nil {
		// no more statements on this line, move to the next
		return c.nextSrcLine()
	}

	return stmt
}

// Have the SourceTree move forward to the next line.
// He will return nil if there isn't one.
func (c *Code) nextSrcLine() Statement {
	sl := c.srcCode.NextLine()

	if sl == nil {
		return nil
	}

	st := c.currLine.NextStatement()

	return st
}
