package fileserv

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var (
	filedir   = flag.String("dir", ".", "directory to serve")
	progFiles = flag.String("progdir", "./source", "")
)

// autoPathingSystem is an http.FileSystem that hides
// my directory structure.
type autoPathingSystem struct {
	http.FileSystem
}

// WrapFileOrg shim myself to build the correct path
// to the various assets served up.
func WrapFileOrg() (filesys http.FileSystem) {

	return autoPathingSystem{http.Dir(*filedir)}
}

// fileRequest holds the parts of a requested file
type fileRequest struct {
	path     string // any directories that should be traversed
	base     string // base filename
	ext      string // any extension specified
	fullSpec string // fully specified path to file
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that builds the actual file name based on his extension and how
// my assets are arranged.
func (fs autoPathingSystem) Open(name string) (hFile http.File, err error) {
	if containsDotFile(name) { // If dot file, return 403 response
		return nil, os.ErrPermission
	}

	fullPath := cleanFileName(name)

	file, err := fs.FileSystem.Open(fullPath)
	if err != nil {
		fmt.Printf("file open failed %s\n", fullPath)
		return nil, err
	}

	return dotFileHidingFile{file}, nil

}

// takes the requested file and extracts out the basename
// and the extension and returns them.  He does only simple
// checking to make sure they are valid.
func cleanFileName(name string) string {
	rq := &fileRequest{}
	parts := strings.Split(name, "/")

	if len(parts) > 1 {
		rq.path = strings.Join(parts[:len(parts)-1], "/")
	}

	if parts[len(parts)-1] == "" {
		rq.base = "gwbasic"
		rq.ext = "html"
	}

	rq.buildBaseName(parts[len(parts)-1])

	rq.buildPath()

	return rq.fullSpec
}

func (rq *fileRequest) buildBaseName(basePart string) {
	if len(basePart) == 0 {
		// blank name gets my default page
		rq.base = "gwbasic"
		rq.ext = "html"
		return
	}

	bparts := strings.Split(basePart, ".")

	switch len(bparts) {
	case 1: // just a name, assume html as extension
		rq.base = basePart
		if strings.Compare(strings.ToLower(rq.base), "gwbasic") == 0 {
			rq.ext = "html"
		}

	case 2: // got base an extension
		rq.base = bparts[0]
		rq.ext = bparts[1]

	default: // multiple extensions?  use only the last one
		rq.base = strings.Join(bparts[:len(bparts)-1], ".")
		rq.ext = bparts[len(bparts)-1]
	}

}

// buildPath() builds the full path to the file based on the
// extension and environment variables that tell me where
// things are stored
func (rq *fileRequest) buildPath() {
	fmt.Printf("About to send %s.%s\n", rq.base, rq.ext)
	switch rq.ext {
	case "html":
		fmt.Printf("got %s-%s\n", rq.base, rq.ext)
		rq.fullSpec = "./assets/html/" + rq.base + "." + rq.ext
	case "js", "map":
		fmt.Printf("sending %s.%s", rq.base, rq.ext)
		rq.fullSpec = "./assets/js/" + rq.base + "." + rq.ext

	case "wasm":
		rq.fullSpec = "./webmodules/" + rq.base + "." + rq.ext
	case "ico":
		rq.fullSpec = "./assets/images/" + rq.base + "." + rq.ext
	case "css":
		rq.fullSpec = "./assets/css/" + rq.base + "." + rq.ext

	default:
		//it isn't one of my special cases, so just build the name
		rq.fullSpec = *progFiles + "/"

		if len(rq.path) > 0 {
			rq.fullSpec = rq.fullSpec + rq.path + "/"
		}

		rq.fullSpec = rq.fullSpec + rq.base

		if len(rq.ext) > 0 {
			rq.fullSpec = rq.fullSpec + "." + rq.ext
		}
	}
}

// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
func containsDotFile(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

// dotFileHidingFile is the http.File use in dotFileHidingFileSystem.
// It is used to wrap the Readdirnames method of http.File so that we can
// remove files and directories that start with a period from its output.
type dotFileHidingFile struct {
	http.File
}

// Readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f dotFileHidingFile) Readdir(n int) (fis []os.FileInfo, err error) {
	fmt.Printf("got request to readdir %d\n", n)
	files, err := f.File.Readdir(n)
	for _, file := range files { // Filters out the dot files
		if !strings.HasPrefix(file.Name(), ".") {
			fis = append(fis, file)
		}
	}
	return
}
