package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID        uuid.UUID       `json:"id"`
	TableID   uuid.UUID       `json:"table_id"`
	Values    json.RawMessage `json:"values"` // Map of field_id -> value
	Position  int             `json:"position"`
	Color     *string         `json:"color,omitempty"` // Optional color for visual highlighting
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
