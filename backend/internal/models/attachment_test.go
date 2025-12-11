package models

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAttachment_IsImage(t *testing.T) {
	t.Run("returns true for image content types", func(t *testing.T) {
		testCases := []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"image/svg+xml",
		}

		for _, contentType := range testCases {
			a := &Attachment{ContentType: contentType}
			assert.True(t, a.IsImage(), "expected %s to be an image", contentType)
		}
	})

	t.Run("returns false for non-image content types", func(t *testing.T) {
		testCases := []string{
			"application/pdf",
			"text/plain",
			"video/mp4",
			"audio/mpeg",
			"application/json",
			"",
		}

		for _, contentType := range testCases {
			a := &Attachment{ContentType: contentType}
			assert.False(t, a.IsImage(), "expected %s to not be an image", contentType)
		}
	})
}

func TestAttachment_ToSummary(t *testing.T) {
	id := uuid.New()
	width := 800
	height := 600
	thumbnailURL := "https://example.com/thumb.jpg"

	attachment := &Attachment{
		ID:           id,
		Filename:     "test.jpg",
		ContentType:  "image/jpeg",
		SizeBytes:    1024,
		Width:        &width,
		Height:       &height,
		URL:          "https://example.com/test.jpg",
		ThumbnailURL: &thumbnailURL,
	}

	summary := attachment.ToSummary()

	assert.Equal(t, id, summary.ID)
	assert.Equal(t, "test.jpg", summary.Filename)
	assert.Equal(t, "image/jpeg", summary.ContentType)
	assert.Equal(t, int64(1024), summary.SizeBytes)
	assert.Equal(t, &width, summary.Width)
	assert.Equal(t, &height, summary.Height)
	assert.Equal(t, "https://example.com/test.jpg", summary.URL)
	assert.Equal(t, &thumbnailURL, summary.ThumbnailURL)
}

func TestAttachment_ToSummary_NilFields(t *testing.T) {
	attachment := &Attachment{
		ID:          uuid.New(),
		Filename:    "doc.pdf",
		ContentType: "application/pdf",
		SizeBytes:   2048,
		URL:         "https://example.com/doc.pdf",
		// Width, Height, ThumbnailURL are nil
	}

	summary := attachment.ToSummary()

	assert.Equal(t, attachment.ID, summary.ID)
	assert.Nil(t, summary.Width)
	assert.Nil(t, summary.Height)
	assert.Nil(t, summary.ThumbnailURL)
}
