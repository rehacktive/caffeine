package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/itchyny/gojq"
)

type Database interface {
	Init()
	Upsert(namespace string, key string, value []byte) (bool, error) // POST /{namespace}/{key} body: value
	Get(namespace string, key string) ([]byte, error)                // GET /{namespace}/{key} return: value
	GetAll(namespace string) (map[string][]byte, error)              // GET /{namespace} return key/value map
	Delete(namespace string, key string) (bool, error)               // DELETE /{namespace}/{id}
	DeleteAll(namespace string) (bool, error)                        // DELETE /{namespace}
	GetNamespaces() []string                                         // GET /
}

type Server struct {
	address string
	router  *mux.Router
	db      Database
}

func (s *Server) Init(db Database) {
	s.db = db
	s.db.Init()

	s.router = mux.NewRouter()
	s.router.HandleFunc("/", s.homeHandler)
	s.router.HandleFunc("/{namespace:[a-zA-Z0-9]+}", s.namespaceHandler)
	s.router.HandleFunc("/{namespace:[a-zA-Z0-9]+}/{key:[a-zA-Z0-9]+}", s.keyvalueHandler)
	s.router.HandleFunc("/search/{namespace:[a-zA-Z0-9]+}", s.searchHandler).Queries("filter", "{filter}")

	srv := &http.Server{
		Handler:      s.router,
		Addr:         s.address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	namespaces, err := jsonWrapper(s.db.GetNamespaces())
	if err != nil {
		respondWithError(w, 400, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, string(namespaces))
}

func (s *Server) namespaceHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		vars := mux.Vars(r)
		data, err := s.db.GetAll(vars["namespace"])
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		namespace, err := jsonWrapper(data)
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, string(namespace))

	case "DELETE":
		// call s.db.deleteAll()
		respondWithError(w, 400, "to be implemented")
	}
}

func (s *Server) keyvalueHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	namespace := vars["namespace"]
	key := vars["key"]
	switch r.Method {
	case "POST":
		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		var parsed interface{}
		err = json.Unmarshal(data, &parsed)
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		ok, err := s.db.Upsert(namespace, key, data)
		if !ok || err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, string(data))
	case "GET":
		data, err := s.db.Get(namespace, key)
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, string(data))
	case "DELETE":
		// call s.db.delete()
		respondWithError(w, 400, "to be implemented")
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	type Result struct {
		Results []interface{} `json:"results"`
	}
	result := Result{
		Results: make([]interface{}, 0),
	}

	switch r.Method {
	case "GET":
		vars := mux.Vars(r)
		query, err := gojq.Parse(vars["filter"])
		if err != nil {
			log.Fatalln(err)
		}
		data, err := s.db.GetAll(vars["namespace"])
		if err != nil {
			respondWithError(w, 400, err.Error())
			return
		}
		for key, value := range data {
			var jsonContent map[string]interface{}
			err := json.Unmarshal(value, &jsonContent)
			if err != nil {
				respondWithError(w, 400, err.Error())
			}
			iter := query.Run(jsonContent)
			for {
				v, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := v.(error); ok {
					respondWithError(w, 400, err.Error())
				}
				result.Results = append(result.Results, map[string]interface{}{key: v})
			}
		}
		jsonResponse, _ := json.Marshal(result)
		respondWithJSON(w, http.StatusOK, string(jsonResponse))
	}
}
