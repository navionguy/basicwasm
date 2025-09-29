package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/evaluator"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/settings"
	"github.com/navionguy/basicwasm/token"
)

var stopTerminal bool

// Start begins interacting with the user
func Start(env *object.Environment) {
	stopTerminal = false
	go runLoop(env, &stopTerminal)
}

// Stop shuts down the command line interface
func Stop() {
	stopTerminal = true
}

// runLoop reads key presses and send them off to be processed
// basically, just loop until the main routine exits
func runLoop(env *object.Environment, done *bool) {

	// send the boot-up "OK" to the console
	env.Terminal().Println("OK")
	for {
		keys := env.Terminal().ReadKeys(1)

		evalKeyCodes(keys, env)

		if *done {
			return
		}
	}
}

// given one or more key codes, turn them into action
func evalKeyCodes(keys []byte, env *object.Environment) {
	k := keys[0]
	switch k {
	case '\r':
		row, col := env.Terminal().GetCursor()
		//		fmt.Printf("cursor at %d:%d\n", row, col)
		env.Terminal().Print("\r\n")
		nr, nc := env.Terminal().GetCursor()
		for (nr == row) && (nc == col) {
			time.Sleep(time.Millisecond)
			nr, nc = env.Terminal().GetCursor()
			//env.Terminal().Log(fmt.Sprintf("cursor at %d:%d\n", nr, nc))
		}
		execCommand(env.Terminal().Read(0, row, 80), env)

		//fmt.Println(term.Read(0, row, 80))
	case 0x7F: // down arrow
		row, col := env.Terminal().GetCursor()
		env.Terminal().Locate(row+1, col)
		env.Terminal().Print("\x1b[P")
	//env.Terminal().Print("\b\a")
	case 0x03: // ctrl-c
		env.Terminal().Println("")
		if env.GetSetting(settings.Auto) != nil {
			env.SaveSetting(settings.Auto, nil)
			env.Terminal().BreakCheck()
			prompt(env)
		}
	default:
		/*var msg []byte
		msg = append(msg, k)
		hex := hex.EncodeToString(msg)
		env.Terminal().Print(hex)*/
		env.Terminal().Print(string(k))
		//cmd = append(cmd, k)
		//fmt.Printf("%s\n", hex.EncodeToString(cmd[len(cmd)-1:]))
	}

}

// we have input terminated with a return key
// should be either a command or a line of source code
func execCommand(input string, env *object.Environment) {

	// go parse the input
	chkForCmd(input, env)

	iter := env.CmdLineIter()

	// if command line is empty, nothing to execute
	if iter.Len() == 0 {
		// if auto is turned on, prompt next line number
		if env.GetSetting(settings.Auto) != nil {
			prompt(env)
		}
		return
	}

	parseCmdExecute(iter, env)
}

// see if I can successfully parse the command line entered
func chkForCmd(input string, env *object.Environment) object.Object {
	l := lexer.New(input)

	if l.NextToken().Type == token.EOL {
		// skip over the starting EOL
		t := l.NextToken()
		if (t.Type == token.INT) || (t.Type == token.INTD) {
			// it starts with a line number, go add to the source code tree
			return addOrReplaceSourceLine(t.Literal, input, env)
		} else {
			// appears to be a command, go parse it
		}
	}
	return nil
}

// I have a source line, add it to the tree
func addOrReplaceSourceLine(line string, input string, env *object.Environment) object.Object {
	val, err := strconv.ParseUint(line, 10, 16)

	if err != nil {
		// can't extract the line number, return an error
		rc := &object.Error{
			Message: berrors.TextForError(berrors.UnDefinedLineNumber),
			Code:    berrors.UnDefinedLineNumber,
		}
		return rc
	}

	l := uint16(val)
	sl := object.NewSourceLine(input, l)
	env.Source.AddSourceLine(sl)
	return nil
}

// once you have a parsed command line, go execute it
func parseCmdExecute(iter *ast.Code, env *object.Environment) {

	for iter.Value() != nil {
		cmd := iter.Value()
		srcIter := env.StatementIter()
		obj := evaluator.Eval(cmd, srcIter, env)

		if handleExitMsgs(obj, env) {
			return
		}
		iter.Next()
	}
	env.CmdComplete()
	prompt(env)
}

// some special objects that can come back from command execution
func handleExitMsgs(rc object.Object, env *object.Environment) bool {
	switch msg := rc.(type) {
	case *object.Error:
		env.Terminal().Println(msg.Message)
	case *object.HaltSignal:
		if len(rc.(*object.HaltSignal).Msg) > 0 {
			env.Terminal().Println(msg.Msg)
		}
	default:
		return false
	}

	env.CmdComplete()
	prompt(env)
	return true
}

func prompt(env *object.Environment) {
	// get the auto settings if they exist
	a := env.GetSetting(settings.Auto)

	if a == nil {
		env.Terminal().Println("OK")
		return
	}

	// unpack the auto settings
	auto := a.(*ast.AutoCommand)
	line := auto.Params[0].(*ast.DblIntegerLiteral).Value
	inc := auto.Params[1].(*ast.DblIntegerLiteral).Value
	fill := " "

	if env.StatementIter().Exists(int(line)) {
		fill = "*"
	}

	env.Terminal().Print(fmt.Sprintf("%d%s", line, fill))
	line += inc
	auto.Params[0].(*ast.DblIntegerLiteral).Value = line
	env.SaveSetting(settings.Auto, auto)
}

// just display the error and then the prompt
func giveError(err string, env *object.Environment) {
	env.Terminal().Println(err)
	env.Terminal().Println("OK")
	env.CmdComplete()
}
