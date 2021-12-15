package service

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"

	"github.com/rehacktive/caffeine/database"
)

type testCase struct {
	name                 string
	method               string
	path                 string
	payload              string
	expectedResponseCode int
	expectedResponse     string
	beforeTest           func(Database)
	dbCheck              func(Database) error
}

var jsonPayload = `{"age":25,"name":"jack"}`
var testNamespace = "ns1"
var testKey = "key1"
var validJsonForSchema = `{
	"firstName":"john",
	"lastName":"never",
	"age":666
}`
var invalidJsonForSchema = `{
	"firstName":"john"
}`

var tests = []testCase{
	{
		name:                 "test home handler",
		method:               http.MethodGet,
		path:                 "/",
		payload:              "",
		expectedResponseCode: http.StatusOK,
		expectedResponse:     `["ns1"]`,
	},
	{
		name:                 "test namespace get",
		method:               http.MethodGet,
		path:                 "/ns/" + testNamespace,
		payload:              "",
		expectedResponseCode: http.StatusOK,
		expectedResponse:     fmt.Sprintf(`[{"key":"%v","value":%v}]`, testKey, jsonPayload),
	},
	{
		name:                 "test namespace get not existing",
		method:               http.MethodGet,
		path:                 "/ns/" + "not_existing_namespace",
		payload:              "",
		expectedResponseCode: http.StatusNotFound,
		expectedResponse:     "",
	},
	{
		name:                 "test namespace delete",
		method:               http.MethodDelete,
		path:                 "/ns/" + testNamespace,
		payload:              "",
		expectedResponseCode: http.StatusAccepted,
		expectedResponse:     "{}",
	},
	{
		name:                 "test namespace delete not existing",
		method:               http.MethodDelete,
		path:                 "/ns/" + "not_existing_namespace",
		payload:              "",
		expectedResponseCode: http.StatusNotFound,
		expectedResponse:     "",
	},
	{
		name:                 "test keyvalue post",
		method:               http.MethodPost,
		path:                 "/ns/test/1",
		payload:              jsonPayload,
		expectedResponseCode: http.StatusCreated,
		expectedResponse:     "",
		dbCheck: func(d Database) error {
			value, err := d.Get("test", "1")
			if err != nil {
				return err
			}
			if string(value) != jsonPayload {
				fmt.Errorf("Expected %v got %s", jsonPayload, string(value))
			}
			return nil
		},
	},
	{
		name:                 "test keyvalue post invalid json",
		method:               http.MethodPost,
		path:                 "/ns/test/1",
		payload:              "{some bad data...",
		expectedResponseCode: http.StatusBadRequest,
		expectedResponse:     "",
	},
	{
		name:                 "test keyvalue get",
		method:               http.MethodGet,
		path:                 "/ns/" + testNamespace + "/" + testKey,
		payload:              "",
		expectedResponseCode: http.StatusOK,
		expectedResponse:     jsonPayload,
	},
	{
		name:                 "test keyvalue get not existing",
		method:               http.MethodGet,
		path:                 "/ns/" + testNamespace + "/not_existing",
		payload:              "",
		expectedResponseCode: http.StatusNotFound,
		expectedResponse:     "",
	},
	{
		name:                 "test keyvalue delete",
		method:               http.MethodDelete,
		path:                 "/ns/" + testNamespace + "/" + testKey,
		payload:              "",
		expectedResponseCode: http.StatusAccepted,
		expectedResponse:     "{}",
	},
	{
		name:                 "test keyvalue delete not existing",
		method:               http.MethodDelete,
		path:                 "/ns/" + testNamespace + "/not_existing",
		payload:              "",
		expectedResponseCode: http.StatusNotFound,
		expectedResponse:     "",
	},
	{
		name:                 "test schema post",
		method:               http.MethodPost,
		path:                 "/schema/user",
		payload:              getUserSchema(),
		expectedResponseCode: http.StatusCreated,
		expectedResponse:     "",
		dbCheck: func(d Database) error {
			value, err := d.Get("user_schema", SchemaId)
			if err != nil {
				return err
			}
			if string(value) != jsonPayload {
				return fmt.Errorf("expected %v got %s", jsonPayload, string(value))
			}
			return nil
		},
	},
	{
		name:                 "test schema get",
		method:               http.MethodGet,
		path:                 "/schema/user",
		payload:              getUserSchema(),
		expectedResponseCode: http.StatusOK,
		expectedResponse:     getUserSchema(),
		beforeTest: func(d Database) {
			d.Upsert("user"+SchemaId, SchemaId, []byte(getUserSchema()))
		},
	},
	{
		name:                 "test schema get not existing",
		method:               http.MethodGet,
		path:                 "/schema/not_existing",
		payload:              "",
		expectedResponseCode: http.StatusNotFound,
		expectedResponse:     "",
	},
	{
		name:                 "test schema delete",
		method:               http.MethodDelete,
		path:                 "/schema/user",
		payload:              "",
		expectedResponseCode: http.StatusAccepted,
		expectedResponse:     "{}",
		beforeTest: func(d Database) {
			d.Upsert("user"+SchemaId, SchemaId, []byte(getUserSchema()))
		},
	},
	{
		name:                 "test post valid json with schema",
		method:               http.MethodPost,
		path:                 "/ns/user/1",
		payload:              validJsonForSchema,
		expectedResponseCode: http.StatusCreated,
		expectedResponse:     validJsonForSchema,
		beforeTest: func(d Database) {
			d.Upsert("user"+SchemaId, SchemaId, []byte(getUserSchema()))
		},
	},
	{
		name:                 "test post invalid json with schema",
		method:               http.MethodPost,
		path:                 "/ns/user/1",
		payload:              invalidJsonForSchema,
		expectedResponseCode: http.StatusBadRequest,
		expectedResponse:     `{ "status": 400, "message": "(root): lastName is required" }`,
		beforeTest: func(d Database) {
			d.Upsert("user"+SchemaId, SchemaId, []byte(getUserSchema()))
		},
	},
}

func setupCaffeineTest(db Database) *TestingRouter {
	db.Init()
	db.Upsert(testNamespace, testKey, []byte(jsonPayload))

	server := Server{
		db: db,
	}

	testingRouter := TestingRouter{Router: mux.NewRouter()}
	testingRouter.AddHandler("/", server.homeHandler)
	testingRouter.AddHandler(NamespacePattern, server.namespaceHandler)
	testingRouter.AddHandler(KeyValuePattern, server.keyValueHandler)
	testingRouter.AddHandler(SchemaPattern, server.schemaHandler)

	return &testingRouter
}

func testHandlers(db Database, t *testing.T) {
	for _, test := range tests {
		testingRouter := setupCaffeineTest(db)
		if test.beforeTest != nil {
			test.beforeTest(db)
		}
		log.Println("running test: ", test.name)
		req, _ := http.NewRequest(test.method, test.path, strings.NewReader(test.payload))
		response := testingRouter.ExecuteRequest(req)
		checkResponseCode(t, test.name, test.expectedResponseCode, response.Code)
		if test.expectedResponse != "" {
			checkResponse(t, test.name, response.Body.String(), test.expectedResponse)
		}
		if test.dbCheck != nil {
			err := test.dbCheck(db)
			if err != nil {
				t.Errorf(err.Error())
			}
		}
	}
}

func Test_UnitTest_MemoryDb(t *testing.T) {
	testHandlers(&database.MemDatabase{}, t)
}

func Test_UnitTest_StorageDb(t *testing.T) {
	db := &database.StorageDatabase{
		RootDirPath: "/tmp/caffeine_test1",
	}
	testHandlers(db, t)
	os.RemoveAll("/tmp/caffeine_test1")
}

func Test_UnitTest_SQLiteDb(t *testing.T) {
	os.MkdirAll("/tmp/caffeine", os.ModePerm)
	db := &database.SQLiteDatabase{
		DirPath: "/tmp/caffeine",
	}
	testHandlers(db, t)
	os.RemoveAll("/tmp/caffeine")
}
