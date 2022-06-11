package gwtoken

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/stretchr/testify/assert"
)

func Test_LiveFile(t *testing.T) {

}

func Test_ParseFile(t *testing.T) {
	tests := []struct {
		prg   []byte
		ascii string
	}{
		{prg: []byte{}},
		{prg: []byte{0x00}},
		{prg: []byte{0xFF, 0x82}},
		{prg: []byte{0xFF, 0x82, 0x12, 0x0a, 0x00}},       // test case line with no statements
		{prg: []byte{0xFF, 0x82, 0x12, 0x0a, 0x00, 0x3a}}, // test case line with no statements
		{prg: []byte{0xFF, 0x82, 0x12, 0x00, 0x00, 0x91, 0x00, 0x00, 0x00, 0x1A}}, // test case line with zero linenum
		{prg: []byte{0xFF, 0x82, 0x12, 0x0A, 0x00, 0x91, 0x20, 0x22, 0x48, 0x65,
			0x6C, 0x6C, 0x6F, 0x20, 0x57, 0x6F, 0x72, 0x6C, 0x64, 0x22, 0x00, 0x00, 0x00, 0x1A}},
		{prg: []byte{0xFF, 0x82, 0x12, 0x0A, 0x00, 0x91, 0x00, 0x00, 0x00, 0x1A}},
	}

	for _, tt := range tests {
		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		mt.ExpMsg = &mocks.Expector{}
		env := object.NewTermEnvironment(mt)
		rdr := bufio.NewReader(bytes.NewReader(tt.prg))
		ParseFile(rdr, env)
	}
}

func Test_TokenTable(t *testing.T) {
	tests := []struct {
		inp []byte
		exp []string
	}{
		{inp: []byte{':'}, exp: []string{":"}},              // special case, colon followed by nothing is legal
		{inp: []byte{':', else_TOK}, exp: []string{"ELSE"}}, // special case, colon followed by else is ELSE
		{inp: []byte{int1Byte_TOK, 0x0a}, exp: []string{"10"}},
		{inp: []byte{int1Byte_TOK}, exp: []string{""}}, // error case
		{inp: []byte{'"'}, exp: []string{`""`}},
		{inp: []byte{eol_TOK, end_TOK, stop_TOK, width_TOK, ':', data_TOK, '"', 'A', '"', dim_TOK, else_TOK, end_TOK, for_TOK, goto_TOK, input_TOK},
			exp: []string{"", "END", "STOP", "WIDTH", ":", "DATA", `"A"`, "DIM", "ELSE", "END", "FOR", "GOTO", "INPUT"}},
		{inp: []byte{let_TOK, next_TOK, read_TOK}, exp: []string{"LET", "NEXT", "READ"}},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		rdr := progRdr{src: src}
		for i := range tt.exp {
			res := rdr.readToken()
			assert.Equal(t, tt.exp[i], res, "TokenTable expected %s, got %s", tt.exp[i], res)

		}
	}
}

func Test_DecryptByte(t *testing.T) {
	tests := []struct {
		inp []byte
		out []byte
	}{
		{inp: []byte{0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4},
			out: []byte{0x82, 0x12, 0x0a, 0x00, 0x91, 0x20, 0x22, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x22, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		pr := protReader{}
		for i, bt := range tt.inp {
			bt = pr.decryptByte(bt)

			assert.Equal(t, tt.out[i], bt, "byte decrypt failed")
		}
	}
}

func Test_DecryptBytes(t *testing.T) {
	tests := []struct {
		inp []byte
		out []byte
	}{
		{inp: []byte{0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4},
			out: []byte{0x82, 0x12, 0x0a, 0x00, 0x91, 0x20, 0x22, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x22, 0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		pr := protReader{}
		bts := pr.decryptBytes(tt.inp)
		for i, bt := range bts {
			assert.Equal(t, tt.out[i], bt, "bytes decrypt failed")
		}
	}
}

func Test_ReadProg(t *testing.T) {
	tests := []struct {
		inp   []byte
		stmts int
	}{
		{inp: []byte{0x7C, 0x12, 0x0A, 0x00, 0x91, 0x20, 0x22, 0x48, 0x65, 0x6C,
			0x6C, 0x6F, 0x22, 0x00, 0x87, 0x12, 0x14, 0x00, 0x59, 0x20, 0xE7,
			0x20, 0x0F, 0x96, 0x00, 0x92, 0x12, 0x1E, 0x00, 0x5A, 0x20, 0xE7,
			0x20, 0x0F, 0x30, 0x00, 0x00, 0x00, 0x1A}, stmts: 6},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		rdr := progRdr{src: src}

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		rdr.readProg(env)
		itr := env.StatementIter()

		assert.Equal(t, tt.stmts, itr.Len(), "Test_readProg expected %d statements, got %d", tt.stmts, itr.Len())
	}
}

func Test_RemStatementShortForm(t *testing.T) {
	tests := []struct {
		inp   []byte
		stmts int
	}{
		{inp: []byte{0x7C, 0x12, 0x0A, 0x00, 0x3A, 0x8F, 0xD9, 0x20, 0x65, 0x6C,
			0x6C, 0x6F, 0x22, 0x00, 0x87, 0x12, 0x14, 0x00, 0x59, 0x20, 0xE7,
			0x20, 0x0F, 0x96, 0x00, 0x92, 0x12, 0x1E, 0x00, 0x5A, 0x20, 0xE7,
			0x20, 0x0F, 0x30, 0x00, 0x00, 0x00, 0x1A}, stmts: 2},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		rdr := progRdr{src: src}

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		rdr.readProg(env)
		itr := env.StatementIter()

		assert.Equal(t, tt.stmts, itr.Len(), "Test_readProg expected %d statements, got %d", tt.stmts, itr.Len())
	}

}

func Test_ReadProtProg(t *testing.T) {
	tests := []struct {
		inp   []byte
		stmts int
	}{
		{inp: []byte{0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4}, stmts: 2},
		{inp: []byte{0xCB, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x1C, 0x18, 0x7B, 0x02, 0xC8, 0x87, 0x78, 0xC5, 0x19, 0xCF, 0x94, 0x74, 0x87, 0xA4, 0x6C, 0x03}, stmts: 4},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		pr := protReader{src: src}
		rdr := progRdr{src: &pr}

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		env := object.NewTermEnvironment(mt)

		rdr.readProg(env)
		itr := env.StatementIter()

		assert.Equal(t, tt.stmts, itr.Len(), "Test_readProgProtected expected %d statements, got %d", tt.stmts, itr.Len())
	}
}

func Test_ReadFD_Bytes(t *testing.T) {
	tests := []struct {
		inp []byte
		exp []string
	}{
		{inp: []byte{fd_TOK}, exp: []string{""}}, // error case
		{inp: []byte{fd_TOK, cvi_TOK}, exp: []string{"CVI"}},
		{inp: []byte{fe_TOK}, exp: []string{""}}, // error case
		{inp: []byte{fe_TOK, files_TOK}, exp: []string{"FILES"}},
		{inp: []byte{ff_TOK}, exp: []string{""}}, // error case
		{inp: []byte{ff_TOK, files_TOK}, exp: []string{"LEFT$"}},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))
		rdr := progRdr{src: src}
		for i := range tt.exp {
			res := rdr.readToken()
			assert.Equal(t, tt.exp[i], res, "TokenTable expected %s, got %s", tt.exp[i], res)

		}
	}

}

func Test_ParseProtProg(t *testing.T) {
	tests := []struct {
		inp   []byte
		stmts int
	}{
		{inp: []byte{0xFE, 0xE3, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x1C, 0x18, 0x7B, 0x02, 0xC8}, stmts: 2}, // test case colon at eol
		{inp: []byte{0xFE, 0xE3, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD}, stmts: 2},                                     // test case dblquote at eol
		{inp: []byte{}, stmts: 0},
		{inp: []byte{0xff}, stmts: 0},
		{inp: []byte{0xfe, 0xCD, 0xA9, 0xBF, 0x54, 0xE2, 0x12, 0xBD, 0x59, 0x20, 0x65, 0x0D, 0x8F, 0xA2, 0x30, 0x98, 0xD3, 0x3E, 0xD3, 0xF1, 0xE6, 0x13, 0xA4}, stmts: 2},
	}

	for _, tt := range tests {
		src := bufio.NewReader(bytes.NewReader(tt.inp))

		var mt mocks.MockTerm
		mocks.InitMockTerm(&mt)
		mt.ExpMsg = &mocks.Expector{}
		env := object.NewTermEnvironment(mt)

		ParseProtectedFile(src, env)

		itr := env.StatementIter()

		assert.Equal(t, tt.stmts, itr.Len(), "Test_ParseProtectedFile")
	}
}

func Test_TokenTestPoints(t *testing.T) {
	inp := []byte{0x81, 0x90, 0xa0, 0xb0, 0xc0, 0xd0, 0xf0}
	exp := []byte{end_TOK, stop_TOK, width_TOK, line_TOK, cls_TOK, usr_TOK, xor_TOK}

	for i := range inp {
		assert.Equal(t, exp[i], inp[i], "token value for %x not correct!", inp[i])
	}
}

func Test_BaseTokens(t *testing.T) {
	tests := []struct {
		inp []byte
		exp string
		tok byte
	}{
		{inp: []byte{0x00}, exp: "", tok: eol_TOK},
		{inp: []byte{0x0b, 0x97, 0x07}, exp: "&O3627", tok: oct_TOK},
		{inp: []byte{0x0c, 0x97, 0x07}, exp: "&H797", tok: hex_TOK},
		{inp: []byte{0x0e, 0x0a, 0x00}, exp: "10", tok: lineNum_TOK},
		{inp: []byte{0x0f, 0x0a}, exp: "10", tok: int1Byte_TOK},
		{inp: []byte{0x11}, exp: "0", tok: const0_TOK},
		{inp: []byte{0x12}, exp: "1", tok: const1_TOK},
		{inp: []byte{0x13}, exp: "2", tok: const2_TOK},
		{inp: []byte{0x14}, exp: "3", tok: const3_TOK},
		{inp: []byte{0x15}, exp: "4", tok: const4_TOK},
		{inp: []byte{0x16}, exp: "5", tok: const5_TOK},
		{inp: []byte{0x17}, exp: "6", tok: const6_TOK},
		{inp: []byte{0x18}, exp: "7", tok: const7_TOK},
		{inp: []byte{0x1a}, exp: "8", tok: const8_TOK},
		{inp: []byte{0x1b}, exp: "9", tok: const9_TOK},
		{inp: []byte{0x1c, 0x72, 0x01}, exp: "370", tok: int2Byte_TOK},
		{inp: []byte{0x1d}, exp: "0", tok: flt4Byte_TOK},
		{inp: []byte{0x1f}, exp: "0", tok: flt8Byte_TOK},

		{inp: []byte{0x81}, exp: "END", tok: end_TOK},
		{inp: []byte{0x82}, exp: "FOR", tok: for_TOK},
		{inp: []byte{0x83}, exp: "NEXT", tok: next_TOK},
		{inp: []byte{0x84}, exp: "DATA", tok: data_TOK},
		{inp: []byte{0x85}, exp: "INPUT", tok: input_TOK},
		{inp: []byte{0x86}, exp: "DIM", tok: dim_TOK},
		{inp: []byte{0x87}, exp: "READ", tok: read_TOK},
		{inp: []byte{0x88}, exp: "LET", tok: let_TOK},
		{inp: []byte{0x89}, exp: "GOTO", tok: goto_TOK},
		{inp: []byte{0x8a}, exp: "RUN", tok: run_TOK},
		{inp: []byte{0x8b}, exp: "IF", tok: if_TOK},
		{inp: []byte{0x8c}, exp: "RESTORE", tok: restore_TOK},
		{inp: []byte{0x8d}, exp: "GOSUB", tok: gosub_TOK},
		{inp: []byte{0x8e}, exp: "RETURN", tok: return_TOK},
		{inp: []byte{0x8f}, exp: "REM", tok: rem_TOK},
		{inp: []byte{':', 0xa1}, exp: "ELSE", tok: ':'},

		{inp: []byte{0x90}, exp: "STOP", tok: stop_TOK},
		{inp: []byte{0x91}, exp: "PRINT", tok: print_TOK},
		{inp: []byte{0x92}, exp: "CLEAR", tok: clear_TOK},
		{inp: []byte{0x93}, exp: "LIST", tok: list_TOK},
		{inp: []byte{0x94}, exp: "NEW", tok: new_TOK},
		{inp: []byte{0x95}, exp: "ON", tok: on_TOK},
		{inp: []byte{0x96}, exp: "WAIT", tok: wait_TOK},
		{inp: []byte{0x97}, exp: "DEF", tok: def_TOK},
		{inp: []byte{0x98}, exp: "POKE", tok: poke_TOK},
		{inp: []byte{0x99}, exp: "CONT", tok: cont_TOK},
		{inp: []byte{0x9a}, exp: "", tok: 0x9a},
		{inp: []byte{0x9b}, exp: "", tok: 0x9b},
		{inp: []byte{0x9c}, exp: "OUT", tok: out_TOK},
		{inp: []byte{0x9d}, exp: "LPRINT", tok: lprint_TOK},
		{inp: []byte{0x9e}, exp: "LLIST", tok: llist_TOK},
		{inp: []byte{0x9f}, exp: "FIX", tok: 0x9f},

		{inp: []byte{0xa0}, exp: "WIDTH", tok: width_TOK},
		{inp: []byte{0xa1}, exp: "ELSE", tok: else_TOK},
		{inp: []byte{0xa2}, exp: "TRON", tok: tron_TOK},
		{inp: []byte{0xa3}, exp: "TROFF", tok: troff_TOK},
		{inp: []byte{0xa4}, exp: "SWAP", tok: swap_TOK},
		{inp: []byte{0xa5}, exp: "ERASE", tok: erase_TOK},
		{inp: []byte{0xa6}, exp: "EDIT", tok: edit_TOK},
		{inp: []byte{0xa7}, exp: "ERROR", tok: error_TOK},
		{inp: []byte{0xa8}, exp: "RESUME", tok: resume_TOK},
		{inp: []byte{0xa9}, exp: "DELETE", tok: delete_TOK},
		{inp: []byte{0xaa}, exp: "AUTO", tok: auto_TOK},
		{inp: []byte{0xab}, exp: "RENUM", tok: renum_TOK},
		{inp: []byte{0xac}, exp: "DEFSTR", tok: defstr_TOK},
		{inp: []byte{0xad}, exp: "DEFINT", tok: defint_TOK},
		{inp: []byte{0xae}, exp: "DEFSNG", tok: defsng_TOK},
		{inp: []byte{0xaf}, exp: "DEFDBL", tok: defdbl_TOK},

		{inp: []byte{0xb0}, exp: "LINE", tok: line_TOK},
		{inp: []byte{0xb1}, exp: "WHILE", tok: while_TOK},
		{inp: []byte{0xb2}, exp: "WEND", tok: wend_TOK},
		{inp: []byte{0xb3}, exp: "CALL", tok: call_TOK},
		{inp: []byte{0xb7}, exp: "WRITE", tok: write_TOK},
		{inp: []byte{0xb8}, exp: "OPTION", tok: option_TOK},
		{inp: []byte{0xb9}, exp: "RANDOMIZE", tok: randomize_TOK},
		{inp: []byte{0xba}, exp: "OPEN", tok: open_TOK},
		{inp: []byte{0xbb}, exp: "CLOSE", tok: close_TOK},
		{inp: []byte{0xbc}, exp: "LOAD", tok: load_TOK},
		{inp: []byte{0xbd}, exp: "MERGE", tok: merge_TOK},
		{inp: []byte{0xbe}, exp: "SAVE", tok: save_TOK},
		{inp: []byte{0xbf}, exp: "COLOR", tok: color_TOK},

		{inp: []byte{0xc0}, exp: "CLS", tok: cls_TOK},
		{inp: []byte{0xc1}, exp: "MOTOR", tok: motor_TOK},
		{inp: []byte{0xc2}, exp: "BSAVE", tok: bsave_TOK},
		{inp: []byte{0xc3}, exp: "BLOAD", tok: bload_TOK},
		{inp: []byte{0xc4}, exp: "SOUND", tok: sound_TOK},
		{inp: []byte{0xc5}, exp: "BEEP", tok: beep_TOK},
		{inp: []byte{0xc6}, exp: "PSET", tok: pset_TOK},
		{inp: []byte{0xc7}, exp: "PRESET", tok: preset_TOK},
		{inp: []byte{0xc8}, exp: "SCREEN", tok: screen_TOK},
		{inp: []byte{0xc9}, exp: "KEY", tok: key_TOK},
		{inp: []byte{0xca}, exp: "LOCATE", tok: locate_TOK},
		{inp: []byte{0xcc}, exp: "TO", tok: to_TOK},
		{inp: []byte{0xcd}, exp: "THEN", tok: then_TOK},
		{inp: []byte{0xce}, exp: "TAB(", tok: tab_TOK},
		{inp: []byte{0xcf}, exp: "STEP", tok: step_TOK},

		{inp: []byte{0xd0}, exp: "USR", tok: usr_TOK},
		{inp: []byte{0xd1}, exp: "FN", tok: fn_TOK},
		{inp: []byte{0xd2}, exp: "SPC(", tok: spc_TOK},
		{inp: []byte{0xd3}, exp: "NOT", tok: not_TOK},
		{inp: []byte{0xd4}, exp: "ERL", tok: erl_TOK},
		{inp: []byte{0xd5}, exp: "ERR", tok: err_TOK},
		{inp: []byte{0xd6}, exp: "STRING$", tok: strings_TOK},
		{inp: []byte{0xd7}, exp: "USING", tok: using_TOK},
		{inp: []byte{0xd8}, exp: "INSTR", tok: instr_TOK},
		{inp: []byte{0xd9}, exp: "'", tok: ticrem_TOK},
		{inp: []byte{0xda}, exp: "VARPTR", tok: varptr_TOK},
		{inp: []byte{0xdb}, exp: "CSRLIN", tok: csrlin_TOK},
		{inp: []byte{0xdc}, exp: "POINT", tok: point_TOK},
		{inp: []byte{0xdd}, exp: "OFF", tok: off_TOK},
		{inp: []byte{0xde}, exp: "INKEY$", tok: inkeys_TOK},

		{inp: []byte{0xe6}, exp: ">", tok: gt_TOK},
		{inp: []byte{0xe7}, exp: "=", tok: eq_TOK},
		{inp: []byte{0xe8}, exp: "<", tok: lt_TOK},
		{inp: []byte{0xe9}, exp: "+", tok: plus_TOK},
		{inp: []byte{0xea}, exp: "-", tok: minus_TOK},
		{inp: []byte{0xeb}, exp: "*", tok: mul_TOK},
		{inp: []byte{0xec}, exp: "/", tok: div_TOK},
		{inp: []byte{0xed}, exp: "^", tok: pwr_TOK},
		{inp: []byte{0xee}, exp: "AND", tok: and_TOK},
		{inp: []byte{0xef}, exp: "OR", tok: or_TOK},

		{inp: []byte{0xf0}, exp: "XOR", tok: xor_TOK},
		{inp: []byte{0xf1}, exp: "EQV", tok: eqv_TOK},
		{inp: []byte{0xf2}, exp: "IMP", tok: imp_TOK},
		{inp: []byte{0xf3}, exp: "MOD", tok: mod_TOK},
		{inp: []byte{0xf4}, exp: `\`, tok: bslash_TOK},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.inp[0], tt.tok, "main page alignment, %x not equal %x", tt.inp[0], tt.tok)
		rdr := progRdr{src: bufio.NewReader(bytes.NewReader(tt.inp))}
		res := rdr.readToken()
		assert.Equal(t, tt.exp, res, "Testing main page, expected %s got %s", tt.exp, res)
	}
}

func Test_FFPageTokens(t *testing.T) {
	tests := []struct {
		inp []byte
		exp string
		tok byte
	}{
		{inp: []byte{0x81}, exp: "LEFT$", tok: lefts_TOK},
		{inp: []byte{0x82}, exp: "RIGHT$", tok: rights_TOK},
		{inp: []byte{0x83}, exp: "MID$", tok: mids_TOK},
		{inp: []byte{0x84}, exp: "SGN", tok: sgn_TOK},
		{inp: []byte{0x85}, exp: "INT", tok: int_TOK},
		{inp: []byte{0x86}, exp: "ABS", tok: abs_TOK},
		{inp: []byte{0x87}, exp: "SQR", tok: sqr_TOK},
		{inp: []byte{0x88}, exp: "RND", tok: rnd_TOK},
		{inp: []byte{0x89}, exp: "SIN", tok: sin_TOK},
		{inp: []byte{0x8a}, exp: "LOG", tok: log_TOK},
		{inp: []byte{0x8b}, exp: "EXP", tok: exp_TOK},
		{inp: []byte{0x8c}, exp: "COS", tok: cos_TOK},
		{inp: []byte{0x8d}, exp: "TAN", tok: tan_TOK},
		{inp: []byte{0x8e}, exp: "ATN", tok: atn_TOK},
		{inp: []byte{0x8f}, exp: "FRE", tok: fre_TOK},
		{inp: []byte{0x90}, exp: "INP", tok: inp_TOK},
		{inp: []byte{0x91}, exp: "POS", tok: pos_TOK},
		{inp: []byte{0x92}, exp: "LEN", tok: len_TOK},
		{inp: []byte{0x93}, exp: "STR$", tok: strs_TOK},
		{inp: []byte{0x94}, exp: "VAL", tok: val_TOK},
		{inp: []byte{0x95}, exp: "ASC", tok: asc_TOK},
		{inp: []byte{0x96}, exp: "CHR$", tok: chrs_TOK},
		{inp: []byte{0x97}, exp: "PEEK", tok: peek_TOK},
		{inp: []byte{0x98}, exp: "SPACE$", tok: spaces_TOK},
		{inp: []byte{0x99}, exp: "OCT$", tok: octs_TOK},
		{inp: []byte{0x9a}, exp: "HEX$", tok: hexs_TOK},
		{inp: []byte{0x9b}, exp: "LPOS", tok: lpos_TOK},
		{inp: []byte{0x9c}, exp: "CINT", tok: cint_TOK},
		{inp: []byte{0x9d}, exp: "CSNG", tok: csng_TOK},
		{inp: []byte{0x9e}, exp: "CDBL", tok: cdbl_TOK},
		{inp: []byte{0x9f}, exp: "FIX", tok: fix_TOK},
		{inp: []byte{0xa0}, exp: "PEN", tok: pen_TOK},
		{inp: []byte{0xa1}, exp: "STICK", tok: stick_TOK},
		{inp: []byte{0xa2}, exp: "STRIG", tok: strig_TOK},
		{inp: []byte{0xa3}, exp: "EOF", tok: eof_TOK},
		{inp: []byte{0xa4}, exp: "LOC", tok: loc_TOK},
		{inp: []byte{0xa5}, exp: "LOF", tok: lof_TOK},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.inp[0], tt.tok, "FF page alignment, %x not equal %x", tt.inp[0], tt.tok)
		rdr := progRdr{src: bufio.NewReader(bytes.NewReader(tt.inp))}
		res := rdr.ffPage()

		assert.Equal(t, tt.exp, res, "Testing FF page, expected %x got %x", tt.exp, res)
	}
}

func Test_FEPageTokens(t *testing.T) {
	tests := []struct {
		inp []byte
		exp string
		tok byte
	}{
		{inp: []byte{0x81}, exp: "FILES", tok: files_TOK},
		{inp: []byte{0x82}, exp: "FIELD", tok: field_TOK},
		{inp: []byte{0x83}, exp: "SYSTEM", tok: system_TOK},
		{inp: []byte{0x84}, exp: "NAME", tok: name_TOK},
		{inp: []byte{0x85}, exp: "LSET", tok: lset_TOK},
		{inp: []byte{0x86}, exp: "RSET", tok: rset_TOK},
		{inp: []byte{0x87}, exp: "KILL", tok: kill_TOK},
		{inp: []byte{0x88}, exp: "PUT", tok: put_TOK},
		{inp: []byte{0x89}, exp: "GET", tok: get_TOK},
		{inp: []byte{0x8a}, exp: "RESET", tok: reset_TOK},
		{inp: []byte{0x8b}, exp: "COMMON", tok: common_TOK},
		{inp: []byte{0x8c}, exp: "CHAIN", tok: chain_TOK},
		{inp: []byte{0x8d}, exp: "DATE$", tok: dates_TOK},
		{inp: []byte{0x8e}, exp: "TIME$", tok: times_TOK},
		{inp: []byte{0x8f}, exp: "PAINT", tok: paint_TOK},
		{inp: []byte{0x90}, exp: "COM", tok: com_TOK},
		{inp: []byte{0x91}, exp: "CIRCLE", tok: circle_TOK},
		{inp: []byte{0x92}, exp: "DRAW", tok: draw_TOK},
		{inp: []byte{0x93}, exp: "PLAY", tok: play_TOK},
		{inp: []byte{0x94}, exp: "TIMER", tok: timer_TOK},
		{inp: []byte{0x95}, exp: "ERDEV", tok: erdev_TOK},
		{inp: []byte{0x96}, exp: "IOCTL", tok: ioctl_TOK},
		{inp: []byte{0x96, '$'}, exp: "IOCTL$", tok: ioctl_TOK},
		{inp: []byte{0x97}, exp: "CHDIR", tok: chdir_TOK},
		{inp: []byte{0x98}, exp: "MKDIR", tok: mkdir_TOK},
		{inp: []byte{0x99}, exp: "RMDIR", tok: rmdir_TOK},
		{inp: []byte{0x9a}, exp: "SHELL", tok: shell_TOK},
		{inp: []byte{0x9b}, exp: "ENVIRON", tok: environ_TOK},
		{inp: []byte{0x9c}, exp: "VIEW", tok: view_TOK},
		{inp: []byte{0x9d}, exp: "WINDOW", tok: window_TOK},
		{inp: []byte{0x9e}, exp: "PMAP", tok: pmap_TOK},
		{inp: []byte{0x9f}, exp: "PALETTE", tok: palette_TOK},
		{inp: []byte{0xa0}, exp: "LCOPY", tok: lcopy_TOK},
		{inp: []byte{0xa1}, exp: "CALLS", tok: calls_TOK},
		{inp: []byte{0xa5}, exp: "PCOPY", tok: pcopy_TOK},
		{inp: []byte{0xa7}, exp: "LOCK", tok: lock_TOK},
		{inp: []byte{0xa8}, exp: "UNLOCK", tok: unlock_TOK},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.inp[0], tt.tok, "FE page alignment, %x not equal %x", tt.inp[0], tt.tok)
		rdr := progRdr{src: bufio.NewReader(bytes.NewReader(tt.inp))}
		res := rdr.fePage()

		assert.Equal(t, tt.exp, res, "Testing FF page, expected %x got %x", tt.exp, res)
	}
}

func Test_FDPageTokens(t *testing.T) {
	tests := []struct {
		inp []byte
		exp string
		tok byte
	}{
		{inp: []byte{0x81}, exp: "CVI", tok: cvi_TOK},
		{inp: []byte{0x82}, exp: "CVS", tok: cvs_TOK},
		{inp: []byte{0x83}, exp: "CVD", tok: cvd_TOK},
		{inp: []byte{0x84}, exp: "MKI$", tok: mkis_TOK},
		{inp: []byte{0x85}, exp: "MKS$", tok: mkss_TOK},
		{inp: []byte{0x86}, exp: "MKD$", tok: mkds_TOK},
		{inp: []byte{0x8b}, exp: "EXTERR", tok: exterr_TOK},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.inp[0], tt.tok, "FD page alignment, %x not equal %x", tt.inp[0], tt.tok)
		rdr := progRdr{src: bufio.NewReader(bytes.NewReader(tt.inp))}
		res := rdr.fdPage()

		assert.Equal(t, tt.exp, res, "Testing FF page, expected %x got %x", tt.exp, res)
	}
}

func Test_Read4ByteFloat(t *testing.T) {
	tests := []struct {
		inp []byte
		exp string
	}{
		{inp: []byte{0x00, 0x00}, exp: "0"},
		{inp: []byte{0x00, 0x00, 0x00, 0x00}, exp: "0"},
		{inp: []byte{0x09, 0xF6, 0x45, 0x71}, exp: "2.359880E-05"},
		{inp: []byte{0x40, 0xF6, 0x45, 0x71}, exp: "2.359890E-05"},
		{inp: []byte{0x2F, 0xFD, 0x6B, 0x88}, exp: "2.359890E+02"},
	}

	for _, tt := range tests {
		rdr := progRdr{src: bufio.NewReader(bytes.NewReader(tt.inp))}
		res := rdr.read4ByteFloat()
		assert.Equal(t, tt.exp, res, "got %s expected %s", res, tt.exp)
	}
}

func Test_Read8ByteFloat(t *testing.T) {
	tests := []struct {
		inp []byte
		exp string
	}{
		{inp: []byte{0x00, 0x00}, exp: "0"},
		{inp: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, exp: "0"},
		{inp: []byte{0xB1, 0xAE, 0x1C, 0x84, 0x8C, 0xE0, 0x12, 0x6D}, exp: "1.094320E-06"},
		{inp: []byte{0x2B, 0xD4, 0xF2, 0x79, 0x40, 0xF6, 0x45, 0x71}, exp: "2.359890E-05"},
		{inp: []byte{0x77, 0xBE, 0x9F, 0x1A, 0x2F, 0xFD, 0x6B, 0x88}, exp: "2.359890E+02"},
	}

	for _, tt := range tests {
		rdr := progRdr{src: bufio.NewReader(bytes.NewReader(tt.inp))}
		res := rdr.read8ByteFloat()
		assert.Equal(t, tt.exp, res, "got %s expected %s", res, tt.exp)
	}
}
