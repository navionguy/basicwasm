package object

import (
	"io"

	"github.com/navionguy/basicwasm/berrors"
)

type LockMode int

const (
	SHARED        LockMode = iota // deny none, allows other process all access except default
	LockRead                      // deny read to other processes, fails if already open in default or read access
	LockWrite                     // deny write to other processes, fails if already open in default or write access
	LockReadWrite                 // deny all, fails if already open in any mode
	Default                       // deny all, no other process can access the file, fails if already open
)

// in-memory implementation of data files
type aFile struct {
	locked LockMode // if locked for exclusive access
	data   []byte   // the local storage for the file
}

// an instance of an open local file
type oFile struct {
	file *aFile
}

// implement a read for the oFile
func (ofl *oFile) ReadByte() (byte, error) {
	return 1, nil
}

// write to the open file
func (ofl *oFile) WriteByte(c byte) error {
	return nil
}

// LocalFiles holds all of the data files accessed by programs.
// In this way, if one program creates a data file, and a later
// program accessed it, the intended contents are preserved
type LocalFiles struct {
	openFiles  map[int]*aFile    // maps the file number to the aFile struct
	localFiles map[string]*aFile // maps the FQ filename to an aFile struct
}

// CreateFileStore initializes all of his internal data
func CreateFileStore() *LocalFiles {
	lf := LocalFiles{}

	// create my two maps
	lf.openFiles = make(map[int]*aFile)
	lf.localFiles = make(map[string]*aFile)

	return &lf
}

// Give the execution layer read only access to the files data
func (lf *LocalFiles) OpenLocalReadOnly(FQFilename string, env *Environment) (io.ByteReader, Object) {
	ofl, err := lf.OpenLocal(FQFilename, env)

	if err != nil {
		return nil, err
	}

	return ofl, nil
}

func (lf *LocalFiles) OpenLocalWriteOnly(FQFilename string, env *Environment) (io.ByteWriter, Object) {
	ofl, err := lf.OpenLocal(FQFilename, env)

	if err != nil {
		return nil, err
	}

	return ofl, nil
}

// OpenLocal takes a filename and if successful returns an oFile object
// if not successful, returns an Error object
// if the file is not currently local, it will try to download it from
// our server and push it into the local store
func (lf *LocalFiles) OpenLocal(FQFilename string, env *Environment) (*oFile, Object) {
	fl := lf.localFiles[FQFilename]

	if fl != nil {
		of := oFile{file: fl}
		return &of, nil
	}

	return nil, StdError(env, berrors.FileNotFound)
}

// CloseFile close the file indicated by the file number
// a value of -1 indicates that all files should be closed
// closing a file that is not open, is not an error
func (lf *LocalFiles) CloseFile(filenum int) {
	if filenum == -1 { // check for CloseFile(all)
		lf.openFiles = make(map[int]*aFile)
		return
	}

	lf.openFiles[filenum] = nil
}
