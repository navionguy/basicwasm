package ast

// CmdLine holds one line entered from the terminal that will be
// immediately parsed and executed.
// Fun fact, running a program from the command is done with the
// "RUN" command, or "RUN line#".  But you can also "GOTO line#"
// or "GOSUB line#".
//
// Burger King is not the only place you can "Have it your way!"
type CmdLine struct {
	input string      // as entered from the terminal
	stmts []Statement // the parts of the command entry
	itr   uint16      // keeps track of where we are in executing
	err   error       // nil unless we encounter an error
}

// Implementation for CmdLine

// Create a new command line holding the string
func NewCmdLine(inp string) *CmdLine {
	cl := &CmdLine{input: inp, itr: 0, err: nil}

	return cl
}

// Add a parsed statement to the line
func (cl *CmdLine) AddStatement(s Statement) {
	cl.stmts = append(cl.stmts, s)
}

// Called when a command line has been fully executed, or failed with an error
func (cl *CmdLine) CmdComplete() {
	cl.stmts = nil
	cl.input = ""
	cl.err = nil
}

// Return the statement count for the command input
func (cl *CmdLine) LineLength() uint16 {
	return uint16(len(cl.stmts))
}

// Return the next statement and bump the index
func (cl *CmdLine) NextStatement() Statement {
	if cl.itr >= uint16(len(cl.stmts)) {
		return nil
	}

	// send back the next statement
	stmt := cl.stmts[cl.itr]
	cl.itr++

	return stmt
}

func (cl *CmdLine) Remaining() uint16 {
	return cl.LineLength() - cl.itr
}

func (cl *CmdLine) String() string {
	return cl.input
}

func (cl *CmdLine) TokenLiteral() string {
	return "CMD"
}
