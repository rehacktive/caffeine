package service

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/rehacktive/caffeine/database"

	"github.com/gorilla/mux"
	"github.com/itchyny/gojq"
	"github.com/rs/cors"
	"github.com/xeipuuv/gojsonschema"
)

type Database interface {
	Init()
	Upsert(namespace string, key string, value []byte) *database.DbError
	Get(namespace string, key string) ([]byte, *database.DbError)
	GetAll(namespace string) (map[string][]byte, *database.DbError)
	Delete(namespace string, key string) *database.DbError
	DeleteAll(namespace string) *database.DbError
	GetNamespaces() []string
}

const (
	NamespacePattern = "/ns/{namespace:[a-zA-Z0-9]+}"
	KeyValuePattern  = "/ns/{namespace:[a-zA-Z0-9]+}/{key:[a-zA-Z0-9]+}"
	SearchPattern    = "/search/{namespace:[a-zA-Z0-9]+}"
	SchemaPattern    = "/schema/{namespace:[a-zA-Z0-9]+}"
	OpenAPIPattern   = "/{openapi|swagger}.json"
	BrokerPattern    = "/broker"
	SwaggerUIPattern = "/swaggerui/"
	SchemaId         = "_schema"

	EVENT_ITEM_ADDED        = "ITEM_ADDED"
	EVENT_ITEM_DELETED      = "ITEM_DELETED"
	EVENT_NAMESPACE_DELETED = "NAMESPACE_DELETED"
)

var (
	ErrInvalidArguments = errors.New("invalid arguments")
)

type Server struct {
	Address string
	router  *mux.Router
	db      Database
	broker  *Broker
}

func (s *Server) Init(db Database) {
	s.db = db
	s.db.Init()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // All origins
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete},
		AllowedHeaders: []string{"X-Content-Type", "text/plain"},
	})

	s.broker = NewServer()

	s.router = mux.NewRouter()
	s.router.HandleFunc("/ns", s.homeHandler)
	s.router.HandleFunc(NamespacePattern, s.namespaceHandler)
	s.router.HandleFunc(KeyValuePattern, s.keyValueHandler)
	s.router.HandleFunc(SearchPattern, s.searchHandler).Queries("filter", "{filter}")
	s.router.HandleFunc(SchemaPattern, s.schemaHandler)
	s.router.HandleFunc(OpenAPIPattern, s.openAPIHandler)
	s.router.PathPrefix(SwaggerUIPattern).Handler(http.StripPrefix(SwaggerUIPattern, http.FileServer(http.Dir("./swagger-ui/"))))
	s.router.Handle(BrokerPattern, s.broker)
	s.router.Use(mux.CORSMethodMiddleware(s.router))

	srv := &http.Server{
		Handler:      c.Handler(s.router),
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
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	namespace := vars["namespace"]

	switch r.Method {
	case http.MethodPost:
		respondWithError(w, http.StatusNotImplemented, "cannot POST to this endpoint!")
	case http.MethodGet:
		data, dbErr := s.db.GetAll(namespace)
		if dbErr != nil {
			switch dbErr.ErrorCode {
			case database.NAMESPACE_NOT_FOUND:
				respondWithError(w, http.StatusBadRequest, dbErr.Error())
			default:
				respondWithError(w, http.StatusInternalServerError, dbErr.Error())
			}
		}
		namespaceData, err := jsonWrapper(data)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, string(namespaceData))

	case http.MethodDelete:
		dbErr := s.db.DeleteAll(namespace)
		if dbErr != nil {
			switch dbErr.ErrorCode {
			case database.NAMESPACE_NOT_FOUND:
				respondWithError(w, http.StatusBadRequest, dbErr.Error())
			default:
				respondWithError(w, http.StatusInternalServerError, dbErr.Error())
			}
		}
		s.Notify(BrokerEvent{
			Event:     EVENT_NAMESPACE_DELETED,
			Namespace: namespace,
			Key:       "",
			Value:     nil,
		})
		respondWithJSON(w, http.StatusAccepted, "{}")
	}
}

func (s *Server) keyValueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	namespace := vars["namespace"]
	key := vars["key"]

	switch r.Method {
	case http.MethodPost:
		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		parsedData, err := s.validate(namespace, data)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		dbErr := s.db.Upsert(namespace, key, data)
		if dbErr != nil {
			switch dbErr.ErrorCode {
			case database.NAMESPACE_NOT_FOUND:
				respondWithError(w, http.StatusBadRequest, dbErr.Error())
			default:
				respondWithError(w, http.StatusInternalServerError, dbErr.Error())
			}
			return
		}
		s.Notify(BrokerEvent{
			Event:     EVENT_ITEM_ADDED,
			Namespace: namespace,
			Key:       key,
			Value:     parsedData,
		})
		respondWithJSON(w, http.StatusCreated, string(data))
	case http.MethodGet:
		data, dbErr := s.db.Get(namespace, key)
		if dbErr != nil {
			switch dbErr.ErrorCode {
			case database.ID_NOT_FOUND:
				respondWithError(w, http.StatusNotFound, dbErr.Error())
			case database.NAMESPACE_NOT_FOUND:
				respondWithError(w, http.StatusBadRequest, dbErr.Error())
			default:
				respondWithError(w, http.StatusInternalServerError, dbErr.Error())
			}
			return
		}
		respondWithJSON(w, http.StatusOK, string(data))
	case http.MethodDelete:
		err := s.db.Delete(namespace, key)
		if err != nil {

			switch err.ErrorCode {
			case database.ID_NOT_FOUND:
				respondWithError(w, http.StatusNotFound, err.Error())
			case database.NAMESPACE_NOT_FOUND:
				respondWithError(w, http.StatusBadRequest, err.Error())
			default:
				respondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		s.Notify(BrokerEvent{
			Event:     EVENT_ITEM_DELETED,
			Namespace: namespace,
			Key:       key,
			Value:     nil,
		})
		respondWithJSON(w, http.StatusAccepted, "{}")
	}
}

func (s *Server) schemaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	vars := mux.Vars(r)
	namespace := vars["namespace"] + SchemaId

	switch r.Method {
	case http.MethodPost:
		defer r.Body.Close()
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		dbErr := s.db.Upsert(namespace, SchemaId, data)
		if dbErr != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		log.Println("added schema for namespace " + vars["namespace"])
		respondWithJSON(w, http.StatusCreated, string(data))
	case http.MethodGet:
		data, dbErr := s.db.Get(namespace, SchemaId)
		if dbErr != nil {
			respondWithError(w, http.StatusNotFound, dbErr.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, string(data))
	case http.MethodDelete:
		dbErr := s.db.Delete(namespace, SchemaId)
		if dbErr != nil {
			respondWithError(w, http.StatusNotFound, dbErr.Error())
			return
		}
		respondWithJSON(w, http.StatusAccepted, "{}")
	}
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
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
	case http.MethodGet:
		vars := mux.Vars(r)
		query, err := gojq.Parse(vars["filter"])
		if err != nil {
			log.Println("error on parsing", err)
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		data, dbErr := s.db.GetAll(vars["namespace"])
		if dbErr != nil {
			log.Println("error on GetAll", err)
			respondWithError(w, http.StatusBadRequest, dbErr.Error())
			return
		}
		for key, value := range data {
			var jsonContent map[string]interface{}
			err := json.Unmarshal(value, &jsonContent)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, err.Error())
				return
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
					return
				}
				result.Results = append(result.Results, map[string]interface{}{"key": key, "value": v})
			}
		}
		jsonResponse, _ := json.Marshal(result)
		respondWithJSON(w, http.StatusOK, string(jsonResponse))
	}
}

func (s *Server) openAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}

	var namespaces []string = s.db.GetNamespaces()

	rootMap := generateOpenAPIMap(namespaces)

	switch r.Method {
	case http.MethodGet:
		output, err := json.MarshalIndent(rootMap, "", "  ")
		output = append(output, '\n')

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		respondWithJSON(w, http.StatusOK, string(output))
	case http.MethodPost:
		respondWithError(w, http.StatusNotImplemented, "cannot POST to this endpoint!")
	case http.MethodDelete:
		respondWithError(w, http.StatusNotImplemented, "cannot DELETE this endpoint!")
	}
}

// utils

func (s *Server) validate(namespace string, data []byte) (interface{}, error) {
	var parsed interface{}

	// if namespace has a schema, validate against it
	schemaJson, dbErr := s.db.Get(namespace+SchemaId, SchemaId)
	if dbErr == nil {
		schemaLoader := gojsonschema.NewBytesLoader(schemaJson)
		documentLoader := gojsonschema.NewBytesLoader(data)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		if err != nil {
			return nil, err
		}

		if result.Valid() {
			json.Unmarshal(data, &parsed)
		} else {
			log.Printf("The document is not valid according to its schema. see errors :")
			errorLog := ""
			for _, desc := range result.Errors() {
				errorLog = errorLog + desc.String()
			}
			log.Println(errorLog)
			return nil, errors.New(errorLog)
		}
	} else {
		// otherwise just validate as json
		err := json.Unmarshal(data, &parsed)
		if err != nil {
			log.Printf("The document is not valid JSON")
			return nil, err
		}
	}
	return parsed, nil
}

func (s *Server) Notify(event BrokerEvent) {
	if s.broker != nil {
		jsonData, _ := json.Marshal(event)
		s.broker.Notifier <- []byte(jsonData)
	}
}
