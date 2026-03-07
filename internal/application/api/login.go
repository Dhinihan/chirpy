package api

import (
	"context"
	"time"

	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/auth"
	"github.com/Dhinihan/chirpy/internal/database"
	"github.com/Dhinihan/chirpy/internal/model/user"
	"github.com/alexedwards/argon2id"
)

type LoginError struct {
	code int
	msg  string
	orig error
}

func login(
	context context.Context,
	apiConfig *admin.ApiConfig,
	email, password string,
) (user.User, *LoginError) {
	found, err := apiConfig.Db.GetUserByEmail(context, email)
	if err != nil {
		return user.User{}, &LoginError{
			401,
			"Incorrect email or password",
			err,
		}
	}
	userEntity := found.ToUser()
	match, err := argon2id.ComparePasswordAndHash(
		password,
		userEntity.HashedPassword,
	)
	if err != nil || !match {
		return user.User{}, &LoginError{
			401,
			"Incorrect email or password",
			err,
		}
	}
	ExpireSeconds := 3600
	token, err := auth.MakeJWT(
		userEntity.ID,
		apiConfig.JwtSecret,
		time.Duration(ExpireSeconds)*time.Second,
	)
	if err != nil {
		return user.User{}, &LoginError{
			500,
			"erro ao gerar token",
			err,
		}
	}
	userEntity.AuthToken = token
	refresh := auth.MakeRefreshToken()
	userEntity.RefreshToken = refresh

	apiConfig.Db.CreateRefreshToken(
		context,
		database.CreateRefreshTokenParams{
			Token:  refresh,
			UserID: userEntity.ID,
		},
	)
	return userEntity, nil
}
