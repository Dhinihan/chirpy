package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Dhinihan/chirpy/internal/application"
	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/auth"
	"github.com/Dhinihan/chirpy/internal/database"
	"github.com/Dhinihan/chirpy/internal/model/chirp"
	"github.com/Dhinihan/chirpy/internal/model/user"
	"github.com/google/uuid"
)

var cfg *admin.ApiConfig

func RegisterHandlers(c *admin.ApiConfig, serverMux *http.ServeMux) {
	serverMux.HandleFunc("GET /api/healthz", handleHealthZ)
	serverMux.HandleFunc("POST /api/users", handleCreateUser)
	serverMux.HandleFunc("PUT /api/users", handleUpadateUser)
	serverMux.HandleFunc("POST /api/login", handleLogin)
	serverMux.HandleFunc("POST /api/refresh", handleRefreshToken)
	serverMux.HandleFunc("POST /api/revoke", handleRevokeToken)
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
		return
	}
	uid, err := auth.ValidateJWT(token, cfg.JwtSecret)
	if err != nil {
		application.RespondWithError(w, 401, "unauthorized", err)
		return
	}
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
		return
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

func handleUpadateUser(w http.ResponseWriter, req *http.Request) {
	var requestData struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	uid, err := application.ExtractAuthBody(
		w,
		req,
		cfg.JwtSecret,
		&requestData,
	)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	hash, err := auth.HashPassword(requestData.Password)
	if err != nil {
		application.RespondWithError(
			w,
			500,
			"Erro ao gerar hash",
			err,
		)
		return
	}
	found, err := cfg.Db.UpdateUserCredentials(
		req.Context(),
		database.UpdateUserCredentialsParams{
			ID:             uid,
			Email:          requestData.Email,
			HashedPassword: hash,
		},
	)
	if err != nil {
		application.RespondWithError(
			w,
			500,
			"Erro ao atualizar usuário",
			err,
		)
		return
	}
	userUpdated := found.ToUser()
	application.RespondWithJson(w, 200, userUpdated)
}

func handleLogin(w http.ResponseWriter, req *http.Request) {
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
	user, lErr := login(
		req.Context(),
		cfg,
		postData.Email,
		postData.Password,
	)
	if lErr != nil {
		application.RespondWithError(
			w,
			lErr.code,
			lErr.msg,
			lErr.orig,
		)
		return
	}
	application.RespondWithJson(w, 200, user)
}

func handleRefreshToken(w http.ResponseWriter, req *http.Request) {
	authToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		application.RespondWithError(
			w,
			401,
			"autorização mal formatada",
			err,
		)
		return
	}
	uid, err := cfg.Db.CheckRefreshToken(req.Context(), authToken)
	if err != nil {
		application.RespondWithError(
			w,
			401,
			"Token inválido",
			err,
		)
		return
	}
	token, err := auth.MakeJWT(uid, cfg.JwtSecret, time.Hour)
	if err != nil {
		application.RespondWithError(
			w,
			500,
			"Erro ao gerar o token",
			err,
		)
		return
	}
	application.RespondWithJson(w, 200, struct {
		Token string `json:"token"`
	}{token})

}

func handleRevokeToken(w http.ResponseWriter, req *http.Request) {
	authToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		application.RespondWithError(
			w,
			401,
			"autorização mal formatada",
			err,
		)
		return
	}
	if err := cfg.Db.RevokeRefreshToken(
		req.Context(),
		authToken,
	); err != nil {
		application.RespondWithError(
			w,
			401,
			"Token inválido",
			err,
		)
		return
	}
	w.WriteHeader(204)
}
