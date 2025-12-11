package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ViewType string

const (
	ViewTypeGrid     ViewType = "grid"
	ViewTypeKanban   ViewType = "kanban"
	ViewTypeCalendar ViewType = "calendar"
	ViewTypeGallery  ViewType = "gallery"
)

func IsValidViewType(vt ViewType) bool {
	switch vt {
	case ViewTypeGrid, ViewTypeKanban, ViewTypeCalendar, ViewTypeGallery:
		return true
	default:
		return false
	}
}

type View struct {
	ID          uuid.UUID       `json:"id"`
	TableID     uuid.UUID       `json:"table_id"`
	Name        string          `json:"name"`
	Type        ViewType        `json:"type"`
	Config      json.RawMessage `json:"config"`
	Position    int             `json:"position"`
	PublicToken *string         `json:"public_token,omitempty"`
	IsPublic    bool            `json:"is_public"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// PublicView is a view with its table and fields for public access
type PublicView struct {
	View    *View    `json:"view"`
	Table   *Table   `json:"table"`
	Fields  []*Field `json:"fields"`
	Records []*Record `json:"records"`
}

// ViewConfig contains view-specific configuration
type ViewConfig struct {
	// For grid view
	Filters []ViewFilter `json:"filters,omitempty"`
	Sorts   []ViewSort   `json:"sorts,omitempty"`

	// For kanban view
	GroupByFieldID string `json:"group_by_field_id,omitempty"`

	// For calendar view
	DateFieldID  string `json:"date_field_id,omitempty"`
	TitleFieldID string `json:"title_field_id,omitempty"`

	// For gallery view
	CoverFieldID string `json:"cover_field_id,omitempty"`

	// Column visibility/order
	VisibleFields []string `json:"visible_fields,omitempty"`
}

type ViewFilter struct {
	FieldID  string `json:"field_id"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type ViewSort struct {
	FieldID   string `json:"field_id"`
	Direction string `json:"direction"` // "asc" or "desc"
}
