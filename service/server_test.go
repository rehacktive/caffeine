package service

import (
	"fmt"
	"net/http"
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
