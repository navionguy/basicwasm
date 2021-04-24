package gwtoken

import (
	"bufio"
	"fmt"

	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"
)

const (
	TOKEN_FILE     = 0xff
	PROTECTED_FILE = 0xfe
)

var subBytesKey1 = []byte{0x0b, 0x0a, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
var xorBytesKey1 = []byte{0x1e, 0x1d, 0xc4, 0x77, 0x26, 0x97, 0xe0, 0x74, 0x59, 0x88, 0x7c}

var xorBytesKey2 = []byte{0xa9, 0x84, 0x8d, 0xcd, 0x75, 0x83, 0x43, 0x63, 0x24, 0x83, 0x19, 0xf7, 0x9a}
var addBytesKey2 = []byte{0x0d, 0x0c, 0x0b, 0x0a, 0x09, 0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}

type basReader interface {
	Peek(int) ([]byte, error)
	Read([]byte) (int, error)
	ReadByte() (byte, error)
	ReadBytes(byte) ([]byte, error)
}

type progRdr struct {
	src     basReader
	eof     bool
	offset  int    // offset of the current line
	linenum int    // current line number being processed
	lineInp string // the text to parse
}

type protReader struct {
	src          *bufio.Reader
	subKey1Index int
	xorKey1Index int
	xorKey2Index int
	addKey2Index int
}

func ParseProtectedFile(src *bufio.Reader, env *object.Environment) {
	prtRdr := protReader{src: src}
	rdr := progRdr{src: &prtRdr}

	bt := make([]byte, 1)
	n, err := src.Read(bt)

	if (err != nil) || (n != 1) {
		env.Terminal().Println("File read error")
		return
	}

	if bt[0] != PROTECTED_FILE {
		env.Terminal().Println("Bad file mode")
		return
	}

	rdr.readProg(env)
}

func ParseFile(src *bufio.Reader, env *object.Environment) {
	bt := make([]byte, 1)
	n, err := src.Read(bt)

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
		val = ":"
		pk, err := rdr.src.Peek(1)
		if err != nil {
			rdr.eof = true
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

// protReader hides an inner bufio.Reader and
// decrypts the bytes as they com out

func (pr *protReader) Peek(n int) ([]byte, error) {
	bt, err := pr.src.Peek(n)

	if len(bt) == 0 {
		// nothing more to do
		return bt, err
	}

	savSub1 := pr.subKey1Index
	xorKey1 := pr.xorKey1Index
	xorKey2 := pr.xorKey2Index
	addKey2 := pr.addKey2Index

	bt = pr.decryptBytes(bt)

	pr.subKey1Index = savSub1
	pr.xorKey1Index = xorKey1
	pr.xorKey2Index = xorKey2
	pr.addKey2Index = addKey2

	return bt, err
}

func (pr *protReader) Read(bt []byte) (int, error) {
	n, err := pr.src.Read(bt)

	if n > 0 {
		bt = pr.decryptBytes(bt[:n])
	}

	return n, err
}

func (pr *protReader) ReadByte() (byte, error) {
	bt, err := pr.src.ReadByte()

	bt = pr.decryptByte(bt)
	return bt, err
}

func (pr *protReader) ReadBytes(delim byte) ([]byte, error) {
	var bts []byte

	bt, err := pr.src.ReadByte()

	if err != nil {
		return bts, err
	}

	bt = pr.decryptByte(bt)
	bts = append(bts, bt)
	for (bt != delim) && (err == nil) {
		bt, err = pr.src.ReadByte()
		bt = pr.decryptByte(bt)
		bts = append(bts, bt)
	}

	return bts, err
}

func (pr *protReader) decryptBytes(bts []byte) []byte {
	for i, bt := range bts {
		bts[i] = pr.decryptByte(bt)
	}

	return bts
}

func (pr *protReader) decryptByte(bt byte) byte {
	bt = (((bt - subBytesKey1[pr.subKey1Index]) ^ xorBytesKey1[pr.xorKey1Index]) ^ xorBytesKey2[pr.xorKey2Index]) + addBytesKey2[pr.addKey2Index]

	pr.subKey1Index = pr.advIndex(pr.subKey1Index, 11)
	pr.xorKey1Index = pr.advIndex(pr.xorKey1Index, 11)
	pr.xorKey2Index = pr.advIndex(pr.xorKey2Index, 13)
	pr.addKey2Index = pr.advIndex(pr.addKey2Index, 13)

	return bt
}

func (pr *protReader) advIndex(id int, max int) int {
	id++

	if id < max {
		return id
	}
	return 0
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
