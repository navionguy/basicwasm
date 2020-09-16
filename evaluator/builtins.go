package evaluator

import (
	"encoding/binary"
	"math"

	"github.com/navionguy/basicwasm/object"
)

const syntaxErr = "Syntax error"
const typeMismatchErr = "Type mismatch"
const overflowErr = "Overflow"
const illegalFuncCallErr = "Illegal function call"

var builtins = map[string]*object.Builtin{
	"ABS": {
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
	"ASC": {
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
	"ATN": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return fixType(math.Atan(float64(arg.Value)))

			case *object.IntDbl:
				return fixType(math.Atan(float64(arg.Value)))

			case *object.Fixed:
				val, _ := arg.Value.Float64()
				return fixType(math.Atan(val))

			case *object.FloatSgl:
				return fixType(math.Atan(float64(arg.Value)))

			case *object.FloatDbl:
				return fixType(math.Atan(arg.Value))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CDBL": {
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
	"CHR$": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			var rc int64
			switch arg := args[0].(type) {
			case *object.Integer:
				rc = int64(arg.Value)

			case *object.IntDbl:
				rc = int64(arg.Value)

			case *object.Fixed:
				dc := arg.Value.Round(0)
				rc = dc.IntPart()

			case *object.FloatSgl:
				rc = int64(math.Round(float64(arg.Value)))

			case *object.FloatDbl:
				rc = int64(math.Round(float64(arg.Value)))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}

			if (rc < 0) || (rc > 255) {
				return newError(env, illegalFuncCallErr)
			}
			return &object.String{Value: string(rc)}
		},
	},
	"CINT": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			var rc int64
			switch arg := args[0].(type) {
			case *object.Integer:
				return newError(env, syntaxErr)

			case *object.IntDbl:
				rc = int64(arg.Value)

			case *object.Fixed:
				dc := arg.Value.Round(0)
				rc = dc.IntPart()

			case *object.FloatSgl:
				rc = int64(math.Round(float64(arg.Value)))

			case *object.FloatDbl:
				rc = int64(math.Round(float64(arg.Value)))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}

			if (rc < -32768) || rc > 32767 {
				return newError(env, overflowErr)
			}

			return &object.Integer{Value: int16(rc)}
		},
	},
	"COS": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.Integer:
				return fixType(math.Cos(float64(arg.Value)))

			case *object.IntDbl:
				return fixType(math.Cos(float64(arg.Value)))

			case *object.Fixed:
				dc, _ := arg.Value.Float64()
				return fixType(math.Cos(dc))

			case *object.FloatSgl:
				return fixType(math.Cos(float64(arg.Value)))

			case *object.FloatDbl:
				return fixType(math.Cos(arg.Value))

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
				return fixType(float32(arg.Value))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CVD": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) < 8 {
					return newError(env, illegalFuncCallErr)
				}

				num := []byte(arg.Value)
				cv := int(binary.LittleEndian.Uint64(num[:8]))
				return fixType(cv)

			case *object.BStr:
				return fixType(int(binary.LittleEndian.Uint64(arg.Value[:8])))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CVI": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) > 2 {
					return newError(env, illegalFuncCallErr)
				}

				num := []byte(arg.Value)
				cv := int16(binary.LittleEndian.Uint16(num[:2]))
				return fixType(cv)

			case *object.BStr:
				return fixType(int16(binary.LittleEndian.Uint16(arg.Value[:2])))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"CVS": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.String:
				if len(arg.Value) > 4 {
					return newError(env, illegalFuncCallErr)
				}

				num := []byte(arg.Value)
				cv := int32(binary.LittleEndian.Uint32(num[:4]))
				return fixType(cv)

			case *object.BStr:
				return fixType(int32(binary.LittleEndian.Uint32(arg.Value[:4])))

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"EXP": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			var exp float64
			switch arg := args[0].(type) {
			case *object.Integer:
				exp = float64(arg.Value)

			case *object.IntDbl:
				exp = float64(arg.Value)

			case *object.Fixed:
				exp, _ = arg.Value.Float64()

			case *object.FloatSgl:
				exp = float64(arg.Value)

			case *object.FloatDbl:
				exp = float64(arg.Value)

			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)

			default:
				return newError(env, typeMismatchErr)
			}

			if (exp < -89) || (exp > 88.029689) {
				return newError(env, overflowErr)
			}

			return &object.FloatSgl{Value: float32(math.Exp(exp))}
		},
	},
	"LEN": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int16(len(arg.Value))}
			case *object.TypedVar:
				return fn.Fn(env, fn, arg.Value)
			default:
				return newError(env, typeMismatchErr)
			}
		},
	},
	"MKD$": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			return bstrEncode(8, env, args[0])
		},
	},
	"MKI$": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			return bstrEncode(2, env, args[0])
		},
	},
	"MKS$": {
		Fn: func(env *object.Environment, fn *object.Builtin, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, syntaxErr)
			}

			return bstrEncode(4, env, args[0])
		},
	},
}

// Some common functionality

func bstrEncode(size int, env *object.Environment, arg object.Object) object.Object {
	// calculate max/min values

	var max int64
	max = (1 << ((int64(size) * 8) - 1)) - 1
	min := -(max + 1)

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

	if (rc < min) || rc > max {
		return newError(env, overflowErr)
	}

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
