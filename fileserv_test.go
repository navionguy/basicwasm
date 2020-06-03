package main

import (
	"net/http"
	"reflect"
	"testing"
)

func Test_parseName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name         string
		args         args
		wantBaseName string
		wantExt      string
		wantOk       bool
	}{
		{name: "Happy Path", args: args{name: "mainmenu.html"}, wantBaseName: "mainmenu", wantExt: "html", wantOk: true},
		{name: "No extension", args: args{name: "mainmenu"}, wantBaseName: "mainmenu", wantExt: "html", wantOk: true},
		{name: "Extra pathing", args: args{name: "/dir/dir/dir/mainmenu.html"}, wantBaseName: "mainmenu", wantExt: "html", wantOk: true},
		{name: "Extra extension", args: args{name: "~/node-modules/xterm/xterm.js.map"}, wantBaseName: "xterm.js", wantExt: "map", wantOk: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := parseName(tt.args.name)
			if got != tt.wantBaseName {
				t.Errorf("parseName() got = %v, want %v", got, tt.wantBaseName)
			}
			if got1 != tt.wantExt {
				t.Errorf("parseName() got1 = %v, want %v", got1, tt.wantExt)
			}
			if got2 != tt.wantOk {
				t.Errorf("parseName() got2 = %v, want %v", got2, tt.wantOk)
			}
		})
	}
}

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
		{name: "wasm", args: args{name: "basename", ext: "wasm"}, wantPath: "./webmodules/basename/lib.wasm", wantOk: true},
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
}

func Test_autoPathingSystem_Open(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name      string
		fs        autoPathingSystem
		args      args
		wantHFile http.File
		wantErr   bool
	}{
		{name: "No extension", args: args{"file"}, wantHFile: nil, wantErr: true},
		{name: "Invalid extension", args: args{"file.snigglefritz"}, wantHFile: nil, wantErr: true},
		{name: "HTML source", fs: autoPathingSystem{http.Dir(".")}, args: args{"mainmenu.html"}, wantHFile: nil, wantErr: false},
		{name: "Root", fs: autoPathingSystem{http.Dir(".")}, args: args{""}, wantHFile: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHFile, err := tt.fs.Open(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("autoPathingSystem.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !reflect.DeepEqual(gotHFile, tt.wantHFile) {
					t.Errorf("autoPathingSystem.Open() = %v, want %v", gotHFile, tt.wantHFile)
				}
			}
		})
	}
}
