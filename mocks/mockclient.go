package mocks

import (
	"errors"
	"net/http"
	"strings"
)

// mocks the parts of http.Client that I use
type MockClient struct {
	Contents   string // file contents to send back
	Url        string // Url to validate
	Err        error  // Error to return on call
	StatusCode int    // Status code to return
}

func (mc *MockClient) Get(url string) (*http.Response, error) {
	// am I expected to error?
	if mc.Err != nil {
		return nil, mc.Err
	}

	// Do I need to validate the Url?
	if (len(mc.Url) != 0) && (strings.Compare(mc.Url, url) != 0) {
		rsp := http.Response{Status: "404 Not Found", StatusCode: mc.StatusCode}
		return &rsp, errors.New("URL not found")
	}

	rdr := readCloser{}
	rdr.rdr = strings.NewReader(mc.Contents)
	if mc.StatusCode == 0 {
		mc.StatusCode = 200
	}
	rsp := http.Response{Status: "200 OK", StatusCode: mc.StatusCode, Body: &rdr}
	return &rsp, mc.Err
}

// implement a io.ReadCloser
type readCloser struct {
	rdr *strings.Reader
}

// Read implements the basic Read method
func (rc *readCloser) Read(p []byte) (int, error) {
	bt, err := rc.rdr.Read(p)
	return bt, err
}

// Closer wraps the Close method
func (rc *readCloser) Close() error {
	return nil
}
