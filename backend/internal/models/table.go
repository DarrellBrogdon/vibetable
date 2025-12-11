package models

import (
	"time"

	"github.com/google/uuid"
)

type Table struct {
	ID        uuid.UUID `json:"id"`
	BaseID    uuid.UUID `json:"base_id"`
	Name      string    `json:"name"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
