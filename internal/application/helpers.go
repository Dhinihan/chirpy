package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Dhinihan/chirpy/internal/auth"
	"github.com/google/uuid"
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

func ExtractBody(
	w http.ResponseWriter,
	req *http.Request,
	requestData any,
) error {
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

func GetAuthUser(
	w http.ResponseWriter,
	req *http.Request,
	secret string,
) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		RespondWithError(w, 401, "unauthorized", err)
		return uuid.UUID{}, errors.New("Erro ao ler token")

	}
	uid, err := auth.ValidateJWT(token, secret)
	if err != nil {
		RespondWithError(w, 401, "unauthorized", err)
		return uuid.UUID{}, errors.New("Token inválido")
	}
	return uid, nil
}

func ExtractAuthBody(
	w http.ResponseWriter,
	req *http.Request,
	secret string,
	requestData any,
) (uuid.UUID, error) {
	defer req.Body.Close()
	uid, err := GetAuthUser(w, req, secret)
	if err != nil {
		return uuid.UUID{}, err
	}
	data, err := io.ReadAll(req.Body)
	if err != nil {
		RespondWithError(w, 500, "Something went wrong", err)
		return uuid.UUID{}, errors.New("Erro ao ler o Body")
	}
	if err := json.Unmarshal(data, requestData); err != nil {
		RespondWithError(
			w,
			400,
			"expected json with 'body' key",
			err,
		)
		return uuid.UUID{}, errors.New("Erro ao decodificar o json")
	}
	return uid, nil
}
