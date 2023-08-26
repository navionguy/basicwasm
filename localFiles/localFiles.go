package localfiles

import (
	"bytes"

	"github.com/navionguy/basicwasm/ast"
)

// in-memory implementation of data files
type aFile struct {
	locked bool
	data   bytes.Buffer // the local storage for the file
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

// Open gets an open statement and tries to open/create the
// file as described by the statement
// ASSUMPTION: the evaluation layer has turned the filename
// as requested into a fully qualified filename
func (lf *LocalFiles) Open(file *ast.OpenStatement) {
	fl := lf.localFiles[file.FQFileName]

	if fl != nil {

	}

	af := aFile{}

	lf.localFiles[file.FQFileName] = &af
}
