package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Dhinihan/chirpy/internal/application"
	"github.com/Dhinihan/chirpy/internal/model/chirp"
)

func RegisterHandlers(serverMux *http.ServeMux) {
	serverMux.HandleFunc("GET /api/healthz", handleHealthZ)
	serverMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
}

func handleHealthZ(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		application.RespondWithError(w, 500, "Something went wrong", err)
		return
	}
	var requestData struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(data, &requestData); err != nil {
		application.RespondWithError(w, 400, "expected json with 'body' key", err)
		return
	}
	valid, msg := chirp.ValidateMessage(requestData.Body)
	if !valid {
		application.RespondWithError(w, 400, msg, nil)
		return
	}
	application.RespondWithJson(w, 200, struct {
		Valid bool `json:"valid"`
	}{Valid: valid})
	return
}

