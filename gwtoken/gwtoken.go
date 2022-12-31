package gwtoken

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"

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

// basReader interface allows me to hide weather a regular reader is in use
// or a protected reader.  Once a reader is chosen, the code stops caring.
type basReader interface {
	Peek(int) ([]byte, error)
	Read([]byte) (int, error)
	ReadByte() (byte, error)
	ReadBytes(byte) ([]byte, error)
}

// progRdr holds the read I get bytes from and the text being de-tokenized
// for the current line
type progRdr struct {
	src     basReader
	eof     bool
	offset  int    // offset of the current line
	linenum int    // current line number being processed
	lineInp string // the text to parse
}

// protReader reads bytes and reverses basic's obfuscation of the programs contents
type protReader struct {
	src          *bufio.Reader
	subKey1Index int
	xorKey1Index int
	xorKey2Index int
	addKey2Index int
}

// injects a protected reader infront of a regular reader
// and then calls readProg to process the input file
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

// create a progRdr object and then call readProg
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

// used for testing
/*func ParseFileToText(src *bufio.Reader, dest *bufio.Writer, env *object.Environment) {
	dest.WriteString("gwtoken.ParseFileToText")
	bt := make([]byte, 1)
	n, err := src.Read(bt)

	if (err != nil) || (n != 1) {
		dest.WriteString("File read error")
		return
	}

	if bt[0] != TOKEN_FILE {
		dest.WriteString("Can't parse non token file.")
		return
	}

	rdr := progRdr{src: src}

	for {
		rdr.readLineHeader()
		if rdr.linenum == 0 {
			return
		}
		rdr.readLine(env)
		dest.WriteString(rdr.lineInp)
	}
}*/

// loop to read line header (line number & offset)
// reads and de-tokens until the end of line
// then does lexical processing and parsing of the line
// that causes the AST for the line to be stored into
// the environment for execution
func (rdr *progRdr) readProg(env *object.Environment) {
	for !rdr.eof {
		rdr.readLineHeader()
		if rdr.linenum == 0 {
			return
		}
		rdr.readLine(env)

		l := lexer.New(rdr.lineInp)
		p := parser.New(l)
		p.ParseProgram(env)
	}
}

// reads tokens until I get back nothing
func (rdr *progRdr) readLine(env *object.Environment) {
	rdr.lineInp = fmt.Sprintf("%d ", rdr.linenum)

	for val := "tmp"; val != ""; {
		val = rdr.readToken()
		rdr.lineInp = rdr.lineInp + val
	}
	//rdr.readInt()
}

// reads one token and uses a (really large) switch stmt
// to return the ASCII representation.
//
// some tokens are multibyte, starting with 0xfd, 0xfe, or 0xff
// those tokens trigger a call to the appropriate sub-function
func (rdr *progRdr) readToken() string {
	var val string
	tok, err := rdr.src.ReadByte()

	if err != nil {
		rdr.eof = true
		return ""
	}

	switch tok {
	case and_TOK:
		val = "AND"
	case auto_TOK:
		val = "AUTO"
	case beep_TOK:
		val = "BEEP"
	case bload_TOK:
		val = "BLOAD"
	case bsave_TOK:
		val = "BSAVE"
	case bslash_TOK:
		val = `\`
	case call_TOK:
		val = "CALL"
	case clear_TOK:
		val = "CLEAR"
	case close_TOK:
		val = "CLOSE"
	case cls_TOK:
		val = "CLS"
	case ':':
		val = ":"
		pk, err := rdr.src.Peek(1)
		if err != nil {
			rdr.eof = true
			return val
		}
		if pk[0] == else_TOK {
			val = rdr.readToken()
		}
		if pk[0] == rem_TOK {
			_ = rdr.readToken()
			val = rdr.readToken()
		}
	case '"':
		val = "\"" + rdr.readString()
	case color_TOK:
		val = "COLOR"
	case const0_TOK, const1_TOK, const2_TOK, const3_TOK, const4_TOK, const5_TOK, const6_TOK, const7_TOK:
		val = fmt.Sprintf("%d", int(tok-const0_TOK))
	case const8_TOK, const9_TOK: // discontinuity in map
		val = fmt.Sprintf("%d", int(tok-const0_TOK-1))
	case cont_TOK:
		val = "CONT"
	case csrlin_TOK:
		val = "CSRLIN"
	case data_TOK:
		val = "DATA"
	case delete_TOK:
		val = "DELETE"
	case def_TOK:
		val = "DEF"
	case defdbl_TOK:
		val = "DEFDBL"
	case defint_TOK:
		val = "DEFINT"
	case defsng_TOK:
		val = "DEFSNG"
	case defstr_TOK:
		val = "DEFSTR"
	case dim_TOK:
		val = "DIM"
	case div_TOK:
		val = "/"
	case edit_TOK:
		val = "EDIT"
	case else_TOK:
		val = "ELSE"
	case end_TOK:
		val = "END"
	case erl_TOK:
		val = "ERL"
	case err_TOK:
		val = "ERR"
	case error_TOK:
		val = "ERROR"
	case eol_TOK:
		val = ""
	case eqv_TOK:
		val = "EQV"
	case erase_TOK:
		val = "ERASE"
	case eq_TOK:
		val = "="
	case fd_TOK:
		return rdr.fdPage()
	case fe_TOK:
		return rdr.fePage()
	case ff_TOK:
		return rdr.ffPage()
	case fix_TOK:
		val = "FIX"
	case flt4Byte_TOK:
		val = rdr.read4ByteFloat()
	case flt8Byte_TOK:
		val = rdr.read8ByteFloat()
	case fn_TOK:
		val = "FN"
	case for_TOK:
		val = "FOR"
	case goto_TOK:
		val = "GOTO"
	case gosub_TOK:
		val = "GOSUB"
	case gt_TOK:
		val = ">"
	case hex_TOK:
		val = rdr.readHexConst()
	case if_TOK:
		val = "IF"
	case imp_TOK:
		val = "IMP"
	case inkeys_TOK:
		val = "INKEY$"
	case input_TOK:
		val = "INPUT"
	case instr_TOK:
		val = "INSTR"
	case int1Byte_TOK:
		bt, err := rdr.src.ReadByte()

		if err != nil {
			rdr.eof = true
			return ""
		}
		val = fmt.Sprintf("%d", int(bt))
	case int2Byte_TOK:
		val = fmt.Sprintf("%d", rdr.readInt())
	case key_TOK:
		val = "KEY"
	case let_TOK:
		val = "LET"
	case line_TOK:
		val = "LINE"
	case lineNum_TOK:
		val = fmt.Sprintf("%d", rdr.readInt())
	case list_TOK:
		val = "LIST"
	case llist_TOK:
		val = "LLIST"
	case load_TOK:
		val = "LOAD"
	case locate_TOK:
		val = "LOCATE"
	case lprint_TOK:
		val = "LPRINT"
	case lt_TOK:
		val = "<"
	case merge_TOK:
		val = "MERGE"
	case minus_TOK:
		val = "-"
	case mod_TOK:
		val = "MOD"
	case motor_TOK:
		val = "MOTOR"
	case mul_TOK:
		val = "*"
	case new_TOK:
		val = "NEW"
	case next_TOK:
		val = "NEXT"
	case not_TOK:
		val = "NOT"
	case oct_TOK:
		val = rdr.readOctConst()
	case off_TOK:
		val = "OFF"
	case on_TOK:
		val = "ON"
	case open_TOK:
		val = "OPEN"
	case option_TOK:
		val = "OPTION"
	case or_TOK:
		val = "OR"
	case out_TOK:
		val = "OUT"
	case plus_TOK:
		val = "+"
	case preset_TOK:
		val = "PRESET"
	case point_TOK:
		val = "POINT"
	case pset_TOK:
		val = "PSET"
	case poke_TOK:
		val = "POKE"
	case print_TOK:
		val = "PRINT"
	case pwr_TOK:
		val = "^"
	case randomize_TOK:
		val = "RANDOMIZE"
	case read_TOK:
		val = "READ"
	case rem_TOK:
		val = "REM"
	case renum_TOK:
		val = "RENUM"
	case restore_TOK:
		val = "RESTORE"
	case resume_TOK:
		val = "RESUME"
	case return_TOK:
		val = "RETURN"
	case run_TOK:
		val = "RUN"
	case save_TOK:
		val = "SAVE"
	case screen_TOK:
		val = "SCREEN"
	case sound_TOK:
		val = "SOUND"
	case spc_TOK:
		val = "SPC("
	case step_TOK:
		val = "STEP"
	case stop_TOK:
		val = "STOP"
	case strings_TOK:
		val = "STRING$"
	case swap_TOK:
		val = "SWAP"
	case tab_TOK:
		val = "TAB("
	case then_TOK:
		val = "THEN"
	case ticrem_TOK:
		val = "'"
	case to_TOK:
		val = "TO"
	case troff_TOK:
		val = "TROFF"
	case tron_TOK:
		val = "TRON"
	case using_TOK:
		val = "USING"
	case usr_TOK:
		val = "USR"
	case varptr_TOK:
		val = "VARPTR"
	case wait_TOK:
		val = "WAIT"
	case wend_TOK:
		val = "WEND"
	case while_TOK:
		val = "WHILE"
	case width_TOK:
		val = "WIDTH"
	case write_TOK:
		val = "WRITE"
	case xor_TOK:
		val = "XOR"
	default:
		if (tok >= 0x20) && (tok <= 0x7f) {
			val = string(tok)
		}
	}
	return val
}

func (rdr *progRdr) fdPage() string {
	var val string
	tok, err := rdr.src.ReadByte()

	if err != nil {
		rdr.eof = true
		return ""
	}

	switch tok {
	case cvd_TOK:
		val = "CVD"
	case cvi_TOK:
		val = "CVI"
	case cvs_TOK:
		val = "CVS"
	case exterr_TOK:
		val = "EXTERR"
	case mkds_TOK:
		val = "MKD$"
	case mkis_TOK:
		val = "MKI$"
	case mkss_TOK:
		val = "MKS$"
	}
	return val
}

func (rdr *progRdr) fePage() string {
	var val string
	tok, err := rdr.src.ReadByte()

	if err != nil {
		rdr.eof = true
		return ""
	}

	switch tok {
	case calls_TOK:
		val = "CALLS"
	case chain_TOK:
		val = "CHAIN"
	case chdir_TOK:
		val = "CHDIR"
	case circle_TOK:
		val = "CIRCLE"
	case com_TOK:
		val = "COM"
	case common_TOK:
		val = "COMMON"
	case dates_TOK:
		val = "DATE$"
	case draw_TOK:
		val = "DRAW"
	case environ_TOK:
		val = "ENVIRON"
	case erdev_TOK:
		val = "ERDEV"
	case field_TOK:
		val = "FIELD"
	case files_TOK:
		val = "FILES"
	case get_TOK:
		val = "GET"
	case ioctl_TOK:
		val = "IOCTL"
		bts, err := rdr.src.Peek(1)
		if err == nil {
			if bts[0] == '$' {
				val = val + "$"
			}
		}
	case kill_TOK:
		val = "KILL"
	case lock_TOK:
		val = "LOCK"
	case paint_TOK:
		val = "PAINT"
	case palette_TOK:
		val = "PALETTE"
	case play_TOK:
		val = "PLAY"
	case pmap_TOK:
		val = "PMAP"
	case put_TOK:
		val = "PUT"
	case lcopy_TOK:
		val = "LCOPY"
	case lset_TOK:
		val = "LSET"
	case mkdir_TOK:
		val = "MKDIR"
	case name_TOK:
		val = "NAME"
	case pcopy_TOK:
		val = "PCOPY"
	case reset_TOK:
		val = "RESET"
	case rmdir_TOK:
		val = "RMDIR"
	case rset_TOK:
		val = "RSET"
	case shell_TOK:
		val = "SHELL"
	case system_TOK:
		val = "SYSTEM"
	case timer_TOK:
		val = "TIMER"
	case times_TOK:
		val = "TIME$"
	case unlock_TOK:
		val = "UNLOCK"
	case view_TOK:
		val = "VIEW"
	case window_TOK:
		val = "WINDOW"
	}
	return val
}

func (rdr *progRdr) ffPage() string {
	var val string
	tok, err := rdr.src.ReadByte()

	if err != nil {
		rdr.eof = true
		return ""
	}

	switch tok {
	case abs_TOK:
		val = "ABS"
	case asc_TOK:
		val = "ASC"
	case atn_TOK:
		val = "ATN"
	case cdbl_TOK:
		val = "CDBL"
	case chrs_TOK:
		val = "CHR$"
	case cint_TOK:
		val = "CINT"
	case cos_TOK:
		val = "COS"
	case csng_TOK:
		val = "CSNG"
	case eof_TOK:
		val = "EOF"
	case exp_TOK:
		val = "EXP"
	case fix_TOK:
		val = "FIX"
	case fre_TOK:
		val = "FRE"
	case hexs_TOK:
		val = "HEX$"
	case inp_TOK:
		val = "INP"
	case int_TOK:
		val = "INT"
	case lefts_TOK:
		val = "LEFT$"
	case len_TOK:
		val = "LEN"
	case loc_TOK:
		val = "LOC"
	case lof_TOK:
		val = "LOF"
	case log_TOK:
		val = "LOG"
	case lpos_TOK:
		val = "LPOS"
	case mids_TOK:
		val = "MID$"
	case octs_TOK:
		val = "OCT$"
	case peek_TOK:
		val = "PEEK"
	case pen_TOK:
		val = "PEN"
	case pos_TOK:
		val = "POS"
	case rights_TOK:
		val = "RIGHT$"
	case rnd_TOK:
		val = "RND"
	case sgn_TOK:
		val = "SGN"
	case sin_TOK:
		val = "SIN"
	case spaces_TOK:
		val = "SPACE$"
	case sqr_TOK:
		val = "SQR"
	case stick_TOK:
		val = "STICK"
	case strs_TOK:
		val = "STR$"
	case strig_TOK:
		val = "STRIG"
	case tan_TOK:
		val = "TAN"
	case val_TOK:
		val = "VAL"
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

// read an octal constant like "&O3627"
// which would stored as 0x0b, 0x97, 0x07
// the laste two bytes are just the integer
// value of the constant
func (rdr *progRdr) readOctConst() string {
	num := rdr.readInt()
	val := "&O" + fmt.Sprintf("%o", num)

	return val
}

// read a hex constant like "&H3712"
// which is also saved as an integer
func (rdr *progRdr) readHexConst() string {
	num := rdr.readInt()
	val := "&H" + fmt.Sprintf("%X", num)

	return val
}

// Handlers for the different progarm statements

// readString reads the contents of a quoted string
func (rdr *progRdr) readString() string {
	// read everything to the next quote
	str, err := rdr.src.ReadBytes('"')

	// if he ran out of bytes before finding a quote, he errors
	if err != nil {
		rdr.eof = true
	}

	// if no data, append a closing quote
	if len(str) == 0 {
		return "\""
	}

	// return convert bytes in array as a string
	return object.DecodeBytes(str)
}

func (rdr *progRdr) read8ByteFloat() string {
	bts := make([]byte, 8)
	n, err := rdr.src.Read(bts)

	if (err != nil) || (n < 8) {
		return "0"
	}

	/* MS Binary Format                                             */
	/* byte order =>    m7 | m6 | m5 | m4 | m3 | m2 | m1 | exponent */
	/* m1 is most significant byte => smmm|mmmm                     */
	/* m7 is the least significant byte                             */
	/*      m = mantissa byte                                       */
	/*      s = sign bit                                            */
	/*      b = bit                                                 */

	if bts[7] == 0 {
		return "0"
	}

	sign := bts[6] & 0x80 /* 1000|0000b  */

	/* IEEE Single Precision Float Format                           */
	/*  byte 8    byte 7    byte 6    byte 5    byte 4    and so on */
	/* seee|eeee eeee|mmmm mmmm|mmmm mmmm|mmmm mmmm|mmmm ...        */
	/*          s = sign bit                                        */
	/*          e = exponent bit                                    */
	/*          m = mantissa bit                                    */

	ieee_exp := int(bts[7]) - 128 - 1 + 1023

	bts[7] = sign | byte(ieee_exp>>4)

	for i := 6; i > 0; i-- {
		bts[i] = bts[i] << 1
		bts[i] |= bts[i-1] >> 7
	}
	bts[0] = bts[0] << 1

	for i := 0; i < 6; i++ {
		bts[i] = (bts[i] >> 4) | (bts[i+1]&0x0f)<<4
	}
	bts[6] = (bts[6] >> 4) | (byte(ieee_exp)&0x0f)<<4

	flt := math.Float64frombits(binary.LittleEndian.Uint64(bts))

	return fmt.Sprintf("%E", flt)
}

func (rdr *progRdr) read4ByteFloat() string {
	bts := make([]byte, 4)
	n, err := rdr.src.Read(bts)

	if (err != nil) || (n < 4) {
		return "0"
	}
	/* MS Binary Format                         */
	/* byte order =>    m3 | m2 | m1 | exponent */
	/* m1 is most significant byte => sbbb|bbbb */
	/* m3 is the least significant byte         */
	/*      m = mantissa byte                   */
	/*      s = sign bit                        */
	/*      b = bit                             */

	if bts[3] == 0 {
		return "0"
	}

	/* IEEE Single Precision Float Format       */
	/*    m3        m2        m1     exponent   */
	/* mmmm|mmmm mmmm|mmmm emmm|mmmm seee|eeee  */
	/*          s = sign bit                    */
	/*          e = exponent bit                */
	/*          m = mantissa bit                */

	sign := bts[2] & 0x80 /* 1000|0000b  */
	ieee_exp := bts[3] - 2
	bts[3] = sign | ieee_exp>>1
	bts[2] = (bts[2] & 0x7f) | (ieee_exp << 7)
	flt := math.Float32frombits(binary.LittleEndian.Uint32(bts))

	return fmt.Sprintf("%E", flt)
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
	fd_TOK       = 0xfd
	fe_TOK       = 0xfe
	ff_TOK       = 0xff
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
	poke_TOK
	cont_TOK // cont
	_        // 0x9a not used
	_        // 0x9b not used
	out_TOK  // out
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
	while_TOK
	wend_TOK
	call_TOK
	_ //Undefined
	_ //Undefined
	_ //Undefined
	write_TOK
	option_TOK
	randomize_TOK
	open_TOK
	close_TOK
	load_TOK
	merge_TOK
	save_TOK
	color_TOK
	cls_TOK // 0xc0
	motor_TOK
	bsave_TOK
	bload_TOK
	sound_TOK
	beep_TOK
	pset_TOK
	preset_TOK
	screen_TOK
	key_TOK
	locate_TOK
	_ //Undefined
	to_TOK
	then_TOK
	tab_TOK
	step_TOK
	usr_TOK // 0xd0
	fn_TOK
	spc_TOK
	not_TOK
	erl_TOK
	err_TOK
	strings_TOK
	using_TOK
	instr_TOK
	ticrem_TOK
	varptr_TOK
	csrlin_TOK
	point_TOK
	off_TOK
	inkeys_TOK
	_
	_
	_
	_
	_
	_
	_
	gt_TOK
	eq_TOK
	lt_TOK
	plus_TOK
	minus_TOK
	mul_TOK
	div_TOK
	pwr_TOK
	and_TOK
	or_TOK
	xor_TOK // 0xf0
	eqv_TOK
	imp_TOK
	mod_TOK
	bslash_TOK
)

// tokens prefixed by 0xfd
const (
	cvi_TOK = iota + 0x81
	cvs_TOK
	cvd_TOK
	mkis_TOK
	mkss_TOK
	mkds_TOK
	_
	_
	_
	_
	exterr_TOK
)

// tokens prefixed by 0xfe
const (
	files_TOK = iota + 0x81
	field_TOK
	system_TOK
	name_TOK
	lset_TOK
	rset_TOK
	kill_TOK
	put_TOK
	get_TOK
	reset_TOK
	common_TOK
	chain_TOK
	dates_TOK
	times_TOK
	paint_TOK
	com_TOK
	circle_TOK
	draw_TOK
	play_TOK
	timer_TOK
	erdev_TOK
	ioctl_TOK
	chdir_TOK
	mkdir_TOK
	rmdir_TOK
	shell_TOK
	environ_TOK
	view_TOK
	window_TOK
	pmap_TOK
	palette_TOK
	lcopy_TOK
	calls_TOK
	_
	_
	_
	pcopy_TOK
	_
	lock_TOK
	unlock_TOK
)

// tokens prefixed by 0xff
const (
	lefts_TOK = iota + 0x81
	rights_TOK
	mids_TOK
	sgn_TOK
	int_TOK
	abs_TOK
	sqr_TOK
	rnd_TOK
	sin_TOK
	log_TOK
	exp_TOK
	cos_TOK
	tan_TOK
	atn_TOK
	fre_TOK
	inp_TOK //0x90
	pos_TOK
	len_TOK
	strs_TOK
	val_TOK
	asc_TOK
	chrs_TOK
	peek_TOK
	spaces_TOK
	octs_TOK
	hexs_TOK
	lpos_TOK
	cint_TOK
	csng_TOK
	cdbl_TOK
	fix_TOK
	pen_TOK //0xa0
	stick_TOK
	strig_TOK
	eof_TOK
	loc_TOK
	lof_TOK
)
