package afile

import (
	"strings"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/gwtypes"
	"github.com/navionguy/basicwasm/object"
)

// sharedFile is in Shared mode
type sharedFile struct {
	filename   string             // the fully qualified (drive:path/filename.ext) for the file
	accessmode gwtypes.AccessMode // access mode for the file
}

func (sf *sharedFile) AccessMode() gwtypes.AccessMode { return sf.accessmode }
func (sf *sharedFile) FQFN() string                   { return sf.filename }
func (sf *sharedFile) LockMode() gwtypes.LockMode     { return gwtypes.Shared }

// lockReadFile denies read access
type lockReadFile struct {
	filename   string             // the fully qualified (drive:path/filename.ext) for the file
	accessmode gwtypes.AccessMode // access mode for the file
}

func (lrf *lockReadFile) FQFN() string                   { return lrf.filename }
func (lrf *lockReadFile) AccessMode() gwtypes.AccessMode { return lrf.accessmode }
func (lrf *lockReadFile) LockMode() gwtypes.LockMode     { return gwtypes.LockRead }

// lockWriteFile denies write access to the file
type lockWriteFile struct {
	filename   string             // the fully qualified (drive:path/filename.ext) for the file
	accessmode gwtypes.AccessMode // access mode for the file
}

func (lwf *lockWriteFile) FQFN() string                   { return lwf.filename }
func (lwf *lockWriteFile) AccessMode() gwtypes.AccessMode { return lwf.accessmode }
func (lwf *lockWriteFile) LockMode() gwtypes.LockMode     { return gwtypes.LockWrite }

// lockReadWriteFile is the deny all mode
type lockReadWriteFile struct {
	filename   string             // the fully qualified (drive:path/filename.ext) for the file
	accessmode gwtypes.AccessMode // access mode for the file
}

func (lrwf *lockReadWriteFile) AccessMode() gwtypes.AccessMode { return lrwf.accessmode }
func (lrwf *lockReadWriteFile) FQFN() string                   { return lrwf.filename }
func (lrwf *lockReadWriteFile) LockMode() gwtypes.LockMode     { return gwtypes.LockReadWrite }

// defaultFile is in the deny all mode
type defaultFile struct {
	filename   string             // the fully qualified (drive:path/filename.ext) for the file
	accessmode gwtypes.AccessMode // access mode for the file
}

func (df *defaultFile) AccessMode() gwtypes.AccessMode { return df.accessmode }
func (df *defaultFile) FQFN() string                   { return df.filename }
func (df *defaultFile) LockMode() gwtypes.LockMode     { return gwtypes.Default }

// OpenFile starts the process of opening a file for access by the program
// being executed.  First step, create the correct type of AnOpenFile based
// on the share mode being requested.
func OpenFile(FQFN string, stmt ast.OpenStatement, env *object.Environment) (gwtypes.AnOpenFile, object.Object) {
	var af gwtypes.AnOpenFile

	switch strings.ToUpper(stmt.Mode) {
	case gwtypes.Shared.String():
		af = &sharedFile{filename: FQFN}
	case gwtypes.LockRead.String():
		af = &lockReadFile{filename: FQFN}
	case gwtypes.LockWrite.String():
		af = &lockWriteFile{filename: FQFN}
	case gwtypes.LockReadWrite.String():
		af = &lockReadWriteFile{filename: FQFN}
	default:
		af = &defaultFile{filename: FQFN}
	}

	// next step, see if file is already open
	return checkFileAlreadyOpen(af, env)
}

// checkFileAlreadyOpen determines if the file is already open in a mode
// that conflicts with this request
func checkFileAlreadyOpen(AnOpenFile gwtypes.AnOpenFile, env *object.Environment) (gwtypes.AnOpenFile, object.Object) {

	// check and see if this file is already open
	// gets back a list of open file handles
	fl := env.FindOpenFiles(AnOpenFile.FQFN())

	if len(fl) == 0 {
		// if no open handles, no conflict
		return AnOpenFile, nil
	}

	return checkForLockConflict(AnOpenFile, fl, env)
}

// checkForLockConflict
// the file being opened is already open, check to ensure the lock mode doesn't
// block opening in the new mode
func checkForLockConflict(AnOpenFile gwtypes.AnOpenFile, files []gwtypes.AnOpenFile, env *object.Environment) (gwtypes.AnOpenFile, object.Object) {

	for _, af := range files {
		switch af.(type) {
		case *lockReadFile:
			_, conf1 := AnOpenFile.(*lockReadFile)
			_, conf2 := AnOpenFile.(*lockReadWriteFile)
			if conf1 || conf2 {
				return nil, object.StdError(env, berrors.PermissionDenied)
			}
		case *lockWriteFile:
			_, conf1 := AnOpenFile.(*lockWriteFile)
			_, conf2 := AnOpenFile.(*lockReadWriteFile)
			if conf1 || conf2 {
				return nil, object.StdError(env, berrors.PermissionDenied)
			}
		case *lockReadWriteFile:
			// a file open in LOCK READ Write mode cannot be opened in any other mode
			return nil, object.StdError(env, berrors.PermissionDenied)
		case *sharedFile:
			_, conf1 := AnOpenFile.(*defaultFile)
			if conf1 {
				return nil, object.StdError(env, berrors.PermissionDenied)
			}
		case *defaultFile:
			_, conf1 := AnOpenFile.(*lockReadFile)
			_, conf2 := AnOpenFile.(*lockWriteFile)
			_, conf3 := AnOpenFile.(*lockReadWriteFile)
			_, conf4 := AnOpenFile.(*sharedFile)
			if conf1 || conf2 || conf3 || conf4 {
				return nil, object.StdError(env, berrors.PermissionDenied)
			}
		}
	}
	return AnOpenFile, nil
}
