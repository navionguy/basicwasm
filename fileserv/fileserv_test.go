package fileserv

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type mockFS struct {
	file       string // filename
	statErr    bool   // return an error when stat is called
	readErr    *bool  // return error from read call
	openAlways bool   // return a file handle no matter what

	// desired Readdir results
	names []string
	err   int
}

func (mf mockFS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (mf mockFS) Open(file string) (http.File, error) {
	if (mf.file != file) && !mf.openAlways {
		return nil, fmt.Errorf("got %s, wanted %s", file, mf.file)
	}
	return mf, nil
}

func (mf mockFS) Read(p []byte) (int, error) {

	if *mf.readErr {
		return 0, io.EOF
	}
	if len(mf.file) > 0 {
		l := len(p)
		if len(mf.file) < l {
			l = len(mf.file)
			*mf.readErr = true // he has read it all
		}
		rc := copy(p, []byte(mf.file[:l]))
		return rc, nil
	}

	return 0, nil
}

func (mf mockFS) Readdir(n int) ([]os.FileInfo, error) {
	if mf.err != 200 {
		return nil, io.EOF
	}

	var mi []os.FileInfo
	for _, nm := range mf.names {
		nmi := mockFI{name: nm}
		mi = append(mi, nmi)
	}

	return mi, nil
}

func (mf mockFS) Seek(offset int64, whence int) (int64, error) {
	var rc int64
	switch whence {
	case io.SeekEnd:
		rc = int64(len(mf.file))
		if len(mf.names) > 0 {
			rc = int64(len(mf.names))
		}
	case io.SeekStart:
		rc = 0
	}
	return rc, nil
}

func (mf mockFS) Stat() (os.FileInfo, error) {
	if mf.statErr {
		return nil, errors.New("a faked error")
	}

	nmi := mockFI{name: mf.file}

	for _, f := range mf.names {
		nmi.files = append(nmi.files, f)
	}

	return nmi, nil
}

func (mf mockFS) Close() error {
	return nil
}

func Test_WrapFileSources(t *testing.T) {
	rt := mux.NewRouter()

	WrapFileSources(rt)

	for key, drv := range drives {

		// if drive has a path, make sure it has a route
		if len(*drv) > 0 {
			trt := rt.Get(key)

			if trt == nil {
				t.Fatalf("drive %s failed to get a route\n", key)
			}

			path, _ := trt.GetPathRegexp()

			assert.Contains(t, path, key, "route doesn't include key")
		}
	}
}

func Test_ContainsDotFile(t *testing.T) {
	tests := []struct {
		name   string
		expect bool
	}{
		{name: "menu.bas", expect: false},
		{name: ".gitignore", expect: true},
		{name: "html/../main.html", expect: true},
	}

	for _, tt := range tests {
		if containsDotFile(tt.name) != tt.expect {
			t.Fatalf("%s should have gotten %T but got %T\n", tt.name, tt.expect, containsDotFile(tt.name))
		}
	}
}

func Test_Open(t *testing.T) {
	tests := []struct {
		name string
		want string
		fail bool
	}{
		{name: ".gitignore", fail: true},
		{name: "menu", want: "menu", fail: false},
		{name: "menu", want: "hello", fail: true},
	}

	for _, tt := range tests {
		fs := fileSource{src: mockFS{file: tt.want}}
		_, err := fs.Open(tt.name)
		if (err != nil) != tt.fail {
			t.Fatalf("Open(%s) should have gotten %T but got %T\n", tt.name, tt.fail, (err != nil))
		}
	}
}

func Test_SendDirectory(t *testing.T) {
	tests := []struct {
		files []string
		want  string
		res   int
	}{
		{files: []string{"hello.bas", "menu.bas"}, want: "", res: 404},
		{files: []string{"hello.bas", ".gitignore", "menu.bas"}, want: "<p>hello.bas</p><p>menu.bas</p>", res: 200},
		{files: []string{"hello.bas", "menu.bas"}, want: "<p>hello.bas</p><p>menu.bas</p>", res: 200},
	}

	for _, tt := range tests {
		fs := mockFS{err: tt.res}
		for _, tf := range tt.files {
			fs.names = append(fs.names, tf)
		}
		ffs := fileSource{}
		df := dotFileHidingFile{fs}
		rr := httptest.NewRecorder()

		ffs.sendDirectory(df, rr)

		bufstr := validateResult(t, rr, tt.res)

		if strings.Compare(bufstr, tt.want) != 0 {
			t.Fatalf("got result: %s\nwanted : %s\n", bufstr, tt.want)
		}
	}
}

func validateResult(t *testing.T, rr *httptest.ResponseRecorder, rc int) string {
	if rr.Result().StatusCode != rc {
		t.Fatalf("got status %d wanted %d\n", rr.Result().StatusCode, rc)
	}

	len := rr.Body.Len()

	if len == 0 {
		return ""
	}

	buf := make([]byte, len)
	_, err := io.ReadFull(rr.Body, buf)

	if err != nil {
		t.Fatalf("got error %s, trying to read body\n", err.Error())
	}

	return string(buf)
}

func Test_ServeFile(t *testing.T) {
	tests := []struct {
		testid  string
		fname   string
		res     int
		want    string
		statErr bool
		readErr bool
		files   []string
	}{
		{fname: "hello.bas", res: 403, want: "", readErr: true},
		{fname: "/", res: 200, want: "<p>hello.bas</p><p>test.bas</p><p>menu.bas</p>", files: []string{"hello.bas", "test.bas", "menu.bas"}},
		{fname: "hello.bas", res: 403, want: "", statErr: true},
		{fname: "hello.bas", res: 404, want: ""},
		{fname: "", res: 200, want: "/"},
		{fname: "hello.bas", res: 200, want: "hello.bas"},
	}

	for _, tt := range tests {
		fs := mockFS{file: tt.fname, err: tt.res, statErr: tt.statErr, readErr: &tt.readErr}
		for _, name := range tt.files {
			fs.names = append(fs.names, name)
		}
		// setup certain errors
		if len(tt.fname) == 0 {
			fs.file = tt.want // empty name should be treated as root
		}
		if tt.res == 404 {
			fs.file = "" // no known files throws an error
		}

		rr := httptest.NewRecorder()
		src := fileSource{src: fs}
		req, err := http.NewRequest("GET", tt.fname, nil)
		assert.Nilf(t, err, "Build rqst failed")
		src.ServeFile(rr, req, tt.fname)

		if rr.Result().StatusCode != tt.res {
			t.Fatalf("got status %d wanted %d\n", rr.Result().StatusCode, tt.res)
		}

		bufstr := validateResult(t, rr, tt.res)

		if strings.Compare(bufstr, tt.want) != 0 {
			t.Fatalf("got result: %s\nwanted : %s\n", bufstr, tt.want)
		}
	}
}

func Test_SetContentType(t *testing.T) {
	tests := []struct {
		name string
		exp  string
	}{
		{name: "xterm.css", exp: "text/css"},
		{name: "xterm.js", exp: "text/javascript"},
	}

	for _, tt := range tests {
		fs := fileSource{filename: tt.name}
		rr := httptest.NewRecorder()

		fs.setContentType(rr)
		got := rr.Result().Header.Get("Content-Type")

		assert.Equalf(t, 0, strings.Compare(got, tt.exp), "SetContentType got %s, expected %s", got, tt.exp)
	}
}

func Test_FileServe(t *testing.T) {
	tests := []struct {
		name string
		rqst string
		resp int
	}{
		{name: "bogus.bas", rqst: "/driveA/bogus.bas", resp: 403},
		{name: "bogus.bas", rqst: "/driveC/bogus.bas", resp: 200},
		{name: "bogus.wasm", rqst: "/webmodules/gwbasic.wasm", resp: 200},
		{name: "bogus.html", rqst: "/assets/html/gwbasic.html", resp: 200},
	}

	for _, tt := range tests {
		// mock out the file sources
		rErr := false
		fileSources["assets"] = &fileSource{src: mockFS{file: tt.name, readErr: &rErr, openAlways: true}, filename: tt.name}
		fileSources["webmodules"] = &fileSource{src: mockFS{file: tt.name, readErr: &rErr, openAlways: true}, filename: tt.name}
		fileSources["driveC"] = &fileSource{src: mockFS{file: tt.name, readErr: &rErr, openAlways: true}, filename: tt.name}
		req, err := http.NewRequest("GET", tt.rqst, nil)
		assert.Nilf(t, err, "Build rqst(%s) failed with: %s\n", tt.name, tt.rqst)

		rr := httptest.NewRecorder()
		FileServ(rr, req)

		bufstr := validateResult(t, rr, tt.resp)

		if (strings.Compare(bufstr, tt.name) != 0) && (tt.resp == 200) {
			t.Fatalf("got result: %s\nwanted : %s\n", bufstr, tt.name)
		}
	}
}

type mockFI struct {
	name  string
	files []string
}

func (mi mockFI) IsDir() bool {
	if len(mi.files) > 0 {
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

func Test_Readdir(t *testing.T) {

	tests := []struct {
		name   string
		err    string
		fnames []string
		want   []string
	}{
		{name: "test file list", fnames: []string{"hello.bas", ".gitignore"}, want: []string{"hello.bas"}},
		{name: "test error handling", err: "test error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := new(mockFS)

			if len(tt.err) > 0 {
				//	fs.err = errors.New(tt.err)
			}

			for _, nm := range tt.fnames {
				fs.names = append(fs.names, nm)
			}

			dfs := dotFileHidingFile{*fs}
			dfs.Readdir(-1)
		})
	}
}
