package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	type args struct {
		password string
	}
	expHash, _ := argon2id.CreateHash(
		"senha qualquer",
		argon2id.DefaultParams,
	)
	tests := []struct {
		name     string
		args     args
		wantHash string
		wantErr  bool
	}{
		{"Hash normal", args{"senha qualquer"}, expHash, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHash, err := HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			resultado, _ := argon2id.ComparePasswordAndHash(
				tt.args.password,
				tt.wantHash,
			)
			if !resultado {
				t.Errorf("HashPassword() = %v, want %v", gotHash, tt.wantHash)
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	type args struct {
		password string
		hash     string
	}
	senha1 := "senha um"
	senha2 := "senha dois"
	hash1, _ := argon2id.CreateHash(senha1, argon2id.DefaultParams)
	hash2, _ := argon2id.CreateHash(senha2, argon2id.DefaultParams)
	tests := []struct {
		name      string
		args      args
		wantMatch bool
		wantErr   bool
	}{
		{"senha1, hash1", args{senha1, hash1}, true, false},
		{"senha2, hash1", args{senha2, hash1}, false, false},
		{"senha1, hash2", args{senha1, hash2}, false, false},
		{"senha2, hash2", args{senha2, hash2}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatch, err := CheckPasswordHash(tt.args.password, tt.args.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gotMatch != tt.wantMatch {
				t.Errorf("CheckPasswordHash() = %v, want %v", gotMatch, tt.wantMatch)
			}
		})
	}
}

func TestMakeJWT(t *testing.T) {
	type args struct {
		userID      uuid.UUID
		tokenSecret string
		expiresIn   time.Duration
	}
	const secret = "segredo correto"
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"Válido",
			args{uuid.New(), secret, time.Hour},
			false,
		},
		{
			"segredo errado",
			args{uuid.New(), "segredo errado", time.Hour},
			true,
		},
		{
			"expirado",
			args{uuid.New(), secret, -1 * time.Hour},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err1 := MakeJWT(
				tt.args.userID,
				tt.args.tokenSecret,
				tt.args.expiresIn,
			)

			uid, err2 := ValidateJWT(got, secret)
			if (err1 != nil || err2 != nil) != tt.wantErr {
				t.Fatalf(
					"MakeJWT() errs=%v,%v , wantErr %v",
					err1,
					err2,
					tt.wantErr,
				)
			}
			if tt.wantErr {
				return
			}
			if uid.String() != tt.args.userID.String() {
				t.Fatalf(
					"%s é diferente de %s",
					uid,
					tt.args.userID,
				)
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	type args struct {
		headerName string
		header     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"Valido",
			args{"Authorization", "bearer token"},
			"token",
			false,
		},
		{
			"outro header",
			args{"Errado", "bearer token"},
			"token",
			true,
		},
		{
			"três strings",
			args{"Authorization", "bearer token errado"},
			"token",
			true,
		},
		{
			"Uma string",
			args{"Authorization", "bearer"},
			"token",
			true,
		},
		{
			"Outro prefixo",
			args{"Authorization", "basic token"},
			"token",
			true,
		},
		{
			"Ignora case",
			args{"Authorization", "beARer token"},
			"token",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := http.Header{}
			headers.Add(tt.args.headerName, tt.args.header)
			got, err := GetBearerToken(headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetBearerToken() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("GetBearerToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
