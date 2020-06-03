package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

// WrapFileOrg shim myself to build the correct path
// to the various assets served up.
func WrapFileOrg() (filesys http.FileSystem) {
	return autoPathingSystem{http.Dir(".")}
}

// takes the requested file and extracts out the basename
// and the extension and returns them.  He does only simple
// checking to make sure they are valid.
func parseName(name string) (basename string, ext string, ok bool) {
	parts := strings.Split(name, "/")

	if parts[len(parts)-1] == "" {
		return "gwbasic", "html", true
	}

	baseparts := strings.Split(parts[len(parts)-1], ".")

	if len(baseparts) == 1 {
		return baseparts[0], "html", true
	}

	if len(baseparts) > 2 {
		base := strings.Join(baseparts[:len(baseparts)-1], ".")
		return base, baseparts[len(baseparts)-1], true
	}

	return baseparts[0], baseparts[1], true
}

// buildPath() builds the full path to the file based on the
// extension and environment variables that tell me where
// things are stored
func buildPath(name string, ext string) (path string, ok bool) {
	fmt.Printf("About to send %s.%s\n", name, ext)
	switch ext {
	case "html":
		return "./assets/html/" + name + "." + ext, true
	case "js", "map":
		fmt.Printf("sending %s.%s", name, ext)
		return "./assets/js/" + name + "." + ext, true

	case "wasm":
		return "./webmodules/" + name + "." + ext, true
	case "ico":
		return "./assets/images/" + name + "." + ext, true
	case "css":
		return "./assets/css/" + name + "." + ext, true
	default:
		fmt.Printf("Request to open %s.%s, no.\n", name, ext)
	}

	return "", false
}

// autoPathingSystem is an http.FileSystem that hides
// my directory structure.
type autoPathingSystem struct {
	http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that builds the actual file name based on his extension and how
// my assets are arranged.
func (fs autoPathingSystem) Open(name string) (hFile http.File, err error) {
	fileName, ext, ok := parseName(name)

	if !ok { // parse failed, return 403 response
		return nil, os.ErrPermission
	}

	fullPath, ok := buildPath(fileName, ext)

	if !ok { // no defined path for file type
		return nil, os.ErrPermission
	}

	file, err := fs.FileSystem.Open(fullPath)
	if err != nil {
		fmt.Printf("file open failed %s\n", fullPath)
		return nil, err
	}
	return file, err
}
