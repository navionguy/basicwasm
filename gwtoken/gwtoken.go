package gwtoken

import (
	"bufio"
	"fmt"
	"io"

	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"
)

const (
	TOKEN_FILE     = 0xff
	PROTECTED_FILE = 0xfe
)

type progRdr struct {
	src     *bufio.Reader
	eof     bool
	offset  int    // offset of the current line
	linenum int    // current line number being processed
	lineInp string // the text to parse
}

func ParseFile(src *bufio.Reader, env *object.Environment) {
	bt := make([]byte, 1)
	lr := io.LimitReader(src, 1)

	n, err := lr.Read(bt)

	if (err != nil) || (n != 1) {
		env.Terminal().Println("File read error")
		return
	}

	if bt[0] != TOKEN_FILE {
		env.Terminal().Println("Bad file mode")
		return
	}

	rdr := progRdr{src: src}

	rdr.readProg(env)
}

func (rdr *progRdr) readProg(env *object.Environment) {
	for !rdr.eof {
		rdr.readLineHeader()
		if rdr.linenum == 0 {
			return
		}
		rdr.readLine()

		l := lexer.New(rdr.lineInp)
		p := parser.New(l)
		p.ParseProgram(env)
	}
}

func (rdr *progRdr) readLine() {
	rdr.lineInp = fmt.Sprintf("%d ", rdr.linenum)

	for val := "tmp"; val != ""; {
		val = rdr.readToken()
		rdr.lineInp = rdr.lineInp + val
	}
}

//
func (rdr *progRdr) readToken() string {
	var val string
	tok, err := rdr.src.ReadByte()

	if err != nil {
		rdr.eof = true
		return ""
	}

	switch tok {
	case colon_TOK:
		pk, err := rdr.src.Peek(1)
		if err != nil {
			rdr.eof = true
			val = ":"
			return val
		}
		if pk[0] == else_TOK {
			val = rdr.readToken()
		}
	case data_TOK:
		val = "DATA"
	case dblQuote_TOK:
		val = rdr.copyString()
		val = `"` + val
	case dim_TOK:
		val = "DIM"
	case else_TOK:
		val = "ELSE"
	case end_TOK:
		val = "END"
	case eol_TOK:
		val = ""
	case for_TOK:
		val = "FOR"
	case goto_TOK:
		val = "GOTO"
	case input_TOK:
		val = "INPUT"
	case int1Byte_TOK:
		bt, err := rdr.src.ReadByte()

		if err != nil {
			rdr.eof = true
			return ""
		}
		val = fmt.Sprintf("%d", int(bt))
	case let_TOK:
		val = "LET"
	case next_TOK:
		val = "NEXT"
	case print_TOK:
		val = "PRINT"
	case read_TOK:
		val = "READ"
	case space_TOK:
		val = " "
	case stop_TOK:
		val = "STOP"
	case width_TOK:
		val = "WIDTH"
	}
	return val
}

// A line header is composed of two 16bit numbers
// the offset to the next line
// the line number
// note that the first offset is a memory address
// and not really useful for finding the second line
func (rdr *progRdr) readLineHeader() {
	// first get the offset of the next line
	rdr.offset = rdr.readInt()
	if rdr.offset == 0 {
		rdr.linenum = 0
		return
	}

	rdr.linenum = rdr.readInt()
	if rdr.linenum == 0 {
		return
	}
}

// reads a 16bit integer in little endian notation
func (rdr *progRdr) readInt() int {
	byts := make([]byte, 2)
	n, err := rdr.src.Read(byts)

	if (err != nil) || (n < 2) {
		rdr.eof = true
		return 0
	}

	num := int(byts[0]) + int(byts[1])<<8

	return num
}

// Handlers for the different progarm statements

func (rdr *progRdr) copyString() string {
	str, err := rdr.src.ReadBytes('"')

	if err != nil {
		rdr.eof = true
	}

	if len(str) == 0 {
		return "\""
	}

	val := string(str)
	return val
}

const (
	eol_TOK      = 0x00
	oct_TOK      = 0x0b
	hex_TOK      = 0x0c
	lineNum_TOK  = 0x0e
	int1Byte_TOK = 0x0f
	const0_TOK   = 0x11
	const1_TOK   = 0x12
	const2_TOK   = 0x13
	const3_TOK   = 0x14
	const4_TOK   = 0x15
	const5_TOK   = 0x16
	const6_TOK   = 0x17
	const7_TOK   = 0x18
	const8_TOK   = 0x1A
	const9_TOK   = 0x1B
	int2Byte_TOK = 0x1C
	flt4Byte_TOK = 0x1d
	flt8Byte_TOK = 0x1f
	space_TOK    = 0x20
	dblQuote_TOK = 0x22
	colon_TOK    = 0x3a
)

const (
	end_TOK = iota + 0x81
	for_TOK
	next_TOK
	data_TOK
	input_TOK
	dim_TOK
	read_TOK
	let_TOK
	goto_TOK
	run_TOK
	if_TOK
	restore_TOK
	gosub_TOK
	return_TOK
	rem_TOK
	stop_TOK // 0x90
	print_TOK
	clear_TOK
	list_TOK
	new_TOK
	on_TOK
	wait_TOK
	def_TOK
	_ // poke
	_ // cont
	_ // not used
	_ // not used
	_ // out
	lprint_TOK
	llist_TOK
	_         // not used
	width_TOK // 0xa0
	else_TOK
	tron_TOK
	troff_TOK
	swap_TOK
	erase_TOK
	edit_TOK
	error_TOK
	resume_TOK
	delete_TOK
	auto_TOK
	renum_TOK
	defstr_TOK
	defint_TOK
	defsng_TOK
	defdbl_TOK
	line_TOK // 0xb0
)
