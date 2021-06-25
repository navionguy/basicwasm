package builtins

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"

	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/token"
)

const syntaxErr = "Syntax error"
const typeMismatchErr = "Type mismatch"
const overflowErr = "Overflow"
const illegalFuncCallErr = "Illegal function call"
const illegalArgErr = "Illegal argument"
const outOfDataErr = "Out of data"
const unDefinedLineNumberErr = "Undefined line number"

var Builtins = map[string]*object.Builtin{
	"ABS": { // absolute value
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				if arg.Value < 0 {
					arg.Value = -arg.Value
				}
				return arg
			case *object.IntDbl:
				if arg.Value < 0 {
					arg.Value = -arg.Value
				}
				return arg
			case *object.Fixed:
				arg.Value = arg.Value.Abs()
				return arg
			case *object.FloatSgl:
				if arg.Value < 0 {
					arg.Value = -arg.Value
				}
				return arg
			case *object.FloatDbl:
				if arg.Value < 0 {
					arg.Value = -arg.Value
				}
				return arg
			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, syntaxErr)
			}
		},
	},
	"ASC": { // ASCII code for first char in string
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) == 0 {
					return newError(env, illegalFuncCallErr)
				}

				b := []byte(arg.Value)

				return &object.Integer{Value: int16(b[0])}

			case *object.BStr:
				return &object.Integer{Value: int16(arg.Value[len(arg.Value)-1])}

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, syntaxErr)
			}
		},
	},
	"ATN": { // Arctangent of value
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return FixType(env, math.Atan(float64(arg.Value)))

			case *object.IntDbl:
				return FixType(env, math.Atan(float64(arg.Value)))

			case *object.Fixed:
				val, _ := arg.Value.Float64()
				return FixType(env, math.Atan(val))

			case *object.FloatSgl:
				return FixType(env, math.Atan(float64(arg.Value)))

			case *object.FloatDbl:
				return FixType(env, math.Atan(arg.Value))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CDBL": { // convert value to double precision
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return &object.IntDbl{Value: int32(arg.Value)}

			case *object.IntDbl:
				return arg

			case *object.Fixed:
				val, ok := arg.Value.Float64()
				if ok {

				}
				return &object.FloatDbl{Value: val}

			case *object.FloatSgl:
				return &object.FloatDbl{Value: float64(arg.Value)}

			case *object.FloatDbl:
				return arg

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CHR$": { // return character at codepoint args[0].Value
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			flt, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			rc := int64(math.Round(flt))

			if (rc < 0) || (rc > 255) {
				return newError(env, illegalFuncCallErr)
			}
			return &object.String{Value: fmt.Sprintf("%c", rc)}
		},
	},
	"CINT": { // convert numeric to integer with rounding, as opposed to FIX()
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			rc, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			if (rc < -32768) || rc > 32767 {
				return newError(env, overflowErr)
			}

			return &object.Integer{Value: int16(math.Round(rc))}
		},
	},
	"COS": { // return the cosine of the arguement
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return FixType(env, math.Cos(float64(arg.Value)))

			case *object.IntDbl:
				return FixType(env, math.Cos(float64(arg.Value)))

			case *object.Fixed:
				dc, _ := arg.Value.Float64()
				return FixType(env, math.Cos(dc))

			case *object.FloatSgl:
				return FixType(env, math.Cos(float64(arg.Value)))

			case *object.FloatDbl:
				return FixType(env, math.Cos(arg.Value))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CSNG": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return arg

			case *object.IntDbl:
				return arg

			case *object.Fixed:
				return arg

			case *object.FloatSgl:
				return arg

			case *object.FloatDbl:
				return FixType(env, float32(arg.Value))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CVD": { // convert string to double precision float
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			str, ok, _ := extractString(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			return FixType(env, int(binary.LittleEndian.Uint64(str[:8])))
		},
	},
	"CVI": { // convert string to integer
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			num, ok, _ := extractString(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			return FixType(env, int16(binary.LittleEndian.Uint16(num[:2])))
		},
	},
	"CVS": { // convert string to single precision float
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			str, ok, _ := extractString(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			return FixType(env, int32(binary.LittleEndian.Uint32(str[:4])))
		},
	},
	"EXP": { // e^^x
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			exp, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			if (exp < -89) || (exp > 88.029689) {
				return newError(env, overflowErr)
			}

			return &object.FloatSgl{Value: float32(math.Exp(exp))}
		},
	},
	"FIX": { // truncate a value, no rounding
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			rc, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			rc = math.Trunc(rc)

			if (rc > math.MinInt16) && (rc < math.MaxInt16) {
				return &object.Integer{Value: int16(rc)}
			}

			if (rc > math.MinInt32) && (rc < math.MaxInt32) {
				return &object.IntDbl{Value: int32(rc)}
			}

			return newError(env, overflowErr)
		},
	},
	"HEX$": { // Convert value to hexidecimal, range -32768 to +65535
		// interesting that covers uint16 and int16
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			rc, ok := extractNumeric((args[0]))

			if !ok {
				return newError(env, typeMismatchErr)
			}

			rc = math.Round(rc)

			if (rc > math.MinInt16) && (rc < math.MaxUint16) {
				return &object.String{Value: fmt.Sprintf("%X", uint16(rc))}
			}

			return newError(env, overflowErr)
		},
	},
	"INPUT$": { // read keystrokes from the keyboard
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 { // TODO: bump if adding file support
				return newError(env, syntaxErr)
			}

			rc, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			rc = math.Round(rc)

			if (rc < 1) || (rc > math.MaxInt8) {
				return newError(env, illegalFuncCallErr)
			}

			bt := env.Terminal().ReadKeys(int(rc))

			st := &object.String{Value: string(bt)}
			tv := &object.TypedVar{Value: st, TypeID: "$"}

			return tv
		},
	},
	"INSTR": { // search for a string inside of another string
		// I actually search them as BStr to be more accepting
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if (len(args) < 2) || (len(args) > 3) {
				return newError(env, syntaxErr)
			}

			a := 0
			strt := 1

			// if there are 3 params, the first is where to start looking
			if len(args) == 3 {
				f, _ := extractNumeric(args[a])
				strt = int(f)
				a++
			}

			// you can't start before the first character
			if strt < 1 {
				return newError(env, illegalArgErr)
			}

			// max start pos is 255
			if strt > 255 {
				return newError(env, illegalFuncCallErr)
			}

			bt, ok, _ := extractString(args[a])
			a++

			sub, ok2, _ := extractString(args[a])

			// if parms 2 & 3 aren't strings of some type, syntax error
			if !ok || !ok2 {
				return newError(env, syntaxErr)
			}

			// if start > length of the string
			// or the substring is longer than the string
			// you get zero
			if (strt > len(bt)) || (len(sub) > len(bt)) {
				return &object.Integer{Value: 0}
			}

			subSize := len(sub)
			for i := strt - 1; i < len(bt); i++ {
				if 0 == bytes.Compare(bt[i:subSize+i], sub) {
					// found him
					return &object.Integer{Value: int16(i + 1)}
				}

				if len(sub) >= len(bt[i:]) {
					i = len(bt)
				}
			}

			return &object.Integer{Value: 0}
		},
	},
	"INT": { // truncate an expression to a whole number
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			rc, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			rc = math.Trunc(rc)

			if (rc > math.MinInt16) && (rc < math.MaxInt16) {
				return &object.Integer{Value: int16(rc)}
			}

			if (rc > math.MinInt32) && (rc < math.MaxInt32) {
				return &object.IntDbl{Value: int32(rc)}
			}

			return newError(env, overflowErr)
		},
	},
	"LEFT$": { // return the left most n characters of x$
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError(env, syntaxErr)
			}

			bstr, ok, str := extractString(args[0])

			fc, ok2 := extractNumeric(args[1])

			if !ok || !ok2 {
				return newError(env, syntaxErr)
			}

			if (fc < 0) || fc > 255 {
				return newError(env, illegalFuncCallErr)
			}

			if str {
				return &object.String{Value: string(bstr[:int16(fc)])}
			}

			return &object.BStr{Value: bstr[:int16(fc)]}
		},
	},
	"LEN": { // return the length of a string
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			bstr, ok, _ := extractString(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			return &object.Integer{Value: int16(len(bstr))}
		},
	},
	"LOG": { // return the natural log of a number
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			x, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			if x <= 0 {
				return newError(env, illegalFuncCallErr)
			}

			return &object.FloatSgl{Value: float32(math.Log(x))}
		},
	},
	"LPOS": { // return printer head position TODO: Implement printing
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			return &object.Integer{Value: 0}
		},
	},
	"MID$": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if (len(args) < 2) || (len(args) > 3) {
				return newError(env, syntaxErr)
			}

			src, ok, isString := extractString(args[0])
			floc, ok2 := extractNumeric(args[1])
			fct := float64(0)
			ok3 := true

			if len(args) == 3 {
				fct, ok3 = extractNumeric(args[2])
			}

			if !ok || !ok2 || !ok3 {
				return newError(env, syntaxErr)
			}

			ct := int(fct) // length of string to return
			loc := int(floc)

			if (loc < 1) || (loc > 255) || (ct < 0) || (ct > 255) {
				return newError(env, illegalFuncCallErr)
			}

			bt := src[loc-1:]

			if ct != 0 {
				bt = bt[:ct]
			}

			if isString {
				return &object.String{Value: string(bt)}
			}

			return &object.BStr{Value: bt}
		},
	},
	"MKD$": { // convert a numeric to a 8 byte BStr
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			return bstrEncode(8, env, args[0])
		},
	},
	"MKI$": { // convert a numeric to a 2 byte BStr
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			return bstrEncode(2, env, args[0])
		},
	},
	"MKS$": { // convert a numeric to a 4 byte string
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			return bstrEncode(4, env, args[0])
		},
	},
	"OCT$": { // convert a numberic to
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			rc, ok := extractNumeric((args[0]))

			if !ok {
				return newError(env, typeMismatchErr)
			}

			rc = math.Round(rc)

			if (rc > math.MinInt16) && (rc < math.MaxUint16) {
				return &object.String{Value: fmt.Sprintf("%o", uint16(rc))}
			}

			return newError(env, overflowErr)
		},
	},
	"RIGHT$": { // return the rightmost n characters of the string
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError(env, syntaxErr)
			}

			src, ok, isString := extractString(args[0])
			floc, ok2 := extractNumeric(args[1])

			if !ok || !ok2 {
				return newError(env, syntaxErr)
			}

			loc := int(floc)

			if (loc < 0) || (loc > 255) {
				return newError(env, illegalFuncCallErr)
			}

			if loc > len(src) {
				return args[0]
			}

			if loc == 0 {
				return &object.String{Value: ""}
			}

			bt := src[len(src)-loc:]

			if isString {
				return &object.String{Value: string(bt)}
			}

			return &object.BStr{Value: bt}
		},
	},
	"RND": { // generate a random number
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) > 1 {
				return newError(env, syntaxErr)
			}

			x, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			return env.Random(int(x))
		},
	},
	"SCREEN": { // read the ascii value at a position on the screen
		// ToDo: add support for screen color
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError(env, illegalFuncCallErr)
			}

			rowarg, ok := extractNumeric(args[0])
			colarg, ok2 := extractNumeric(args[1])

			if !ok || !ok2 {
				return newError(env, typeMismatchErr)
			}

			row := int(rowarg)
			col := int(colarg)

			bt := []byte(env.Terminal().Read(col, row, 1))

			return &object.Integer{Value: int16(bt[0])}
		},
	},
	"SGN": { // return the sign of the argument -1 = neg, 0 = zero, 1 = pos
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			arg, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			if arg < 0 {
				return &object.Integer{Value: -1}
			}

			if arg == 0 {
				return &object.Integer{Value: 0}
			}

			return &object.Integer{Value: 1}
		},
	},
	"SIN": { // calculate sine of arg in radians
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			arg, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			return &object.FloatSgl{Value: float32(math.Sin(arg))}
		},
	},
	"SPACE$": { // return number of spaces == round(arg[0])
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			arg, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			val := int(math.Round(arg))

			if (val < 0) || (val > 255) {
				return newError(env, illegalFuncCallErr)
			}

			sp := ""
			for i := 0; i < val; i++ {
				sp += " "
			}

			return &object.String{Value: sp}
		},
	},
	"SQR": { // calculate square root of argument
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			arg, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			val := float32(math.Sqrt(arg))

			return &object.FloatSgl{Value: val}
		},
	},
	"STR$": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			arg, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			st := fmt.Sprint(arg)

			return &object.String{Value: st}
		},
	},
	"STRING$": { // (x, y) build a string of length x consisting of character y repeated
		// if y is string, repeat the first character
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError(env, syntaxErr)
			}

			arg, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			if arg < 0 || arg > 255 {
				return newError(env, illegalFuncCallErr)
			}

			arg2, ok := extractNumeric(args[1])

			var bt []byte
			if ok {
				if arg2 < 0 || arg2 > 255 {
					return newError(env, illegalFuncCallErr)
				}

				bt = append(bt, byte(arg2))
			} else {
				bt, _, _ = extractString(args[1])
			}

			bt = bt[0:1]
			for i := 0; i < int(math.Round(arg))-1; i++ {
				bt = append(bt, bt[0])
			}

			if bt[0] == 0 {
				return &object.BStr{Value: bt}
			}
			return &object.String{Value: string(bt)}
		},
	},
	"TAN": { // compute the tangent of x in radians
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			arg, ok := extractNumeric(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}

			return &object.FloatSgl{Value: float32(math.Tan(arg))}
		},
	},
	"VAL": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			arg, ok, _ := extractString(args[0])

			if !ok {
				return newError(env, typeMismatchErr)
			}
			val, err := strconv.ParseFloat(string(arg), 32)

			if err != nil {
				val = 0
			}

			return &object.FloatSgl{Value: float32(val)}
		},
	},
}

// Some common functionality

// MKD$, MKI$, and MKS$ all return values as a Bstr
func bstrEncode(size int, env *object.Environment, arg object.Object) object.Object {
	var rc int64
	switch ar := arg.(type) {
	case *object.Integer:
		rc = int64(ar.Value)

	case *object.IntDbl:
		rc = int64(ar.Value)

	case *object.Fixed:
		dc := ar.Value.Round(0)
		rc = dc.IntPart()

	case *object.FloatSgl:
		rc = int64(math.Round(float64(ar.Value)))

	case *object.FloatDbl:
		rc = int64(math.Round(float64(ar.Value)))

	default:
		return newError(env, typeMismatchErr)
	}

	// calculate max/min values
	max := int64(1<<((int64(size)*8)-1) - 1)
	min := -(max + 1)

	if (rc < min) || rc > max {
		return newError(env, overflowErr)
	}

	return buildBstr(size, rc)
}

// now that I have created the integer part
// use the binary package to serialize rc
// as a byte series, little Endian
func buildBstr(size int, rc int64) object.Object {
	bt := make([]byte, size)

	switch size {
	case 2:
		binary.LittleEndian.PutUint16(bt, uint16(rc))
	case 4:
		binary.LittleEndian.PutUint32(bt, uint32(rc))
	case 8:
		binary.LittleEndian.PutUint64(bt, uint64(rc))
	}
	return &object.BStr{Value: bt}

}

// given any of the numeric values, return a float64 representation
// bool = false means non-numeric
func extractNumeric(obj object.Object) (float64, bool) {
	var rc float64

	switch arg := obj.(type) {
	case *object.Integer:
		return float64(arg.Value), true

	case *object.IntDbl:
		return float64(arg.Value), true

	case *object.Fixed:
		rc, _ = arg.Value.Float64()
		return rc, true

	case *object.FloatSgl:
		return float64(arg.Value), true

	case *object.FloatDbl:
		return float64(arg.Value), true

	case *object.TypedVar:
		f, ok := extractNumeric(arg.Value)

		if !ok {
			return 0, false
		}

		return f, true

	default:
		return 0, false
	}
}

// returns []bytes, extractedYN, isString
func extractString(obj object.Object) ([]byte, bool, bool) {

	switch arg := obj.(type) {
	case *object.String:
		bt := []byte(arg.Value)
		return bt, true, true

	case *object.BStr:
		return arg.Value, true, false

	case *object.TypedVar:
		bt, ok, str := extractString(arg.Value)
		return bt, ok, str

	default:
		return nil, false, false
	}
}

func newError(env *object.Environment, format string, a ...interface{}) *object.Error {
	msg := fmt.Sprintf(format, a...)
	tk, ok := env.Get(token.LINENUM)

	if ok {
		msg += fmt.Sprintf(" in %d", tk.(*object.IntDbl).Value)
	}

	return &object.Error{Message: msg}
}

func FixType(env *object.Environment, val interface{}) object.Object {
	// check the integer types
	res := tryInteger(val)

	if res != nil {
		return res
	}

	res = tryFloat(val)

	if res != nil {
		return res
	}

	return newError(env, typeMismatchErr)
}

// see if value is an integer type
func tryInteger(val interface{}) object.Object {
	// is he a 16bit integer
	i16, ok := val.(int16)
	if ok {
		return &object.Integer{Value: i16}
	}

	return tryInt32(val)
}

func tryInt32(val interface{}) object.Object {
	//is he a 32bit integer
	i32, ok := val.(int32)
	if ok {
		return shrinkI32(i32)
	}

	return tryInt64(val)
}

func tryInt64(val interface{}) object.Object {
	i, ok := val.(int)

	if ok {
		return shrinkI64(i)
	}

	return nil
}

// will 32bit integer actually fit in int16
func shrinkI32(i32 int32) object.Object {

	i16 := int16(i32)
	if int32(i16) == i32 {
		// yes, return the smaller object
		return &object.Integer{Value: i16}
	}

	return &object.IntDbl{Value: i32}
}

func shrinkI64(i64 int) object.Object {
	i32 := int32(i64)
	if int(i32) == i64 {
		return shrinkI32(i32)
	}
	// have to return him as a 64 bit float

	return &object.FloatDbl{Value: float64(i64)}
}

func tryFloat(val interface{}) object.Object {
	f32, ok := val.(float32)

	if ok {
		return shrinkF32(f32)
	}

	return tryFloat64(val)
}

func tryFloat64(val interface{}) object.Object {
	f64, ok := val.(float64)

	if ok {
		return shrinkF64(f64)
	}

	return nil
}

func shrinkF64(f64 float64) object.Object {
	f32 := float32(f64)
	if float64(f32) == f64 {
		return shrinkF32(float32(f64))
	}

	i64 := int(f64)
	if float64(i64) == f64 {
		return shrinkI64(i64)
	}

	return &object.FloatDbl{Value: f64}
}

func shrinkF32(f32 float32) object.Object {
	i32 := int32(f32)
	if float32(i32) == f32 {
		return shrinkI32(i32)
	}

	return &object.FloatSgl{Value: f32}
}
