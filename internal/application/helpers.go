package application

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	RespondWithJson(w, code, struct {
		Error string `json:"error"`
	}{Error: msg})
}

func RespondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	js, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(js))
}
