package store

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

// mockStorage implements storage.Storage for testing
type mockStorage struct {
	uploadErr   error
	downloadErr error
	deleteErr   error
	existsErr   error
	data        []byte
	exists      bool
}

func (m *mockStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	if m.uploadErr != nil {
		return m.uploadErr
	}
	// Read the data
	m.data, _ = io.ReadAll(data)
	return nil
}

func (m *mockStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if m.downloadErr != nil {
		return nil, m.downloadErr
	}
	return io.NopCloser(bytes.NewReader(m.data)), nil
}

func (m *mockStorage) Delete(ctx context.Context, key string) error {
	return m.deleteErr
}

func (m *mockStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return "http://localhost/files/" + key, nil
}

func (m *mockStorage) Exists(ctx context.Context, key string) (bool, error) {
	if m.existsErr != nil {
		return false, m.existsErr
	}
	return m.exists, nil
}

func TestNewAttachmentStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewAttachmentStore(mock, nil, nil, nil, nil, "http://localhost")
	assert.NotNil(t, store)
}

func TestAttachmentStore_GetAttachment(t *testing.T) {
	ctx := context.Background()

	t.Run("returns attachment when found and user has access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAttachmentStore(mock, baseStore, nil, nil, nil, "http://localhost:8080")

		attachmentID := uuid.New()
		recordID := uuid.New()
		fieldID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		// Query for attachment
		mock.ExpectQuery("SELECT id, record_id, field_id, filename, content_type, size_bytes").
			WithArgs(attachmentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "field_id", "filename", "content_type", "size_bytes",
				"storage_key", "thumbnail_key", "width", "height", "created_by", "created_at",
			}).AddRow(
				attachmentID, recordID, fieldID, "test.txt", "text/plain", int64(100),
				"storage/key", nil, nil, nil, userID, now,
			))

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		attachment, err := store.GetAttachment(ctx, attachmentID, userID)
		require.NoError(t, err)
		assert.Equal(t, attachmentID, attachment.ID)
		assert.Equal(t, "test.txt", attachment.Filename)
		assert.Contains(t, attachment.URL, attachmentID.String())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when attachment not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAttachmentStore(mock, nil, nil, nil, nil, "http://localhost")
		attachmentID := uuid.New()

		mock.ExpectQuery("SELECT id, record_id, field_id, filename, content_type, size_bytes").
			WithArgs(attachmentID).
			WillReturnError(pgx.ErrNoRows)

		attachment, err := store.GetAttachment(ctx, attachmentID, uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, attachment)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAttachmentStore_ListAttachmentsForField(t *testing.T) {
	ctx := context.Background()

	t.Run("returns attachments for field", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAttachmentStore(mock, baseStore, nil, nil, nil, "http://localhost:8080")

		recordID := uuid.New()
		fieldID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		attachmentID := uuid.New()
		now := time.Now().UTC()

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// List attachments
		mock.ExpectQuery("SELECT id, record_id, field_id, filename, content_type, size_bytes").
			WithArgs(recordID, fieldID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "field_id", "filename", "content_type", "size_bytes",
				"storage_key", "thumbnail_key", "width", "height", "created_by", "created_at",
			}).AddRow(
				attachmentID, recordID, fieldID, "test.txt", "text/plain", int64(100),
				"storage/key", nil, nil, nil, userID, now,
			))

		attachments, err := store.ListAttachmentsForField(ctx, recordID, fieldID, userID)
		require.NoError(t, err)
		assert.Len(t, attachments, 1)
		assert.Equal(t, "test.txt", attachments[0].Filename)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no attachments", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAttachmentStore(mock, baseStore, nil, nil, nil, "http://localhost")

		recordID := uuid.New()
		fieldID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		mock.ExpectQuery("SELECT id, record_id, field_id, filename, content_type, size_bytes").
			WithArgs(recordID, fieldID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "field_id", "filename", "content_type", "size_bytes",
				"storage_key", "thumbnail_key", "width", "height", "created_by", "created_at",
			}))

		attachments, err := store.ListAttachmentsForField(ctx, recordID, fieldID, userID)
		require.NoError(t, err)
		assert.NotNil(t, attachments)
		assert.Empty(t, attachments)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when record not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAttachmentStore(mock, nil, nil, nil, nil, "http://localhost")
		recordID := uuid.New()

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnError(pgx.ErrNoRows)

		attachments, err := store.ListAttachmentsForField(ctx, recordID, uuid.New(), uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, attachments)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAttachmentStore_CreateAttachment(t *testing.T) {
	ctx := context.Background()

	t.Run("creates attachment successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		mockStor := &mockStorage{}
		store := NewAttachmentStore(mock, baseStore, nil, nil, mockStor, "http://localhost:8080")

		recordID := uuid.New()
		fieldID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		attachmentID := uuid.New()
		now := time.Now().UTC()

		data := bytes.NewReader([]byte("test content"))

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Insert attachment
		mock.ExpectQuery("INSERT INTO attachments").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "field_id", "filename", "content_type", "size_bytes",
				"storage_key", "thumbnail_key", "width", "height", "created_by", "created_at",
			}).AddRow(
				attachmentID, recordID, fieldID, "test.txt", "text/plain", int64(12),
				"storage/key", nil, nil, nil, userID, now,
			))

		attachment, err := store.CreateAttachment(ctx, recordID, fieldID, userID, "test.txt", "text/plain", 12, data)
		require.NoError(t, err)
		assert.Equal(t, attachmentID, attachment.ID)
		assert.Equal(t, "test.txt", attachment.Filename)
		assert.Contains(t, attachment.URL, attachmentID.String())

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when user has no edit access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAttachmentStore(mock, baseStore, nil, nil, nil, "http://localhost")

		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		attachment, err := store.CreateAttachment(ctx, recordID, uuid.New(), userID, "test.txt", "text/plain", 12, nil)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, attachment)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAttachmentStore_DeleteAttachment(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes attachment when user has edit access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		mockStor := &mockStorage{}
		store := NewAttachmentStore(mock, baseStore, nil, nil, mockStor, "http://localhost")

		attachmentID := uuid.New()
		recordID := uuid.New()
		fieldID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		// GetAttachment - query for attachment
		mock.ExpectQuery("SELECT id, record_id, field_id, filename, content_type, size_bytes").
			WithArgs(attachmentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "field_id", "filename", "content_type", "size_bytes",
				"storage_key", "thumbnail_key", "width", "height", "created_by", "created_at",
			}).AddRow(
				attachmentID, recordID, fieldID, "test.txt", "text/plain", int64(100),
				"storage/key", nil, nil, nil, userID, now,
			))

		// Get base ID for record (in GetAttachment)
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check (in GetAttachment)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Get base ID for record again (in DeleteAttachment)
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check (in DeleteAttachment)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Delete
		mock.ExpectExec("DELETE FROM attachments WHERE id").
			WithArgs(attachmentID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteAttachment(ctx, attachmentID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAttachmentStore(mock, baseStore, nil, nil, nil, "http://localhost")

		attachmentID := uuid.New()
		recordID := uuid.New()
		fieldID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		mock.ExpectQuery("SELECT id, record_id, field_id, filename, content_type, size_bytes").
			WithArgs(attachmentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "field_id", "filename", "content_type", "size_bytes",
				"storage_key", "thumbnail_key", "width", "height", "created_by", "created_at",
			}).AddRow(
				attachmentID, recordID, fieldID, "test.txt", "text/plain", int64(100),
				"storage/key", nil, nil, nil, userID, now,
			))

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		err = store.DeleteAttachment(ctx, attachmentID, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAttachmentStore_DownloadAttachment(t *testing.T) {
	ctx := context.Background()

	t.Run("downloads attachment successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		mockStor := &mockStorage{data: []byte("file content")}
		store := NewAttachmentStore(mock, baseStore, nil, nil, mockStor, "http://localhost")

		attachmentID := uuid.New()
		recordID := uuid.New()
		fieldID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		// GetAttachment
		mock.ExpectQuery("SELECT id, record_id, field_id, filename, content_type, size_bytes").
			WithArgs(attachmentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "field_id", "filename", "content_type", "size_bytes",
				"storage_key", "thumbnail_key", "width", "height", "created_by", "created_at",
			}).AddRow(
				attachmentID, recordID, fieldID, "test.txt", "text/plain", int64(12),
				"storage/key", nil, nil, nil, userID, now,
			))

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		reader, attachment, err := store.DownloadAttachment(ctx, attachmentID, userID)
		require.NoError(t, err)
		assert.NotNil(t, reader)
		assert.NotNil(t, attachment)
		assert.Equal(t, "test.txt", attachment.Filename)

		// Read the content
		content, _ := io.ReadAll(reader)
		assert.Equal(t, []byte("file content"), content)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
