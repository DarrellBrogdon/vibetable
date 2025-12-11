package models

import (
	"time"

	"github.com/google/uuid"
)

// Attachment represents a file attached to a record field
type Attachment struct {
	ID           uuid.UUID  `json:"id"`
	RecordID     uuid.UUID  `json:"record_id"`
	FieldID      uuid.UUID  `json:"field_id"`
	Filename     string     `json:"filename"`
	ContentType  string     `json:"content_type"`
	SizeBytes    int64      `json:"size_bytes"`
	StorageKey   string     `json:"-"` // Not exposed to API
	ThumbnailKey *string    `json:"-"` // Not exposed to API
	Width        *int       `json:"width,omitempty"`
	Height       *int       `json:"height,omitempty"`
	CreatedBy    uuid.UUID  `json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`

	// Computed fields for API response
	URL          string     `json:"url,omitempty"`
	ThumbnailURL *string    `json:"thumbnail_url,omitempty"`
}

// AttachmentSummary is a lightweight version for embedding in records
type AttachmentSummary struct {
	ID           uuid.UUID `json:"id"`
	Filename     string    `json:"filename"`
	ContentType  string    `json:"content_type"`
	SizeBytes    int64     `json:"size_bytes"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	URL          string    `json:"url"`
	ThumbnailURL *string   `json:"thumbnail_url,omitempty"`
}

// IsImage returns true if the attachment is an image
func (a *Attachment) IsImage() bool {
	switch a.ContentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml":
		return true
	default:
		return false
	}
}

// ToSummary converts an Attachment to AttachmentSummary
func (a *Attachment) ToSummary() AttachmentSummary {
	return AttachmentSummary{
		ID:           a.ID,
		Filename:     a.Filename,
		ContentType:  a.ContentType,
		SizeBytes:    a.SizeBytes,
		Width:        a.Width,
		Height:       a.Height,
		URL:          a.URL,
		ThumbnailURL: a.ThumbnailURL,
	}
}
