package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

type TestingRouter struct {
	Router *mux.Router
}

func (tr *TestingRouter) AddHandler(path string, handler func(http.ResponseWriter, *http.Request), queryParamsPairs ...string) {
	tr.Router.HandleFunc(path, handler).Queries(queryParamsPairs...)
}

func (tr *TestingRouter) ExecuteRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	tr.Router.ServeHTTP(rr, req)

	return rr
}

func (tr *TestingRouter) CheckResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}
