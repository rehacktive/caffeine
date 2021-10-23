package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func respondWithJSON(w http.ResponseWriter, code int, jsonContent string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write([]byte(jsonContent + "\n"))
	if err != nil {
		log.Println("error sending response: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, fmt.Sprintf("{\"error\": \"%v\"}", message))
}

func jsonWrapper(payload interface{}) (content []byte, err error) {
	unboxed, ok := payload.(map[string][]byte)
	if !ok {
		content, err = json.Marshal(payload)
		return
	}
	r := make([]map[string]interface{}, 0)
	for k, v := range unboxed {
		var parsed interface{}
		err = json.Unmarshal(v, &parsed)
		if err != nil {
			return
		}
		r = append(r, map[string]interface{}{"key": k, "value": parsed})
	}
	content, err = json.Marshal(r)
	return
}
