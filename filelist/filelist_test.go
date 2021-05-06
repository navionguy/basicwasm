package filelist

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockFI struct {
	name  string
	files []string
}

func (mi mockFI) IsDir() bool {
	if len(mi.files) > 1 {
		return true
	}
	return false
}

func (mi mockFI) ModTime() time.Time {
	return time.Now()
}

func (mi mockFI) Mode() os.FileMode {
	return os.ModeDir
}

func (mi mockFI) Name() string {
	return mi.name
}

func (mi mockFI) Size() int64 {
	return int64(len(mi.name))
}

func (mi mockFI) Sys() interface{} {
	return nil
}

var testDir = []mockFI{
	{name: "short.bas"},
	{name: "test.bas"},
	{name: "alongername.bas"},
	{name: "subdir", files: []string{"test1.bas", "test2.bas"}},
}

const marshaledFiles = `[{"name":"short.bas","isdir":false},{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":false},{"name":"subdir","isdir":true}]`

const buildFiles = `[{"name":"test.bas","isdir":false},{"name":"alongername.bas","isdir":true}]`

func Test_FilesJSON(t *testing.T) {
	fl := NewFileList()
	assert.NotNil(t, fl, "Test_FilesJSON fl was nil")

	files := fl.JSON()
	assert.Len(t, files, 0, "Test_FilesJSON files len should have been 0")

	fl.loadTestDir()
	files = fl.JSON()
	assert.Contains(t, string(files), marshaledFiles, "Test_FilesJSON unexpected results")

	fl.Build(bufio.NewReader(bytes.NewReader([]byte(buildFiles))))
	assert.Len(t, fl.Files, 2, "Test_FilesJSON Build sent back %d elements, expected 2", len(fl.Files))
}

func Test_FileSort(t *testing.T) {
	fl := NewFileList()
	fl.loadTestDir()
	fs := &fileSorter{list: fl}

	assert.Len(t, fl.Files, fs.Len(), "Test_FileSort fs.Len() returned %d", fs.Len())

	sort.Sort(fs)

	assert.True(t, fs.list.Files[0].Subdir, "Subdirectory didn't float to the start of the list.")
	assert.EqualValues(t, "alongername.bas", fs.list.Files[1].Name, "Files not alphabetical")

	res := fl.JSON()
	fmt.Println(string(res))
}

func (fl *FileList) loadTestDir() {
	for _, fn := range testDir {
		fl.AddFile(fn)
	}
}
