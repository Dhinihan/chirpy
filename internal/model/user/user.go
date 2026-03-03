package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func NewUser(email string) User {
	return User{ID: uuid.New(), Email: email}
}

func (u *User) Sync(createdAt time.Time, updatedAt time.Time) {
	u.CreatedAt = createdAt
	u.UpdatedAt = updatedAt
}
