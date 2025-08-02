package mocks

import (
	"net/http"
	"net/http/httptest"
)

type MockServ struct {
	sever      *httptest.Server
	statuscode int
	body       []byte
	ExpURL     string
	Requests   []string
}

// NewMockServer returns a pointer to a ready to use mock http server
// caller should call close when finished, to shut it down
func NewMockServer(statuscode int, body []byte) *MockServ {
	ms := MockServ{statuscode: statuscode, body: body}

	ms.sever = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(statuscode)
		res.Write([]byte(body))

		ms.Requests = append(ms.Requests, req.URL.String())
	}))
	ms.ExpURL = ms.sever.URL

	return &ms
}

// Close the mock server
func (ms *MockServ) Close() {
	ms.sever.Close()
}
