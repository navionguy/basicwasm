package mocks

import (
	"bytes"
	"errors"
)

type MockRdr struct {
	data []byte
	rdr  *bytes.Reader
}

func NewReader(b []byte) *MockRdr {
	rdr := MockRdr{data: b}

	if len(b) > 0 {
		rdr.rdr = bytes.NewReader(rdr.data)
	}

	return &rdr
}

func (rdr *MockRdr) Read(p []byte) (n int, err error) {
	if len(rdr.data) == 0 {
		return 0, errors.New("i live to fail")
	}

	return rdr.rdr.Read(p)
}
