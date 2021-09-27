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

	p.ParseCmd(env)

	if len(p.Errors()) > 0 {
		for _, m := range p.Errors() {
			env.Terminal().Println(m)
		}
		return
	}

	iter := env.CmdLineIter()

	// if command line is empty, nothing to execute
	if iter.Len() == 0 {
		if env.GetAuto() != nil {
			prompt(env)
		}
		return
	}

	for iter.Value() != nil {
		cmd := iter.Value()
		srcIter := env.StatementIter()
		obj := evaluator.Eval(cmd, srcIter, env)

		if handleExitMsgs(obj, env) {
			return
		}

		// TODO: move this into eval code
		// see if cmd is trying to start execution
		switch cmd.(type) {
		case *ast.GotoStatement:
			stmt := &ast.Program{}
			strt, err := strconv.Atoi(cmd.(*ast.GotoStatement).Goto)

			if err != nil {
				giveError(err.Error(), env)
				return
			}
			errmsg := srcIter.Jump(strt)

			if err != nil {
				giveError(errmsg, env)
				return
			}
			evaluator.Eval(stmt, srcIter, env)
			break
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
		env.Terminal().Println(rc.(*object.HaltSignal).Msg)
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
