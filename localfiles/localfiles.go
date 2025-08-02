package localfiles

import (
	"io"

	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/fileserv"
	"github.com/navionguy/basicwasm/object"
)

// an instance of a local file
type aLocalFile struct {
	FQFilename string  // the name as requested from the server
	readonly   bool    // readonly files will one day be the first to be dropped if running low on memory
	data       *[]byte // the actual data in the file

}

// localFiles holds all of the data files accessed by programs.
// In this way, if one program creates a data file, and a later
// program accesses it, the intended contents are preserved
type localFiles struct {
	dir map[string]*aLocalFile // maps the FQ filename to an AnOpenFile struct
}

var lf localFiles // files stored locally

// support the object.Object interface
func (lf *aLocalFile) Type() object.ObjectType {
	return object.FILE_OBJ
}

func (lf *aLocalFile) Inspect() string {
	return lf.FQFilename
}

// Open is called by the evaluator is trying to open a data file.
func Open(FQFN string, env *object.Environment) object.Object {
	// make sure local file store has been created
	if lf.dir == nil {
		lf.dir = make(map[string]*aLocalFile)
	}

	// check local file map for FQFN
	alf := lf.dir[FQFN]

	// if the local file exists, send it back
	if alf != nil {
		return alf
	}

	return fetchFile(FQFN, env)
}

// fetchFile tries to download the file from the server
func fetchFile(FQFN string, env *object.Environment) object.Object {

	// go request the file from the server
	rdr, err := fileserv.GetFile(FQFN, env)

	// check for an error object coming back
	if err != nil {
		// couldn't get file for some reason
		return err
	}

	fl := make([]byte, rdr.Size())
	lf := aLocalFile{FQFilename: FQFN, readonly: false, data: &fl}

	return &lf
}

// storeFile takes the io.Reader returned from the file request and reads the contents
// into memory.
// NOTE! This is only called for data file requests.  Program file requests are not stored
// since they are consumed once by the parser and then live in the AST.
func storeFile(filename string, file io.Reader, env *object.Environment) object.Object {

	return lf.saveFile(filename, file, env)
}

// saveFile does the work of adding the file to local storage
// I don't check to see if I already have the file stored.
// That shouldn't happen, but if it does, it may signal my copy is old.
func (lf *localFiles) saveFile(filename string, file io.Reader, env *object.Environment) object.Object {

	data, err := io.ReadAll(file)

	if err != nil {
		return object.StdError(env, berrors.DeviceIOError)
	}

	alf := aLocalFile{FQFilename: filename, readonly: true, data: &data}
	lf.dir[filename] = &alf

	return &alf
}
