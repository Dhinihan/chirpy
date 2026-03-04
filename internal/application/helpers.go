package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func ExtractBody(w http.ResponseWriter, req *http.Request, requestData any) error {
	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		RespondWithError(w, 500, "Something went wrong", err)
		return errors.New("Erro ao ler o Body")
	}
	if err := json.Unmarshal(data, requestData); err != nil {
		RespondWithError(w, 400, "expected json with 'body' key", err)
		return errors.New("Erro ao decodificar o json")
	}
	return nil
}
