package filelist

import (
	"encoding/json"
	"os"
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

// Build list takes the json form and builds a full FileList
func (fl *FileList) Build(jsn []byte) error {
	fl.Files = fl.Files[:0]

	err := json.Unmarshal(jsn, &fl.Files)
	return err
}
