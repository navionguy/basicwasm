package main

import (
	"fmt"
	"syscall/js"
)

type environment struct {
	term js.Value
}

var env environment

func registerCallbacks() {

	fmt.Println("gwbasic::registerCallbacks")
	env.term = js.Global().Get("term")

	env.locate(1, 1)
	env.println("10 REM A simple program")
	env.println("20 PRINT \"Hello World!\"")
	env.println("OK\r\nRUN\r\nHello World!\r\nOK")
	//env.println("1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890")
	//env.locate(23, 79)
	//for i := 0; i < 24; i++ {
	//	env.println(fmt.Sprintf("%d", i))
	//}
	fmt.Println("wrote")
}

func (e *environment) println(msg string) {
	e.print(msg + "\r\n")
}

func (e *environment) locate(row, col int) {
	cmd := fmt.Sprintf("\x1B[%dd\x1b[%d`", row, col)
	e.print(cmd)
}

func (e *environment) print(msg string) {
	e.term.Call("write", msg)
}

func main() {
	c := make(chan struct{}, 0)
	registerCallbacks()
	<-c
}
