package cli

import (
	"fmt"
	"strconv"

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

func runLoop(env *object.Environment) {
	//var cmd []byte

	env.Terminal().Println("OK")
	for {
		key := env.Terminal().ReadKeys(1)
		k := key[0]
		switch k {
		case '\r':
			row, _ := env.Terminal().GetCursor()
			//fmt.Printf("cursor at %d:%d\n", row, col)
			env.Terminal().Print("\r\n")
			execCommand(env.Terminal().Read(0, row, 80), env)
			//fmt.Println(term.Read(0, row, 80))
		case 0x7F:
			row, col := env.Terminal().GetCursor()

			/*if col > 0 {
				col--
			}*/
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
}

func execCommand(input string, env *object.Environment) {
	l := lexer.New(input)
	p := parser.New(l)

	/*if checkForLineNum(input) {
		// fresh line of code
		bExc = false
		p.ParseProgram(env)
	} else {*/
	p.ParseCmd(env)
	//}

	if len(p.Errors()) > 0 {
		for _, m := range p.Errors() {
			env.Terminal().Println(m)
		}
		return
	}

	iter := env.Program.CmdLineIter()

	// if command line is empty, nothing to execute
	if iter.Len() == 0 {
		if env.GetAuto() != nil {
			prompt(env)
		}
		return
	}

	for iter.Value() != nil {
		cmd := iter.Value()
		srcIter := env.Program.StatementIter()
		obj := evaluator.Eval(cmd, srcIter, env)

		err, ok := obj.(*object.Error)

		if ok {
			env.Terminal().Println(err.Message)
			env.Program.CmdComplete()
			prompt(env)
			return
		}

		// see if cmd is trying to start execution
		switch cmd.(type) {
		case *ast.GotoStatement:
			stmt := &ast.Program{}
			strt, err := strconv.Atoi(cmd.(*ast.GotoStatement).Goto)

			if err != nil {
				giveError(err, env)
				return
			}
			err = srcIter.Jump(strt)

			if err != nil {
				giveError(err, env)
				return
			}
			evaluator.Eval(stmt, srcIter, env)
			break
		}
		iter.Next()
	}
	env.Program.CmdComplete()
	prompt(env)

}

func prompt(env *object.Environment) {
	auto := env.GetAuto()

	if auto == nil {
		env.Terminal().Println("OK")
		return
	}

	fill := " "
	if env.Program.StatementIter().Exists(auto.Start) {
		fill = "*"
	}

	env.Terminal().Print(fmt.Sprintf("%d%s", auto.Start, fill))
	auto.Start += auto.Increment
	env.SetAuto(auto)
}

// just display the error and then the prompt
func giveError(err error, env *object.Environment) {
	env.Terminal().Println(err.Error())
	env.Terminal().Println("OK")
	env.Program.CmdComplete()
	return
}
