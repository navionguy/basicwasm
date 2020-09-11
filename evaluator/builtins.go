package evaluator

import (
	"math"

	"github.com/navionguy/basicwasm/object"
)

const SYNTAX_ERR = "Syntax error"
const TYPEMIS_ERR = "Type mismatch"

var builtins = map[string]*object.Builtin{
	"ABS": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, SYNTAX_ERR)
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
			default:
				return arg
			}
		},
	},
	"ASC": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			arg, ok := args[0].(*object.String)

			if (len(args) != 1) || !ok {
				return newError(env, SYNTAX_ERR)
			}

			if len(arg.Value) == 0 {
				return newError(env, "Illegal Function Call")
			}

			b := []byte(arg.Value)

			return &object.Integer{Value: int16(b[0])}
		},
	},
	"ATN": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, SYNTAX_ERR)
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
			default:
				return newError(env, TYPEMIS_ERR)
			}
		},
	},
	"CDBL": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, SYNTAX_ERR)
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
			default:
				return newError(env, TYPEMIS_ERR)
			}
		},
	}, "CINT": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, SYNTAX_ERR)
			}

			var rc int64
			switch arg := args[0].(type) {
			case *object.Integer:
				return arg
			case *object.IntDbl:
				rc = int64(arg.Value)
				break
			case *object.Fixed:
				dc := arg.Value.Round(0)
				rc = dc.IntPart()
				break

			case *object.FloatSgl:
				rc = int64(math.Round(float64(arg.Value)))
				break

			case *object.FloatDbl:
				rc = int64(math.Round(float64(arg.Value)))
				break

			default:
				return newError(env, TYPEMIS_ERR)
			}

			if (rc < -32768) || rc > 32767 {
				return newError(env, "Overflow")
			}

			return &object.Integer{Value: int16(rc)}
		},
	},
	"LEN": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, SYNTAX_ERR)
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int16(len(arg.Value))}
			default:
				return newError(env, "Type mismatch")
			}
		},
	},
}
