package models

import (
	"time"

	"github.com/google/uuid"
)

type Form struct {
	ID               uuid.UUID   `json:"id"`
	TableID          uuid.UUID   `json:"table_id"`
	Name             string      `json:"name"`
	Description      *string     `json:"description,omitempty"`
	PublicToken      *string     `json:"public_token,omitempty"`
	IsActive         bool        `json:"is_active"`
	SuccessMessage   string      `json:"success_message"`
	RedirectURL      *string     `json:"redirect_url,omitempty"`
	SubmitButtonText string      `json:"submit_button_text"`
	CreatedBy        uuid.UUID   `json:"created_by"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Fields           []FormField `json:"fields,omitempty"`
}

type FormField struct {
	ID         uuid.UUID `json:"id"`
	FormID     uuid.UUID `json:"form_id"`
	FieldID    uuid.UUID `json:"field_id"`
	Label      *string   `json:"label,omitempty"`
	HelpText   *string   `json:"help_text,omitempty"`
	IsRequired bool      `json:"is_required"`
	IsVisible  bool      `json:"is_visible"`
	Position   int       `json:"position"`
	// Denormalized field info for public forms
	FieldName string `json:"field_name,omitempty"`
	FieldType string `json:"field_type,omitempty"`
	FieldOptions any `json:"field_options,omitempty"`
}

// PublicForm is a stripped-down version for public access
type PublicForm struct {
	ID               uuid.UUID         `json:"id"`
	Name             string            `json:"name"`
	Description      *string           `json:"description,omitempty"`
	SuccessMessage   string            `json:"success_message"`
	RedirectURL      *string           `json:"redirect_url,omitempty"`
	SubmitButtonText string            `json:"submit_button_text"`
	Fields           []PublicFormField `json:"fields"`
}

type PublicFormField struct {
	FieldID      uuid.UUID `json:"field_id"`
	Label        string    `json:"label"`
	HelpText     *string   `json:"help_text,omitempty"`
	IsRequired   bool      `json:"is_required"`
	FieldType    string    `json:"field_type"`
	FieldOptions any       `json:"field_options,omitempty"`
	Position     int       `json:"position"`
}
