package fileserv

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type fileSource struct {
	src      http.FileSystem
	filename string
}

// These are the command line flags that tell where to find runtime resources
var (
	assetsDir = flag.String("assets", "./assets/", "web page assets")
	moduleDir = flag.String("webmodules", "./webmodules/", "web assembly file(s)")
	drives    = map[string]*string{
		"driveA": flag.String("driveA", "", ""),
		"driveB": flag.String("driveB", "", ""),
		"driveC": flag.String("driveC", "./source", "current directory on start-up"),
		// TODO: add the rest of the possible drive letter flags
	}
)

// WrapFileSources builds mux routes to all my resources
// css files, images, javascript files and of course
// the basic interpreter wasm file.
// Then he maps all the drive letters that point to a
// file store.
//
// ToDo: drive the resource mapping from a table
func WrapFileSources(rtr *mux.Router) {
	drv := *assetsDir + "css/"
	fs := &fileSource{src: http.Dir(drv)}
	fs.wrapSource(rtr, "/css/{file}.{ext}", "text/css")

	drv = *assetsDir + "images/"
	fs = &fileSource{src: http.Dir(drv)}
	fs.wrapSource(rtr, "/images/{file}.{ext}", "text/plain")

	drv = *assetsDir + "js/"
	fs = &fileSource{src: http.Dir(drv)}
	fs.wrapSource(rtr, "/js/{file}.{ext}", "application/x-javascript; charset=utf-8")

	drv = *moduleDir
	fs = &fileSource{src: http.Dir(drv)}
	fs.wrapSource(rtr, "/wasm/{file}.{ext}", "application/wasm")

	for key, drv := range drives {
		if len(*drv) > 0 {
			fs := &fileSource{src: http.Dir(*drv)}
			path := "/" + key
			fs.fullyWrapSource(rtr, path)
			fs.wrapSubDirs(rtr, *drv, path)
		}
	}
}

// given a path, create a handler function that will extract the
// parts of the path and then call the source directory to work
// on the file
func (fs *fileSource) wrapSource(rtr *mux.Router, path string, mimetype string) {
	rtr.HandleFunc(path, func(rw http.ResponseWriter, r *http.Request) {
		vs := mux.Vars(r)
		file := vs["file"]
		ext := vs["ext"]

		if len(ext) > 0 {
			file = file + "." + ext
		}
		fs.serveFile(rw, r, file, mimetype)
	}).Name(path)

}

// Since the gorilla mux doesn't support wildcard routes I have to map
// all the possibilities independantly.
// 		http://hostname:port/driveC
// 		http://hostname:port/driveC/
// 		http://hostname:port/driveC/program
// 		http://hostname:port/driveC/program.ext
func (fs *fileSource) fullyWrapSource(rtr *mux.Router, path string) {
	fs.wrapSource(rtr, path, "text/plain; charset=ASCII")
	fs.wrapSource(rtr, path+"/", "text/plain; charset=ASCII")
	fs.wrapSource(rtr, path+"/{file}.{ext}", "text/plain; charset=ASCII")
	fs.wrapSource(rtr, path+"/{file}", "text/plain; charset=ASCII")
}

// After wrapping a directory, I want to wrap any sub-directories
// he might have.
//
func (fs *fileSource) wrapSubDirs(rtr *mux.Router, dir string, path string) {
	hfile, err := fs.src.Open("/")

	// if I can't open him, nothing more to do
	if err != nil {
		return
	}
	defer hfile.Close()

	files, err := hfile.Readdir(-1)

	// he might not be a directory
	if err != nil {
		return
	}

	// he is a directory
	// go wrap any sub directories
	fs.wrapADir(rtr, dir, path, files)

}

// loops through filenames looking for directories
// wraps the directories and then calls wrapSubDirs on them
// to understand recursion, you must understand recursion
//
func (fs fileSource) wrapADir(rtr *mux.Router, dir string, path string, files []os.FileInfo) {
	for _, finfo := range files {
		if containsDotFile(finfo.Name()) {
			continue
		}

		tFile, err := fs.src.Open(finfo.Name())

		if err != nil {
			continue
		}

		info, err := tFile.Stat()

		if err != nil {
			continue
		}
		tFile.Close()

		if !info.IsDir() {
			continue
		}

		fname := info.Name()
		subdir := dir + "/" + fname
		subpath := path + "/" + fname
		nfs := &fileSource{src: http.Dir(subdir)}
		nfs.fullyWrapSource(rtr, subpath)
		nfs.wrapSubDirs(rtr, subdir, subpath)
	}
}

// serveFile opens up the file and sends its contents
//
func (fs fileSource) serveFile(w http.ResponseWriter, r *http.Request, fname string, mimetype string) {
	if len(fname) == 0 {
		fname = "/"
	}

	hfile, err := fs.Open(fname)

	if err != nil {
		w.WriteHeader(404)
		return
	}

	st, err := hfile.Stat()

	if err != nil {
		w.WriteHeader(500)
		return
	}

	if st.IsDir() {
		fs.sendDirectory(hfile, w)
		return
	}

	buf := make([]byte, int(st.Size()))
	_, err = hfile.Read(buf)

	if err != nil {
		w.WriteHeader(503)
		return
	}

	if len(mimetype) > 0 {
		w.Header().Set("Content-Type", mimetype)
	}
	w.Write(buf)

}

// sendDirectory sends all the filenames found in hfile
// he does block any that start with '.'
func (fs fileSource) sendDirectory(hfile http.File, w http.ResponseWriter) {
	files, err := hfile.Readdir(-1)

	if err != nil {
		w.WriteHeader(404)
		return
	}

	for _, finfo := range files {
		if !containsDotFile(finfo.Name()) {
			w.Write([]byte(fmt.Sprintf("<p>%s</p>", finfo.Name())))
		}
	}
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that builds the actual file name based on his extension and how
// my assets are arranged.
func (fs fileSource) Open(name string) (hFile http.File, err error) {
	if containsDotFile(name) { // If dot file, return 403 response
		return nil, os.ErrPermission
	}

	file, err := fs.src.Open(name)
	if err != nil {
		return nil, err
	}

	return dotFileHidingFile{file}, nil

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
	files, err := f.File.Readdir(n)
	for _, file := range files { // Filters out the dot files
		if !strings.HasPrefix(file.Name(), ".") {
			fis = append(fis, file)
		}
	}
	return
}
