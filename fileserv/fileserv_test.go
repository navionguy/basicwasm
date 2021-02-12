package fileserv

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
)

type mockFS struct {
	// file path wanted
	want string

	// desired Readdir results
	names []string
	err   error
}

func (mf mockFS) Open(file string) (http.File, error) {
	if mf.want != file {
		return nil, fmt.Errorf("got %s, wanted %s", file, mf.want)
	}
	return nil, nil
}

func (mf mockFS) Read([]byte) (int, error) {
	return 0, nil
}

func (mf mockFS) Readdir(n int) ([]os.FileInfo, error) {
	if mf.err != nil {
		return nil, mf.err
	}

	var mi []os.FileInfo
	for _, nm := range mf.names {
		nmi := mockFI{name: nm}
		mi = append(mi, nmi)
	}

	return mi, nil
}

func (mf mockFS) Seek(int64, int) (int64, error) {
	return 0, nil
}

func (mf mockFS) Stat() (os.FileInfo, error) {
	nmi := mockFI{name: mf.names[0]}

	return nmi, nil
}

func (mf mockFS) Close() error {
	return nil
}

func Test_autoPathingSystem_Open(t *testing.T) {
	fs := new(mockFS)
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fs        http.FileSystem
		args      args
		wantHFile string
		wantErr   bool
	}{
		// I have to set the root director for the pathing object to my parent
		// since I'm down in the fileserv directory
		//
		{name: "index file", args: args{"menu/"}, wantHFile: "./source/menu/prog/calc.bas", wantErr: false},
		{name: "bas file not in subdir", args: args{"menu/prog/calc.bas"}, wantHFile: "./source/menu/prog/calc.bas", wantErr: true},
		{name: "bas file in subdir", fs: autoPathingSystem{fs}, args: args{"menu/prog/calc.bas"}, wantHFile: "./source/menu/prog/calc.bas", wantErr: false},
		{name: "bas file", args: args{"menu/hello.bas"}, wantHFile: "./source/menu/hello.bas", wantErr: false},
		{name: "css file", args: args{"print.css"}, wantHFile: "./assets/css/print.css", wantErr: false},
		{name: "ico file", args: args{"favicon.ico"}, wantHFile: "./assets/images/favicon.ico", wantErr: false},
		{name: "Wasm file", args: args{"gwbasic.wasm"}, wantHFile: "./assets/webmodules/gwbasic.wasm", wantErr: false},
		{name: "Javascript source", args: args{"xterm.js"}, wantHFile: "./assets/js/xterm.js", wantErr: false},
		{name: "Multi-extension", fs: autoPathingSystem{fs}, args: args{"compound.name.html"}, wantHFile: "./assets/html/compound.name.html", wantErr: false},
		{name: "Secured file", args: args{".github"}, wantHFile: "", wantErr: true},
		{name: "No extension", fs: autoPathingSystem{fs}, args: args{"gwbasic"}, wantHFile: "./assets/html/gwbasic.html", wantErr: false},
		{name: "Invalid extension", args: args{"file.snigglefritz"}, wantHFile: "", wantErr: true},
		{name: "HTML source", args: args{"gwbasic.html"}, wantHFile: "", wantErr: false},
		{name: "Root", fs: autoPathingSystem{fs}, args: args{""}, wantHFile: "./assets/html/gwbasic.html", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.wantHFile) > 0 {
				fs.want = tt.wantHFile
			}

			if tt.fs == nil {
				*filedir = ".."
				tt.fs = WrapFileOrg()
			}

			_, err := tt.fs.Open(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("autoPathingSystem.Open(%s) error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
		})
	}
}

type mockFI struct {
	name string
}

func (mi mockFI) IsDir() bool {
	return true
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
	return 0
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
				fs.err = errors.New(tt.err)
			}

			for _, nm := range tt.fnames {
				fs.names = append(fs.names, nm)
			}

			dfs := dotFileHidingFile{*fs}
			dfs.Readdir(-1)
		})
	}
}
