package cli

import (
	"time"

	"github.com/navionguy/basicwasm/evaluator"
	"github.com/navionguy/basicwasm/keybuffer"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"
	"github.com/navionguy/basicwasm/terminal"
	"github.com/navionguy/basicwasm/token"
)

//Start begins interacting with the user
func Start(term *terminal.Terminal) {
	go runLoop(term)
}

func runLoop(term *terminal.Terminal) {
	//var cmd []byte
	env := object.NewTermEnvironment(term)

	term.Println("OK")
	for {
		k, ok := keybuffer.ReadByte()

		if !ok {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		switch k {
		case '\r':
			row, _ := term.GetCursor()
			//fmt.Printf("cursor at %d:%d\n", row, col)
			term.Print("\r\n")
			execCommand(term.Read(0, row, 80), term, env)
			//fmt.Println(term.Read(0, row, 80))
		default:
			term.Print(string(k))
			//cmd = append(cmd, k)
			//fmt.Printf("%s\n", hex.EncodeToString(cmd[len(cmd)-1:]))
		}
	}
}

func execCommand(input string, term *terminal.Terminal, env *object.Environment) {
	l := lexer.New(input)
	tk := l.NextToken()
	if tk.Type == token.LINENUM {
		// fresh line of code
		term.Print("LINENUM")
	} else {
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			term.Println("Syntax error")
			return
		}
		evaluator.Eval(program, program.StatementIter(), env)
	}

	term.Println("OK")
}
