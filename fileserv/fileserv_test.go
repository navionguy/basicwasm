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

const (
	sawOpen    = "sawOpen"
	sawReadDir = "sawReadDir"
	sawStat    = "sawStat"
	sawName    = "sawName"
)

type mockFS struct {
	file       string // filename
	statErr    bool   // return an error when stat is called
	readErr    *bool  // return error from read call
	openAlways bool   // return a file handle no matter what
	events     map[string]bool

	// desired Readdir results
	names []string
	err   int
}

func (mf mockFS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (mf mockFS) Open(file string) (http.File, error) {
	if mf.events != nil {
		mf.events[sawOpen] = true
	}
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
	if mf.events != nil {
		mf.events[sawReadDir] = true
	}
	if mf.err != http.StatusOK {
		return nil, io.EOF
	}

	var mi []os.FileInfo
	for _, nm := range mf.names {
		nmi := mockFI{name: nm, mom: &mf}
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
	if mf.events != nil {
		mf.events[sawStat] = true
	}
	if mf.statErr {
		return nil, errors.New("a faked error")
	}

	nmi := mockFI{name: mf.file, mom: &mf}

	for _, f := range mf.names {
		nmi.files = append(nmi.files, f)
	}

	return nmi, nil
}

func (mf *mockFS) SawName() {
	fmt.Println("sawName")
	if mf.events != nil {
		mf.events[sawName] = true
	}
}

func (mf mockFS) Close() error {
	return nil
}

type mockFI struct {
	name  string
	files []string
	mom   *mockFS
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
	if mi.mom != nil {
		mi.mom.SawName()
	}
	return mi.name
}

func (mi mockFI) Size() int64 {
	return int64(len(mi.name))
}

func (mi mockFI) Sys() interface{} {
	return nil
}

func Test_WrapSource(t *testing.T) {
	rt := mux.NewRouter()
	fs := fileSource{src: http.Dir("../source")}
	fs.wrapSource(rt, "/driveC/{file}.{ext}", "text/plain; charset=ASCII")

	ts := httptest.NewServer(rt)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/driveC/start.bas")

	assert.Nil(t, err, "http.Get got error")
	assert.NotEmpty(t, res, "http.Get no body returned")
}

func Test_WrapSubDirs(t *testing.T) {
	tests := []struct {
		tname      string
		fname      string
		rc         int
		files      []string
		events     []string
		openAlways bool
		statErr    bool
	}{
		{tname: "WrapSubDirs fail stat", fname: "/", openAlways: true, rc: http.StatusOK, files: []string{"test.bas"}, events: []string{"sawOpen", "sawReadDir", "sawName", "sawStat"}, statErr: true},
		{tname: "WrapSubDirs fail fopen", fname: "/", rc: http.StatusOK, files: []string{"test.bas"}, events: []string{"sawOpen", "sawReadDir", "sawName"}},
		{tname: "WrapSubDirs fopen", fname: "/", openAlways: true, rc: http.StatusOK, files: []string{"test.bas"}, events: []string{"sawOpen", "sawReadDir", "sawName", "sawStat"}},
		{tname: "WrapSubDirs fail dotfile", fname: "/", rc: http.StatusOK, files: []string{".test"}, events: []string{"sawOpen", "sawReadDir", "sawName"}},
		{tname: "WrapSubDirs fail readdir", fname: "/", rc: http.StatusTeapot, events: []string{"sawOpen", "sawReadDir"}},
		{tname: "WrapSubDirs fail open", fname: "bogus", rc: http.StatusOK, events: []string{"sawOpen"}},
	}

	for _, tt := range tests {
		fs := mockFS{file: tt.fname, err: tt.rc, names: tt.files, openAlways: tt.openAlways, statErr: tt.statErr}
		fs.events = make(map[string]bool)
		src := fileSource{src: fs}
		rt := mux.NewRouter()
		src.wrapSubDirs(rt, "/", "/")

		assert.Equal(t, len(tt.events), len(fs.events), "Test %s unexpectedly got %d events", tt.tname, len(fs.events))
	}
}

func Test_WrapFileSources(t *testing.T) {
	rt := mux.NewRouter()
	fix := "../source"
	drives["driveC"] = &fix

	WrapFileSources(rt)

	for key, drv := range drives {

		// if drive has a path, make sure it has a route
		if len(*drv) > 0 {
			trt := rt.Get("/" + key)
			assert.NotEmpty(t, trt, "drive %s failed to get a route\n", key)

			path, _ := trt.GetPathRegexp()
			assert.Contains(t, path, key, "route doesn't include key")

			ts := httptest.NewServer(rt)
			defer ts.Close()

			res, err := http.Get(ts.URL + "/driveC/")

			assert.Nil(t, err, "http.Get got error")
			assert.NotEmpty(t, res, "http.Get no body returned")
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
		ifs := mockFS{file: tt.want}
		ifs.events = make(map[string]bool)

		fs := fileSource{src: ifs}

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
		mtype string
		res   int
	}{
		{files: []string{"hello.bas", "menu.bas"}, want: "", res: 404},
		{files: []string{"hello.bas", ".gitignore", "menu.bas"}, want: "<p>hello.bas</p><p>menu.bas</p>", res: 200},
		{files: []string{"hello.bas", "menu.bas"}, want: "<p>hello.bas</p><p>menu.bas</p>", res: 200},
	}

	for _, tt := range tests {
		fs := mockFS{err: tt.res}
		fs.events = make(map[string]bool)
		for _, tf := range tt.files {
			fs.names = append(fs.names, tf)
		}
		ffs := fileSource{}
		df := dotFileHidingFile{fs}
		rr := httptest.NewRecorder()

		ffs.sendDirectory(df, rr)

		bufstr := validateResult(t, rr, tt.res, tt.mtype)

		if strings.Compare(bufstr, tt.want) != 0 {
			t.Fatalf("got result: %s\nwanted : %s\n", bufstr, tt.want)
		}
	}
}

func validateResult(t *testing.T, rr *httptest.ResponseRecorder, rc int, mtype string) string {
	if rr.Result().StatusCode != rc {
		t.Fatalf("got status %d wanted %d\n", rr.Result().StatusCode, rc)
	}

	if rr.Body.Len() == 0 {
		return ""
	}

	buf := make([]byte, rr.Body.Len())
	_, err := io.ReadFull(rr.Body, buf)

	if err != nil {
		t.Fatalf("got error %s, trying to read body\n", err.Error())
	}

	if len(mtype) > 0 {
		assert.Equal(t, mtype, rr.HeaderMap.Get("content-type"), "expected mime type %s, got %s", mtype, rr.HeaderMap.Get("content-type"))
	}

	return string(buf)
}

func Test_ServeFile(t *testing.T) {
	tests := []struct {
		testid  string
		fname   string
		mtype   string
		res     int
		want    string
		statErr bool
		readErr bool
		files   []string
	}{
		{testid: "read fail", fname: "hello.bas", mtype: "text/plain; charset=ASCII", res: 503, want: "", readErr: true},
		{testid: "dir", fname: "/", mtype: "text/html; charset=utf-8", res: 200, want: "<p>hello.bas</p><p>test.bas</p><p>menu.bas</p>", files: []string{"hello.bas", "test.bas", "menu.bas"}},
		{testid: "stat Error", fname: "hello.bas", res: 500, want: "", statErr: true},
		{testid: "file not found", fname: "hello.bas", res: 404, want: ""},
		{testid: "read from root", fname: "", res: 200, mtype: "text/plain; charset=ASCII", want: "/"},
		{testid: "read file", fname: "hello.bas", mtype: "text/plain; charset=ASCII", res: 200, want: "hello.bas"},
	}

	for _, tt := range tests {
		fs := mockFS{file: tt.fname, err: tt.res, statErr: tt.statErr, readErr: &tt.readErr}
		fs.events = make(map[string]bool)
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
		src.serveFile(rr, req, tt.fname, tt.mtype)

		if rr.Result().StatusCode != tt.res {
			t.Fatalf("got status %d wanted %d\n", rr.Result().StatusCode, tt.res)
		}

		bufstr := validateResult(t, rr, tt.res, tt.mtype)

		if strings.Compare(bufstr, tt.want) != 0 {
			t.Fatalf("got result: %s\nwanted : %s\n", bufstr, tt.want)
		}
	}
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
			fs.events = make(map[string]bool)

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
