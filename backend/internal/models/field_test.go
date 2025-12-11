package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidFieldTypes(t *testing.T) {
	types := ValidFieldTypes()

	assert.Len(t, types, 11)
	assert.Contains(t, types, FieldTypeText)
	assert.Contains(t, types, FieldTypeNumber)
	assert.Contains(t, types, FieldTypeCheckbox)
	assert.Contains(t, types, FieldTypeDate)
	assert.Contains(t, types, FieldTypeSingleSelect)
	assert.Contains(t, types, FieldTypeMultiSelect)
	assert.Contains(t, types, FieldTypeLinkedRecord)
	assert.Contains(t, types, FieldTypeFormula)
	assert.Contains(t, types, FieldTypeRollup)
	assert.Contains(t, types, FieldTypeLookup)
	assert.Contains(t, types, FieldTypeAttachment)
}

func TestIsValidFieldType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  bool
	}{
		{"valid text", FieldTypeText, true},
		{"valid number", FieldTypeNumber, true},
		{"valid checkbox", FieldTypeCheckbox, true},
		{"valid date", FieldTypeDate, true},
		{"valid single_select", FieldTypeSingleSelect, true},
		{"valid multi_select", FieldTypeMultiSelect, true},
		{"valid linked_record", FieldTypeLinkedRecord, true},
		{"valid formula", FieldTypeFormula, true},
		{"valid rollup", FieldTypeRollup, true},
		{"valid lookup", FieldTypeLookup, true},
		{"valid attachment", FieldTypeAttachment, true},
		{"invalid empty", FieldType(""), false},
		{"invalid unknown", FieldType("unknown"), false},
		{"invalid boolean", FieldType("boolean"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidFieldType(tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFieldTypeConstants(t *testing.T) {
	assert.Equal(t, FieldType("text"), FieldTypeText)
	assert.Equal(t, FieldType("number"), FieldTypeNumber)
	assert.Equal(t, FieldType("checkbox"), FieldTypeCheckbox)
	assert.Equal(t, FieldType("date"), FieldTypeDate)
	assert.Equal(t, FieldType("single_select"), FieldTypeSingleSelect)
	assert.Equal(t, FieldType("multi_select"), FieldTypeMultiSelect)
	assert.Equal(t, FieldType("linked_record"), FieldTypeLinkedRecord)
	assert.Equal(t, FieldType("formula"), FieldTypeFormula)
	assert.Equal(t, FieldType("rollup"), FieldTypeRollup)
	assert.Equal(t, FieldType("lookup"), FieldTypeLookup)
	assert.Equal(t, FieldType("attachment"), FieldTypeAttachment)
}

func TestIsComputedField(t *testing.T) {
	tests := []struct {
		name      string
		fieldType FieldType
		expected  bool
	}{
		{"formula is computed", FieldTypeFormula, true},
		{"rollup is computed", FieldTypeRollup, true},
		{"lookup is computed", FieldTypeLookup, true},
		{"text is not computed", FieldTypeText, false},
		{"number is not computed", FieldTypeNumber, false},
		{"checkbox is not computed", FieldTypeCheckbox, false},
		{"date is not computed", FieldTypeDate, false},
		{"single_select is not computed", FieldTypeSingleSelect, false},
		{"multi_select is not computed", FieldTypeMultiSelect, false},
		{"linked_record is not computed", FieldTypeLinkedRecord, false},
		{"attachment is not computed", FieldTypeAttachment, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsComputedField(tt.fieldType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
