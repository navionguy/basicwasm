package gwtypes

// Access Mode for file i/o
type AccessMode int

const (
	Input  AccessMode = iota // Sequential input mode
	Output                   // Sequential output mode
	Append                   // Position to end of file for writing
	Random                   // Random input/output mode
)

func (am AccessMode) String() string {
	return []string{"INPUT", "OUTPUT", "APPEND", "RANDOM"}[am]
}

// Lock Mode for file i/o
type LockMode int

const (
	Shared        LockMode = iota // deny none, allows other process all access except default
	LockRead                      // deny read to other processes, fails if already open in default or read access
	LockWrite                     // deny write to other processes, fails if already open in default or write access
	LockReadWrite                 // deny all, fails if already open in any mode
	Default                       // deny all, no other process can access the file, fails if already open
)

func (lm LockMode) String() string {
	return []string{"SHARED", "LOCK READ", "LOCK WRITE", "LOCK READ WRITE", "DEFAULT"}[lm]
}

// AnOpenFile holds the data for an open file that lives in the
// in-memory implementation of data files
type AnOpenFile interface {
	AccessMode() AccessMode // the access mode for this open file
	FQFN() string           // the fully qualified (drive:path/filename.ext) for the file
	LockMode() LockMode     // the lock mode for this open file
}
