// Logic to evaluate statements and commands.  Updates variables saved in the environment and outputs
// the browser UI.
package evaluator

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/builtins"
	"github.com/navionguy/basicwasm/decimal"
	"github.com/navionguy/basicwasm/filelist"
	"github.com/navionguy/basicwasm/fileserv"
	"github.com/navionguy/basicwasm/lexer"
	"github.com/navionguy/basicwasm/object"
	"github.com/navionguy/basicwasm/parser"
	"github.com/navionguy/basicwasm/settings"
	"github.com/navionguy/basicwasm/token"
)

// Eval evaluates the current node in the AST.  It generally returns nil, but
// it can return an error object or a halt object.
func Eval(node ast.Node, code *ast.Code, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(code, env)

	case *ast.AutoCommand:
		return evalAutoCommand(node, code, env)

	case *ast.BeepStatement:
		evalBeepStatement(node, code, env)

	case *ast.BuiltinExpression:
		return evalBuiltinExpression(node, code, env)

	case *ast.ChainStatement:
		return evalChainStatement(node, code, env)

	case *ast.ChDirStatement:
		return evalChDirStatement(node, code, env)

	case *ast.ClearCommand:
		evalClearCommand(node, code, env)

	case *ast.ClsStatement:
		evalClsStatement(node, code, env)

	case *ast.ColorStatement:
		return evalColorStatement(node, code, env)

	case *ast.ContCommand:
		return evalContCommand(node, code, env)

	case *ast.Csrlin:
		return evalCsrLinExpression(code, env)

	case *ast.DataStatement:
		return nil

	case *ast.DimStatement:
		evalDimStatement(node, code, env)

	case *ast.BlockExpression:
		return evalBlockExpression(node, code, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, code, env)

	case *ast.CommonStatement:
		evalCommonStatement(node, code, env)

	case *ast.EndStatement:
		return evalEndStatement(node, code, env)

	case *ast.ErrorStatement:
		return evalErrorStatement(node, code, env)

	case *ast.ExpressionStatement:
		if node.Token.Literal == ":" {
			return nil
		}
		return Eval(node.Expression, code, env)

	case *ast.FilesCommand:
		return evalFilesCommand(node, code, env)

	case *ast.ForStatment:
		return evalForStatement(node, code, env)

	case *ast.GosubStatement:
		return evalGosubStatement(node, code, env)

	case *ast.GotoStatement:
		return evalGotoStatement(node, code, env)

	case *ast.GroupedExpression:
		return Eval(node.Exp, code, env)

	case *ast.HexConstant:
		return evalHexConstant(node, code, env)

	case *ast.KeyStatement:
		return evalKeyStatement(node, code, env)

	case *ast.LetStatement:
		val := Eval(node.Value, code, env)
		if isError(val) {
			return val
		}
		// life gets more complicated, not less
		if !strings.ContainsAny(node.Name.Token.Literal, "[($%!#") {
			env.Set(node.Name.Token.Literal, val)
			return nil
		}
		return saveVariable(code, env, node.Name, val)

	case *ast.LineNumStmt:
		ln := &object.IntDbl{Value: node.Value}
		env.Set(token.LINENUM, ln)
		if env.GetTrace() {
			env.Terminal().Print(fmt.Sprintf("[%d]", node.Value))
		}

	case *ast.ListStatement:
		evalListStatement(node, code, env)

	case *ast.LoadCommand:
		return evalLoadCommand(node, code, env)

	case *ast.LocateStatement:
		return evalLocateStatement(node, code, env)

	case *ast.NextStatement:
		return evalNextStatement(node, code, env)

	case *ast.NewCommand:
		return evalNewCommand(node, code, env)

	case *ast.OctalConstant:
		return evalOctalConstant(node, env)

	case *ast.OnErrorGoto:
		return evalOnErrorStatement(node, code, env)

	case *ast.OnGoStatement:
		return evalOnGoStatement(node, code, env)

	case *ast.PrintStatement:
		return evalPrintStatement(node, code, env)

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
		return evalImpliedLetStatement(node, code, env)

	case *ast.PrefixExpression:
		right := Eval(node.Right, code, env)
		return evalPrefixExpression(node.Operator, right, code, env)

	case *ast.InfixExpression:
		left := Eval(node.Left, code, env)
		right := Eval(node.Right, code, env)
		return evalInfixExpression(node.Operator, left, right, env)

	case *ast.IfStatement:
		return evalIfStatement(node, code, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		obj := &object.Function{Parameters: params, Env: env, Body: body}
		env.Set(node.Token.Literal, obj)
		return nil

	case *ast.ReadStatement:
		return evalReadStatement(node, code, env)

	case *ast.RestoreStatement:
		return evalRestoreStatement(node, code, env)

	case *ast.ResumeStatement:
		return evalResumeStatement(node, code, env)

	case *ast.RemStatement:
		return nil // nothing to be done

	case *ast.ReturnStatement:
		return evalReturnStatement(node, code, env)

	case *ast.RunCommand:
		return evalRunCommand(node, code, env)

	case *ast.ScreenStatement:
		return evalScreenStatement(node, code, env)

	case *ast.StopStatement:
		return evalStopStatement(node, code, env)

	case *ast.CallExpression:
		function := Eval(node.Function, code, env)
		if isError(function) {
			// looking up the function failed, must be undefined
			return object.StdError(env, berrors.UndefinedFunction)
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

	case *ast.ViewPrintStatement:
		return evalViewPrintStatement(node, code, env)

	case *ast.ViewStatement:
		return evalViewStatement(node, code, env)
	default:
		msg := fmt.Sprintf("unsupported codepoint at line %d, %T", code.CurLine(), node)
		env.Terminal().Println(msg)
		return &object.HaltSignal{}
	}

	return nil
}

// evaluate the parameters to the auto command and determine the starting
// line number.  Once it is ready, save it into the environment for operating
func evalAutoCommand(cmd *ast.AutoCommand, code *ast.Code, env *object.Environment) object.Object {
	// if I have too many params, that's an error
	if len(cmd.Params) > 2 {
		return object.StdError(env, berrors.Syntax)
	}

	// If I have params, use them
	if len(cmd.Params) > 0 {
		return evalAutoCommandParams(cmd, code, env)
	}

	// just start a default auto mode
	def := ast.AutoCommand{Params: []ast.Expression{&ast.DblIntegerLiteral{Value: 10}, &ast.DblIntegerLiteral{Value: 10}}, On: true}

	env.SaveSetting(settings.Auto, &def)

	return nil
}

// evaluate and use params to auto command
func evalAutoCommandParams(cmd *ast.AutoCommand, code *ast.Code, env *object.Environment) object.Object {
	auto := ast.AutoCommand{}

	// if no starting value, assume default
	if cmd.Params[0] == nil {
		auto.Params = append(auto.Params, &ast.DblIntegerLiteral{Value: 10})
	} else {
		// go figure out the first parameter
		rc := evalAutoCommandParam1(&auto, cmd, code, env)

		// if he got an error, time to leave
		if rc != nil {
			return rc
		}
	}

	if len(cmd.Params) == 2 {
		return evalAutoCommandParam2(&auto, cmd, code, env)
	}

	// use the default step
	auto.Params = append(auto.Params, &ast.DblIntegerLiteral{Value: 10})

	// save to the settings
	env.SaveSetting(settings.Auto, &auto)

	return nil
}

// evalute the first parameter of the Auto Command
func evalAutoCommandParam1(auto, cmd *ast.AutoCommand, code *ast.Code, env *object.Environment) object.Object {
	// if the first param is '.', then we use current line number

	cl, ok := cmd.Params[0].(*ast.Identifier)

	if ok && (cl.Value == ".") {
		// valid "start with current line number"
		line := env.Get(token.LINENUM)

		val := line.(*object.IntDbl).Value
		if val > 0 {
			auto.Params = append(auto.Params, &ast.DblIntegerLiteral{Value: val})
		} else {
			auto.Params = append(auto.Params, &ast.DblIntegerLiteral{Value: 10})
		}
		return nil
	}

	// there is an expression, go evaluate it
	val := evalExpressionNode(cmd.Params[0], code, env)
	l, err := coerceDblInteger(val, env)

	if err != nil {
		return err
	}

	auto.Params = append(auto.Params, &ast.DblIntegerLiteral{Value: l})
	return nil
}

// evaluate the second parameter of the Auto Command
func evalAutoCommandParam2(auto, cmd *ast.AutoCommand, code *ast.Code, env *object.Environment) object.Object {
	// eval the second parameter
	val := evalExpressionNode(cmd.Params[1], code, env)

	switch prm := val.(type) {
	case *object.Integer:
		auto.Params = append(auto.Params, &ast.DblIntegerLiteral{Value: int32(prm.Value)})
	case *object.IntDbl:
		auto.Params = append(auto.Params, &ast.DblIntegerLiteral{Value: prm.Value})

	default:
		return object.StdError(env, berrors.Syntax)
	}

	// save the final values
	env.SaveSetting(settings.Auto, auto)
	return nil
}

// just sound the bell
func evalBeepStatement(beep *ast.BeepStatement, code *ast.Code, env *object.Environment) {
	env.Terminal().SoundBell()
}

// evaluate the user defined expression
func evalBlockExpression(Exp *ast.BlockExpression, code *ast.Code, env *object.Environment) object.Object {
	return Eval(Exp.Exp, code, env)
}

// evals a user defined function, need to rename this
func evalBlockStatement(block *ast.BlockStatement, code *ast.Code, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, code, env)

		if result != nil {
			rt := result.Type()
			if rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

// execute a built in function
func evalBuiltinExpression(builtin *ast.BuiltinExpression, code *ast.Code, env *object.Environment) object.Object {

	// if I can't find the function, it isn't really built in
	blt, ok := builtins.Builtins[builtin.TokenLiteral()]

	if !ok {
		return object.StdError(env, berrors.Syntax)
	}

	// evaluate all of his parameters
	params := evalExpressions(builtin.Params, code, env)

	// return functions result
	return blt.Fn(env, blt, params...)
}

// tries to load a new program and start it's execution.
func evalChainStatement(chain *ast.ChainStatement, code *ast.Code, env *object.Environment) object.Object {
	// if no filename was given, report the error
	if chain.Path == nil {
		return object.StdError(env, berrors.Syntax)
	}

	// eval the path to get a string
	res := Eval(chain.Path, code, env)

	// make sure the result is a string
	fn, ok := res.(*object.String)

	if !ok {
		return object.StdError(env, berrors.TypeMismatch)
	}
	return evalChainLoad(fn.Value, code, chain, env)
}

// attempt to pull down the  desired file
func evalChainLoad(file string, code *ast.Code, chain *ast.ChainStatement, env *object.Environment) object.Object {
	rdr, err := fileserv.GetFile(file, env)

	if err != nil {
		return err
	}

	return evalChainParse(rdr, code, chain, env)
}

// parse the file into an executable AST
func evalChainParse(rdr *bufio.Reader, code *ast.Code, chain *ast.ChainStatement, env *object.Environment) object.Object {
	env.NewProgram()
	env.ClearVars()
	fileserv.ParseFile(rdr, env)
	env.ConstData().Restore() // start at the first DATA statement

	// go figure out how to start execution
	return evalChainStart(chain, code, env)
}

// either start execution, or move to the first statement in the new AST
func evalChainStart(chain *ast.ChainStatement, code *ast.Code, env *object.Environment) object.Object {

	// if we are not running, time to start
	if !env.ProgramRunning() {
		return evalChainExecute(chain, env)
	}

	// time to trigger a restart
	return &object.RestartSignal{}
}

// executing a command entry, start program execution
func evalChainExecute(chain *ast.ChainStatement, env *object.Environment) object.Object {
	pcode := env.StatementIter()
	env.ConstData().Restore()

	rc := evalRunStart(pcode, env)

	return rc
}

// executing a directory change
func evalChDirStatement(chdir *ast.ChDirStatement, code *ast.Code, env *object.Environment) object.Object {
	// should be one, and only one, parameter
	if len(chdir.Path) != 1 {
		return object.StdError(env, berrors.Syntax)
	}

	// evaluate my path expression
	rc := evalExpressionNodeTyped(chdir.Path[0], code, env, &object.String{})

	// I should get a string value
	path, ok := rc.(*object.String)
	if !ok {
		return object.StdError(env, berrors.Syntax)
	}

	fp := fileserv.BuildFullPath(strings.ToLower(path.Value), env)
	fmt.Printf("evalChDirStatement(%s)\n", path)

	return fileserv.SetCWD(fp, env)
}

// clear all variables and close all files
func evalClearCommand(clear *ast.ClearCommand, code *ast.Code, env *object.Environment) {
	env.ClearVars() // environment handles all the details
	env.ClearFiles()
	env.ClearCommon()
}

// just tell the termianl to clear the screen
func evalClsStatement(cls *ast.ClsStatement, code *ast.Code, env *object.Environment) {
	env.Terminal().Cls()
}

// change screen foreground/background color
func evalColorStatement(color *ast.ColorStatement, code *ast.Code, env *object.Environment) object.Object {
	// get the current screen mode
	scr := evalColorMode(env)

	plt := evalColorPalette(scr, env)

	// TODO this will one day support more modes
	switch scr.Settings[ast.ScrnMode] {
	case ast.ScrnModeMDA, ast.ScrnModeCGA: // Text mode only
		return evalColorScreen0(color, plt.CurPalette, code, env)
	default:
		return object.StdError(env, berrors.IllegalFuncCallErr)
	}
}

func evalColorMode(env *object.Environment) *ast.ScreenStatement {
	// what is the current screen mode?
	ss := env.GetSetting(settings.Screen)

	// if I didn't get one from the settings
	// need to create a default one
	if ss == nil {
		t := &ast.ScreenStatement{}
		t.InitValue()
		return t
	}

	// unwrap the actual statement
	return ss.(*ast.ScreenStatement)
}

// for screen mode 0, the three parameters are foreground, background, border
// ToDo: actually support border color if I ever see it used
func evalColorScreen0(color *ast.ColorStatement, plt ast.ColorPalette, code *ast.Code, env *object.Environment) object.Object {
	// reset to normal video mode
	// TODO preserve foreground color
	env.Terminal().Print("\x1B[0m")
	// go evaluate all my parameters
	parms := evalExpressions(color.Parms, code, env)

	// apply all them parameters
	for i, p := range parms {
		// I ignore border colors for now
		if i == 2 {
			return nil
		}

		// convert it to an int16
		ind, err := coerceIndex(p, env)

		if err != nil {
			return err
		}

		// currently only does foreground and background
		evalColorSet((ind & 0x0f), (i == 1), plt, env)
	}

	return nil
}

// use the map to calculate  final output
func evalColorSet(color int16, bkGrnd bool, plt ast.ColorPalette, env *object.Environment) {
	rc := plt[color]

	if bkGrnd {
		rc += 10 // move into background range
	}

	// set the color
	if rc != 40 {
		env.Terminal().Print(fmt.Sprintf("\x1B[%dm", rc))
	}
}

func evalColorPalette(scr *ast.ScreenStatement, env *object.Environment) *ast.PaletteStatement {
	// fetch the current palette settings
	pset := env.GetSetting(settings.Palette)

	// if no settings are saved, get the default values
	if pset == nil {
		// go build default palette and save it
		plt := evalPaletteDefault(scr.Settings[0])
		env.SaveSetting(settings.Palette, plt)
		return plt
	}

	return pset.(*ast.PaletteStatement)
}

// common statement allows data to survive across a chain
func evalCommonStatement(com *ast.CommonStatement, code *ast.Code, env *object.Environment) {
	for _, id := range com.Vars {
		env.Common(id.Value)
	}
}

// user wants to continue execution, if we can
func evalContCommand(cont *ast.ContCommand, code *ast.Code, env *object.Environment) object.Object {
	// if a program is currently running, you can't use this command
	if env.ProgramRunning() {
		return object.StdError(env, berrors.CantContinue)
	}

	// see if there is a saved continuation point
	np := env.GetSetting(settings.Restart)

	if np == nil {
		return object.StdError(env, berrors.CantContinue)
	}

	// recover the ast.Code object
	cd := np.(*ast.Code)

	if cd == nil {
		return object.StdError(env, berrors.CantContinue)
	}

	// move the code iterator to the continuation point

	return evalContStart(cd, env)
}

func evalContStart(code *ast.Code, env *object.Environment) object.Object {
	// see if I should move to the next statement
	evalContChkInput(code)

	env.SetRun(true)
	rc := evalStatements(code, env)
	env.SetRun(false)
	return rc
}

// skips moving to the next statement if current statement is an Input statement or function
// Input functions will re-prompt and then accept input
func evalContChkInput(code *ast.Code) {
	switch code.Value() {
	default:
		code.Next()
	}
}

// return an integer that tells the current line # for the cursor
func evalCsrLinExpression(code *ast.Code, env *object.Environment) object.Object {
	res := object.Integer{}

	row, _ := env.Terminal().GetCursor()
	res.Value = int16(row) + 1

	return &res
}

func evalStatements(code *ast.Code, env *object.Environment) object.Object {
	var rc object.Object

	// make sure there are statements to evaluate
	t := code.Len()
	ok := t > 0
	// loop until you run out of code
	for halt := false; ok && !halt; {
		if code.Value() != nil {
			rc = Eval(code.Value(), code, env)
		} else {
			rc = object.StdError(env, berrors.Syntax)
		}

		// Eval should *almost* always return nil
		// the exceptions are:
		// RESTART - user has entered a CONTinue command
		// HALT - user hit CTRL-C to stop execution
		// ERROR - a runtime error has occurred

		if rc != nil {
			halt, code, rc = evalStatementResult(rc, code, env)

			if !halt {
				halt = !code.Next()
			}
		} else {
			if env.Terminal().BreakCheck() {
				rc = evalStatementsBreakChk(code, env)
				halt = true
			} else {
				halt = !code.Next()
			}
		}
	}

	return rc
}

// figure out three things:
// should execution stop (bool)
// did I get a CTRL-C that the eval loop didn't see? rc == "HALT" (this may be redundant)
// if I should keep going, where should I start from
// .
// This is *really* clunky code, I should rewrite once I'm smarter
func evalStatementResult(rc object.Object, code *ast.Code, env *object.Environment) (bool, *ast.Code, object.Object) {

	halt := false

	switch rc.Type() {
	case object.ObjectType("RESTART"):
		code = env.StatementIter()
		halt = (code.Len() == 0)
	case object.ObjectType("ERROR"):
		// but wait!  Is there an ON ERROR rule in place
		halt = evalErrorHandler(code, env)
	case object.ObjectType("HALT"):
		halt = true
		env.SaveSetting(settings.Restart, code)
		rc = nil
	default:
		halt = true
		rc = object.StdError(env, berrors.Syntax)
	}

	return halt, code, rc
}

// got an error, is an ON ERROR rule in place
func evalErrorHandler(code *ast.Code, env *object.Environment) bool {
	// if the setting doesn't exist, he returns a turned off value (zero)
	oer, _ := env.GetSetting(settings.OnError).(*ast.OnErrorGoto)

	if (oer == nil) || (oer.Jump == 0) {
		// no rule in place
		return true
	}

	// save a resume point
	env.SaveSetting(settings.Restart, &ast.ResumeStatement{ResmPt: code.GetReturnPoint()})
	// jump to the defined error handler
	code.Jump(oer.Jump)

	return false
}

// check for a user break - Ctrl-C, returns a halt if it was seen
func evalStatementsBreakChk(code *ast.Code, env *object.Environment) object.Object {
	/*	if !env.Terminal().BreakCheck() {
		return nil
	}*/
	msg := "Break"

	if env.ProgramRunning() {
		msg = fmt.Sprintf("%s in line %d", msg, code.CurLine())
	}

	hlt := object.HaltSignal{Msg: msg}
	env.SaveSetting(settings.Restart, code)

	return &hlt
}

// read constant values out of data statements into variables
func evalReadStatement(rd *ast.ReadStatement, code *ast.Code, env *object.Environment) object.Object {
	var value object.Object

	// if no vars, that's a problem
	if len(rd.Vars) == 0 {
		return object.StdError(env, berrors.Syntax)
	}

	for _, item := range rd.Vars {
		name, ok := item.(*ast.Identifier)

		if !ok {
			return object.StdError(env, berrors.Syntax)
		}

		cst := env.ConstData().Next()

		if cst == nil {
			return object.StdError(env, berrors.OutOfData)
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

		rc := saveVariable(code, env, name, value)
		if rc != nil {
			return rc
		}
	}
	return nil
}

// evalRestoreStatement makes sure you can re-read data statements
func evalRestoreStatement(rst *ast.RestoreStatement, code *ast.Code, env *object.Environment) object.Object {
	if rst.Line >= 0 {
		// he wants to restore to a certain line
		if env.ConstData().RestoreTo(rst.Line) {
			return nil
		}

		return object.StdError(env, berrors.UnDefinedLineNumber)
	}

	// restore to the beginning
	env.ConstData().Restore()

	return nil
}

// user wants to resume after an ON ERROR routine
func evalResumeStatement(res *ast.ResumeStatement, code *ast.Code, env *object.Environment) object.Object {
	if len(res.ResmDir) > 1 {
		// too many directives, error
		return evalResumeError(berrors.Syntax, env)
	}

	// get the restart setting
	oe := env.GetSetting(settings.Restart)
	oer, _ := oe.(*ast.ResumeStatement)
	rtp := oer.ResmPt

	// if no direction given, go back to statement that caused the error and hope
	// some day I should figure out how to unit test this
	if len(res.ResmDir) == 0 {
		code.JumpBeforeRetPoint(rtp)
		return nil
	}

	// There is some kind of directive telling me where to resume, figure it out
	switch dir := res.ResmDir[0].(type) {
	case *ast.Identifier:
		if strings.EqualFold(dir.Value, "NEXT") {
			code.JumpToRetPoint(rtp)
			return nil
		} else {
			return evalResumeError(berrors.Syntax, env)
		}
	case *ast.DblIntegerLiteral:
		if dir.Value == 0 {
			code.JumpBeforeRetPoint(rtp)
			return nil
		}
		if dir.Value > 0 {
			if code.Exists(int(dir.Value)) {
				code.Jump(int(dir.Value))
				return nil
			} else {
				return evalResumeError(berrors.UnDefinedLineNumber, env)
			}
		}

	}

	// if I fall out of the switch, I didn't recognize the directive
	return evalResumeError(berrors.Syntax, env)
}

// got an error in the error handler
func evalResumeError(err int, env *object.Environment) object.Object {
	// have to turn off error handle or I will spin forever
	env.ClrSetting(settings.OnError)

	return object.StdError(env, err)
}

// evalReturnStatement gets you back to where the sub-routine was called
// alternatively, allows you to recover from an event trap <- ToDo
func evalReturnStatement(ret *ast.ReturnStatement, code *ast.Code, env *object.Environment) object.Object {
	// get code iterator pointing to where I need to be
	rt := env.Pop()

	// check for stack underflow
	if rt == nil {
		return object.StdError(env, berrors.ReturnWoGosub)
	}

	code.JumpToRetPoint(*rt)
	return nil
}

// actually run the program
// ToDo: close open data files (as soon as I support data files)
func evalRunCommand(run *ast.RunCommand, code *ast.Code, env *object.Environment) object.Object {
	if run.LoadFile != nil {
		// load the source file then run it
		return evalRunLoad(run, code, env)
	}

	return evalRunCheckStartLineNum(run, env)
}

// pull the file down from the server
func evalRunLoad(run *ast.RunCommand, code *ast.Code, env *object.Environment) object.Object {
	val := Eval(run.LoadFile, code, env)

	fname, ok := val.(*object.String)

	if !ok {
		object.StdError(env, berrors.TypeMismatch)
	}

	return evalRunFetch(fname.Value, run, env)
}

func evalRunFetch(file string, run *ast.RunCommand, env *object.Environment) object.Object {
	rdr, err := fileserv.GetFile(file, env)

	if err != nil {
		object.StdError(env, berrors.Syntax)
		return nil
	}
	return evalRunParse(rdr, run, env)
}

// Parse the code in the reader try to run it.
func evalRunParse(rdr *bufio.Reader, run *ast.RunCommand, env *object.Environment) object.Object {
	// create a new program code space
	env.NewProgram()
	if !run.KeepOpen {
		env.ClearFiles()
	}

	// parse the loaded file into an AST for evaluation
	fileserv.ParseFile(rdr, env)

	return evalRunCheckStartLineNum(run, env)
}

func evalRunCheckStartLineNum(run *ast.RunCommand, env *object.Environment) object.Object {
	//	env.Terminal().Println("evalRunCheckStartLineNum")
	pcode := env.StatementIter()
	env.ConstData().Restore()

	if run.StartLine > 0 {
		err := pcode.Jump(run.StartLine)

		if err > 0 {
			return object.StdError(env, err)
		}
	}

	return evalRunStart(pcode, env)
}

// actually go execute the code
func evalRunStart(code *ast.Code, env *object.Environment) object.Object {
	env.SetRun(true)
	rc := Eval(&ast.Program{}, code, env)
	env.SetRun(false)

	return rc
}

// set the screen mode
func evalScreenStatement(scrn *ast.ScreenStatement, code *ast.Code, env *object.Environment) object.Object {
	// get the current settings object
	cur := evalScreenGetCurrent(env)

	// apply any settings in the statement to the current settings
	for i := range scrn.Params {
		if scrn.Params[i] == nil {
			continue
		}
		id, err := coerceIndex(Eval(scrn.Params[i], code, env), env)

		if err != nil {
			return err
		}

		// only valid values are 0,1,2,7,8,9,10
		if (id < 0) || ((id > 2) && (id < 7)) || (id > 10) {
			return object.StdError(env, berrors.IllegalFuncCallErr)
		}

		cur.Settings[i] = int(id)
	}

	// save the new SCREEN settings
	env.SaveSetting(settings.Screen, cur)
	return nil
}

// get the current setting, if it exists
func evalScreenGetCurrent(env *object.Environment) *ast.ScreenStatement {
	cur := env.GetSetting(settings.Screen)

	// if no current setting, return default
	if cur == nil {
		return evalScreenDefaults()
	}

	return cur.(*ast.ScreenStatement)
}

// build and return the default, power-on settings
func evalScreenDefaults() *ast.ScreenStatement {
	return &ast.ScreenStatement{Settings: [4]int{0, 1, 0, 0}}
}

// halt execution, if running, leave file opens, tell user where we are
func evalStopStatement(stop *ast.StopStatement, code *ast.Code, env *object.Environment) object.Object {
	msg := "Break"

	if env.ProgramRunning() {
		msg = fmt.Sprintf("%s in line %d", msg, code.CurLine())
	}

	halt := object.HaltSignal{Msg: msg}
	return &halt
}

// turn off tracing
func evalTroffCommand(env *object.Environment) {
	env.SetTrace(false)
}

// turn on tracing
func evalTronCommand(env *object.Environment) {
	env.SetTrace(true)
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
		id, err := coerceIndex(d, env)

		if err != nil {
			return nil
		}

		i = &object.Integer{Value: id}
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

// stop execution and close any open files
func evalEndStatement(end *ast.EndStatement, code *ast.Code, env *object.Environment) object.Object {
	env.ClearFiles()
	return &object.HaltSignal{}
}

// ERROR statement, user wants to signal an error has occurred
func evalErrorStatement(ers *ast.ErrorStatement, code *ast.Code, env *object.Environment) object.Object {
	rc := evalExpressionNode(ers.ErrNum, code, env)

	// got to catch any potential errors from evaluating the expression
	switch val := rc.(type) {
	case *object.Error:
		return val
	default:
		errnum, err := coerceDblInteger(val, env)

		if err != nil {
			return err
		}

		return evalErrorStatementNumber(errnum, env)
	}
}

// decide if I got valid error number
func evalErrorStatementNumber(errnum int32, env *object.Environment) object.Object {
	// first check the range
	if (errnum < 0) || (errnum > 255) {
		return object.StdError(env, berrors.Syntax)
	}

	return object.StdError(env, int(errnum))
}

// FILES instruct the system to list filenames for current directory
// FILES "path" lists all files in the specified directory
func evalFilesCommand(files *ast.FilesCommand, code *ast.Code, env *object.Environment) object.Object {
	// I should have at most one parameter
	if len(files.Path) > 1 {
		return object.StdError(env, berrors.Syntax)
	}

	// assume no param, get the CWD
	pth := fileserv.GetCWD(env)

	// if there is a param, retrieve the value
	if len(files.Path) > 0 {
		// eval the expression expecting a string to come back
		res := evalExpressionNodeTyped(files.Path[0], code, env, &object.String{})

		// If I don't get a string, we have a syntax error
		if res == nil {
			return object.StdError(env, berrors.Syntax)
		}

		// use the parameter value instead of CWD
		pth = fileserv.BuildFullPath(res.(*object.String).Value, env)
	}

	dir, err := fileserv.GetFile(pth, env)
	if err != nil {
		return err
	}

	list := filelist.NewFileList()
	err = list.Build(dir, env)
	if err != nil {
		env.Terminal().Log("list build failed!")
		//catchNotDir(files.Path, err, env)
		return err
	}

	displayFiles(list, env)

	return nil
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

type forStmtParams struct {
	forBlock object.ForBlock // gets added to the array of nested for loops
	initial  object.Object   // the starting value of the for loop
	stepSign bool            // is the step value positive or negative
}

// FOR statement begins a for-loop
func evalForStatement(four *ast.ForStatment, code *ast.Code, env *object.Environment) object.Object {
	// check for obvious problems
	if (four.Init == nil) || (len(four.Final) == 0) {
		return object.StdError(env, berrors.Syntax)
	}

	// initialize my counter
	rc := Eval(four.Init, code, env)

	if rc != nil {
		return object.StdError(env, berrors.Syntax)
	}

	initial := env.Get(four.Init.Name.Value)

	// need to save off the ForStatement and the current code value
	blk := object.ForBlock{Code: code.GetReturnPoint(), Four: four}

	forStmt := forStmtParams{forBlock: blk, initial: initial}

	return evalForCalcStep(forStmt, code, env)
}

// calculate the sign (+/-) of the step
// need that to ensure initial value doesn't already
// exceed the desired final value
func evalForCalcStep(four forStmtParams, code *ast.Code, env *object.Environment) object.Object {
	// build the default step value of one
	step := []ast.Expression{&ast.IntegerLiteral{Value: 1}}

	// if step is specified, go get it
	if four.forBlock.Four.Step != nil {
		step = four.forBlock.Four.Step
	}

	// eval the step expression and then get sign
	// first time through the loop I ignore zero step
	stp := evalExpressions(step, code, env)
	four.stepSign, _ = evalNextStepSign(stp[0], env)

	return evalForTestSkip(four, code, env)
}

func evalForTestSkip(four forStmtParams, code *ast.Code, env *object.Environment) object.Object {
	_, start := evalNextComplete(four.stepSign, four.initial, four.forBlock.Four, code, env)

	if start {
		return evalForStartLoop(four.forBlock, code, env)
	}

	return evalForSkipLoop(four.forBlock.Four, code, env)
}

//
func evalForStartLoop(fb object.ForBlock, code *ast.Code, env *object.Environment) object.Object {

	// add ForBlock to the list of running for loops
	env.ForLoops = append(env.ForLoops, fb)

	return nil
}

// evalForSkipLoop initial condition exceeds final
// just skip over statements until you find a NEXT
func evalForSkipLoop(four *ast.ForStatment, code *ast.Code, env *object.Environment) object.Object {
	// iterate over the code until we find the next NEXT
	for more := code.Next(); more == true; {
		switch typ := code.Value().(type) {
		case *ast.ForStatment:
			// found an inner FOR loop, skip over it
			rc := evalForSkipLoop(typ, code, env)
			if rc != nil {
				return rc
			}
		case *ast.NextStatement:
			// found a NEXT, but do the variables match?
			if (len(typ.Id.String()) != 0) && !strings.EqualFold(typ.Id.String(), four.Init.Name.Value) {
				// nope, that's a problem
				return object.StdError(env, berrors.NextWithoutFor)
			}
			return nil
		}
		more = code.Next()
	}
	return object.StdError(env, berrors.ForWoNext)
}

// push current code position then jump to new position
func evalGosubStatement(gosub *ast.GosubStatement, code *ast.Code, env *object.Environment) object.Object {
	// should only have one destination
	if len(gosub.Gosub) != 1 || gosub.Gosub[0].Type != token.INT {
		return object.StdError(env, berrors.Syntax)
	}

	line, _ := strconv.Atoi(gosub.Gosub[0].Literal)

	// save the return address and jump to the sub-routine
	env.Push(code.GetReturnPoint())

	if !env.ProgramRunning() {
		return evalGotoStart(line, env)
	}

	err := code.Jump(line)

	if err > 0 {
		return object.StdError(env, err)
	}

	if env.GetTrace() {
		env.Terminal().Print(fmt.Sprintf("[%d]", line))
	}

	return nil
}

// Transfer control to the indicated line number
// If we aren't currently running, get started!
func evalGotoStatement(node *ast.GotoStatement, code *ast.Code, env *object.Environment) object.Object {
	// should only have one destination
	if len(node.JmpTo) != 1 || node.JmpTo[0].Type != token.INT {
		return object.StdError(env, berrors.Syntax)
	}

	line, _ := strconv.Atoi(node.JmpTo[0].Literal)

	if env.ProgramRunning() {
		return evalGotoJump(line, code, env)
	}

	return evalGotoStart(line, env)
}

// we are running, jump to new line
func evalGotoJump(line int, code *ast.Code, env *object.Environment) object.Object {

	err := code.Jump(line)

	if err > 0 {
		return object.StdError(env, err)
	}

	return nil
}

// 'GOTO' entered from command line, start running at target line
func evalGotoStart(line int, env *object.Environment) object.Object {
	code := env.StatementIter()
	err := code.Jump(line)

	// if I get a msg, line wasn't found
	if err > 0 {
		return object.StdError(env, err)
	}

	// go run the program
	env.SetRun(true)
	rc := evalStatements(code, env)
	env.SetRun(false)
	return rc
}

func evalHexConstant(stmt *ast.HexConstant, code *ast.Code, env *object.Environment) object.Object {
	dst, err := strconv.ParseInt(stmt.Value, 16, 16)

	if err != nil {
		st := err.Error()
		if strings.Contains(st, "value out of range") {
			return object.StdError(env, berrors.Overflow)
		}
		return object.StdError(env, berrors.Syntax)
	}

	return &object.Integer{Value: int16(dst)}
}

// appears to be an "Implied" LET statement
func evalImpliedLetStatement(node *ast.Identifier, code *ast.Code, env *object.Environment) object.Object {

	// if it is a builtin function, just return it
	if builtin, ok := builtins.Builtins[node.Value]; ok {
		return builtin
	}

	id := env.Get(node.Value)

	return id
}

// defines, enables, disables and lists keyboard macros
func evalKeyStatement(node *ast.KeyStatement, code *ast.Code, env *object.Environment) object.Object {
	// get the current key definitions
	keyDefs := evalKeyStatementGetKeyDefs(env)

	switch node.Param.(type) {
	case *ast.OnExpression:
		keyDefs.Disp = true
	case *ast.OffExpression:
		keyDefs.Disp = false
	case *ast.ListExpression: // list out the current defined values
		env.Terminal().Print(keyDefs.String())
	default:
		err := evalKeyStatementFKeyParams(keyDefs, node, code, env)
		if err != nil {
			return err
		}
	}

	// save any changes made
	env.SaveSetting(settings.KeyMacs, keyDefs)

	return nil
}

// retrieve or create the current key definitions
func evalKeyStatementGetKeyDefs(env *object.Environment) *ast.KeySettings {
	// create empty settings
	keyDefs := &ast.KeySettings{}
	keyDefs.Keys = make(map[string]string)

	// go get the keyboard macro settings
	keys := env.GetSetting(settings.KeyMacs)

	if keys != nil {
		// settings have been saved before, use saved values
		keyDefs = keys.(*ast.KeySettings)
	}

	return keyDefs
}

// user wants to redefine macro for F1 to F10
func evalKeyStatementFKeyParams(keys *ast.KeySettings, node *ast.KeyStatement, code *ast.Code, env *object.Environment) object.Object {
	// do the error checking first
	// one and only one parameter for this form
	if len(node.Data) != 1 {
		return object.StdError(env, berrors.Syntax)
	}
	ind := evalExpressionNodeTyped(node.Param, code, env, &object.Integer{})

	if ind == nil {
		return object.StdError(env, berrors.Syntax)
	}

	i := ind.(*object.Integer).Value
	// index 1-10 is a keybord function key
	if (i >= 1) && (i <= 10) {
		return evalKeyStatementMapFKey(i, keys, node.Data[0], code, env)
	}
	// index 15-20 are user defined keys for the ON KEY statements
	if (i >= 15) && (i <= 20) {
		return evalKeyStatmentCustomKey(i, keys, node.Data[0], code, env)
	}

	// index is invalid
	return object.StdError(env, berrors.IllegalFuncCallErr)
}

// define a string to insert for a function key
func evalKeyStatementMapFKey(key int16, keys *ast.KeySettings, val ast.Expression, code *ast.Code, env *object.Environment) object.Object {
	s := evalExpressionNodeTyped(val, code, env, &object.String{})

	if s == nil {
		return object.StdError(env, berrors.Syntax)
	}

	k := fmt.Sprintf("F%d", key)
	keys.Keys[k] = s.(*object.String).Value

	return nil
}

func evalKeyStatmentCustomKey(key int16, keys *ast.KeySettings, val ast.Expression, code *ast.Code, env *object.Environment) object.Object {
	b := evalExpressionNodeTyped(val, code, env, &object.String{})

	if b == nil {
		return object.StdError(env, berrors.Syntax)
	}
	return nil
}

// list some or all of the current program
func evalListStatement(stmt *ast.ListStatement, code *ast.Code, env *object.Environment) {
	var out bytes.Buffer
	// get a code iterator
	cd := env.StatementIter()

	// assume my default limits
	start := 0
	stop := cd.MaxLineNum()

	// figure out any limits to the listing
	// is there a starting line?
	if len(stmt.Start) > 0 {
		start, _ = strconv.Atoi(stmt.Start)
		if len(stmt.Lrange) == 0 {
			stop = start
		}
	}

	// is there a stopping point
	if len(stmt.Stop) > 0 {
		stop, _ = strconv.Atoi(stmt.Stop)
	}

	// couple of flags to control the listing loop
	midLine := false // tells me I've printed a line # and the first statement (need to insert colons)
	bList := false   // set true when I see a line # in the printing range

	// roll through lines until I'm done
	for more := true; more; {
		stmt := cd.Value() // fetch the next statment

		// check to see if we are starting a new line
		lnm, ok := stmt.(*ast.LineNumStmt)

		if ok {
			if int(lnm.Value) > stop {
				// I've passed the end line #, I'm done
				break
			}

			// output anything in the buffer from a previous line, if I'm printing yet
			if bList {
				env.Terminal().Println(strings.TrimRight(out.String(), " "))
				out.Truncate(0)
			}
			bList = (int(lnm.Value) >= start)
			midLine = false // just wrote a line number, not in the middle of a line
		}

		if bList {
			if midLine {
				// seperate the statements
				out.WriteString(": ")
			}
			out.WriteString(stmt.String())
			if lnm == nil {
				midLine = true // in a line until your not
			}
		}

		more = cd.Next()
	}
	env.Terminal().Println(out.String())
}

// evalLoadCommand - load and parse the target program
func evalLoadCommand(stmt *ast.LoadCommand, code *ast.Code, env *object.Environment) object.Object {
	// get the target file name
	res := Eval(stmt.Path, code, env)
	str, ok := res.(*object.String)

	if !ok {
		return object.StdError(env, berrors.TypeMismatch)
	}

	return evalLoadGetFile(str.Value, stmt, code, env)
}

// calls the file server looking for a source file
func evalLoadGetFile(file string, stmt *ast.LoadCommand, code *ast.Code, env *object.Environment) object.Object {
	rdr, err := fileserv.GetFile(file, env)

	if err != nil {
		// server sent an error, get out
		return err
	}

	return evalLoadParse(rdr, stmt, code, env)
}

// parse in the loaded file
func evalLoadParse(rdr *bufio.Reader, stmt *ast.LoadCommand, code *ast.Code, env *object.Environment) object.Object {
	// flush the old program
	env.NewProgram()
	fileserv.ParseFile(rdr, env)

	if !stmt.KeppOpen {
		// he does not want to start execution
		return nil
	}

	newCode := env.StatementIter()
	return evalRunStart(newCode, env)
}

// eval where to LOCATE the cursor
func evalLocateStatement(stmt *ast.LocateStatement, code *ast.Code, env *object.Environment) object.Object {
	// check if I have too many parameters or not enough
	if (len(stmt.Parms) > 5) || (len(stmt.Parms) == 0) {
		return object.StdError(env, berrors.Syntax)
	}

	// evaluate movement of cursor
	return evalLocateCursorMove(stmt, code, env)
}

// figure out if/how the cursor needs to move
func evalLocateCursorMove(stmt *ast.LocateStatement, code *ast.Code, env *object.Environment) object.Object {
	row, col := env.Terminal().GetCursor()

	// if no params for new position, I'm done here
	if (stmt.Parms[0] == nil) && (stmt.Parms[1] == nil) {
		return nil
	}

	// calculte new row
	if stmt.Parms[0] != nil {
		nr := evalExpressions(stmt.Parms[0:1], code, env)
		newRow, err := coerceIndex(nr[0], env)

		if err != nil {
			return err
		}
		row = int(newRow)
	}

	// calculte new col
	if (len(stmt.Parms) > 1) && (stmt.Parms[1] != nil) {
		nc := evalExpressions(stmt.Parms[1:2], code, env)
		newCol, err := coerceIndex(nc[0], env)

		if err != nil {
			return err
		}
		col = int(newCol)
	}

	env.Terminal().Locate(row, col)

	return evalLocateCursorShow(stmt, code, env)
}

// check if he set the cursor visibility
func evalLocateCursorShow(stmt *ast.LocateStatement, code *ast.Code, env *object.Environment) object.Object {
	// did he specify the parmeter?
	if (len(stmt.Parms) < 3) || (stmt.Parms[2] == nil) {
		return nil
	}

	return object.StdError(env, berrors.Syntax)
}

// convert a string octal constant into an integer
func evalOctalConstant(stmt *ast.OctalConstant, env *object.Environment) object.Object {

	dst, err := strconv.ParseInt(stmt.Value, 8, 16)

	if err != nil {
		st := err.Error()
		if strings.Contains(st, "value out of range") {
			return object.StdError(env, berrors.Overflow)
		}
		return object.StdError(env, berrors.Syntax)
	}

	return &object.Integer{Value: int16(dst)}
}

// evalNewCommand clears the code space and all the variables
func evalNewCommand(cmd *ast.NewCommand, code *ast.Code, env *object.Environment) object.Object {
	env.NewProgram()
	env.ClearVars()

	// send a halt signal if we are executing a program
	var htl object.HaltSignal

	return &htl
}

// evalNextStatement decides if I should do it all over again
func evalNextStatement(stmt *ast.NextStatement, code *ast.Code, env *object.Environment) object.Object {
	// make sure we are actually in a FOR loop
	if len(env.ForLoops) == 0 {
		return object.StdError(env, berrors.NextWithoutFor)
	}

	blk := env.ForLoops[len(env.ForLoops)-1:]

	// does the next specify the variable?
	if stmt.Id.Token.Literal != "" {
		// make sure they match
		if !strings.EqualFold(stmt.Id.Token.Literal, blk[0].Four.Init.Name.Token.Literal) {
			return object.StdError(env, berrors.NextWithoutFor)
		}
	}

	return evalNextStep(blk[0], code, env)
}

// time to bump the counter by the step value
func evalNextStep(four object.ForBlock, code *ast.Code, env *object.Environment) object.Object {
	// get the counter variable
	cntr := env.Get(four.Four.Init.Name.Token.Literal)

	// counter for loop wasn't saved, how did that happen
	if cntr == nil {
		return object.StdError(env, berrors.InternalErr)
	}

	// get the step value or set it to one if nothing specified
	var step []object.Object
	var pos, zero bool
	if four.Four.Step != nil {
		step = evalExpressions(four.Four.Step, code, env)
		pos, zero = evalNextStepSign(step[0], env)
	} else {
		step = append(step, &object.Integer{Value: 1})
		pos = true
	}

	// if step is zero, stop
	if zero {
		return nil
	}

	// add step to the cntr
	cntr = evalInfixExpression("+", cntr, step[0], env)
	// save off the counter variable
	env.Set(four.Four.Init.Name.Token.Literal, cntr)

	obj, jump := evalNextComplete(pos, cntr, four.Four, code, env)

	if jump {
		// go back to where the four loop started
		code.JumpToRetPoint(four.Code)
		return nil
	}

	return obj
}

// evalNextStepSign returns step>0, step == zero and step value
func evalNextStepSign(stepper object.Object, env *object.Environment) (bool, bool) {
	pos := evalInfixBooleanExpression(">", stepper, &object.Integer{Value: 0}, env)
	neg := evalInfixBooleanExpression("<", stepper, &object.Integer{Value: 0}, env)

	if !pos && !neg {
		return true, true
	}

	return pos, false
}

// return possible err and keep going true/false
func evalNextComplete(pos bool, cntr object.Object, four *ast.ForStatment, code *ast.Code, env *object.Environment) (object.Object, bool) {
	// compute the final value
	fnl := evalExpressions(four.Final, code, env)

	if len(fnl) == 0 {
		return object.StdError(env, berrors.Syntax), false
	}

	_, ok := fnl[0].(*object.Error)
	if ok {
		return fnl[0], false
	}

	// check if counter has passed final value
	var res bool
	if pos {
		res = evalInfixBooleanExpression(">", cntr, fnl[0], env)
	} else {
		res = evalInfixBooleanExpression("<", cntr, fnl[0], env)
	}

	return nil, !res
}

// make sure the line is valid and then save it in the settings
func evalOnErrorStatement(node *ast.OnErrorGoto, code *ast.Code, env *object.Environment) object.Object {
	// make sure the statment is complete
	if (!strings.EqualFold(node.Token.Literal, "ON ERROR GOTO")) || (node.Jump < 0) {
		return object.StdError(env, berrors.Syntax)
	}

	// if the jump to line is zero, he is turning off the error trap
	if node.Jump == 0 {
		env.ClrSetting(settings.OnError)
		return nil
	}

	// make sure the error handler actually exists
	if !code.Exists(node.Jump) {
		return object.StdError(env, berrors.UnDefinedLineNumber)
	}

	// save the node to the settings
	env.SaveSetting(settings.OnError, node)
	return nil
}

// evalOnGoStatement, can be ON x GOTO or ON x GOSUB
func evalOnGoStatement(node *ast.OnGoStatement, code *ast.Code, env *object.Environment) object.Object {
	// make sure I have an expression
	if node.Exp == nil {
		return object.StdError(env, berrors.Syntax)
	}

	// figure out the index
	rc := evalExpressionNode(node.Exp, code, env)

	// did I get an error?
	_, ok := rc.(*object.Error)
	if ok {
		return rc
	}

	// Get the result as an index
	ind, rc := coerceDblInteger(rc, env)

	// if rc is not nil, it is an error
	if rc != nil {
		return rc
	}

	return evalOnGoJump(ind, node, code, env)
}

// figure out where to jump, how to jump then jump
func evalOnGoJump(ind int32, node *ast.OnGoStatement, code *ast.Code, env *object.Environment) object.Object {
	// if index to small/large, just continue
	if (ind <= 0) || (int(ind) > len(node.Jumps)) {
		return nil
	}

	// pull the destination expression
	jmpto := evalExpressionNode(node.Jumps[ind-1], code, env)

	// extract out the line number
	jmp, rc := coerceDblInteger(jmpto, env)

	if rc != nil {
		return rc
	}

	if !code.Exists(int(jmp)) {
		return object.StdError(env, berrors.UnDefinedLineNumber)
	}

	switch node.MidTok.Literal {
	case "GOTO":
		return Eval(&ast.GotoStatement{JmpTo: []token.Token{{Type: token.INT, Literal: strconv.Itoa(int(jmp))}}}, code, env)
	case "GOSUB":
		return Eval(&ast.GosubStatement{Gosub: []token.Token{{Type: token.INT, Literal: strconv.Itoa(int(jmp))}}}, code, env)
	}
	return object.StdError(env, berrors.Syntax)
}

// Build the default Palette struct
func evalPaletteDefault(scrmode int) *ast.PaletteStatement {
	plt := ast.PaletteStatement{}
	plt.BasePalette = make(map[int16]int)

	switch scrmode {
	case 0, 1: // just load the standard colors
		plt.BasePalette[object.GWBlack] = object.XBlack // [0]90
		plt.BasePalette[object.GWBlue] = object.XBlue
		plt.BasePalette[object.GWGreen] = object.XGreen
		plt.BasePalette[object.GWCyan] = object.XCyan
		plt.BasePalette[object.GWRed] = object.XRed
		plt.BasePalette[object.GWMagenta] = object.XMagenta
		plt.BasePalette[object.GWBrown] = object.XYellow
		plt.BasePalette[object.GWWhite] = object.XWhite - 60 // [7]37
		plt.BasePalette[object.GWGray] = object.XBlack - 60  // [0]30
		plt.BasePalette[object.GWLtBlue] = object.XBlue - 60
		plt.BasePalette[object.GWLtGreen] = object.XGreen - 60
		plt.BasePalette[object.GWLtCyan] = object.XCyan - 60
		plt.BasePalette[object.GWLtRed] = object.XRed - 60
		plt.BasePalette[object.GWLtMagenta] = object.XMagenta - 60
		plt.BasePalette[object.GWYellow] = object.XYellow - 60
		plt.BasePalette[object.GWBrtWhite] = object.XWhite // [15]97
	}

	plt.CurPalette = plt.BasePalette

	return &plt
}

// Process parameters of a Print statement
func evalPrintStatement(node *ast.PrintStatement, code *ast.Code, env *object.Environment) object.Object {
	var rc object.Object
	// go print items, if there are any
	if len(node.Items) > 0 {
		rc = evalPrintItems(node, code, env)
	}

	// if I got anything, it is an error
	if rc != nil {
		return rc
	}

	// if last seperator is ; no CR/LF
	if (len(node.Seperators) > 0) && (node.Seperators[len(node.Seperators)-1] == ";") {
		return nil
	}

	// end with a newline
	env.Terminal().Println("")

	return nil
}

// Print the individual items
func evalPrintItems(node *ast.PrintStatement, code *ast.Code, env *object.Environment) object.Object {
	var obj object.Object
	fmt := ""

	for i, item := range node.Items {
		switch node := item.(type) {
		/*case *ast.BuiltinExpression:
		rc := evalBuiltinExpression(node, code, env)
		return rc.Inspect()*/

		case *ast.CallExpression:
			obj = Eval(node, code, env)

		case *ast.Identifier:
			obj = evalPrintIdentifier(node, code, env)
		case *ast.InfixExpression:
			l := evalExpressionNode(node.Left, code, env)
			r := evalExpressionNode(node.Right, code, env)
			obj = evalInfixExpression(node.Operator, l, r, env)
		case *ast.StringLiteral:
			obj = &object.String{Value: node.Value}
		case *ast.IntegerLiteral:
			obj = &object.Integer{Value: node.Value}
		case *ast.DblIntegerLiteral:
			obj = &object.IntDbl{Value: node.Value}
		case *ast.FixedLiteral:
			obj = &object.Fixed{Value: node.Value}
		case *ast.UsingExpression:
			obj = evalUsingExpression(node, code, env)
			form, ok := obj.(*object.String)
			if ok {
				fmt = form.Value
				continue // skip any printing
			}
		}
		_, ok := obj.(*object.Error)

		if ok {
			return obj
		}

		if len(fmt) == 0 {
			evalPrintItemValue(obj, env)
		} else {
			err := evalPrintItemUsing(fmt, obj, env)
			if err != nil {
				return err
			}
		}

		// if seperated by a comma, that means tab
		if node.Seperators[i] == "," {
			env.Terminal().Print("\t")
		}
	}

	return nil
}

// evalPrintItemUsing uses the suppllied format string to Sprintf the object into a string
// and then prints it.
// TODO support more than just numerics
func evalPrintItemUsing(form string, item object.Object, env *object.Environment) object.Object {
	out := fmt.Sprintf("fmt snap %T", item)
	switch val := item.(type) {
	case *object.Integer:
		out = fmt.Sprintf(form, float32(val.Value))
	case *object.IntDbl:
		out = fmt.Sprintf(form, float32(val.Value))
	case *object.Fixed:
		v, _ := val.Value.Float64()
		out = fmt.Sprintf(form, v)
	case *object.FloatSgl:
		out = fmt.Sprintf(form, val.Value)
	case *object.FloatDbl:
		out = fmt.Sprintf(form, val.Value)
	}
	env.Terminal().Print(out)
	return nil
}

// figure out what a print item is, and turn it into a string
func evalPrintItemValue(item object.Object, env *object.Environment) {
	out := fmt.Sprintf("oh snap %T", item)
	switch val := item.(type) {
	case *object.String:
		out = val.Inspect()
	case *object.FloatDbl:
		out = val.Inspect()
	case *object.FloatSgl:
		out = val.Inspect()
	case *object.Fixed:
		out = val.Value.String()
	case *object.Integer:
		out = val.Inspect()
	case *object.IntDbl:
		out = val.Inspect()
	}
	env.Terminal().Print(out)
}

// get the value of the identifier
func evalPrintIdentifier(item *ast.Identifier, code *ast.Code, env *object.Environment) object.Object {
	id := evalIdentifier(item, code, env)
	return id
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
	switch obj := right.(type) {
	case *object.Integer:
		obj.Value = -obj.Value
	case *object.IntDbl:
		obj.Value = -obj.Value
	case *object.FloatDbl:
		obj.Value = -obj.Value
	case *object.FloatSgl:
		obj.Value = -obj.Value
	case *object.Fixed:
		obj.Value = obj.Value.Neg()
	default:
		return object.StdError(env, berrors.Syntax)
	}

	return right
}

// pass parms to evalInfixExpression if result is zero, return false otherwise return true
func evalInfixBooleanExpression(operator string, left, right object.Object, env *object.Environment) bool {
	exp := evalInfixExpression(operator, left, right, env)

	if exp == nil {
		return false
	}

	switch val := exp.(type) {
	case *object.Integer:
		return (val.Value != 0)
	case *object.Fixed:
		return !val.Value.IsZero()

	}
	return true
}

// evaluate an infix expression, returned value type should match either left or right type
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
		if rightVal == 0 {
			return object.StdError(env, berrors.DivByZero)
		}
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
		val, err := left.Div(right)
		if err != 0 {
			return object.StdError(env, err)
		}
		return &object.Fixed{Value: val}
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
		if rightVal == 0 {
			return object.StdError(env, berrors.DivByZero)
		}
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
		if rightVal == 0 {
			return object.StdError(env, berrors.DivByZero)
		}
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

func evalIfStatement(ie *ast.IfStatement, code *ast.Code, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, code, env)
	if isError(condition) {
		return condition
	}

	if condition.(*object.Integer).Value == 0 { // that's a false
		if ie.Alternative == nil {
			return nil // continues at next statement
		}
		return Eval(ie.Alternative, code, env)
	}

	return Eval(ie.Consequence, code, env)
}

// take an array of expressions and evaluate them
func evalExpressions(exps []ast.Expression, code *ast.Code, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		//evaluated := Eval(e, code, env)
		evaluated := evalExpressionNode(e, code, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

// eval an expression looking for a specific type, return nil if wrong type
func evalExpressionNodeTyped(node ast.Node, code *ast.Code, env *object.Environment, want object.Object) object.Object {
	val := evalExpressionNode(node, code, env)

	// see if the type is right
	if val.Type() != want.Type() {
		return nil
	}

	return val
}

// evaluate a single node value
func evalExpressionNode(node ast.Node, code *ast.Code, env *object.Environment) object.Object {
	if node == nil {
		return object.StdError(env, berrors.IllegalFuncCallErr)
	}

	rc := Eval(node, code, env)

	if rc == nil {
		return object.StdError(env, berrors.IllegalFuncCallErr)
	}

	return rc
}

// apply either a user defined function or a builtin function
func applyFunction(fn object.Object, args []object.Object, code *ast.Code, env *object.Environment) object.Object {

	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		obj := Eval(fn.Body, code, extendedEnv)
		return obj

	case *object.Builtin:
		obj := fn.Fn(env, fn, args...)
		return obj

	default:
		return object.StdError(env, berrors.UndefinedFunction)

	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)
	for paramIdx, param := range fn.Parameters {
		env.Set(param.Value, args[paramIdx])
	}
	return env
}

// check if the Identifier has a known value saved in the environment
func evalIdentifier(node *ast.Identifier, code *ast.Code, env *object.Environment) object.Object {

	val := env.Get(node.Value)

	// if it isn't an array, it is the value
	if node.Value[len(node.Value)-1] != ']' {
		return val
	}

	// if there is no index into the array, that's an error
	if node.Index == nil {
		return object.StdError(env, berrors.Syntax)
	}

	// evaluate the index and return it
	return evalIndexArray(node.Index, val, nil, code, env)
}

// evaluate the expression to index into array and save newVal
// if the len(index) recurse down into array until you get to the last index value
// once you find it, if newVal is nil return the current value
// if newVal is not nil, push newVal into correct element and get out

// original caller will save the whole mess back to the environment
func evalIndexArray(index []*ast.IndexExpression, array, newVal object.Object, code *ast.Code, env *object.Environment) object.Object {

	// get the first index value
	indObj := Eval(index[0].Index, code, env)
	if isError(indObj) {
		return indObj
	}

	// coerce the index into an int 16
	ind, err := coerceIndex(indObj, env)

	if err != nil {
		return object.StdError(env, berrors.Syntax)
	}

	// TODO array indices in BASIC can be based at zero as well!
	ind -= 1

	vals, ok := array.(*object.Array)

	if !ok {
		return object.StdError(env, berrors.Syntax)
	}

	// check if their are more dimensions to the array
	if len(index) > 1 {
		return evalIndexArray(index[1:], vals.Elements[ind], newVal, code, env)
	}

	if (ind < 0) || (int(ind) >= len(vals.Elements)) {
		return object.StdError(env, berrors.Syntax)
	}

	if newVal != nil {
		vals.Elements[ind] = newVal
	}
	return vals.Elements[ind]
}

// saveVariable into the environment
func saveVariable(code *ast.Code, env *object.Environment, name *ast.Identifier, val object.Object) object.Object {
	sname := name.Value

	typeid, isarray := parseVarName(sname)

	if !checkTypes(typeid, val) {
		return object.StdError(env, berrors.Syntax)
	}

	cv := env.Get(sname)

	// if not dealing with an array, just save the new value
	if !isarray {
		env.Set(sname, val)
		return nil
	}

	cvarray, ok := cv.(*object.Array)

	if ok && checkTypes(typeid, val) {
		val := evalIndexArray(name.Index, cvarray, val, code, env)

		_, ok := val.(*object.Error)

		if ok {
			return val
		}
	}

	env.Set(sname, cv)
	return nil
}

// coerce a idx value into an int16 array index if you can
// since originally writing this I have co-opted it for use in other
// situations, ie: parameters to basic statements
func coerceIndex(idx object.Object, env *object.Environment) (int16, object.Object) {
	switch fx := idx.(type) {
	case *object.Integer:
		return fx.Value, nil
	case *object.Fixed:
		fx2 := fx.Value.Round(0)
		ti := fx2.IntPart()
		return int16(ti), nil
	case *object.FloatSgl:
		return int16(math.Round(float64(fx.Value))), nil
	case *object.FloatDbl:
		return int16(math.Round(fx.Value)), nil
	}

	return 0, object.StdError(env, berrors.Syntax)
}

// coerce a numeric value into an int32 DblInteger value
func coerceDblInteger(idx object.Object, env *object.Environment) (int32, object.Object) {
	switch fx := idx.(type) {
	case *object.Integer:
		return int32(fx.Value), nil
	case *object.Fixed:
		fx2 := fx.Value.Round(0)
		ti := fx2.IntPart()
		return int32(ti), nil
	case *object.FloatSgl:
		return int32(math.Round(float64(fx.Value))), nil
	case *object.FloatDbl:
		return int32(math.Round(fx.Value)), nil
	}

	return 0, object.StdError(env, berrors.Syntax)
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
	case object.INTEGER_DBL:
		if strings.Contains(typeid, "!") || (typeid == "") {
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
	tk := env.Get(token.LINENUM)

	if tk != nil {
		msg += fmt.Sprintf(" in %d", tk.(*object.IntDbl).Value)
	}

	return &object.Error{Message: msg}
}

// return true if object is of type want
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

// evalUsingExpression and return string object with format string
func evalUsingExpression(stmt *ast.UsingExpression, code *ast.Code, env *object.Environment) object.Object {
	frm := evalExpressionNode(stmt.Format, code, env)
	inp := ""
	switch obj := frm.(type) {
	case *object.String:
		inp = obj.Value
	default:
		return object.StdError(env, berrors.Syntax)
	}
	l := lexer.New(inp)
	p := parser.New(l)
	fmt := p.ParseUsingRunTime()

	return &object.String{Value: fmt}
}

func evalViewPrintStatement(stmt *ast.ViewPrintStatement, code *ast.Code, env *object.Environment) object.Object {
	// if no params, that means I should clear whatever portal is set
	if len(stmt.Parms) == 0 {
		// reset to full
		evalViewPrintOff(env)
		return nil
	}

	// did I get all three parameters
	if len(stmt.Parms) == 3 {
		// quick syntax check
		_, ok := stmt.Parms[1].(*ast.ToStatement)

		if !ok {
			return object.StdError(env, berrors.Syntax)
		}
		return evalViewPrintOn(stmt, code, env)
	}

	return object.StdError(env, berrors.MissingOp)
}

// clears any output limits
func evalViewPrintOff(env *object.Environment) {
	// the xtermjs sequence is `CSI Ps;Ps r` size is [top;bottom]  CSI is `ESC[`
	env.Terminal().Print("\x1b[1;24r")
}

// going to turn ON a view range, get the start and end values
func evalViewPrintOn(stmt *ast.ViewPrintStatement, code *ast.Code, env *object.Environment) object.Object {

	// now eval the two expressions
	// low value first
	//low := evalExpressionNode(stmt.Parms[0], code, env)
	low, rc := coerceIndex(evalExpressionNode(stmt.Parms[0], code, env), env)

	if rc != nil {
		return rc
	}

	// then the high value
	high, rc := coerceIndex(evalExpressionNode(stmt.Parms[2], code, env), env)

	return evalViewPrintRange(low, high, env)
}

// check the view range for validaty
func evalViewPrintRange(low, high int16, env *object.Environment) object.Object {
	// bounds check the values
	if (low < 1) || (high < 1) || (low > 25) || (high > 25) || (low >= high) {
		return object.StdError(env, berrors.Syntax)
	}

	// were good, set the view port
	// the xtermjs sequence is `CSI Ps;Ps r` size is [top;bottom]  CSI is `ESC[`
	cmd := fmt.Sprintf("\x1b[%d;%dr", low, high)
	env.Terminal().Print(cmd)

	return nil
}

func evalViewStatement(stmt *ast.ViewStatement, code *ast.Code, env *object.Environment) object.Object {
	return nil
}
