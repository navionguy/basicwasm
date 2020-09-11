package evaluator

import (
	"github.com/navionguy/basicwasm/object"
)

var builtins = map[string]*object.Builtin{
	"ABS": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, "Syntax error")
			}

			switch arg := args[0].(type) {
			case *object.Integer:
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
				return newError(env, "Syntax error")
			}

			if len(arg.Value) == 0 {
				return newError(env, "Illegal Function Call")
			}

			b := []byte(arg.Value)

			return &object.Integer{Value: int16(b[0])}
		},
	},
	"LEN": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, "Syntax error")
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
