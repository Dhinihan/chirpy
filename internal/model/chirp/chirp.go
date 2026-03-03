package chirp

import (
	"time"

	"github.com/Dhinihan/chirpy/internal/model/user"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func NewChirp(user user.User, body string) Chirp {
	return Chirp{ID: uuid.New(), Body: body, UserID: user.ID}
}

func (c *Chirp) Sync(createdAt time.Time, updatedAt time.Time) {
	c.CreatedAt = createdAt
	c.UpdatedAt = updatedAt
}
