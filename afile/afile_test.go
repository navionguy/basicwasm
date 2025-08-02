package afile

import (
	"testing"

	"github.com/navionguy/basicwasm/ast"
	"github.com/navionguy/basicwasm/berrors"
	"github.com/navionguy/basicwasm/gwtypes"
	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/stretchr/testify/assert"
)

func TestCheckFileAlreadyOpen(t *testing.T) {
	tests := []struct {
		file   gwtypes.AnOpenFile
		others []gwtypes.AnOpenFile
		error  bool
	}{
		{file: &lockReadFile{filename: "notes.txt"}},
		{file: &lockReadFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&lockReadFile{filename: "notes.txt"}}, error: true},
	}

	for _, tt := range tests {
		var trm mocks.MockTerm
		env := object.NewTermEnvironment(trm)
		for _, fl := range tt.others {
			env.AddOpenFile(5, fl)
		}
		checkFileAlreadyOpen(tt.file, env)
	}
}

func TestCheckFileLockConflict(t *testing.T) {
	tests := []struct {
		file   gwtypes.AnOpenFile
		others []gwtypes.AnOpenFile
		error  bool
	}{
		// check the happy path, no open files by that name
		{file: &lockReadFile{filename: "notes.txt"}},
		{file: &lockWriteFile{filename: "notes.txt"}},
		{file: &lockReadWriteFile{filename: "notes.txt"}},
		{file: &sharedFile{filename: "notes.txt"}},
		{file: &defaultFile{filename: "notes.txt"}},

		// for a lockReadFile, it will error if the file is:
		//		already open in LOCK READ mode
		//		already open in LOCK READ WRITE mode
		// 		already open in default mode
		{file: &lockReadFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&lockReadFile{filename: "notes.txt"}}, error: true},
		{file: &lockReadFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&lockReadWriteFile{filename: "notes.txt"}}, error: true},
		{file: &lockReadFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&defaultFile{filename: "notes.txt"}}, error: true},

		// for a lockWriteFile, it will error if the file is:
		//		already open in LOCK WRITE mode
		//		already open in LOCK READ WRITE mode
		//		already open in default mode
		{file: &lockWriteFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&lockWriteFile{filename: "notes.txt"}}, error: true},
		{file: &lockWriteFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&lockReadWriteFile{filename: "notes.txt"}}, error: true},
		{file: &lockWriteFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&defaultFile{filename: "notes.txt"}}, error: true},

		// for a LockReadWrite
		//		already open in LOCK WRITE mode
		//		already open in LOCK READ WRITE mode
		//		already open in default mode
		{file: &lockReadWriteFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&lockReadWriteFile{filename: "notes.txt"}}, error: true},
		{file: &lockReadWriteFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&lockReadFile{filename: "notes.txt"}}, error: true},
		{file: &lockReadWriteFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&defaultFile{filename: "notes.txt"}}, error: true},

		// for a shared file, it will error if the file is
		//		already open in default mode
		{file: &sharedFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&defaultFile{filename: "notes.txt"}}, error: true},

		// for a default lock mode succeeds if not open in any other mode
		{file: &defaultFile{filename: "notes.txt"}, others: []gwtypes.AnOpenFile{&sharedFile{filename: "notes.txt"}}, error: true},
	}

	// checkForLockFileConflict only returns one kind of error
	stdErr := berrors.TextForError(berrors.PermissionDenied)

	for _, tt := range tests {
		var trm mocks.MockTerm
		env := object.NewTermEnvironment(trm)
		f, err := checkForLockConflict(tt.file, tt.others, env)

		if tt.error {
			assert.Nil(t, f, "mode conflict still returned a file")
			assert.EqualValues(t, stdErr, err.Inspect())
		}
	}
}

func TestFileModes(t *testing.T) {
	tests := []struct {
		file  gwtypes.AnOpenFile
		amode gwtypes.AccessMode
		fqfn  string
		lmode gwtypes.LockMode
	}{
		{file: &sharedFile{filename: "test.dat", accessmode: gwtypes.Append}, amode: gwtypes.Append, fqfn: "test.dat", lmode: gwtypes.Shared},
		{file: &lockReadFile{filename: "test2.dat", accessmode: gwtypes.Input}, amode: gwtypes.Input, fqfn: "test2.dat", lmode: gwtypes.LockRead},
		{file: &lockWriteFile{filename: "test3.dat", accessmode: gwtypes.Output}, amode: gwtypes.Output, fqfn: "test3.dat", lmode: gwtypes.LockWrite},
		{file: &lockReadWriteFile{filename: "test4.dat", accessmode: gwtypes.Random}, amode: gwtypes.Random, fqfn: "test4.dat", lmode: gwtypes.LockReadWrite},
		{file: &defaultFile{filename: "test5.dat", accessmode: gwtypes.Random}, amode: gwtypes.Random, fqfn: "test5.dat", lmode: gwtypes.Default},
	}

	for _, tt := range tests {
		assert.EqualValues(t, tt.fqfn, tt.file.FQFN())
		assert.Equal(t, tt.amode, tt.file.AccessMode())
		assert.EqualValues(t, tt.lmode.String(), tt.file.LockMode().String())
	}

}

func TestOpenFile(t *testing.T) {
	fn := ast.FileNumber{Numbr: &ast.IntegerLiteral{Value: 1}}

	tests := []struct {
		FqFn     string
		OpenStmt ast.OpenStatement
		fail     bool
	}{
		{FqFn: "d:\\data\\users.txt", OpenStmt: ast.OpenStatement{FileName: "USERS.TXT", FileNumber: fn, Mode: "SHARED", Access: "RANDOM"}},
		{FqFn: "d:\\data\\users.txt", OpenStmt: ast.OpenStatement{FileName: "USERS.TXT", FileNumber: fn, Mode: "LOCK READ", Access: "RANDOM"}},
		{FqFn: "d:\\data\\users.txt", OpenStmt: ast.OpenStatement{FileName: "USERS.TXT", FileNumber: fn, Mode: "LOCK WRITE", Access: "RANDOM"}},
		{FqFn: "d:\\data\\users.txt", OpenStmt: ast.OpenStatement{FileName: "USERS.TXT", FileNumber: fn, Mode: "LOCK READ WRITE", Access: "RANDOM"}},
		{FqFn: "d:\\data\\users.txt", OpenStmt: ast.OpenStatement{FileName: "USERS.TXT", FileNumber: fn, Mode: "SHARED", Access: ""}},
		{FqFn: "d:\\data\\users.txt", OpenStmt: ast.OpenStatement{FileName: "USERS.TXT", FileNumber: fn, Mode: "DEFAULT", Access: ""}},
	}

	for _, tt := range tests {
		var trm mocks.MockTerm
		env := object.NewTermEnvironment(trm)
		af, err := OpenFile(tt.FqFn, tt.OpenStmt, env)

		if !tt.fail {
			assert.Equal(t, tt.FqFn, af.FQFN())
		} else {
			assert.Nil(t, af, "OpenFile failed to fail")
			assert.NotNil(t, err, "OpenFile failed to return error")
		}
	}
}
