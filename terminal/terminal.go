package terminal

import (
	"fmt"
	"strings"
	"syscall/js"
)

// Terminal holds the terminal instance and provides io abilities
type Terminal struct {
	term js.Value
	buff js.Value
}

// New creates a new Terminal object
func New(t js.Value) *Terminal {
	env := &Terminal{term: t}

	t.Call("setOption", "scrollback", 0)
	return env
}

// Println prints the string follow by CRLF
func (t *Terminal) Println(msg string) {
	t.Print(msg + "\r\n")
}

// Locate moves the cursor to the passed row/col
// NOTE: When issuing the Locate sequence, the upper left
// screen position is 1,1
func (t *Terminal) Locate(row, col int) {
	cmd := fmt.Sprintf("\x1B[%dd\x1b[%d`", row, col)
	t.Print(cmd)
}

// Print sends the passed string to the terminal at the current cursor position
func (t *Terminal) Print(msg string) {
	t.term.Call("write", msg)
}

// SoundBell plays the current bell sound
func (t *Terminal) SoundBell() {
	t.term.Get("_core").Get("_soundService").Call("playBellSound")
}

// Cls clears the terminal of all text
func (t *Terminal) Cls() {
	t.term.Call("clear")
}

// GetCursor retrieves the current cursor position
// NOTE: Cursor position is based on the upper left
// position being 0,0
func (t *Terminal) GetCursor() (int, int) { // row,col
	col := t.term.Get("buffer").Get("_buffers").Get("_activeBuffer").Get("x").Int()
	row := t.term.Get("buffer").Get("_buffers").Get("_activeBuffer").Get("y").Int()

	return row, col
}

// Read pulls the contents of the selected area
func (t *Terminal) Read(col, row, len int) string {
	t.term.Call("select", col, row, len)
	inp := strings.TrimRight(t.term.Call("getSelection").String(), " ")
	t.term.Call("clearSelection")
	return inp
}
