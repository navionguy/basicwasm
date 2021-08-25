package mocks

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	sawOpen    = "sawOpen"
	sawReadDir = "sawReadDir"
	sawStat    = "sawStat"
	sawName    = "sawName"
)

type MockFS struct {
	File       string // Filename
	StatErr    bool   // return an error when stat is called
	ReadErr    *bool  // return error from read call
	OpenAlways bool   // return a file handle no matter what
	Events     map[string]bool

	// desired Readdir results
	Names []string
	Err   int
}

func (mf MockFS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (mf MockFS) Open(file string) (http.File, error) {
	if mf.Events != nil {
		mf.Events[sawOpen] = true
	}
	if (mf.File != file) && !mf.OpenAlways {
		return nil, fmt.Errorf("got %s, wanted %s", file, mf.File)
	}
	return mf, nil
}

func (mf MockFS) Read(p []byte) (int, error) {

	if *mf.ReadErr {
		return 0, io.EOF
	}
	if len(mf.File) > 0 {
		l := len(p)
		if len(mf.File) < l {
			l = len(mf.File)
			*mf.ReadErr = true // he has read it all
		}
		rc := copy(p, []byte(mf.File[:l]))
		return rc, nil
	}

	return 0, nil
}

func (mf MockFS) Readdir(n int) ([]os.FileInfo, error) {
	if mf.Events != nil {
		mf.Events[sawReadDir] = true
	}
	if mf.Err != http.StatusOK {
		return nil, io.EOF
	}

	var mi []os.FileInfo
	for _, nm := range mf.Names {
		nmi := MockFI{Fname: nm, Mom: &mf}
		mi = append(mi, nmi)
	}

	return mi, nil
}

func (mf MockFS) Seek(offset int64, whence int) (int64, error) {
	var rc int64
	switch whence {
	case io.SeekEnd:
		rc = int64(len(mf.File))
		if len(mf.Names) > 0 {
			rc = int64(len(mf.Names))
		}
	case io.SeekStart:
		rc = 0
	}
	return rc, nil
}

func (mf MockFS) Stat() (os.FileInfo, error) {
	if mf.Events != nil {
		mf.Events[sawStat] = true
	}
	if mf.StatErr {
		return nil, errors.New("a faked error")
	}

	nmi := MockFI{Fname: mf.File, Mom: &mf}

	for _, f := range mf.Names {
		nmi.Files = append(nmi.Files, f)
	}

	return nmi, nil
}

func (mf *MockFS) SawName() {
	if mf.Events != nil {
		mf.Events[sawName] = true
	}
}

func (mf MockFS) Close() error {
	return nil
}
