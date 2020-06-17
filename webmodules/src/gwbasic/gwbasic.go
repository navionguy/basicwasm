package main

import (
	"fmt"
	"syscall/js"

	"github.com/navionguy/basicwasm/cli"
	"github.com/navionguy/basicwasm/keybuffer"
	"github.com/navionguy/basicwasm/terminal"
)

func registerCallbacks() {

	fmt.Println("gwbasic::registerCallbacks")

	//kybd := make(chan byte, 0)
	term := terminal.New(js.Global().Get("term"))

	term.Println("OK")
	cli.Start(term)

	js.Global().Set("keyPress", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		keybuffer.SaveKeyStroke([]byte(inputs[0].String()))
		return nil
	}))
}

func main() {
	c := make(chan struct{}, 0)
	registerCallbacks()
	<-c
}
