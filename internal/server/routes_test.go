package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	http.HandlerFunc(s.HelloWorldHandler).ServeHTTP(w, req)

	// Assertions
	if w.Code != http.StatusOK {
		t.Errorf("expected status OK; got %v", w.Code)
	}
	expected := "{\"message\":\"Hello World\"}"
	if expected != w.Body.String() {
		t.Errorf("expected response body to be %v; got %v", expected, w.Body.String())
	}
}
