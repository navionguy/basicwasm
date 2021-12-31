package cli

import (
	"fmt"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/evaluator"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"
)

//Start begins interacting with the user
func Start(env *object.Environment) {
	go runLoop(env)
}

// runLoop reads key presses and send them off to be processed
// basically, just loop until the main routine exits
func runLoop(env *object.Environment) {

	// send the boot-up "OK" to the console
	env.Terminal().Println("OK")
	for {
		keys := env.Terminal().ReadKeys(1)

		evalKeyCodes(keys, env)
	}
}

// given one or more key codes, turn them into action
func evalKeyCodes(keys []byte, env *object.Environment) {
	k := keys[0]
	switch k {
	case '\r':
		row, _ := env.Terminal().GetCursor()
		//fmt.Printf("cursor at %d:%d\n", row, col)
		env.Terminal().Print("\r\n")
		execCommand(env.Terminal().Read(0, row, 80), env)
		//fmt.Println(term.Read(0, row, 80))
	case 0x7F: // down arrow
		row, col := env.Terminal().GetCursor()
		env.Terminal().Locate(row+1, col)
		env.Terminal().Print("\x1b[P")
	//env.Terminal().Print("\b\a")
	case 0x03: // ctrl-c
		env.Terminal().Println("")
		if env.GetAuto() != nil {
			env.SetAuto(nil)
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
	p := parseCmdLine(input, env)

	// check for error during parse
	if p == nil {
		return
	}

	iter := env.CmdLineIter()

	// if command line is empty, nothing to execute
	if iter.Len() == 0 {
		// if auto is turned on, prompt next line number
		if env.GetAuto() != nil {
			prompt(env)
		}
		return
	}

	parseCmdExecute(iter, env)
}

// see if I can successfully parse the command line entered
func parseCmdLine(input string, env *object.Environment) *parser.Parser {
	l := lexer.New(input)
	p := parser.New(l)

	p.ParseCmd(env)

	// if all went well return the parser
	if len(p.Errors()) == 0 {
		return p
	}

	// display error messages
	for _, m := range p.Errors() {
		env.Terminal().Println(m)
	}

	// nothing to evaluate
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
	switch rc.(type) {
	case *object.Error:
		env.Terminal().Println(rc.(*object.Error).Message)
	case *object.HaltSignal:
		if len(rc.(*object.HaltSignal).Msg) > 0 {
			env.Terminal().Println(rc.(*object.HaltSignal).Msg)
		}
	default:
		return false
	}

	env.CmdComplete()
	prompt(env)
	return true
}

func prompt(env *object.Environment) {
	auto := env.GetAuto()

	if auto == nil {
		env.Terminal().Println("OK")
		return
	}

	fill := " "
	if env.StatementIter().Exists(auto.Start) {
		fill = "*"
	}

	env.Terminal().Print(fmt.Sprintf("%d%s", auto.Start, fill))
	auto.Start += auto.Increment
	env.SetAuto(auto)
}

// just display the error and then the prompt
func giveError(err string, env *object.Environment) {
	env.Terminal().Println(err)
	env.Terminal().Println("OK")
	env.CmdComplete()
	return
}
