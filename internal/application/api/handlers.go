package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Dhinihan/chirpy/internal/application"
	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/database"
	"github.com/Dhinihan/chirpy/internal/model/chirp"
	"github.com/Dhinihan/chirpy/internal/model/user"
)

var cfg *admin.ApiConfig

func RegisterHandlers(c *admin.ApiConfig, serverMux *http.ServeMux) {
	serverMux.HandleFunc("GET /api/healthz", handleHealthZ)
	serverMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	serverMux.HandleFunc("POST /api/users", handleCreateUser)
	cfg = c
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
		CleanedBody string `json:"cleaned_body"`
	}{chirp.CleanMessage(requestData.Body)})
	return
}

func handleCreateUser(w http.ResponseWriter, req *http.Request) {
	data, err := io.ReadAll(req.Body)
	defer req.Body.Close()
	if err != nil {
		application.RespondWithError(w, 400, "Não foi possível ler a requisição", err)
		return
	}
	var postData struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &postData); err != nil {
		application.RespondWithError(w, 400, "Não foi possível ler a requisição", err)
		return
	}
	user := user.NewUser(postData.Email)
	created, err := cfg.Db.CreateUser(req.Context(), database.CreateUserParams{
		ID:    user.ID,
		Email: user.Email,
	})
	if err != nil {
		application.RespondWithError(w, 500, "Não foi possível criar o usuário", err)
		return
	}
	user.Sync(created.CreatedAt, created.UpdatedAt)
	jsonUser, err := json.Marshal(user)
	if err != nil {
		application.RespondWithError(w, 500, "Erro inesperado ao montar a resposta", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	fmt.Fprintln(w, string(jsonUser))
}
