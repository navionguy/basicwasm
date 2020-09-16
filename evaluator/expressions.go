package evaluator

import (
	"fmt"

	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/object"
)

type expression func(string, object.Object, object.Object, *object.Environment) object.Object

var typeConverters = map[string]expression{
	// the ones that don't actually need conversion
	object.INTEGER_OBJ + object.INTEGER_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalIntegerInfixExpression(operator, int(left.(*object.Integer).Value), int(right.(*object.Integer).Value), env)
	},

	object.STRING_OBJ + object.STRING_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalStringInfixExpression(operator, left, right, env)
	},

	object.FIXED_OBJ + object.FIXED_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFixedInfixExpression(operator, left.(*object.Fixed).Value, right.(*object.Fixed).Value, env)
	},

	object.FLOATSGL_OBJ + object.FLOATSGL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatInfixExpression(operator, left.(*object.FloatSgl).Value, right.(*object.FloatSgl).Value, env)
	},

	object.FLOATDBL_OBJ + object.FLOATDBL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatDblInfixExpression(operator, left.(*object.FloatDbl).Value, right.(*object.FloatDbl).Value, env)
	},

	// Now we start the one's that require type conversion
	// Fixed point integers, faster than Float, more precise than integers
	object.INTEGER_OBJ + object.FIXED_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		dleft := decimal.NewFromInt32(int32(left.(*object.Integer).Value))
		return evalFixedInfixExpression(operator, dleft, right.(*object.Fixed).Value, env)
	},

	object.FIXED_OBJ + object.INTEGER_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		dright := decimal.NewFromInt32(int32(right.(*object.Integer).Value))
		return evalFixedInfixExpression(operator, left.(*object.Fixed).Value, dright, env)
	},

	// Floats, more precise? than Fixed?

	object.FLOATSGL_OBJ + object.INTEGER_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatInfixExpression(operator, left.(*object.FloatSgl).Value, float32(right.(*object.Integer).Value), env)
	},

	object.FLOATSGL_OBJ + object.FIXED_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		f, _ := right.(*object.Fixed).Value.Float64()
		return evalFloatInfixExpression(operator, left.(*object.FloatSgl).Value, float32(f), env)
	},

	object.INTEGER_OBJ + object.FLOATSGL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatInfixExpression(operator, float32(left.(*object.Integer).Value), right.(*object.FloatSgl).Value, env)
	},

	object.FIXED_OBJ + object.FLOATSGL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		f, _ := left.(*object.Fixed).Value.Float64()
		return evalFloatInfixExpression(operator, float32(f), right.(*object.FloatSgl).Value, env)
	},

	// FloatDbl, even slower and more precise

	object.FLOATDBL_OBJ + object.INTEGER_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatDblInfixExpression(operator, left.(*object.FloatDbl).Value, float64(right.(*object.Integer).Value), env)
	},

	object.FLOATDBL_OBJ + object.FIXED_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		f, _ := right.(*object.Fixed).Value.Float64()
		return evalFloatDblInfixExpression(operator, left.(*object.FloatDbl).Value, f, env)
	},

	object.FLOATDBL_OBJ + object.FLOATSGL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatDblInfixExpression(operator, left.(*object.FloatDbl).Value, float64(right.(*object.FloatSgl).Value), env)
	},

	object.INTEGER_OBJ + object.FLOATDBL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatDblInfixExpression(operator, float64(left.(*object.Integer).Value), right.(*object.FloatDbl).Value, env)
	},

	object.FIXED_OBJ + object.FLOATDBL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		f, _ := left.(*object.Fixed).Value.Float64()
		return evalFloatDblInfixExpression(operator, f, right.(*object.FloatDbl).Value, env)
	},

	object.FLOATSGL_OBJ + object.FLOATDBL_OBJ: func(operator string, left, right object.Object, env *object.Environment) object.Object {
		return evalFloatDblInfixExpression(operator, float64(left.(*object.FloatSgl).Value), right.(*object.FloatDbl).Value, env)
	},

	/*

		object. + object.: func(operator string, left, right object.Object, env *object.Environment) object.Object {
			return evalFixedInfixExpression(operator, left, right, env)
		},

	*/
}

func fixType(x interface{}) object.Object {
	i16, ok := x.(int16)

	if ok {
		return &object.Integer{Value: i16}
	}

	i32, ok := x.(int32)

	if ok {
		return &object.IntDbl{Value: i32}
	}

	i, ok := x.(int)

	if ok {
		if i == int(int16(i)) {
			return &object.Integer{Value: int16(i)}
		}

		// too big, convert to IntDbl
		if i == int(int32(i)) {
			return &object.IntDbl{Value: int32(i)}
		}

		fx := tryFixed(float64(i))

		if fx != nil {
			return fx
		}

		// int larger than 32 bits, make him a float
		return &object.FloatDbl{Value: float64(i)}
	}

	f, ok := x.(float64)

	if ok {
		i := int16(f)
		if f == float64(i) {
			return &object.Integer{Value: i}
		}

		fxd := tryFixed(f)

		if fxd != nil {
			return fxd
		}

		if f == float64(float32(f)) {
			return &object.FloatSgl{Value: float32(f)}
		}
		return &object.FloatDbl{Value: f}
	}

	fs, ok := x.(float32)

	if ok {
		return &object.FloatSgl{Value: fs}
	}

	return &object.Error{Message: "unknown type"}
}

func tryFixed(val float64) object.Object {
	dec, err := decimal.NewFromString(fmt.Sprintf("%.8f", val))

	if err != nil {
		// can't convert him, give up
		return nil
	}

	fxd := &object.Fixed{Value: dec.Round(-7)}

	return fxd
}
