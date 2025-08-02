package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_Routes(t *testing.T) {
	tests := []struct {
		name    string
		request string
		route   string
		body    string
	}{
		{"main page", "/", rootRt, "gwbasic.wasm"},
	}

	for _, tt := range tests {
		rq, err := http.NewRequest("GET", tt.request, nil)

		if err != nil {
			t.Fatal(err)
		}

		rtr := startup()
		rr := httptest.NewRecorder()

		hd := rtr.Get(tt.name).GetHandler()

		hd.ServeHTTP(rr, rq)

		if rr.Code != http.StatusOK {
			t.Errorf("handler returned unexpected status code, got %v wanted %v\n", rr.Code, http.StatusOK)
		}

		body := rr.Body.String()

		// I get back the html file so to check it
		// I go looking for the refererence to the wasm file
		if !strings.Contains(body, tt.body) {
			t.Errorf("returned body did not contain %s, %s\n", tt.body, body)
		}
	}
}
