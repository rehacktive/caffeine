package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	log "github.com/sirupsen/logrus"
)

type Payload struct {
	User string      `json:"user_id"`
	Data interface{} `json:"data"`
}

func (p *Payload) wrap() (content []byte, err error) {
	content, err = json.Marshal(p)
	return
}

func respondWithJSON(w http.ResponseWriter, code int, jsonContent string) {
	w.Header().Set("Content-Type", "application/json")

	// w.WriteHeader(code)
	_, err := w.Write([]byte(jsonContent))
	if err != nil {
		log.Println("error sending response: ", err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, fmt.Sprintf(
		`{ "status": %v, "message": "%v" }`, code, message))
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
	// ensure the ascending order
	sort.SliceStable(r, func(i, j int) bool {
		return r[i]["key"].(string) < r[j]["key"].(string)
	})
	content, err = json.Marshal(r)
	return
}
