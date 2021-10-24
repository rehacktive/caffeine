package service

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
	Upsert(namespace string, key string, value []byte) error
	Get(namespace string, key string) ([]byte, error)
	GetAll(namespace string) (map[string][]byte, error)
	Delete(namespace string, key string) error
	DeleteAll(namespace string) error
	GetNamespaces() []string
}

const (
	NamespacePattern = "/ns/{namespace:[a-zA-Z0-9]+}"
	KeyValuePattern  = "/ns/{namespace:[a-zA-Z0-9]+}/{key:[a-zA-Z0-9]+}"
	SearchPattern    = "/search/{namespace:[a-zA-Z0-9]+}"
)

type Server struct {
	Address string
	router  *mux.Router
	db      Database
}

func (s *Server) Init(db Database) {
	s.db = db
	s.db.Init()

	s.router = mux.NewRouter()
	s.router.HandleFunc("/ns", s.homeHandler).Methods(http.MethodGet, http.MethodOptions)
	s.router.HandleFunc(NamespacePattern, s.namespaceHandler).Methods(http.MethodGet, http.MethodDelete, http.MethodOptions)
	s.router.HandleFunc(KeyValuePattern, s.keyvalueHandler).Methods(http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions)
	s.router.HandleFunc(SearchPattern, s.searchHandler).Queries("filter", "{filter}").Methods(http.MethodGet, http.MethodOptions)
	s.router.Use(mux.CORSMethodMiddleware(s.router))

	srv := &http.Server{
		Handler:      s.router,
		Addr:         s.Address,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	namespaces, err := jsonWrapper(s.db.GetNamespaces())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, string(namespaces))
}

func (s *Server) namespaceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	namespace := vars["namespace"]

	switch r.Method {
	case "GET":
		data, err := s.db.GetAll(namespace)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		namespaceData, err := jsonWrapper(data)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, string(namespaceData))

	case "DELETE":
		err := s.db.DeleteAll(namespace)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		respondWithJSON(w, http.StatusAccepted, "{}")
	}
}

func (s *Server) keyvalueHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	key := vars["key"]

	switch r.Method {
	case "POST":
		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		var parsed interface{}
		err = json.Unmarshal(data, &parsed)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		err = s.db.Upsert(namespace, key, data)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithJSON(w, http.StatusCreated, string(data))
	case "GET":
		data, err := s.db.Get(namespace, key)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, string(data))
	case "DELETE":
		err := s.db.Delete(namespace, key)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		respondWithJSON(w, http.StatusAccepted, "{}")
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		return
	}

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
			log.Println("error on parsing", err)
			respondWithError(w, http.StatusBadRequest, err.Error())
		}
		data, err := s.db.GetAll(vars["namespace"])
		if err != nil {
			log.Println("error on GetAll", err)
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		for key, value := range data {
			var jsonContent map[string]interface{}
			err := json.Unmarshal(value, &jsonContent)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
			}
			iter := query.Run(jsonContent)
			for {
				v, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := v.(error); ok {
					log.Println("error on query", err)
					respondWithError(w, http.StatusInternalServerError, err.Error())
				}
				result.Results = append(result.Results, map[string]interface{}{"key": key, "value": v})
			}
		}
		jsonResponse, _ := json.Marshal(result)
		respondWithJSON(w, http.StatusOK, string(jsonResponse))
	}
}
