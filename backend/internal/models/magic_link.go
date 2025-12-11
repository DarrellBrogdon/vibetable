package models

import (
	"time"

	"github.com/google/uuid"
)

type MagicLink struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Token     string     `json:"-"` // Never expose token in JSON
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
