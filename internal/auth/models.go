package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Version      int64      `json:"version"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	DeletedAt    *time.Time `json:"deleted_at"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SignUpRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignResponse struct {
	User   User   `json:"user"`
	Tokens Tokens `json:"tokens"`
}

type RefreshToken struct {
	ID      uuid.UUID
	Version int64
	UserID  uuid.UUID

	TokenHash string

	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}
