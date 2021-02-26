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

var (
	assetsDir = flag.String("assets", "./assets/", "web page assets")
	moduleDir = flag.String("webmodules", "./webmodules/", "web assembly file(s)")
	drives    = map[string]*string{
		"driveA": flag.String("driveA", "", ""),
		"driveB": flag.String("driveB", "", ""),
		"driveC": flag.String("driveC", "./source", "current directory on start-up"),
		// TODO: add the rest of the possible drive letter flags
	}

	fileSources = map[string]*fileSource{
		"assets":     {src: http.Dir(*assetsDir)},
		"webmodules": {src: http.Dir(*moduleDir)},
	}
)

// WrapFileSources wrap all the file sources
func WrapFileSources(rtr *mux.Router) {
	for key, drv := range drives {
		if len(*drv) > 0 {
			fileSources[key] = &fileSource{src: http.Dir(*drv)}
			rt := "/" + key + "/"
			rtr.HandleFunc(rt, FileServ).Name(key)
		}
	}
}

// FileServ sends the requested file
func FileServ(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path, "/")
	src := fileSources[parts[1]]

	//
	if (len(parts[1]) == 0) || (src == nil) {
		w.WriteHeader(403)
		return
	}

	vs := mux.Vars(r)

	switch parts[1] {
	case "assets":
		src.filename = fmt.Sprintf("%s/%s/", vs["type"], vs["file"])
		src.setContentType(w)

	case "webmodules":
		src.filename = vs["file"]
	default:
		src.filename = r.URL.Path
	}

	src.ServeFile(w, r, src.filename)
}

func (fs fileSource) setContentType(w http.ResponseWriter) {
	parts := strings.Split(fs.filename, ".")

	if len(parts) == 1 {
		return // just go with text/plain
	}

	ext := parts[len(parts)-1]
	ext = strings.TrimRight(ext, "/")

	switch ext {
	case "js":
		w.Header().Set("Content-Type", "text/javascript")
	case "css":
		w.Header().Set("Content-Type", "text/css")
	}
}

func (fs fileSource) ServeFile(w http.ResponseWriter, r *http.Request, fname string) {
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
		w.WriteHeader(403)
		return
	}

	if st.IsDir() {
		fs.sendDirectory(hfile, w)
		return
	}

	buf := make([]byte, int(st.Size()))
	_, err = hfile.Read(buf)

	if err != nil {
		w.WriteHeader(403)
		return
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
