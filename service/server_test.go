package service

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/rehacktive/caffeine/database"
)

func TestHomeHandler(t *testing.T) {
	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("namespace", "key", []byte("{}"))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler("/", server.homeHandler)

	req, _ := http.NewRequest("GET", "/", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != `["namespace"]` {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func TestNamespaceHandlerGet(t *testing.T) {
	json := `{"name":"jack"}`

	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(json))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(NamespacePattern, server.namespaceHandler)

	req, _ := http.NewRequest("GET", "/ns/test", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusOK, response.Code)

	expected := fmt.Sprintf(`[{"key":"1","value":%v}]`, json)
	if body := response.Body.String(); body != expected {
		t.Errorf("Expected %v got %s", expected, body)
	}
}

func TestNamespaceHandlerGet_NotExisting(t *testing.T) {
	json := `{"name":"jack"}`

	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(json))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(NamespacePattern, server.namespaceHandler)

	req, _ := http.NewRequest("GET", "/ns/not_existing_ns", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusNotFound, response.Code)
}

func TestNamespaceHandlerDelete(t *testing.T) {
	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(`{"name":"jack","age":25}`))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(NamespacePattern, server.namespaceHandler)

	req, _ := http.NewRequest("DELETE", "/ns/test", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusAccepted, response.Code)

	expected := `{}`
	if body := response.Body.String(); body != expected {
		t.Errorf("Expected %v got %s", expected, body)
	}
}

func TestNamespaceHandlerDelete_NotExisting(t *testing.T) {
	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(`{"name":"jack","age":25}`))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(NamespacePattern, server.namespaceHandler)

	req, _ := http.NewRequest("DELETE", "/ns/not_existing_ns", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusNotFound, response.Code)
}

func TestKeyValueHandlerPost(t *testing.T) {
	json := `{"name":"jack"}`

	mockDb := &database.MemDatabase{}
	mockDb.Init()

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(KeyValuePattern, server.keyValueHandler)

	req, _ := http.NewRequest("POST", "/ns/test/1", strings.NewReader(json))
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusCreated, response.Code)

	if body := response.Body.String(); body != json {
		t.Errorf("Expected %v got %s", json, body)
	}

	value, err := mockDb.Get("test", "1")
	if err != nil {
		t.Errorf("Unexpected error %s", err)
	}
	if string(value) != json {
		t.Errorf("Expected %v got %s", json, string(value))
	}
}

func TestKeyValueHandlerGet(t *testing.T) {
	json := `{"name":"jack"}`

	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(json))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(KeyValuePattern, server.keyValueHandler)

	req, _ := http.NewRequest("GET", "/ns/test/1", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != json {
		t.Errorf("Expected %v got %s", json, body)
	}
}

func TestKeyValueHandlerGet_NotExisting(t *testing.T) {
	json := `{"name":"jack"}`

	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(json))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(KeyValuePattern, server.keyValueHandler)

	req, _ := http.NewRequest("GET", "/ns/test/2", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusNotFound, response.Code)
}

func TestKeyValueHandlerDelete(t *testing.T) {
	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(`{"name":"jack"}`))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(KeyValuePattern, server.keyValueHandler)

	req, _ := http.NewRequest("DELETE", "/ns/test/1", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusAccepted, response.Code)

	expected := `{}`
	if body := response.Body.String(); body != expected {
		t.Errorf("Expected %v got %s", expected, body)
	}
}

func TestKeyValueHandlerDelete_NotExisting(t *testing.T) {
	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(`{"name":"jack"}`))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(KeyValuePattern, server.keyValueHandler)

	req, _ := http.NewRequest("DELETE", "/ns/test/2", nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusAccepted, response.Code)
}

func TestSearchHandler(t *testing.T) {
	json1 := `{"name":"jack"}`
	json2 := `{"name":"john"}`

	mockDb := &database.MemDatabase{}
	mockDb.Init()
	mockDb.Upsert("test", "1", []byte(json1))
	mockDb.Upsert("test", "2", []byte(json2))

	server := Server{
		db: mockDb,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler(SearchPattern, server.searchHandler, "filter", "{filter}")

	req, _ := http.NewRequest("GET", `/search/test?filter="select(.name==\"john\")"`, nil)
	response := testingRouter.ExecuteRequest(req)

	testingRouter.CheckResponseCode(t, http.StatusOK, response.Code)

	// expected := fmt.Sprintf(`{"results":[{"key":"1","value":%v}]}`, json1)
	// if body := response.Body.String(); body != expected {
	// 	t.Errorf("Expected %v got %s", expected, body)
	// }
}
