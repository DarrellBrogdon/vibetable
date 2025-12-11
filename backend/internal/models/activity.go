package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Activity action types
const (
	ActionCreate = "create"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// Activity entity types
const (
	EntityTypeRecord = "record"
	EntityTypeField  = "field"
	EntityTypeTable  = "table"
	EntityTypeView   = "view"
	EntityTypeBase   = "base"
)

type Activity struct {
	ID         uuid.UUID       `json:"id"`
	BaseID     uuid.UUID       `json:"base_id"`
	TableID    *uuid.UUID      `json:"table_id,omitempty"`
	RecordID   *uuid.UUID      `json:"record_id,omitempty"`
	UserID     uuid.UUID       `json:"user_id"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityName *string         `json:"entity_name,omitempty"`
	Changes    json.RawMessage `json:"changes,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`

	// Joined fields (not in database)
	User *User `json:"user,omitempty"`
}

// ActivityChanges represents the changes made in an update action
type ActivityChanges struct {
	FieldID   string      `json:"field_id,omitempty"`
	FieldName string      `json:"field_name,omitempty"`
	OldValue  interface{} `json:"old_value,omitempty"`
	NewValue  interface{} `json:"new_value,omitempty"`
}
