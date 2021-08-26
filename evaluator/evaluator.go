package evaluator

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/builtins"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/filelist"
	"github.com/navionguy/basicwasm/fileserv"
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

// Eval returns the object at a node
func Eval(node ast.Node, code *ast.Code, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(node, code, env)

	case *ast.AutoCommand:
		evalAutoCommand(node, code, env)

	case *ast.BeepStatement:
		evalBeepStatement(node, code, env)

	case *ast.ChainStatement:
		evalChainStatement(node, code, env)

	case *ast.ClsStatement:
		evalClsStatement(node, code, env)

	case *ast.DataStatement:
		return nil

	case *ast.DimStatement:
		evalDimStatement(node, code, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, code, env)

	case *ast.FilesCommand:
		evalFilesCommand(node, code, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, code, env)

	case *ast.GroupedExpression:
		return Eval(node.Exp, code, env)

	case *ast.HexConstant:
		return evalHexConstant(node, code, env)

	case *ast.LetStatement:
		val := Eval(node.Value, code, env)
		if isError(val) {
			return val
		}
		// life gets more complicated, not less
		if !strings.ContainsAny(node.Name.Token.Literal, "[($%!#") {
			env.Set(node.Name.Token.Literal, val)
			return val
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

	case *ast.OctalConstant:
		return evalOctalConstant(node, code, env)

	case *ast.PrintStatement:
		evalPrintStatement(node, code, env)

		// Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.DblIntegerLiteral:
		return &object.IntDbl{Value: node.Value}

	case *ast.FixedLiteral:
		val, err := decimal.NewFromString(node.Value.String())

		if err != nil {
			return newError(env, err.Error())
		}
		return &object.Fixed{Value: val}

	case *ast.FloatSingleLiteral:
		return &object.FloatSgl{Value: node.Value}

	case *ast.FloatDoubleLiteral:
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

	case *ast.ReadStatement:
		return evalReadStatement(node, code, env)

	case *ast.RestoreStatement:
		return evalRestoreStatement(node, code, env)

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

// just sound the bell
func evalBeepStatement(beep *ast.BeepStatement, code *ast.Code, env *object.Environment) {
	env.Terminal().SoundBell()
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

// tries to load a new program and start it's execution.
func evalChainStatement(chain *ast.ChainStatement, code *ast.Code, env *object.Environment) {
	rdr := evalChainLoad(code, chain, env)

	// if file doesn't load, get out
	if rdr == nil {
		return
	}
}

// attempt to pull down the  desired file
func evalChainLoad(code *ast.Code, chain *ast.ChainStatement, env *object.Environment) *bufio.Reader {
	rdr, err := fileserv.GetFile(chain.File, env)

	if err != nil {
		env.Terminal().Println(err.Error())
		return nil
	}

	evalChainParse(rdr, code, chain, env)

	return rdr
}

// parse the file into an executable AST
func evalChainParse(rdr *bufio.Reader, code *ast.Code, chain *ast.ChainStatement, env *object.Environment) {
	fileserv.ParseFile(rdr, env)
	env.Program.ConstData().Restore() // start at the first DATA statement

	// go figure out how to start execution
	evalChainStart(chain, code, env)
}

// either start execution, or move to the first statement in the new AST
func evalChainStart(chain *ast.ChainStatement, code *ast.Code, env *object.Environment) {

	// if we are not running, time to start
	if !env.ProgramRunning() {
		evalChainExecute(chain, env)
	}
}

// executing a command entry, start program execution
func evalChainExecute(chain *ast.ChainStatement, env *object.Environment) object.Object {
	pcode := env.Program.StatementIter()
	env.Program.ConstData().Restore()

	rc := evalRunStart(pcode, env)

	return rc
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

// read constant values out of data statements into variables
func evalReadStatement(rd *ast.ReadStatement, code *ast.Code, env *object.Environment) object.Object {
	var value object.Object

	for _, item := range rd.Vars {
		name, ok := item.(*ast.Identifier)

		if !ok {
			return newError(env, syntaxErr)
		}

		cst := env.Program.ConstData().Next()

		if cst == nil {
			return newError(env, outOfDataErr)
		}

		switch val := (*cst).(type) {
		case *ast.StringLiteral:
			value = &object.String{Value: val.Value}
		case *ast.IntegerLiteral:
			value = &object.Integer{Value: val.Value}
		case *ast.DblIntegerLiteral:
			value = &object.IntDbl{Value: val.Value}
		case *ast.FixedLiteral:
			value = &object.Fixed{Value: val.Value}
		case *ast.FloatSingleLiteral:
			value = &object.FloatSgl{Value: val.Value}
		case *ast.FloatDoubleLiteral:
			value = &object.FloatDbl{Value: val.Value}
		default:
			value = Eval(val, code, env)

			// yes, the default case would work for all cases
			// but I wanted to be clear what was going on
		}

		saveVariable(code, env, name, value)
	}
	return value
}

// evalRestoreStatement makes sure you can re-read data statements
func evalRestoreStatement(rst *ast.RestoreStatement, code *ast.Code, env *object.Environment) object.Object {
	if rst.Line >= 0 {
		// he wants to restore to a certain line
		if env.Program.ConstData().RestoreTo(rst.Line) {
			return nil
		}

		return newError(env, unDefinedLineNumberErr)
	}

	// restore to the beginning
	env.Program.ConstData().Restore()

	return nil
}

// actually run the program
// ToDo: close open data files (as soon as I support data files)
func evalRunCommand(run *ast.RunCommand, code *ast.Code, env *object.Environment) object.Object {
	if run.LoadFile != "" {
		// load the source file then run it
		return evalRunLoad(run, env)
	}

	return evalRunCheckStartLineNum(run, env)
}

// pull the file down from the server
func evalRunLoad(run *ast.RunCommand, env *object.Environment) object.Object {
	rdr, err := fileserv.GetFile(run.LoadFile, env)

	if err != nil {
		env.Terminal().Println(err.Error())
		return nil
	}

	return evalRunParse(rdr, run, env)
}

// Parse the code in the reader try to run it.
func evalRunParse(rdr *bufio.Reader, run *ast.RunCommand, env *object.Environment) object.Object {
	// create a new program code space
	env.Program.New()
	if !run.KeepOpen {
		// ToDo: once I implement files, I'll need a way to close them all at once
	}

	// parse the loaded file into an AST for evaluation
	fileserv.ParseFile(rdr, env)

	return evalRunCheckStartLineNum(run, env)
}

func evalRunCheckStartLineNum(run *ast.RunCommand, env *object.Environment) object.Object {
	pcode := env.Program.StatementIter()
	env.Program.ConstData().Restore()

	if run.StartLine > 0 {
		pcode.Jump(run.StartLine)
	}

	return evalRunStart(pcode, env)
}

// actually go execute the code
func evalRunStart(code *ast.Code, env *object.Environment) object.Object {
	env.SetRun(true)
	rc := Eval(env.Program, code, env)
	env.SetRun(false)

	return rc
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

// FILES instruct the system to list filenames for current directory
// FILES "path" lists all files in the specified directory
func evalFilesCommand(files *ast.FilesCommand, code *ast.Code, env *object.Environment) {

	dir, err := fileserv.GetFile(files.Path, env)
	if err != nil {
		env.Terminal().Println(err.Error())
		return
	}

	list := filelist.NewFileList()
	err = list.Build(dir)
	if err != nil {
		catchNotDir(files.Path, err, env)
		return
	}

	displayFiles(list, env)
}

func catchNotDir(path string, err error, env *object.Environment) {
	if err.Error() != "NotDir" {
		env.Terminal().Println(err.Error())
		return
	}

	cwd := fileserv.GetCWD(env)
	env.Terminal().Println(cwd)
	env.Terminal().Println(path)
}

// mimic the way in which GWBasic display directory contents
func displayFiles(files *filelist.FileList, env *object.Environment) {
	col := 0
	for _, fl := range files.Files {
		output := fileserv.FormatFileName(fl.Name, fl.Subdir)

		env.Terminal().Print(output)
		col++
		if col == 4 {
			env.Terminal().Println("")
			col = 0
		}
	}
	env.Terminal().Println("")
}

func evalGotoStatement(jmp string, code *ast.Code, env *object.Environment) object.Object {
	v, err := strconv.Atoi(strings.Trim(jmp, " "))

	if err != nil {
		return newError(env, syntaxErr)
	}
	v2 := v

	err = code.Jump(v2)

	if err != nil {
		return newError(env, err.Error())
	}

	return &object.Integer{Value: int16(v)}
}

func evalHexConstant(stmt *ast.HexConstant, code *ast.Code, env *object.Environment) object.Object {
	dst, err := strconv.ParseInt(stmt.Value, 16, 16)

	if err != nil {
		st := err.Error()
		if strings.Contains(st, "value out of range") {
			return newError(env, overflowErr)
		}
		return newError(env, syntaxErr)
	}

	return &object.Integer{Value: int16(dst)}
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
			if int(lnm.Value) > stop {
				break
			}

			if bList {
				env.Terminal().Println(strings.TrimRight(out.String(), " "))
				out.Truncate(0)
			}
			bList = (int(lnm.Value) >= start)
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

func evalOctalConstant(stmt *ast.OctalConstant, code *ast.Code, env *object.Environment) object.Object {

	dst, err := strconv.ParseInt(stmt.Value, 8, 16)

	if err != nil {
		st := err.Error()
		if strings.Contains(st, "value out of range") {
			return newError(env, overflowErr)
		}
		return newError(env, syntaxErr)
	}

	return &object.Integer{Value: int16(dst)}
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
	case object.INTEGER_DBL:
		value := right.(*object.IntDbl).Value
		return &object.IntDbl{Value: -value}
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
		return builtins.FixType(env, leftVal+rightVal)
	case "-":
		return builtins.FixType(env, leftVal-rightVal)
	case "*":
		return builtins.FixType(env, leftVal*rightVal)
	case "/":
		return builtins.FixType(env, float64(leftVal)/float64(rightVal))
	case "\\":
		// I'm learning stuff I never knew about GWBasic
		return &object.Integer{Value: int16(leftVal) / int16(rightVal)}
	case "MOD":
		return builtins.FixType(env, leftVal%rightVal)
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
		return builtins.FixType(env, leftVal+rightVal)
	case "-":
		return builtins.FixType(env, leftVal-rightVal)
	case "*":
		return builtins.FixType(env, leftVal*rightVal)
	case "/":
		return builtins.FixType(env, leftVal/rightVal)
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
		return builtins.FixType(env, leftVal+rightVal)
	case "-":
		return builtins.FixType(env, leftVal-rightVal)
	case "*":
		return builtins.FixType(env, leftVal*rightVal)
	case "/":
		return builtins.FixType(env, leftVal/rightVal)
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
		obj := Eval(fn.Body, code, extendedEnv)
		return obj

	case *object.Builtin:
		return fn.Fn(env, fn, args...)

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

func evalIdentifier(node *ast.Identifier, code *ast.Code, env *object.Environment) object.Object {

	label := node.Value
	val, ok := env.Get(node.Value)
	if ok && (label[len(label)-1] != ']') {
		return val
	}

	if builtin, ok := builtins.Builtins[node.Value]; ok {
		return builtin
	}

	if ok && (label[len(label)-1] == ']') && (node.Index != nil) {
		arrValue := evalIndexArray(node.Index, val, nil, code, env)

		return arrValue
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

// saveVariable into the environment
func saveVariable(code *ast.Code, env *object.Environment, name *ast.Identifier, val object.Object) object.Object {
	sname := name.Value
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

	sname := name.Value
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
			env.Set(name.Value, arr)
			return arr
		}

		return newError(env, "index out of range")
	}

	// just a typed variable
	tv := &object.TypedVar{Value: val, TypeID: typeid}
	env.Set(name.Token.Literal, tv)
	return tv
}

func coerceIndex(idx object.Object) int16 {
	switch fx := idx.(type) {
	case *object.Fixed:
		fx2 := fx.Value.Round(0)
		ti := fx2.IntPart()
		return int16(ti)
	case *object.FloatSgl:
		return int16(math.Round(float64(fx.Value)))
	case *object.FloatDbl:
		return int16(math.Round(fx.Value))
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
	parts := strings.Split(name, "[")
	altparts := strings.Split(name, "(")

	isarray := false
	if len(parts) > 1 {
		isarray = true
	}

	if len(altparts) > 1 {
		isarray = true
		parts = altparts
	}

	base := parts[0]
	typeid := base[len(base)-1:]

	if !strings.ContainsAny(typeid, "$%#!") {
		return "", isarray
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
