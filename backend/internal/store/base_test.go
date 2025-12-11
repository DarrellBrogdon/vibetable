package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestNewBaseStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewBaseStore(mock)
	assert.NotNil(t, store)
}

func TestBaseStore_ListBasesForUser(t *testing.T) {
	ctx := context.Background()

	t.Run("returns bases for user", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "name", "created_by", "created_at", "updated_at", "role"}).
			AddRow(baseID, "Test Base", userID, now, now, models.RoleOwner)

		mock.ExpectQuery("SELECT b.id, b.name, b.created_by, b.created_at, b.updated_at, bc.role").
			WithArgs(userID).
			WillReturnRows(rows)

		bases, err := store.ListBasesForUser(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, bases, 1)
		assert.Equal(t, "Test Base", bases[0].Name)
		assert.Equal(t, models.RoleOwner, *bases[0].Role)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no bases", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()

		rows := pgxmock.NewRows([]string{"id", "name", "created_by", "created_at", "updated_at", "role"})
		mock.ExpectQuery("SELECT b.id, b.name, b.created_by, b.created_at, b.updated_at, bc.role").
			WithArgs(userID).
			WillReturnRows(rows)

		bases, err := store.ListBasesForUser(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, bases)
		assert.Empty(t, bases)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBaseStore_CreateBase(t *testing.T) {
	ctx := context.Background()

	t.Run("creates base successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		mock.ExpectBegin()

		baseRows := pgxmock.NewRows([]string{"id", "name", "created_by", "created_at", "updated_at"}).
			AddRow(baseID, "New Base", userID, now, now)
		mock.ExpectQuery("INSERT INTO bases").
			WithArgs("New Base", userID).
			WillReturnRows(baseRows)

		mock.ExpectExec("INSERT INTO base_collaborators").
			WithArgs(baseID, userID).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		mock.ExpectCommit()

		base, err := store.CreateBase(ctx, "New Base", userID)
		require.NoError(t, err)
		assert.Equal(t, baseID, base.ID)
		assert.Equal(t, "New Base", base.Name)
		assert.Equal(t, models.RoleOwner, *base.Role)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBaseStore_GetBase(t *testing.T) {
	ctx := context.Background()

	t.Run("returns base when user has access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "name", "created_by", "created_at", "updated_at", "role"}).
			AddRow(baseID, "Test Base", userID, now, now, models.RoleOwner)

		mock.ExpectQuery("SELECT b.id, b.name, b.created_by, b.created_at, b.updated_at, bc.role").
			WithArgs(baseID, userID).
			WillReturnRows(rows)

		base, err := store.GetBase(ctx, baseID, userID)
		require.NoError(t, err)
		assert.Equal(t, baseID, base.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when user has no access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()

		rows := pgxmock.NewRows([]string{"id", "name", "created_by", "created_at", "updated_at", "role"})
		mock.ExpectQuery("SELECT b.id, b.name, b.created_by, b.created_at, b.updated_at, bc.role").
			WithArgs(baseID, userID).
			WillReturnRows(rows)

		base, err := store.GetBase(ctx, baseID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, base)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBaseStore_GetUserRole(t *testing.T) {
	ctx := context.Background()

	t.Run("returns role when user is collaborator", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()

		rows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(rows)

		role, err := store.GetUserRole(ctx, baseID, userID)
		require.NoError(t, err)
		assert.Equal(t, models.RoleEditor, role)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when not collaborator", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()

		rows := pgxmock.NewRows([]string{"role"})
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(rows)

		role, err := store.GetUserRole(ctx, baseID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Empty(t, role)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBaseStore_UpdateBase(t *testing.T) {
	ctx := context.Background()

	t.Run("updates base when user can edit", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()
		now := time.Now().UTC()

		// GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Update query
		updateRows := pgxmock.NewRows([]string{"id", "name", "created_by", "created_at", "updated_at"}).
			AddRow(baseID, "Updated Name", userID, now, now)
		mock.ExpectQuery("UPDATE bases SET name").
			WithArgs(baseID, "Updated Name").
			WillReturnRows(updateRows)

		base, err := store.UpdateBase(ctx, baseID, "Updated Name", userID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", base.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to update", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()

		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		base, err := store.UpdateBase(ctx, baseID, "New Name", userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, base)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBaseStore_DeleteBase(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes base when user is owner", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()

		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		mock.ExpectExec("DELETE FROM bases WHERE id").
			WithArgs(baseID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteBase(ctx, baseID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when non-owner tries to delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()

		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		err = store.DeleteBase(ctx, baseID, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestBaseStore_ListCollaborators(t *testing.T) {
	ctx := context.Background()

	t.Run("returns collaborators for base", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewBaseStore(mock)
		userID := uuid.New()
		baseID := uuid.New()
		collabID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// List collaborators
		collabRows := pgxmock.NewRows([]string{"id", "base_id", "user_id", "role", "created_at", "u_id", "email", "name", "u_created_at", "u_updated_at"}).
			AddRow(collabID, baseID, userID, models.RoleOwner, now, userID, "test@example.com", name, now, now)
		mock.ExpectQuery("SELECT bc.id, bc.base_id, bc.user_id, bc.role, bc.created_at").
			WithArgs(baseID).
			WillReturnRows(collabRows)

		collaborators, err := store.ListCollaborators(ctx, baseID, userID)
		require.NoError(t, err)
		assert.Len(t, collaborators, 1)
		assert.Equal(t, models.RoleOwner, collaborators[0].Role)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestErrForbidden(t *testing.T) {
	assert.NotNil(t, ErrForbidden)
	assert.Equal(t, "forbidden", ErrForbidden.Error())
}
