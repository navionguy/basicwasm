package filelist

import (
	"bufio"
	"bytes"
	"fmt"
	"sort"
	"testing"

	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/stretchr/testify/assert"
)

var testDir = []mocks.MockFI{
	{Fname: "short.bas"},
	{Fname: "test.bas"},
	{Fname: "alongername.bas"},
	{Fname: "subdir", Files: []string{"test1.bas", "test2.bas"}},
	{Fname: "aamenu.bas"},
	{Fname: "asubdir", Files: []string{"test2.bas", "test1.bas"}},
}

const marshaledFiles = `[{"name":"short.bas","isdir":false},{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":false},{"name":"subdir","isdir":true},{"name":"aamenu.bas","isdir":false},{"name":"asubdir","isdir":true}]`

const buildFiles = `[{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":true}]`

func Test_Build(t *testing.T) {
	fl := NewFileList()

	mt := &mocks.MockTerm{}
	env := object.NewTermEnvironment(mt)

	obj := fl.Build(bufio.NewReader(bytes.NewReader(nil)), env)

	_, ok := obj.(*object.Error)
	assert.True(t, ok, "Test_Build didn't catch invalid json")
}

func TestBuildError(t *testing.T) {
	rdr := bufio.NewReader(mocks.NewReader(nil))
	fl := NewFileList()

	mt := &mocks.MockTerm{}
	env := object.NewTermEnvironment(mt)

	obj := fl.Build(rdr, env)

	_, ok := obj.(*object.Error)
	assert.True(t, ok, "TestBuildError didn't get an error")

	rdr = bufio.NewReader(mocks.NewReader([]byte(`{"entry":}`)))

	obj = fl.Build(rdr, env)

	_, ok = obj.(*object.Error)
	assert.True(t, ok, "TestBuildError with invalid json didn't get an error")
}

func Test_FilesJSON(t *testing.T) {
	fl := NewFileList()
	assert.NotNil(t, fl, "Test_FilesJSON fl was nil")

	files := fl.JSON()
	assert.Len(t, files, 0, "Test_FilesJSON files len should have been 0")

	fl.loadTestDir()
	files = fl.JSON()
	assert.EqualValuesf(t, string(files), marshaledFiles, "Test_FilesJSON unexpected results, got %s,\n wanted %s", string(files), marshaledFiles)

	mt := &mocks.MockTerm{}
	env := object.NewTermEnvironment(mt)

	fl.Build(bufio.NewReader(bytes.NewReader([]byte(buildFiles))), env)
	assert.Len(t, fl.Files, 2, "Test_FilesJSON Build sent back %d elements, expected 2", len(fl.Files))
}

func Test_FileSort(t *testing.T) {
	sorted := []string{"asubdir", "subdir", "aamenu.bas", "alongername.bas"}
	fl := NewFileList()
	fl.loadTestDir()
	fs := &fileSorter{list: fl}

	assert.Len(t, fl.Files, fs.Len(), "Test_FileSort fs.Len() returned %d", fs.Len())

	sort.Sort(fs)

	assert.True(t, fs.list.Files[0].Subdir, "Subdirectory didn't float to the start of the list.")

	for i, name := range sorted {
		assert.EqualValuesf(t, name, fs.list.Files[i].Name, "Sorted was expecting %s, found %s", name, fs.list.Files[i].Name)
	}
	res := fl.JSON()
	fmt.Println(string(res))
}

func (fl *FileList) loadTestDir() {
	for _, fn := range testDir {
		fl.AddFile(fn)
	}
}
