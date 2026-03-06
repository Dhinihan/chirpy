package user

import (
	"errors"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Email          string    `json:"email"`
	HashedPassword string    `json:"-"`
}

func NewUser(email, password string) (User, error) {
	if email == "" || password == "" {
		return User{}, errors.New("email e senha são obrigatórios")
	}
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return User{}, err
	}
	return User{ID: uuid.New(), Email: email, HashedPassword: hash}, nil
}

func (u *User) Sync(createdAt time.Time, updatedAt time.Time) {
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
}
