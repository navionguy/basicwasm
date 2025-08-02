package mocks

import "github.com/navionguy/basicwasm/gwtypes"

type MockFile struct {
	FileName string
	AccMode  gwtypes.AccessMode
	LckMode  gwtypes.LockMode
}

func (maf *MockFile) AccessMode() gwtypes.AccessMode {
	return maf.AccMode
}

func (maf *MockFile) FQFN() string {
	return maf.FileName
}

func (maf *MockFile) LockMode() gwtypes.LockMode {
	return maf.LckMode
}
func MockAnOpenFile(name string) MockFile {
	maf := MockFile{FileName: name}

	return maf
}
