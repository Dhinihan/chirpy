package api

import (
	"net/http"
	"time"

	"github.com/Dhinihan/chirpy/internal/application"
	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/auth"
	"github.com/Dhinihan/chirpy/internal/database"
	"github.com/Dhinihan/chirpy/internal/model/chirp"
	"github.com/Dhinihan/chirpy/internal/model/user"
	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

var cfg *admin.ApiConfig

func RegisterHandlers(c *admin.ApiConfig, serverMux *http.ServeMux) {
	serverMux.HandleFunc("GET /api/healthz", handleHealthZ)
	serverMux.HandleFunc("POST /api/users", handleCreateUser)
	serverMux.HandleFunc("POST /api/login", handleLogin)
	serverMux.HandleFunc("POST /api/chirps", handleCreateChirp)
	serverMux.HandleFunc("GET /api/chirps", handleGetAllChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpID}", handleGetChirp)
	cfg = c
}

func handleHealthZ(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func handleCreateChirp(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		application.RespondWithError(w, 401, "unauthorized", err)
	}
	uid, err := auth.ValidateJWT(token, cfg.JwtSecret)
	var requestData struct {
		Body string `json:"body"`
	}
	if err := application.ExtractBody(w, req, &requestData); err != nil {
		application.RespondWithError(
			w,
			400,
			"Erro ao ler requisição",
			nil,
		)
		return
	}
	valid, msg := chirp.ValidateMessage(requestData.Body)
	if !valid {
		application.RespondWithError(w, 400, msg, nil)
		return
	}
	dataFound, err := cfg.Db.GetUser(req.Context(), uid)
	if err != nil {
		application.RespondWithError(
			w,
			404,
			"Usuário não encontrado",
			err,
		)
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

func handleGetAllChirps(w http.ResponseWriter, req *http.Request) {
	found, err := cfg.Db.GetAllChirps(req.Context())
	if err != nil {
		application.RespondWithError(
			w,
			500,
			"Erro ao buscar os chirps",
			err,
		)
		return
	}
	chirps := make([]chirp.Chirp, len(found))
	for i, v := range found {
		chirps[i] = v.ToChirp()
	}
	application.RespondWithJson(w, 200, chirps)
}

func handleGetChirp(w http.ResponseWriter, req *http.Request) {
	uid, err := uuid.Parse(req.PathValue("chirpID"))
	if err != nil {
		application.RespondWithError(w, 400, "id inválido", err)
		return
	}
	found, err := cfg.Db.GetChirp(req.Context(), uid)
	if err != nil {
		application.RespondWithError(
			w,
			404,
			"Chirp não encontrado",
			err,
		)
		return
	}
	chirp := found.ToChirp()
	application.RespondWithJson(w, 200, chirp)
}

func handleCreateUser(w http.ResponseWriter, req *http.Request) {
	var postData struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := application.ExtractBody(w, req, &postData); err != nil {
		application.RespondWithError(
			w,
			400,
			"Erro ao processar requisição",
			err,
		)
		return
	}
	user, err := user.NewUser(postData.Email, postData.Password)
	if err != nil {
		application.RespondWithError(
			w,
			500,
			"Erro ao processar a senha",
			err,
		)
	}
	created, err := cfg.Db.CreateUser(
		req.Context(),
		database.CreateUserParams{
			ID:             user.ID,
			Email:          user.Email,
			HashedPassword: user.HashedPassword,
		},
	)
	if err != nil {
		application.RespondWithError(w, 500, "Não foi possível criar o usuário", err)
		return
	}
	user.Sync(created.CreatedAt, created.UpdatedAt)
	application.RespondWithJson(w, 201, user)
}

func handleLogin(w http.ResponseWriter, req *http.Request) {
	var postData struct {
		Password      string `json:"password"`
		Email         string `json:"email"`
		ExpireSeconds int    `json:"expire_in_seconds"`
	}
	if err := application.ExtractBody(w, req, &postData); err != nil {
		application.RespondWithError(
			w,
			400,
			"Erro ao processar requisição",
			err,
		)
		return
	}
	found, err := cfg.Db.GetUserByEmail(req.Context(), postData.Email)
	if err != nil {
		application.RespondWithError(
			w,
			401,
			"Incorrect email or password",
			err,
		)
		return
	}
	user := found.ToUser()
	match, err := argon2id.ComparePasswordAndHash(
		postData.Password,
		user.HashedPassword,
	)
	if err != nil || !match {
		application.RespondWithError(
			w,
			401,
			"Incorrect email or password",
			err,
		)
		return
	}
	if postData.ExpireSeconds <= 0 || postData.ExpireSeconds > 3600 {
		postData.ExpireSeconds = 3600
	}
	token, err := auth.MakeJWT(
		user.ID,
		cfg.JwtSecret,
		time.Duration(postData.ExpireSeconds)*time.Second,
	)
	if err != nil {
		application.RespondWithError(
			w,
			500,
			"erro ao gerar token",
			err,
		)
		return
	}
	user.AuthToken = token
	application.RespondWithJson(w, 200, user)
}
