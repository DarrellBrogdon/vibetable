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
	"github.com/vibetable/backend/internal/realtime"
)

func TestNewTableStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewTableStore(mock, baseStore)
	assert.NotNil(t, store)
}

func TestTableStore_ListTablesForBase(t *testing.T) {
	ctx := context.Background()

	t.Run("returns tables for base", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock list tables
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(baseID).
			WillReturnRows(tableRows)

		tables, err := store.ListTablesForBase(ctx, baseID, userID)
		require.NoError(t, err)
		assert.Len(t, tables, 1)
		assert.Equal(t, "Test Table", tables[0].Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no tables", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock empty result
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(baseID).
			WillReturnRows(tableRows)

		tables, err := store.ListTablesForBase(ctx, baseID, userID)
		require.NoError(t, err)
		assert.NotNil(t, tables)
		assert.Empty(t, tables)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when no access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()

		// Mock GetUserRole returns empty (no access)
		roleRows := pgxmock.NewRows([]string{"role"})
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		tables, err := store.ListTablesForBase(ctx, baseID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, tables)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTableStore_CreateTable(t *testing.T) {
	ctx := context.Background()

	t.Run("creates table successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock max position
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(-1)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(baseID).
			WillReturnRows(posRows)

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "New Table", 0, now, now)
		mock.ExpectQuery("INSERT INTO tables").
			WithArgs(baseID, "New Table", 0).
			WillReturnRows(insertRows)

		table, err := store.CreateTable(ctx, baseID, "New Table", userID)
		require.NoError(t, err)
		assert.Equal(t, tableID, table.ID)
		assert.Equal(t, "New Table", table.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()

		// Mock GetUserRole returns viewer
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		table, err := store.CreateTable(ctx, baseID, "New Table", userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, table)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTableStore_GetTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns table when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()

		// Mock get table
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		table, err := store.GetTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.Equal(t, tableID, table.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when table doesn't exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		tableID := uuid.New()

		// Mock get table returns no rows
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnError(pgx.ErrNoRows)

		table, err := store.GetTable(ctx, tableID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, table)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTableStore_UpdateTable(t *testing.T) {
	ctx := context.Background()

	t.Run("updates table successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()

		// Mock GetTable
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Old Name", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole for GetTable
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetUserRole for update permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock update
		updateRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "New Name", 0, now, now)
		mock.ExpectQuery("UPDATE tables SET name").
			WithArgs(tableID, "New Name").
			WillReturnRows(updateRows)

		table, err := store.UpdateTable(ctx, tableID, "New Name", userID)
		require.NoError(t, err)
		assert.Equal(t, "New Name", table.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to update", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()

		// Mock GetTable
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Old Name", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole for GetTable (viewer can view)
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetUserRole for update permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		table, err := store.UpdateTable(ctx, tableID, "New Name", userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, table)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTableStore_DeleteTable(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes table successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()

		// Mock GetTable
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole for GetTable
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetUserRole for delete permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock delete
		mock.ExpectExec("DELETE FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteTable(ctx, tableID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()

		// Mock GetTable
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole for GetTable
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetUserRole for delete permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		err = store.DeleteTable(ctx, tableID, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTableStore_ReorderTables(t *testing.T) {
	ctx := context.Background()

	t.Run("reorders tables successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID1 := uuid.New()
		tableID2 := uuid.New()

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock transaction
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE tables SET position").
			WithArgs(0, tableID1, baseID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectExec("UPDATE tables SET position").
			WithArgs(1, tableID2, baseID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectCommit()

		err = store.ReorderTables(ctx, baseID, []uuid.UUID{tableID1, tableID2}, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to reorder", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewTableStore(mock, baseStore)
		userID := uuid.New()
		baseID := uuid.New()

		// Mock GetUserRole returns viewer
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		err = store.ReorderTables(ctx, baseID, []uuid.UUID{uuid.New()}, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestTableStore_SetHub(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewTableStore(mock, baseStore)

	// Initially hub is nil
	assert.Nil(t, store.hub)

	hub := &realtime.Hub{}
	store.SetHub(hub)

	assert.Equal(t, hub, store.hub)
}

