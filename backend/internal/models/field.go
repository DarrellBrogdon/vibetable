package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type FieldType string

const (
	FieldTypeText         FieldType = "text"
	FieldTypeNumber       FieldType = "number"
	FieldTypeCheckbox     FieldType = "checkbox"
	FieldTypeDate         FieldType = "date"
	FieldTypeSingleSelect FieldType = "single_select"
	FieldTypeMultiSelect  FieldType = "multi_select"
	FieldTypeLinkedRecord FieldType = "linked_record"
	FieldTypeFormula      FieldType = "formula"
	FieldTypeRollup       FieldType = "rollup"
	FieldTypeLookup       FieldType = "lookup"
	FieldTypeAttachment   FieldType = "attachment"
)

// SelectOption represents an option for single/multi select fields
type SelectOption struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// FieldOptions stores type-specific configuration
type FieldOptions struct {
	// Number field options
	Precision *int    `json:"precision,omitempty"`
	Format    *string `json:"format,omitempty"`

	// Date field options
	IncludeTime *bool `json:"include_time,omitempty"`

	// Select field options
	Options []SelectOption `json:"options,omitempty"`

	// Linked record options
	LinkedTableID *uuid.UUID `json:"linked_table_id,omitempty"`

	// Formula field options
	Expression *string `json:"expression,omitempty"` // The formula expression
	ResultType *string `json:"result_type,omitempty"` // 'text', 'number', 'date', 'boolean'

	// Rollup field options
	RollupLinkedFieldID *string `json:"rollup_linked_field_id,omitempty"` // The linked_record field to rollup from
	RollupFieldID       *string `json:"rollup_field_id,omitempty"`        // Field in linked table to aggregate
	AggregationFunction *string `json:"aggregation_function,omitempty"`   // COUNT, SUM, AVG, MIN, MAX, COUNTA

	// Lookup field options
	LookupLinkedFieldID *string `json:"lookup_linked_field_id,omitempty"` // The linked_record field to lookup from
	LookupFieldID       *string `json:"lookup_field_id,omitempty"`        // Field in linked table to pull value from

	// Attachment field options
	AllowedTypes []string `json:"allowed_types,omitempty"` // ['image/*', 'application/pdf']
	MaxSizeBytes *int64   `json:"max_size_bytes,omitempty"` // Default 10MB
}

type Field struct {
	ID        uuid.UUID       `json:"id"`
	TableID   uuid.UUID       `json:"table_id"`
	Name      string          `json:"name"`
	FieldType FieldType       `json:"field_type"`
	Options   json.RawMessage `json:"options"`
	Position  int             `json:"position"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ValidFieldTypes returns all valid field types
func ValidFieldTypes() []FieldType {
	return []FieldType{
		FieldTypeText,
		FieldTypeNumber,
		FieldTypeCheckbox,
		FieldTypeDate,
		FieldTypeSingleSelect,
		FieldTypeMultiSelect,
		FieldTypeLinkedRecord,
		FieldTypeFormula,
		FieldTypeRollup,
		FieldTypeLookup,
		FieldTypeAttachment,
	}
}

// IsComputedField returns true if the field type is computed (formula, rollup, lookup)
func IsComputedField(ft FieldType) bool {
	return ft == FieldTypeFormula || ft == FieldTypeRollup || ft == FieldTypeLookup
}

// IsValidFieldType checks if a field type is valid
func IsValidFieldType(ft FieldType) bool {
	for _, valid := range ValidFieldTypes() {
		if ft == valid {
			return true
		}
	}
	return false
}
