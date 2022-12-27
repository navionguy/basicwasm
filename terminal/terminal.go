package terminal

import (
	"fmt"
	"strings"
	"syscall/js"
	"time"

	"github.com/navionguy/basicwasm/keybuffer"
)

// Terminal holds the terminal instance and provides io abilities
type Terminal struct {
	term  js.Value
	buff  js.Value
	kbuff *keybuffer.KeyBuffer
}

// New creates a new Terminal object
func New(t js.Value) *Terminal {
	env := &Terminal{term: t, kbuff: keybuffer.GetKeyBuffer()}

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
	js.Global().Get("document").Call("getElementById", "chatAudio").Call("play")
}

// Log basicwasm information via call to javascript function
func (t *Terminal) Log(msg string) {
	js.Global().Call("consoleMsg", msg)
}

// Cls clears the terminal of all text
func (t *Terminal) Cls() {
	t.term.Call("clear")
	t.Locate(1, 1)       // set the cursor
	t.Print("\x1B[80'~") // clear to end of line
}

// GetCursor retrieves the current cursor position
// NOTE: Cursor position is based on the upper left
// position being 0,0
func (t *Terminal) GetCursor() (int, int) { // row,col
	col := t.term.Get("buffer").Get("active").Get("cursorX").Int()
	row := t.term.Get("buffer").Get("active").Get("cursorY").Int()

	return row, col
}

// Read pulls the contents of the selected area
func (t *Terminal) Read(col, row, len int) string {
	t.term.Call("select", col, row, len)
	inp := strings.TrimRight(t.term.Call("getSelection").String(), " ")
	t.term.Call("clearSelection")
	return inp
}

// ReadKeys reads the requested number of keystrokes
func (t *Terminal) ReadKeys(count int) []byte {
	var keys []byte

	for i := 0; i < count; {
		bt, err := t.kbuff.ReadByte()

		if err != nil {
			time.Sleep(100 * time.Millisecond)
		} else {
			keys = append(keys, bt)
			i++
		}
	}

	/*if len(keys) > 0 {
		t.Log("Keypress")
		for _, bt := range keys {
			msg := fmt.Sprintf("%x", bt)
			t.Log(msg)
		}
	}*/

	return keys
}

func (t *Terminal) BreakCheck() bool {
	bc := t.kbuff.BreakSeen()
	t.kbuff.ClearBreak()

	return bc
}
