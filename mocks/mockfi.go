package mocks

import (
	"os"
	"time"
)

type MockFI struct {
	Fname string
	Files []string
	Mom   *MockFS // *may* point to the FileSystem mock that holds this file
}

func (mi MockFI) IsDir() bool {
	if len(mi.Files) > 1 {
		return true
	}
	return false
}

func (mi MockFI) ModTime() time.Time {
	return time.Now()
}

func (mi MockFI) Mode() os.FileMode {
	return os.ModeDir
}

func (mi MockFI) Name() string {
	if mi.Mom != nil {
		mi.Mom.SawName()
	}
	return mi.Fname
}

func (mi MockFI) Size() int64 {
	return int64(len(mi.Fname))
}

func (mi MockFI) Sys() interface{} {
	return nil
}
