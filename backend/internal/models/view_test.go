package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidViewType(t *testing.T) {
	tests := []struct {
		name     string
		viewType ViewType
		expected bool
	}{
		{"valid grid", ViewTypeGrid, true},
		{"valid kanban", ViewTypeKanban, true},
		{"invalid empty", ViewType(""), false},
		{"invalid unknown", ViewType("unknown"), false},
		{"invalid list", ViewType("list"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidViewType(tt.viewType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestViewTypeConstants(t *testing.T) {
	assert.Equal(t, ViewType("grid"), ViewTypeGrid)
	assert.Equal(t, ViewType("kanban"), ViewTypeKanban)
	assert.Equal(t, ViewType("calendar"), ViewTypeCalendar)
	assert.Equal(t, ViewType("gallery"), ViewTypeGallery)
}

func TestIsValidViewType_AllTypes(t *testing.T) {
	// Test all valid types
	assert.True(t, IsValidViewType(ViewTypeGrid))
	assert.True(t, IsValidViewType(ViewTypeKanban))
	assert.True(t, IsValidViewType(ViewTypeCalendar))
	assert.True(t, IsValidViewType(ViewTypeGallery))
}
