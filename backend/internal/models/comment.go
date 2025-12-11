package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID         uuid.UUID  `json:"id"`
	RecordID   uuid.UUID  `json:"record_id"`
	UserID     uuid.UUID  `json:"user_id"`
	Content    string     `json:"content"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	IsResolved bool       `json:"is_resolved"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Joined fields (not in database)
	User    *User      `json:"user,omitempty"`
	Replies []*Comment `json:"replies,omitempty"`
}
