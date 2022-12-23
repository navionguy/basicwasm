package main

import (
	"syscall/js"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/cli"
	"github.com/navionguy/basicwasm/keybuffer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/settings"
	"github.com/navionguy/basicwasm/terminal"
)

func registerCallbacks() {
	// reach into the document and get my servers address
	document := js.Global().Get("document")

	// Shout out to Mr. Schiedermayer's function, WhoIsMyMomma()
	momma := document.Call("getElementById", "momma").Get("innerHTML")
	term := terminal.New(js.Global().Get("term"))

	env := object.NewTermEnvironment(term)
	env.SaveSetting(settings.ServerURL, &ast.StringLiteral{Value: momma.String()})

	cli.Start(env)
	env.Terminal().Log("cli started")

	js.Global().Set("keyPress", js.FuncOf(func(this js.Value, inputs []js.Value) interface{} {
		kbuff := keybuffer.GetKeyBuffer()
		kbuff.SaveKeyStroke([]byte(inputs[0].String()))
		return nil
	}))
}

func main() {
	c := make(chan struct{}, 0)
	registerCallbacks()
	<-c
}
