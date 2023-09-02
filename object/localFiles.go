package object

import (
	"bufio"
	"bytes"
)

// in-memory implementation of data files
type aFile struct {
	locked bool         // if locked for exclusive access
	data   bytes.Buffer // the local storage for the file
}

// an instance of an open local file
type oFile struct {
	rdr   bufio.Reader
	wrtr  bufio.Writer
	rdWrt bufio.ReadWriter // supports random access to a file
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

// CheckLocal gets a fully qualified file name
// and checks to see if it exists in memory
func (lf *LocalFiles) CheckLocal(FQFilename string) bool {
	return lf.localFiles[FQFilename] != nil
}

func (lf *LocalFiles) SaveLocal(FQFilename string, rdr *bufio.Reader) {

}

func (lf *LocalFiles) OpenLocalReadOnly(FQFilename string) *bufio.Reader {
	fl := lf.localFiles[FQFilename]

	if fl != nil {
		return nil
	}

	return nil
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
