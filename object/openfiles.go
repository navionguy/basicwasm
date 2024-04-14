package object

import (
	"strings"

	"github.com/navionguy/basicwasm/gwtypes"
)

func (e *Environment) AddOpenFile(num int16, file gwtypes.AnOpenFile) {
	e.files[num] = file
}

// CloseAllFiles closes all open files
func (e *Environment) CloseAllFiles() {
	e.files = make(map[int16]gwtypes.AnOpenFile)
}

// CloseFile closes a file based on its handle
func (e *Environment) CloseFile(f int16) bool {
	if e.files[f] == nil {
		return false
	}
	e.files[f] = nil

	return true
}

// find all the currently open instances of files that
// match the fully qualified file name provided
func (e *Environment) FindOpenFiles(file string) []gwtypes.AnOpenFile {
	var l []gwtypes.AnOpenFile

	for _, af := range e.files {
		if strings.EqualFold(af.FQFN(), file) {
			l = append(l, af)
		}
	}

	return l
}
