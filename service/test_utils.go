package service

import (
	"io/ioutil"
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

func checkResponseCode(t *testing.T, testName string, expected, actual int) {
	if expected != actual {
		t.Errorf("%v: Expected response code %d. Got %d\n", testName, expected, actual)
	}
}

func checkResponse(t *testing.T, testName string, response string, expected string) {
	if response != expected {
		t.Errorf("%v: Expected %s  Got %s", testName, expected, response)
	}
}

func checkErr(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

// utils
func getUserSchema() string {
	jsonSchema, err := ioutil.ReadFile("../schema_sample/user_schema.json")
	if err != nil {
		panic(err)
	}
	return string(jsonSchema)
}
