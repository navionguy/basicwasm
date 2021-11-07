package main

import (
	"fmt"
	"syscall/js"

	"github.com/navionguy/basicwasm/cli"
	"github.com/navionguy/basicwasm/keybuffer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/terminal"
)

func registerCallbacks() {

	fmt.Println("gwbasic::registerCallBacks")

	//kybd := make(chan byte, 0)
	document := js.Global().Get("document")
	momma := document.Call("getElementById", "momma").Get("innerHTML")
	term := terminal.New(js.Global().Get("term"))

	env := object.NewTermEnvironment(term)
	env.Set(object.SERVER_URL, &object.String{Value: momma.String()})
	env.Set(object.WORK_DRIVE, &object.String{Value: `C:\`})

	cli.Start(env)
	env.Terminal().Log("cli started")

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
