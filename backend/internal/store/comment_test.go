package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestNewCommentStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewCommentStore(mock, nil, nil, nil)
	assert.NotNil(t, store)
}

func TestCommentStore_GetComment(t *testing.T) {
	ctx := context.Background()

	t.Run("returns comment when found and user has access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		commentID := uuid.New()
		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// Query for comment
		commentRows := pgxmock.NewRows([]string{
			"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
			"u_id", "email", "name", "u_created_at", "u_updated_at",
		}).AddRow(
			commentID, recordID, userID, "Test comment", nil, false, now, now,
			userID, "test@example.com", name, now, now,
		)

		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(commentID).
			WillReturnRows(commentRows)

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		comment, err := store.GetComment(ctx, commentID, userID)
		require.NoError(t, err)
		assert.Equal(t, commentID, comment.ID)
		assert.Equal(t, "Test comment", comment.Content)
		assert.NotNil(t, comment.User)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when comment not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		commentID := uuid.New()

		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(commentID).
			WillReturnError(pgx.ErrNoRows)

		comment, err := store.GetComment(ctx, commentID, uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, comment)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCommentStore_ListComments(t *testing.T) {
	ctx := context.Background()

	t.Run("returns comments for record", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		commentID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// List comments
		rows := pgxmock.NewRows([]string{
			"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
			"u_id", "email", "name", "u_created_at", "u_updated_at",
		}).AddRow(
			commentID, recordID, userID, "Test comment", nil, false, now, now,
			userID, "test@example.com", name, now, now,
		)

		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(recordID).
			WillReturnRows(rows)

		comments, err := store.ListComments(ctx, recordID, userID)
		require.NoError(t, err)
		assert.Len(t, comments, 1)
		assert.Equal(t, "Test comment", comments[0].Content)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when record not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		recordID := uuid.New()

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnError(pgx.ErrNoRows)

		comments, err := store.ListComments(ctx, recordID, uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, comments)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no comments", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
				"u_id", "email", "name", "u_created_at", "u_updated_at",
			}))

		comments, err := store.ListComments(ctx, recordID, userID)
		require.NoError(t, err)
		assert.NotNil(t, comments)
		assert.Empty(t, comments)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCommentStore_CreateComment(t *testing.T) {
	ctx := context.Background()

	t.Run("creates comment successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		commentID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// Insert comment
		mock.ExpectQuery("INSERT INTO comments").
			WithArgs(recordID, userID, "New comment", (*uuid.UUID)(nil)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
			}).AddRow(commentID, recordID, userID, "New comment", nil, false, now, now))

		// Fetch user info
		mock.ExpectQuery("SELECT id, email, name, created_at, updated_at FROM users").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
				AddRow(userID, "test@example.com", name, now, now))

		comment, err := store.CreateComment(ctx, recordID, userID, "New comment", nil)
		require.NoError(t, err)
		assert.Equal(t, commentID, comment.ID)
		assert.Equal(t, "New comment", comment.Content)
		assert.NotNil(t, comment.User)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("validates parent comment belongs to same record", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		parentID := uuid.New()
		differentRecordID := uuid.New()

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// Check parent comment - belongs to different record
		mock.ExpectQuery("SELECT record_id FROM comments WHERE id").
			WithArgs(&parentID).
			WillReturnRows(pgxmock.NewRows([]string{"record_id"}).AddRow(differentRecordID))

		comment, err := store.CreateComment(ctx, recordID, userID, "Reply", &parentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent comment belongs to a different record")
		assert.Nil(t, comment)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when parent comment not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		parentID := uuid.New()

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		mock.ExpectQuery("SELECT record_id FROM comments WHERE id").
			WithArgs(&parentID).
			WillReturnError(pgx.ErrNoRows)

		comment, err := store.CreateComment(ctx, recordID, userID, "Reply", &parentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parent comment not found")
		assert.Nil(t, comment)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCommentStore_UpdateComment(t *testing.T) {
	ctx := context.Background()

	t.Run("updates comment when user is author", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		commentID := uuid.New()
		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// GetComment - query for comment
		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(commentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
				"u_id", "email", "name", "u_created_at", "u_updated_at",
			}).AddRow(commentID, recordID, userID, "Old content", nil, false, now, now, userID, "test@example.com", name, now, now))

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// Update comment
		mock.ExpectQuery("UPDATE comments SET content").
			WithArgs(commentID, "Updated content").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
			}).AddRow(commentID, recordID, userID, "Updated content", nil, false, now, now))

		comment, err := store.UpdateComment(ctx, commentID, userID, "Updated content")
		require.NoError(t, err)
		assert.Equal(t, "Updated content", comment.Content)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when user is not author", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		commentID := uuid.New()
		recordID := uuid.New()
		authorID := uuid.New()
		otherUserID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Author")

		// GetComment
		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(commentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
				"u_id", "email", "name", "u_created_at", "u_updated_at",
			}).AddRow(commentID, recordID, authorID, "Content", nil, false, now, now, authorID, "author@example.com", name, now, now))

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, otherUserID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		comment, err := store.UpdateComment(ctx, commentID, otherUserID, "Try to update")
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, comment)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCommentStore_DeleteComment(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes comment when user is author", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		commentID := uuid.New()
		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// GetComment
		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(commentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
				"u_id", "email", "name", "u_created_at", "u_updated_at",
			}).AddRow(commentID, recordID, userID, "Content", nil, false, now, now, userID, "test@example.com", name, now, now))

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// Delete
		mock.ExpectExec("DELETE FROM comments WHERE id").
			WithArgs(commentID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteComment(ctx, commentID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("deletes comment when user is editor (not author)", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		commentID := uuid.New()
		recordID := uuid.New()
		authorID := uuid.New()
		editorID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Author")

		// GetComment
		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(commentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
				"u_id", "email", "name", "u_created_at", "u_updated_at",
			}).AddRow(commentID, recordID, authorID, "Content", nil, false, now, now, authorID, "author@example.com", name, now, now))

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, editorID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// Editor is not author, so get base ID again to check role
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// Check edit role
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, editorID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Delete
		mock.ExpectExec("DELETE FROM comments WHERE id").
			WithArgs(commentID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteComment(ctx, commentID, editorID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCommentStore_ResolveComment(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves comment", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewCommentStore(mock, baseStore, nil, nil)

		commentID := uuid.New()
		recordID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// GetComment
		mock.ExpectQuery("SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id").
			WithArgs(commentID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
				"u_id", "email", "name", "u_created_at", "u_updated_at",
			}).AddRow(commentID, recordID, userID, "Content", nil, false, now, now, userID, "test@example.com", name, now, now))

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// Update resolved status
		mock.ExpectQuery("UPDATE comments SET is_resolved").
			WithArgs(commentID, true).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "record_id", "user_id", "content", "parent_id", "is_resolved", "created_at", "updated_at",
			}).AddRow(commentID, recordID, userID, "Content", nil, true, now, now))

		comment, err := store.ResolveComment(ctx, commentID, userID, true)
		require.NoError(t, err)
		assert.True(t, comment.IsResolved)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCommentStore_GetCommentCount(t *testing.T) {
	ctx := context.Background()

	t.Run("returns comment count", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewCommentStore(mock, nil, nil, nil)
		recordID := uuid.New()

		mock.ExpectQuery("SELECT COUNT").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(5))

		count, err := store.GetCommentCount(ctx, recordID)
		require.NoError(t, err)
		assert.Equal(t, 5, count)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
