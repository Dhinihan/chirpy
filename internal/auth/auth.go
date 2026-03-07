package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (hash string, err error) {
	params := argon2id.DefaultParams
	hash, err = argon2id.CreateHash(password, params)
	if err != nil {
		return "", fmt.Errorf("Erro ao montar hash:\n%w", err)
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (match bool, err error) {
	match, err = argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("Erro ao comparar hash:\n%w", err)
	}
	return match, nil
}

func MakeJWT(
	userID uuid.UUID,
	tokenSecret string,
	expiresIn time.Duration,
) (string, error) {
	exp := jwt.NewNumericDate(time.Now().Add(expiresIn))
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Issuer:    "chirpy-access",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: exp,
			Subject:   userID.String(),
		},
	)

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("Token invalido:\n%w", err)
	}
	return uuid.Parse(claims.Subject)
}

func GetBearerToken(headers http.Header) (string, error) {
	authToken := headers.Get("Authorization")
	atParts := strings.Split(authToken, " ")
	stdMsg := "%s não está no formato 'Bearer <token>'"
	stdErr := fmt.Errorf(stdMsg, authToken)
	if len(atParts) != 2 {
		return "", stdErr
	}
	if strings.ToLower(atParts[0]) != "bearer" {
		return "", stdErr
	}
	return atParts[1], nil
}

func MakeRefreshToken() string {
	random := make([]byte, 32)
	rand.Read(random)
	return hex.EncodeToString(random)
}
