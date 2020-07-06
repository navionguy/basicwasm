package evaluator

import (
	"github.com/navionguy/basicwasm/object"
)

var builtins = map[string]*object.Builtin{
	"LEN": &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError(env, "wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int16(len(arg.Value))}
			default:
				return newError(env, "argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
}
