package evaluator

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/token"
)

const (
	ERROR   = object.ERROR_OBJ
	INTEGER = object.INTEGER_OBJ
)

// Eval returns the object at a node
func Eval(node ast.Node, code *ast.Code, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(node, code, env)

	case *ast.AutoCommand:
		evalAutoCommand(node, code, env)

	case *ast.ClsStatement:
		evalClsStatement(node, code, env)

	case *ast.DimStatement:
		evalDimStatement(node, code, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, code, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, code, env)

	case *ast.LetStatement:
		val := Eval(node.Value, code, env)
		if isError(val) {
			return val
		}
		// life gets more complicated, not less
		if !strings.ContainsAny(node.Name.String(), "[$%!#") {
			env.Set(node.Name.String(), val)
			break
		}
		return saveVariable(code, env, node.Name, val)

	case *ast.LineNumStmt:
		ln := &object.IntDbl{Value: node.Value}
		env.Set(token.LINENUM, ln)
		if env.GetTrace() {
			env.Terminal().Print(fmt.Sprintf("[%d]", node.Value))
		}
		return ln

	case *ast.ListStatement:
		evalListStatement(code, node, env)
		return nil

	case *ast.GotoStatement:
		return evalGotoStatement(strings.Trim(node.Goto, " "), code, env)

	case *ast.EndStatement:
		code.Jump(code.Len())
		return &object.Integer{Value: 0}

	case *ast.PrintStatement:
		evalPrintStatement(node, code, env)

		// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.FixedLiteral:
		val, err := decimal.NewFromString(node.Value.String())

		if err != nil {
			return newError(env, err.Error())
		}
		return &object.Fixed{Value: val}

	case *ast.FloatSingleLiteral:
		return &object.FloatSgl{Value: node.Value}

	case *ast.FLoatDoubleLiteral:
		return &object.FloatDbl{Value: node.Value}

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.Identifier:
		return evalIdentifier(node, code, env)

	case *ast.PrefixExpression:
		right := Eval(node.Right, code, env)
		return evalPrefixExpression(node.Operator, right, code, env)

	case *ast.InfixExpression:
		left := Eval(node.Left, code, env)
		right := Eval(node.Right, code, env)
		return evalInfixExpression(node.Operator, left, right, env)

	case *ast.IfExpression:
		return evalIfExpression(node, code, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		obj := &object.Function{Parameters: params, Env: env, Body: body}
		env.Set(node.Token.Literal, obj)
		return obj

	case *ast.RemStatement:
		return nil // nothing to be done

	case *ast.RunCommand:
		return evalRunCommand(node, code, env)

	case *ast.CallExpression:
		function := Eval(node.Function, code, env)
		if isError(function) {
			// looking up the function failed, must be undefined
			return newError(env, "Undefined user function")
		}

		args := evalExpressions(node.Arguments, code, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args, code, env)

	case *ast.TroffCommand:
		evalTroffCommand(env)

	case *ast.TronCommand:
		evalTronCommand(env)
	}

	return nil
}

// evaluate the parameters to the auto command and determine the starting
// line number.  Once it is ready, save it into the environment for operating
func evalAutoCommand(cmd *ast.AutoCommand, code *ast.Code, env *object.Environment) {

	// if no start specified, assume 10 for know
	if cmd.Start == -1 {
		cmd.Start = 10
	}

	// does he want to start with the current line #, and do I have one?
	cl := code.CurLine()
	if cmd.Curr && (cl != 0) {
		cmd.Start = cl
	}

	// we just poke him into the environment
	env.SetAuto(cmd)
}

func evalBlockStatement(block *ast.BlockStatement, code *ast.Code, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, code, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalStatements(stmts *ast.Program, code *ast.Code, env *object.Environment) object.Object {
	var result object.Object

	// get an iterator across the program statements
	//iter := stmts.StatementIter()
	// make sure there are some
	t := code.Len()
	ok := t > 0
	// loop until you run out of code
	for ; ok; ok = code.Next() {
		result = Eval(code.Value(), code, env)

		err, ok := result.(*object.Error)

		if ok {
			env.Terminal().Println(err.Message)
		}
	}
	return result
}

// actually run the program
// ToDo: implement loading a program by name
// ToDo: close open data files (as soon as I support data files)
func evalRunCommand(run *ast.RunCommand, code *ast.Code, env *object.Environment) object.Object {
	// ToDo: need to clear the variable space
	if run.StartLine > 0 {
		code.Jump(run.StartLine)
	}

	pcode := env.Program.StatementIter()

	return Eval(env.Program, pcode, env)
}

// turn off tracing
func evalTroffCommand(env *object.Environment) {
	env.SetTrace(false)
}

// turn on tracing
func evalTronCommand(env *object.Environment) {
	env.SetTrace(true)
}

func evalClsStatement(cls *ast.ClsStatement, code *ast.Code, env *object.Environment) {
	env.Terminal().Cls()
}

func evalDimStatement(dim *ast.DimStatement, code *ast.Code, env *object.Environment) {

	for i, id := range dim.Vars {
		typeid, _ := parseVarName(id.Token.Literal)

		obj := allocArray(typeid, id.Index, code, env)
		env.Set(dim.Vars[i].Token.Literal, obj)
	}

}

func allocArray(typeid string, dims []*ast.IndexExpression, code *ast.Code, env *object.Environment) object.Object {
	d := Eval(dims[0].Index, code, env)
	if isError(d) {
		return d
	}
	i, ok := d.(*object.Integer)

	if !ok {
		i = &object.Integer{Value: coerceIndex(d)}
	}

	elms := make([]object.Object, i.Value)
	obj := object.Array{TypeID: typeid, Elements: elms}

	// if more dimensions exist, recurse down them
	if len(dims) > 1 {
		for i := range obj.Elements {
			obj.Elements[i] = allocArray(typeid, dims[1:], code, env)

			// check for possible errors
			err, ok := obj.Elements[i].(*object.Error)

			if ok {
				return err
			}
		}
		return &obj
	}

	// I'm at the last dimension value
	// create initial values for everybody

	for j := range obj.Elements {
		obj.Elements[j] = allocArrayValue(typeid)
	}

	return &obj
}

func allocArrayValue(typeid string) object.Object {
	var obj object.Object

	switch typeid {
	case "", "%":
		obj = &object.Integer{Value: 0}
	case "$":
		obj = &object.String{Value: ""}
	case "#":
		obj = &object.FloatDbl{Value: 0}
	case "!":
		obj = &object.FloatSgl{Value: 0}
	case "FIXED":
		obj = &object.Fixed{Value: decimal.Zero}
	}

	return obj
}

func evalGotoStatement(jmp string, code *ast.Code, env *object.Environment) object.Object {
	v, err := strconv.Atoi(strings.Trim(jmp, " "))

	if err != nil {
		return newError(env, "invalid line number: %s", jmp)
	}
	v2 := v

	err = code.Jump(v2)

	if err != nil {
		return newError(env, err.Error())
	}

	return &object.Integer{Value: int16(v)}
}

func evalListStatement(code *ast.Code, stmt *ast.ListStatement, env *object.Environment) {
	var out bytes.Buffer
	cd := env.Program.StatementIter()
	start := 0
	stop := cd.MaxLineNum()

	// figure out any limits to the listing
	if len(stmt.Start) > 0 {
		start, _ = strconv.Atoi(stmt.Start)
		if len(stmt.Lrange) == 0 {
			stop = start
		}
	}

	if len(stmt.Stop) > 0 {
		stop, _ = strconv.Atoi(stmt.Stop)
	}

	bMidLine := false
	bList := false
	for bMore := true; bMore; {
		stmt := cd.Value()

		lnm, ok := stmt.(*ast.LineNumStmt)

		if ok {
			if lnm.Value > stop {
				break
			}

			if bList {
				env.Terminal().Println(strings.TrimRight(out.String(), " "))
				out.Truncate(0)
			}
			bList = (lnm.Value >= start)
			bMidLine = false
		} else {
			if bMidLine {
				out.WriteString(": ")
			}
			bMidLine = true
		}

		if bList {
			out.WriteString(stmt.String())
		}

		bMore = cd.Next()
	}
	env.Terminal().Println(out.String())
}

func evalPrintStatement(node *ast.PrintStatement, code *ast.Code, env *object.Environment) {
	for i, item := range node.Items {

		env.Terminal().Print(Eval(item, code, env).Inspect())

		if node.Seperators[i] == "," {
			env.Terminal().Print("\t")
		}
	}
	if node.Seperators[len(node.Seperators)-1] == ";" {
		return
	}

	// end with a newline
	env.Terminal().Println("")
}

func evalPrefixExpression(operator string, right object.Object, code *ast.Code, env *object.Environment) object.Object {
	switch operator {
	case "-":
		return evalMinusPrefixOperatorExpression(right, env)
	default:
		return newError(env, "unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusPrefixOperatorExpression(right object.Object, env *object.Environment) object.Object {
	switch right.Type() {
	case object.INTEGER_OBJ:
		value := right.(*object.Integer).Value
		return &object.Integer{Value: -value}
	case object.FIXED_OBJ:
		value := right.(*object.Fixed).Value
		return &object.Fixed{Value: value.Neg()}
	case object.FLOATSGL_OBJ:
		value := right.(*object.FloatSgl).Value
		return &object.FloatSgl{Value: -value}
	case object.FLOATDBL_OBJ:
		value := right.(*object.FloatDbl).Value
		return &object.FloatDbl{Value: -value}
	}
	return newError(env, "unsupport negative on %s", right.Type())
}

func evalInfixExpression(operator string, left, right object.Object, env *object.Environment) object.Object {
	fn, ok := typeConverters[string(left.Type())+string(right.Type())]

	if !ok {
		return newError(env, "type mis-match")
	}
	return fn(operator, left, right, env)
}

func evalStringInfixExpression(operator string, left, right object.Object, env *object.Environment) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}

	case "=":
		return &object.Integer{Value: bool2int16(leftVal == rightVal)}

	default:
		return newError(env, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, leftVal, rightVal int, env *object.Environment) object.Object {
	switch operator {
	case "+":
		return fixType(leftVal + rightVal)
	case "-":
		return fixType(leftVal - rightVal)
	case "*":
		return fixType(leftVal * rightVal)
	case "/":
		return fixType(float64(leftVal) / float64(rightVal))
	case "\\":
		// I'm learning stuff I never knew about GWBasic
		return &object.Integer{Value: int16(leftVal) / int16(rightVal)}
	case "MOD":
		return fixType(leftVal % rightVal)
	case "<":
		return &object.Integer{Value: bool2int16(leftVal < rightVal)}
	case "<=":
		return &object.Integer{Value: bool2int16(leftVal <= rightVal)}
	case ">":
		return &object.Integer{Value: bool2int16(leftVal > rightVal)}
	case ">=":
		return &object.Integer{Value: bool2int16(leftVal >= rightVal)}
	case "<>":
		return &object.Integer{Value: bool2int16(leftVal != rightVal)}
	case "=":
		return &object.Integer{Value: bool2int16(leftVal == rightVal)}
	default:
		return newError(env, "unsupported operator %s", operator)
	}
}

func evalFixedInfixExpression(operator string, left, right decimal.Decimal, env *object.Environment) object.Object {

	switch operator {
	case "+":
		return &object.Fixed{Value: left.Add(right)}
	case "-":
		return &object.Fixed{Value: left.Sub(right)}
	case "*":
		return &object.Fixed{Value: left.Mul(right)}
	case "/":
		decimal.DivisionPrecision = 6
		return &object.Fixed{Value: left.Div(right)}
	case "<":
		return &object.Integer{Value: bool2int16(left.Cmp(right) == -1)}
	case "<=":
		return &object.Integer{Value: bool2int16(left.Cmp(right) != 1)}
	case ">":
		return &object.Integer{Value: bool2int16(left.Cmp(right) == 1)}
	case ">=":
		return &object.Integer{Value: bool2int16(left.Cmp(right) != -1)}
	case "<>":
		return &object.Integer{Value: bool2int16(left.Cmp(right) != 0)}
	case "=":
		return &object.Integer{Value: bool2int16(left.Cmp(right) == 0)}
	default:
		return newError(env, "unsupported operator %s", operator)
	}
}

func evalFloatInfixExpression(operator string, leftVal, rightVal float32, env *object.Environment) object.Object {

	switch operator {
	case "+":
		return fixType(leftVal + rightVal)
	case "-":
		return fixType(leftVal - rightVal)
	case "*":
		return fixType(leftVal * rightVal)
	case "/":
		return fixType(leftVal / rightVal)
	case "<":
		return &object.Integer{Value: bool2int16(leftVal < rightVal)}
	case "<=":
		return &object.Integer{Value: bool2int16(leftVal <= rightVal)}
	case ">":
		return &object.Integer{Value: bool2int16(leftVal > rightVal)}
	case ">=":
		return &object.Integer{Value: bool2int16(leftVal >= rightVal)}
	case "<>":
		return &object.Integer{Value: bool2int16(leftVal != rightVal)}
	case "=":
		return &object.Integer{Value: bool2int16(leftVal == rightVal)}
	default:
		return newError(env, "unsupported operator %s", operator)
	}
}

func evalFloatDblInfixExpression(operator string, leftVal, rightVal float64, env *object.Environment) object.Object {

	switch operator {
	case "+":
		return fixType(leftVal + rightVal)
	case "-":
		return fixType(leftVal - rightVal)
	case "*":
		return fixType(leftVal * rightVal)
	case "/":
		return fixType(leftVal / rightVal)
	case "<":
		return &object.Integer{Value: bool2int16(leftVal < rightVal)}
	case "<=":
		return &object.Integer{Value: bool2int16(leftVal <= rightVal)}
	case ">":
		return &object.Integer{Value: bool2int16(leftVal > rightVal)}
	case ">=":
		return &object.Integer{Value: bool2int16(leftVal >= rightVal)}
	case "<>":
		return &object.Integer{Value: bool2int16(leftVal != rightVal)}
	case "=":
		return &object.Integer{Value: bool2int16(leftVal == rightVal)}
	default:
		return newError(env, "unsupported operator %s", operator)
	}
}

func evalIfExpression(ie *ast.IfExpression, code *ast.Code, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, code, env)
	if isError(condition) {
		return condition
	}

	if condition.(*object.Integer).Value == 0 { // that's a false
		return Eval(ie.Alternative, code, env)
	}

	return Eval(ie.Consequence, code, env)
}

func evalExpressions(exps []ast.Expression, code *ast.Code, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, code, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func applyFunction(fn object.Object, args []object.Object, code *ast.Code, env *object.Environment) object.Object {

	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, code, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(env, args...)

	default:
		return newError(env, "not a function: %s", fn.Type())

	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

func evalIdentifier(node *ast.Identifier, code *ast.Code, env *object.Environment) object.Object {

	tk := node.Token.Literal
	val, ok := env.Get(node.Value)
	if ok && (tk[len(tk)-1] != ']') {
		return val
	}

	if ok && (tk[len(tk)-1] == ']') && (node.Index != nil) {
		arrValue := evalIndexArray(node.Index, val, nil, code, env)

		return arrValue
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError(env, "Syntax error")
	//return newError(env, "identifier not found: "+node.Value)
}

func evalIndexArray(index []*ast.IndexExpression, array, newVal object.Object, code *ast.Code, env *object.Environment) object.Object {

	indVal := Eval(index[0].Index, code, env)
	if isError(indVal) {
		return indVal
	}

	var arrValue object.Object
	if len(index) == 1 {
		// if this array was multi-dimensional
		// I've reached the end
		arrValue = evalIndexExpression(array, indVal, newVal, env)
	} else {
		// not to the final object, don't send the new value
		arrValue = evalIndexExpression(array, indVal, nil, env)
	}

	_, ok := arrValue.(*object.Array)

	if (len(index) > 1) && ok {
		arrValue = evalIndexArray(index[1:], arrValue, newVal, code, env)
	}

	if (len(index) > 1) && !ok {
		arrValue = newError(env, "Subscript out of range")
	}

	return arrValue
}

func evalIndexExpression(array, index, newVal object.Object, env *object.Environment) object.Object {
	switch {
	case array.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(array, index, newVal, env)
	case array.Type() == object.ARRAY_OBJ && index.Type() != object.INTEGER_OBJ:
		return evalArrayIndexExpression(array, &object.Integer{Value: coerceIndex(index)}, newVal, env)
	default:
		er := newError(env, "index operator not supported: %s", array.Type())
		return er
	}
}

func evalArrayIndexExpression(array, index, newVal object.Object, env *object.Environment) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int16(len(arrayObject.Elements) - 1)
	if idx < 0 || idx > max {
		return newError(env, "Subscript out of range")
	}

	if arrayObject.Elements[idx] != nil {
		if newVal != nil {
			arrayObject.Elements[idx] = newVal
		}
		return arrayObject.Elements[idx]
	}

	return newError(env, "Subscript out of range")
}

func saveVariable(code *ast.Code, env *object.Environment, name *ast.Identifier, val object.Object) object.Object {
	sname := name.String()
	cv, ok := env.Get(sname)

	if !ok {
		// variable doesn't exist, time to create
		return saveNewVariable(code, env, name, val)
	}

	typeid, isarray := parseVarName(sname)

	cvarray, cvisarray := cv.(*object.Array)

	if isarray && cvisarray && checkTypes(typeid, val) {
		arrVar := evalIndexArray(name.Index, cvarray, val, code, env)
		env.Set(sname, cvarray)
		return arrVar
	}

	if !checkTypes(typeid, val) {
		return newError(env, "type mis-match")
	}

	tv := &object.TypedVar{Value: val, TypeID: typeid}
	env.Set(sname, tv)
	return tv
}

func saveNewVariable(code *ast.Code, env *object.Environment, name *ast.Identifier, val object.Object) object.Object {

	sname := name.String()
	typeid, isarray := parseVarName(sname)

	// handle it if he is an array
	if isarray {
		var dim []*ast.IndexExpression
		dim = append(dim, &ast.IndexExpression{Index: &ast.IntegerLiteral{Value: 10}})
		arr := allocArray(typeid, dim, code, env).(*object.Array)
		index := Eval(name.Index[0].Index, code, env)

		if isError(index) {
			return index
		}

		indValue, isInt := index.(*object.Integer)

		if !isInt {
			indValue = &object.Integer{Value: coerceIndex(index)}
		}

		if 10 > indValue.Value {
			arr.Elements[indValue.Value] = val
			env.Set(sname, arr)
			return arr
		}

		return newError(env, "index out of range")
	}

	// just a typed variable
	tv := &object.TypedVar{Value: val, TypeID: typeid}
	env.Set(sname, tv)
	return tv
}

func coerceIndex(idx object.Object) int16 {
	switch idx.Type() {
	case object.FIXED_OBJ:
		fx, _ := idx.(*object.Fixed)
		fx2 := fx.Value.Round(0)
		ti := fx2.IntPart()
		return int16(ti)
	case object.FLOATSGL_OBJ:
		fx, _ := idx.(*object.FloatSgl)
		return int16(fx.Value)
	case object.FLOATDBL_OBJ:
		fx, _ := idx.(*object.FloatDbl)
		return int16(fx.Value)
	}

	return 0
}

func checkTypes(typeid string, val object.Object) bool {

	tv, typed := val.(*object.TypedVar)

	// if TypedVar, reach in and get his value object
	if typed {
		val = tv.Value
	}

	switch val.Type() {
	case object.INTEGER_OBJ:
		if strings.ContainsAny(typeid, "%#!") || (typeid == "") {
			return true
		}
	case object.FIXED_OBJ:
		if strings.ContainsAny(typeid, "#!") || (typeid == "") {
			return true
		}
	case object.FLOATDBL_OBJ:
		if (typeid == "#") || (typeid == "") {
			return true
		}
	case object.FLOATSGL_OBJ:
		if strings.ContainsAny(typeid, "#!") || (typeid == "") {
			return true
		}
	case object.STRING_OBJ:
		if typeid == "$" {
			return true
		}
	}

	return false
}

func parseVarName(name string) (string, bool) {
	l := len(name) - 1
	isarray := false
	typeid := ""

	// check if name is an array
	if name[l] == ']' {
		isarray = true
		l = l - 2
	}

	switch name[l] {
	case '$':
		typeid = "$"
	case '%':
		typeid = "%"
	case '#':
		typeid = "#"
	case '!':
		typeid = "!"
	}

	return typeid, isarray
}

func newError(env *object.Environment, format string, a ...interface{}) *object.Error {
	msg := fmt.Sprintf(format, a...)
	tk, ok := env.Get(token.LINENUM)

	if ok {
		msg += fmt.Sprintf(" in %d", tk.(*object.IntDbl).Value)
	}

	return &object.Error{Message: msg}
}

func checkType(obj object.Object, want object.ObjectType) bool {
	if obj != nil {
		return obj.Type() == want
	}
	return false
}

func isError(obj object.Object) bool {
	return checkType(obj, object.ERROR_OBJ)
}

func bool2int16(b bool) int16 {
	// The compiler currently only optimizes this form.
	// See issue 6011.
	var i int16
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}
