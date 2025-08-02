package localfiles

import (
	"net/http"
	"testing"

	"github.com/navionguy/basicwasm/mocks"
	"github.com/navionguy/basicwasm/object"
	"github.com/stretchr/testify/assert"
)

func TestALocalFile(t *testing.T) {
	af := aLocalFile{FQFilename: "test.dat"}

	assert.EqualValues(t, "test.dat", af.Inspect(), "Inspect() failed")
	assert.EqualValues(t, object.FILE_OBJ, af.Type())
}

func TestFetchFile(t *testing.T) {
	tests := []struct {
		filename string
		contents string
		result   object.Object
	}{
		{filename: "TEST.DAT", contents: "This is some test data.", result: &aLocalFile{}},
	}

	for _, tt := range tests {
		var trm mocks.MockTerm
		mocks.InitMockTerm(&trm)
		trm.ExpMsg = &mocks.Expector{}

		env := object.NewTermEnvironment(trm)

		cl := mocks.MockClient{Contents: tt.contents, Url: "http://localhost:8080/drivec/test.dat", StatusCode: http.StatusOK}
		env.SetClient(&cl)

		res := fetchFile(tt.filename, env)

		assert.Equal(t, tt.result.Type(), res.Type(), "unexpected fetchFile() result")
	}
}

func TestOpenFile(t *testing.T) {
	tests := []struct {
		filename string
		contents []byte
		preload  bool
	}{
		{filename: "TEST.DAT", contents: []byte("This is some test data.")},
		{filename: "TEST.DAT", contents: []byte("This is some test data."), preload: true},
	}

	for _, tt := range tests {
		var trm mocks.MockTerm
		env := object.NewTermEnvironment(trm)
		if tt.preload {
			lf.dir = make(map[string]*aLocalFile)
			alf := &aLocalFile{FQFilename: tt.filename, readonly: true, data: &tt.contents}
			lf.dir[tt.filename] = alf
		}

		Open(tt.filename, env)
	}
}

func TestSaveFile(t *testing.T) {
	tests := []struct {
		filename string
		contents []byte
	}{
		{filename: "TEST.DAT", contents: []byte("This is some test data.")},
	}

	lf.dir = make(map[string]*aLocalFile)

	for _, tt := range tests {
		var trm mocks.MockTerm
		env := object.NewTermEnvironment(trm)
		lf.saveFile(tt.filename, mocks.NewReader(tt.contents), env)
	}
}

func TestStoreFile(t *testing.T) {
	tests := []struct {
		filename string
		contents string
	}{
		{filename: "TEST.DAT", contents: "This is some test data."},
		{filename: "TESTFAIL.DAT"},
	}

	lf.dir = nil

	for _, tt := range tests {
		var trm mocks.MockTerm
		env := object.NewTermEnvironment(trm)
		b := []byte(tt.contents)
		rdr := mocks.NewReader(b)
		lf.dir = make(map[string]*aLocalFile)

		storeFile(tt.filename, rdr, env)
	}
}
