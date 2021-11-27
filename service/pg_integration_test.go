package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

var httpClient http.Client

// Integration test for Postgres

func setupContainers() (string, error) {
	composeFilePaths := []string{"../docker-compose.yaml"}
	identifier := strings.ToLower(uuid.New().String())

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.
		WithCommand([]string{"up", "-d", "--build"}).
		Invoke()
	err := execError.Error
	if err != nil {
		return "", fmt.Errorf("could not run compose file: %v - %v", composeFilePaths, err)
	}
	return identifier, nil
}

func stopContainers(identifier string) error {
	composeFilePaths := []string{"../docker-compose.yaml"}

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.Down()
	err := execError.Error
	if err != nil {
		return fmt.Errorf("could not run compose file: %v - %v", composeFilePaths, err)
	}
	return nil
}

func TestPgIntegration(t *testing.T) {
	httpClient = http.Client{Timeout: time.Duration(5) * time.Second}

	id, err := setupContainers()
	if err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	// do tests
	doTests(t)

	err = stopContainers(id)
	if err != nil {
		panic(err)
	}
}

func doTests(t *testing.T) {
	// add three key/values
	for i := 1; i < 4; i++ {
		code, response, err := call(http.MethodPost, fmt.Sprintf("http://localhost:8000/ns/test/%v", i), jsonPayload)
		checkErr(t, err)
		checkResponseCode(t, fmt.Sprintf("insert id %v", i), code, http.StatusCreated)
		checkResponse(t, fmt.Sprintf("insert id %v", i), jsonPayload, response)
	}

	// check namespace exists
	code, response, err := call(http.MethodGet, "http://localhost:8000/ns", "")
	checkErr(t, err)
	checkResponseCode(t, "check namespace", http.StatusOK, code)
	checkResponse(t, "check namespace", response, `["test"]`)

	// check all key/values exist
	code, response, err = call(http.MethodGet, "http://localhost:8000/ns/test", "")
	checkErr(t, err)
	checkResponseCode(t, "check values", http.StatusOK, code)
	checkResponse(t, "check values", response, `[{"key":"1","value":{"age":25,"name":"jack"}},{"key":"2","value":{"age":25,"name":"jack"}},{"key":"3","value":{"age":25,"name":"jack"}}]`)

	// delete key 2
	code, response, err = call(http.MethodDelete, "http://localhost:8000/ns/test/2", "")
	checkErr(t, err)
	checkResponseCode(t, "check delete", http.StatusAccepted, code)
	checkResponse(t, "check delete", response, `{}`)

	// check again key/values
	code, response, err = call(http.MethodGet, "http://localhost:8000/ns/test", "")
	checkErr(t, err)
	checkResponseCode(t, "check values", http.StatusOK, code)
	checkResponse(t, "check values", response, `[{"key":"1","value":{"age":25,"name":"jack"}},{"key":"3","value":{"age":25,"name":"jack"}}]`)

	// delete all namespace
	code, response, err = call(http.MethodDelete, "http://localhost:8000/ns/test", "")
	checkErr(t, err)
	checkResponseCode(t, "check delete", http.StatusAccepted, code)
	checkResponse(t, "check delete", response, `{}`)

	// check namespace doesn't exist
	code, response, err = call(http.MethodGet, "http://localhost:8000/ns", "")
	checkErr(t, err)
	checkResponseCode(t, "check namespace", http.StatusOK, code)
	checkResponse(t, "check namespace", response, `[]`)
}

func call(method string, URL string, payload string) (code int, response string, err error) {
	req, err := http.NewRequest(method, URL, strings.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Add("Accept", `application/json`)
	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return resp.StatusCode, string(body), nil
}
