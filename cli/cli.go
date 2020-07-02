package cli

import (
	"fmt"
	"time"

	"github.com/navionguy/basicwasm/keybuffer"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/terminal"
)

//Start begins interacting with the user
func Start(term *terminal.Terminal) {
	go runLoop(term)
}

func runLoop(term *terminal.Terminal) {
	//var cmd []byte

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
			term.Println("")
			execCommand(term.Read(0, row, 80), term)
			//fmt.Println(term.Read(0, row, 80))
		default:
			term.Print(string(k))
			//cmd = append(cmd, k)
			//fmt.Printf("%s\n", hex.EncodeToString(cmd[len(cmd)-1:]))
		}
	}
	fmt.Println("cli stopping")
}

func execCommand(input string, term *terminal.Terminal) {
	l := lexer.New(input)
	for tk := l.NextToken(); tk.Literal != "EOF"; tk = l.NextToken() {
		term.Print(tk.Literal)
	}
	term.Println(":")
	term.Println("OK")
}
