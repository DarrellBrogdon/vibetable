package store

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/storage"
)

type AttachmentStore struct {
	db          DBTX
	baseStore   *BaseStore
	tableStore  *TableStore
	recordStore *RecordStore
	storage     storage.Storage
	baseURL     string
}

func NewAttachmentStore(db DBTX, baseStore *BaseStore, tableStore *TableStore, recordStore *RecordStore, stor storage.Storage, baseURL string) *AttachmentStore {
	return &AttachmentStore{
		db:          db,
		baseStore:   baseStore,
		tableStore:  tableStore,
		recordStore: recordStore,
		storage:     stor,
		baseURL:     baseURL,
	}
}

// getBaseIDForRecord returns the base ID for a record
func (s *AttachmentStore) getBaseIDForRecord(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	var baseID uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT t.base_id
		FROM records r
		JOIN tables t ON r.table_id = t.id
		WHERE r.id = $1
	`, recordID).Scan(&baseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return baseID, err
}

// CreateAttachment creates a new attachment for a record field
func (s *AttachmentStore) CreateAttachment(ctx context.Context, recordID, fieldID, userID uuid.UUID, filename, contentType string, sizeBytes int64, data io.Reader) (*models.Attachment, error) {
	// Verify user has edit access
	baseID, err := s.getBaseIDForRecord(ctx, recordID)
	if err != nil {
		return nil, err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	// Generate storage key
	storageKey := storage.GenerateKey(fieldID, recordID, filename)

	// Upload file to storage
	if err := s.storage.Upload(ctx, storageKey, data, contentType); err != nil {
		return nil, err
	}

	// Create database record
	var a models.Attachment
	err = s.db.QueryRow(ctx, `
		INSERT INTO attachments (record_id, field_id, filename, content_type, size_bytes, storage_key, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, record_id, field_id, filename, content_type, size_bytes, storage_key, thumbnail_key, width, height, created_by, created_at
	`, recordID, fieldID, filename, contentType, sizeBytes, storageKey, userID).Scan(
		&a.ID, &a.RecordID, &a.FieldID, &a.Filename, &a.ContentType, &a.SizeBytes,
		&a.StorageKey, &a.ThumbnailKey, &a.Width, &a.Height, &a.CreatedBy, &a.CreatedAt,
	)
	if err != nil {
		// Try to clean up uploaded file
		_ = s.storage.Delete(ctx, storageKey)
		return nil, err
	}

	// Add URL using the download endpoint
	a.URL = fmt.Sprintf("%s/api/v1/attachments/%s/download", s.baseURL, a.ID.String())

	return &a, nil
}

// GetAttachment returns an attachment by ID
func (s *AttachmentStore) GetAttachment(ctx context.Context, attachmentID, userID uuid.UUID) (*models.Attachment, error) {
	var a models.Attachment
	err := s.db.QueryRow(ctx, `
		SELECT id, record_id, field_id, filename, content_type, size_bytes, storage_key, thumbnail_key, width, height, created_by, created_at
		FROM attachments WHERE id = $1
	`, attachmentID).Scan(
		&a.ID, &a.RecordID, &a.FieldID, &a.Filename, &a.ContentType, &a.SizeBytes,
		&a.StorageKey, &a.ThumbnailKey, &a.Width, &a.Height, &a.CreatedBy, &a.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Verify user has access
	baseID, err := s.getBaseIDForRecord(ctx, a.RecordID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	// Add URL using the download endpoint
	a.URL = fmt.Sprintf("%s/api/v1/attachments/%s/download", s.baseURL, a.ID.String())
	// Note: thumbnails would need their own endpoint if implemented

	return &a, nil
}

// ListAttachmentsForField returns all attachments for a record's field
func (s *AttachmentStore) ListAttachmentsForField(ctx context.Context, recordID, fieldID, userID uuid.UUID) ([]*models.Attachment, error) {
	// Verify user has access
	baseID, err := s.getBaseIDForRecord(ctx, recordID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, record_id, field_id, filename, content_type, size_bytes, storage_key, thumbnail_key, width, height, created_by, created_at
		FROM attachments
		WHERE record_id = $1 AND field_id = $2
		ORDER BY created_at ASC
	`, recordID, fieldID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*models.Attachment
	for rows.Next() {
		var a models.Attachment
		if err := rows.Scan(
			&a.ID, &a.RecordID, &a.FieldID, &a.Filename, &a.ContentType, &a.SizeBytes,
			&a.StorageKey, &a.ThumbnailKey, &a.Width, &a.Height, &a.CreatedBy, &a.CreatedAt,
		); err != nil {
			return nil, err
		}

		// Add URL using the download endpoint
		a.URL = fmt.Sprintf("%s/api/v1/attachments/%s/download", s.baseURL, a.ID.String())

		attachments = append(attachments, &a)
	}

	if attachments == nil {
		attachments = []*models.Attachment{}
	}

	return attachments, rows.Err()
}

// DeleteAttachment deletes an attachment
func (s *AttachmentStore) DeleteAttachment(ctx context.Context, attachmentID, userID uuid.UUID) error {
	// Get attachment to check access and get storage key
	a, err := s.GetAttachment(ctx, attachmentID, userID)
	if err != nil {
		return err
	}

	// Verify user has edit access
	baseID, err := s.getBaseIDForRecord(ctx, a.RecordID)
	if err != nil {
		return err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return err
	}
	if !role.CanEdit() {
		return ErrForbidden
	}

	// Delete from database
	result, err := s.db.Exec(ctx, `DELETE FROM attachments WHERE id = $1`, attachmentID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Delete from storage (best effort - don't fail if storage delete fails)
	_ = s.storage.Delete(ctx, a.StorageKey)
	if a.ThumbnailKey != nil {
		_ = s.storage.Delete(ctx, *a.ThumbnailKey)
	}

	return nil
}

// DownloadAttachment returns a reader for the attachment file
func (s *AttachmentStore) DownloadAttachment(ctx context.Context, attachmentID, userID uuid.UUID) (io.ReadCloser, *models.Attachment, error) {
	a, err := s.GetAttachment(ctx, attachmentID, userID)
	if err != nil {
		return nil, nil, err
	}

	reader, err := s.storage.Download(ctx, a.StorageKey)
	if err != nil {
		return nil, nil, err
	}

	return reader, a, nil
}
