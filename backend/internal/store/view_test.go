package store

import (
	"context"
	"encoding/json"
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

func TestNewViewStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	tableStore := NewTableStore(mock, baseStore)
	store := NewViewStore(mock, baseStore, tableStore)
	assert.NotNil(t, store)
}

func TestViewStore_ListViewsForTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns views for table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		config := json.RawMessage(`{}`)

		// Mock GetTable (which checks access)
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole for GetTable
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock list views
		viewRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "Grid View", models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(viewRows)

		views, err := store.ListViewsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.Len(t, views, 1)
		assert.Equal(t, "Grid View", views[0].Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no views", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
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

		// Mock GetUserRole
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock empty result
		viewRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(viewRows)

		views, err := store.ListViewsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.NotNil(t, views)
		assert.Empty(t, views)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when table not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		tableID := uuid.New()

		// Mock GetTable returns not found
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnError(pgx.ErrNoRows)

		views, err := store.ListViewsForTable(ctx, tableID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, views)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestViewStore_CreateView(t *testing.T) {
	ctx := context.Background()

	t.Run("creates view successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		config := json.RawMessage(`{"columns": []}`)

		// Mock GetTable
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(tableRows)

		// Mock GetUserRole for GetTable
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(roleRows1)

		// Mock GetUserRole for create permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(roleRows2)

		// Mock max position
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(-1)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(pgxmock.AnyArg()).
			WillReturnRows(posRows)

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "My View", models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("INSERT INTO views").
			WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(insertRows)

		view, err := store.CreateView(ctx, tableID, "My View", models.ViewTypeGrid, config, userID)
		require.NoError(t, err)
		assert.Equal(t, viewID, view.ID)
		assert.Equal(t, "My View", view.Name)
		assert.Equal(t, models.ViewTypeGrid, view.Type)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creates view with nil config defaults to empty object", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		defaultConfig := json.RawMessage("{}")

		// Mock GetTable
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole for GetTable
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetUserRole for create permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock max position
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(1)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(tableID).
			WillReturnRows(posRows)

		// Mock insert with default config
		insertRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "Kanban", models.ViewTypeKanban, defaultConfig, 2, nil, false, now, now)
		mock.ExpectQuery("INSERT INTO views").
			WithArgs(tableID, "Kanban", models.ViewTypeKanban, defaultConfig, 2).
			WillReturnRows(insertRows)

		view, err := store.CreateView(ctx, tableID, "Kanban", models.ViewTypeKanban, nil, userID)
		require.NoError(t, err)
		assert.Equal(t, viewID, view.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
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

		// Mock GetUserRole for create permission - returns viewer
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		view, err := store.CreateView(ctx, tableID, "View", models.ViewTypeGrid, nil, userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, view)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestViewStore_GetView(t *testing.T) {
	ctx := context.Background()

	t.Run("returns view when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		config := json.RawMessage(`{}`)

		// Mock get view
		viewRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "Grid View", models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(viewID).
			WillReturnRows(viewRows)

		// Mock GetTable (access check)
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Mock GetUserRole
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		view, err := store.GetView(ctx, viewID, userID)
		require.NoError(t, err)
		assert.Equal(t, viewID, view.ID)
		assert.Equal(t, "Grid View", view.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when view doesn't exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		viewID := uuid.New()

		// Mock get view returns no rows
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(viewID).
			WillReturnError(pgx.ErrNoRows)

		view, err := store.GetView(ctx, viewID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, view)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestViewStore_UpdateView(t *testing.T) {
	ctx := context.Background()

	t.Run("updates view name successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		config := json.RawMessage(`{}`)
		newName := "Updated View"

		// Mock GetView
		viewRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "Old View", models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(viewID).
			WillReturnRows(viewRows)

		// Mock GetTable for GetView
		tableRows1 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows1)

		// Mock GetUserRole for GetView
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetTable for update permission check
		tableRows2 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows2)

		// Mock GetUserRole for GetTable in update
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock GetUserRole for update permission
		roleRows3 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows3)

		// Mock update
		updateRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, newName, models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("UPDATE views").
			WithArgs(newName, config, viewID).
			WillReturnRows(updateRows)

		view, err := store.UpdateView(ctx, viewID, &newName, nil, userID)
		require.NoError(t, err)
		assert.Equal(t, newName, view.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to update", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		config := json.RawMessage(`{}`)
		newName := "Updated View"

		// Mock GetView
		viewRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "Old View", models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(viewID).
			WillReturnRows(viewRows)

		// Mock GetTable for GetView
		tableRows1 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows1)

		// Mock GetUserRole for GetView (viewer can view)
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetTable for update permission check
		tableRows2 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows2)

		// Mock GetUserRole for GetTable in update
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock GetUserRole for update permission - viewer
		roleRows3 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows3)

		view, err := store.UpdateView(ctx, viewID, &newName, nil, userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, view)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestViewStore_DeleteView(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes view successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		config := json.RawMessage(`{}`)

		// Mock GetView
		viewRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "Grid View", models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(viewID).
			WillReturnRows(viewRows)

		// Mock GetTable for GetView
		tableRows1 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows1)

		// Mock GetUserRole for GetView
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetTable for delete permission check
		tableRows2 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows2)

		// Mock GetUserRole for GetTable in delete
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock GetUserRole for delete permission
		roleRows3 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows3)

		// Mock delete
		mock.ExpectExec("DELETE FROM views WHERE id").
			WithArgs(viewID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteView(ctx, viewID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewViewStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		now := time.Now().UTC()
		config := json.RawMessage(`{}`)

		// Mock GetView
		viewRows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "public_token", "is_public", "created_at", "updated_at"}).
			AddRow(viewID, tableID, "Grid View", models.ViewTypeGrid, config, 0, nil, false, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at").
			WithArgs(viewID).
			WillReturnRows(viewRows)

		// Mock GetTable for GetView
		tableRows1 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows1)

		// Mock GetUserRole for GetView (viewer can view)
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock GetTable for delete permission check
		tableRows2 := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(tableRows2)

		// Mock GetUserRole for GetTable in delete
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock GetUserRole for delete permission - viewer
		roleRows3 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows3)

		err = store.DeleteView(ctx, viewID, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestViewStore_SetHub(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	tableStore := NewTableStore(mock, baseStore)
	store := NewViewStore(mock, baseStore, tableStore)

	// Initially hub is nil
	assert.Nil(t, store.hub)

	hub := &realtime.Hub{}
	store.SetHub(hub)

	assert.Equal(t, hub, store.hub)
}

