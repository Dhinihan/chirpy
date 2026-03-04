package api

import (
	"net/http"

	"github.com/Dhinihan/chirpy/internal/application"
	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/database"
	"github.com/Dhinihan/chirpy/internal/model/chirp"
	"github.com/Dhinihan/chirpy/internal/model/user"
	"github.com/google/uuid"
)

var cfg *admin.ApiConfig

func RegisterHandlers(c *admin.ApiConfig, serverMux *http.ServeMux) {
	serverMux.HandleFunc("GET /api/healthz", handleHealthZ)
	serverMux.HandleFunc("POST /api/users", handleCreateUser)
	serverMux.HandleFunc("POST /api/chirps", handleCreateChirp)
	cfg = c
}

func handleHealthZ(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func handleCreateChirp(w http.ResponseWriter, req *http.Request) {
	var requestData struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	if err := application.ExtractBody(w, req, &requestData); err != nil {
		return
	}
	valid, msg := chirp.ValidateMessage(requestData.Body)
	if !valid {
		application.RespondWithError(w, 400, msg, nil)
		return
	}
	dataFound, err := cfg.Db.GetUser(req.Context(), requestData.UserId)
	if err != nil {
		application.RespondWithError(w, 404, "Usuário não encontrado", err)
	}
	user := dataFound.ToUser()
	chirp := chirp.NewChirp(user, chirp.CleanMessage(requestData.Body))
	created, err := cfg.Db.CreateChirp(req.Context(), database.CreateChirpParams{
		ID:     chirp.ID,
		Body:   chirp.Body,
		UserID: chirp.UserID,
	})
	if err != nil {
		application.RespondWithError(w, 500, "Não foi possível criar o usuário", err)
		return
	}
	chirp.Sync(created.CreatedAt, created.UpdatedAt)
	application.RespondWithJson(w, 201, chirp)
}

func handleCreateUser(w http.ResponseWriter, req *http.Request) {
	var postData struct {
		Email string `json:"email"`
	}
	if err := application.ExtractBody(w, req, &postData); err != nil {
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
	application.RespondWithJson(w, 201, user)
}
