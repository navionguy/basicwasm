package fileserv

import (
	"fmt"
	"net/http"
	"testing"
)

/*
func Test_parseName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name         string
		args         args
		wantBaseName string
		wantExt      string
	}{
		{name: "Happy Path", args: args{name: "mainmenu.html"}, wantBaseName: "mainmenu", wantExt: "html"},
		{name: "No extension", args: args{name: "mainmenu"}, wantBaseName: "mainmenu", wantExt: "html"},
		{name: "Extra pathing", args: args{name: "/dir/dir/dir/mainmenu.html"}, wantBaseName: "mainmenu", wantExt: "html"},
		{name: "Extra extension", args: args{name: "~/node-modules/xterm/xterm.js.map"}, wantBaseName: "xterm.js", wantExt: "map"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := parseName(tt.args.name)
			if got != tt.wantBaseName {
				t.Errorf("parseName() got = %v, want %v", got, tt.wantBaseName)
			}
			if got1 != tt.wantExt {
				t.Errorf("parseName() got1 = %v, want %v", got1, tt.wantExt)
			}
		})
	}
}*/
/*
func Test_buildPath(t *testing.T) {
	type args struct {
		name string
		ext  string
	}
	tests := []struct {
		name     string
		args     args
		wantPath string
		wantOk   bool
	}{
		{name: "Unsupported extension", args: args{name: "basename", ext: "foobar"}, wantPath: "", wantOk: false},
		{name: "html", args: args{name: "basename", ext: "html"}, wantPath: "./assets/html/basename.html", wantOk: true},
		{name: "js", args: args{name: "basename", ext: "js"}, wantPath: "./assets/js/basename.js", wantOk: true},
		{name: "ico", args: args{name: "basename", ext: "ico"}, wantPath: "./assets/images/basename.ico", wantOk: true},
		{name: "css", args: args{name: "basename", ext: "css"}, wantPath: "./assets/css/basename.css", wantOk: true},
		{name: "wasm", args: args{name: "basename", ext: "wasm"}, wantPath: "./webmodules/basename.wasm", wantOk: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotOk := buildPath(tt.args.name, tt.args.ext)
			if gotPath != tt.wantPath {
				t.Errorf("buildPath() gotPath = %v, want %v", gotPath, tt.wantPath)
			}
			if gotOk != tt.wantOk {
				t.Errorf("buildPath() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}*/

type mockFS struct {
	want string
}

func (mf mockFS) Open(file string) (http.File, error) {
	if mf.want != file {
		return nil, fmt.Errorf("got %s, wanted %s", file, mf.want)
	}
	return nil, nil
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
