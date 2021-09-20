package berrors

const (
	NextWithoutFor = iota + 1
	Syntax
	ReturnWoGosub
	OutOfData
	IllegalFuncCallErr
	Overflow
	OutOfMemory
	UnDefinedLineNumber
	SubscriptRange
	DuplicateDefinition // 10
	DivByZero
	IllegalDirect
	TypeMismatch
	StringSpace
	String2Long
	StringForm2Complex
	CantContinue
	UndefinedFunction
	NoResume
	ResumeWoError //20
	Unprintable
	MissingOp
	LineOverflow
	DeviceTimeout
	DeviceFault
	ForWoNext
	OutOfPaper
	UnprintableErr
	WhileWoWend
	WendWoWhile // 30
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_ // 40
	_
	_
	_
	_
	_
	_
	_
	_
	_
	FieldOverflow // 50
	InternalErr
	BadFileNum
	FileNotFound
)

// TextForError returns the error text based on error number
func TextForError(err int) string {
	switch err {
	case FileNotFound:
		return "File not found"
	case IllegalDirect:
		return "Illegal direct"
	case NextWithoutFor:
		return "NEXT without FOR"
	case OutOfData:
		return "Out of DATA"
	case ReturnWoGosub:
		return "RETURN without GOSUB"
	case Syntax:
		return "Syntax error"
	case TypeMismatch:
		return "Type mismatch"
	case UnDefinedLineNumber:
		return "Undefined line number"
	}

	return "Unprintable error"
}
