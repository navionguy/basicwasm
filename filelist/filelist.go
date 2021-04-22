package filelist

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// dirEntry holds information about a directory entry from the file system
type dirEntry struct {
	Name   string `json:"name"`
	Subdir bool   `json:"isdir"`
}

// FileList holds the array of entries
type FileList struct {
	Files []dirEntry
}

type fileSorter struct {
	list *FileList
}

// NewFileList builds a new list of files in a directory
func NewFileList() *FileList {
	return &FileList{}
}

// JSON returns the file list formatted in HTML compatable JSON
func (fl *FileList) JSON() []byte {
	if len(fl.Files) == 0 {
		return nil
	}
	res, _ := json.Marshal(fl.Files)
	return res
}

// AddFile takes a directory entry and adds it to the file list
func (fl *FileList) AddFile(file os.FileInfo) {
	nf := dirEntry{Name: file.Name(), Subdir: file.IsDir()}
	fl.Files = append(fl.Files, nf)
}

// Build list takes the json form and builds a full FileList and sorts it
func (fl *FileList) Build(dir *bufio.Reader) error {
	fl.Files = fl.Files[:0]

	jsn, err := ioutil.ReadAll(dir)

	if err != nil {
		return err
	}

	if !json.Valid(jsn) {
		return errors.New("NotDir")
	}

	err = json.Unmarshal(jsn, &fl.Files)
	if err != nil {
		return err
	}
	fs := &fileSorter{list: fl}
	sort.Sort(fs)

	return nil
}

// Len is a part of the sort.Interface
// returns the number file entries
func (fs *fileSorter) Len() int {
	return len(fs.list.Files)
}

// Swap is part of sort.Interface
// change two elements
func (fs *fileSorter) Swap(i, j int) {
	fs.list.Files[i], fs.list.Files[j] = fs.list.Files[j], fs.list.Files[i]
}

// Less is part of sort.Interface
//
func (fs *fileSorter) Less(i, j int) bool {
	if fs.list.Files[i].Subdir && !fs.list.Files[j].Subdir {
		return true
	}

	if !fs.list.Files[i].Subdir && fs.list.Files[j].Subdir {
		return false
	}

	if strings.Compare(fs.list.Files[i].Name, fs.list.Files[j].Name) == -1 {
		return true
	}

	return false
}
