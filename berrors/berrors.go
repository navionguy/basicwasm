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
	_
	_
	_
	DeviceIOError
	_
	_
	_ // 60
	_
	_
	_
	_
	_
	_
	_
	_
	_
	PermissionDenied // 70
	_
	_
	_
	_
	_
	PathNotFound
	ServerError
)

// TextForError returns the error text based on error number
func TextForError(err int) string {
	switch err {
	case CantContinue:
		return "Can't continue"
	case DivByZero:
		return "Division by zero"
	case FileNotFound:
		return "File not found"
	case DeviceIOError:
		return "Device I/O Error"
	case IllegalDirect:
		return "Illegal direct"
	case IllegalFuncCallErr:
		return "Illegal function call"
	case NextWithoutFor:
		return "NEXT without FOR"
	case OutOfData:
		return "Out of DATA"
	case Overflow:
		return "Overflow"
	case ReturnWoGosub:
		return "RETURN without GOSUB"
	case Syntax:
		return "Syntax error"
	case TypeMismatch:
		return "Type mismatch"
	case UndefinedFunction:
		return "Undefined user function"
	case UnDefinedLineNumber:
		return "Undefined line number"
	case PermissionDenied:
		return "Permission Denied"
	case PathNotFound:
		return "Path not found"
	case ServerError:
		return "Server error"
	}

	return "Unprintable error"
}
