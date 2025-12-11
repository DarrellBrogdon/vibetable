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

func TestNewFieldStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	tableStore := NewTableStore(mock, baseStore)
	store := NewFieldStore(mock, baseStore, tableStore)
	assert.NotNil(t, store)
}

func TestFieldStore_getBaseIDForTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns base ID when table exists", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		tableID := uuid.New()
		baseID := uuid.New()

		rows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(rows)

		result, err := store.getBaseIDForTable(ctx, tableID)
		require.NoError(t, err)
		assert.Equal(t, baseID, result)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when table doesn't exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		tableID := uuid.New()

		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnError(pgx.ErrNoRows)

		result, err := store.getBaseIDForTable(ctx, tableID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Equal(t, uuid.Nil, result)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFieldStore_ListFieldsForTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns fields for table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID := uuid.New()
		now := time.Now().UTC()
		options := json.RawMessage(`{}`)

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock list fields
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, "Name", models.FieldTypeText, options, 0, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(fieldRows)

		fields, err := store.ListFieldsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.Len(t, fields, 1)
		assert.Equal(t, "Name", fields[0].Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no fields", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock empty result
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(fieldRows)

		fields, err := store.ListFieldsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.NotNil(t, fields)
		assert.Empty(t, fields)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFieldStore_CreateField(t *testing.T) {
	ctx := context.Background()

	t.Run("creates field successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID := uuid.New()
		now := time.Now().UTC()
		options := json.RawMessage(`{"precision": 2}`)

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock max position
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(-1)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(tableID).
			WillReturnRows(posRows)

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, "Amount", models.FieldTypeNumber, options, 0, now, now)
		mock.ExpectQuery("INSERT INTO fields").
			WithArgs(tableID, "Amount", models.FieldTypeNumber, options, 0).
			WillReturnRows(insertRows)

		field, err := store.CreateField(ctx, tableID, "Amount", models.FieldTypeNumber, options, userID)
		require.NoError(t, err)
		assert.Equal(t, fieldID, field.ID)
		assert.Equal(t, "Amount", field.Name)
		assert.Equal(t, models.FieldTypeNumber, field.FieldType)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creates field with nil options defaults to empty object", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID := uuid.New()
		now := time.Now().UTC()
		defaultOptions := json.RawMessage(`{}`)

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock max position
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(2)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(tableID).
			WillReturnRows(posRows)

		// Mock insert - note nil options becomes {}
		insertRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, "Notes", models.FieldTypeText, defaultOptions, 3, now, now)
		mock.ExpectQuery("INSERT INTO fields").
			WithArgs(tableID, "Notes", models.FieldTypeText, defaultOptions, 3).
			WillReturnRows(insertRows)

		field, err := store.CreateField(ctx, tableID, "Notes", models.FieldTypeText, nil, userID)
		require.NoError(t, err)
		assert.Equal(t, fieldID, field.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole returns viewer
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		field, err := store.CreateField(ctx, tableID, "Name", models.FieldTypeText, nil, userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, field)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFieldStore_GetField(t *testing.T) {
	ctx := context.Background()

	t.Run("returns field when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID := uuid.New()
		now := time.Now().UTC()
		options := json.RawMessage(`{}`)

		// Mock get field
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, "Name", models.FieldTypeText, options, 0, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at").
			WithArgs(fieldID).
			WillReturnRows(fieldRows)

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		field, err := store.GetField(ctx, fieldID, userID)
		require.NoError(t, err)
		assert.Equal(t, fieldID, field.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when field doesn't exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		fieldID := uuid.New()

		// Mock get field returns no rows
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at").
			WithArgs(fieldID).
			WillReturnError(pgx.ErrNoRows)

		field, err := store.GetField(ctx, fieldID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, field)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFieldStore_UpdateField(t *testing.T) {
	ctx := context.Background()

	t.Run("updates field name successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID := uuid.New()
		now := time.Now().UTC()
		options := json.RawMessage(`{}`)
		newName := "Updated Name"

		// Mock GetField
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, "Old Name", models.FieldTypeText, options, 0, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at").
			WithArgs(fieldID).
			WillReturnRows(fieldRows)

		// Mock getBaseIDForTable for GetField
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetField
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getBaseIDForTable for update
		baseRows2 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows2)

		// Mock GetUserRole for update permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock update
		updateRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, newName, models.FieldTypeText, options, 0, now, now)
		mock.ExpectQuery("UPDATE fields SET name").
			WithArgs(fieldID, newName, options).
			WillReturnRows(updateRows)

		field, err := store.UpdateField(ctx, fieldID, &newName, nil, userID)
		require.NoError(t, err)
		assert.Equal(t, newName, field.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to update", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID := uuid.New()
		now := time.Now().UTC()
		options := json.RawMessage(`{}`)
		newName := "Updated Name"

		// Mock GetField
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, "Old Name", models.FieldTypeText, options, 0, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at").
			WithArgs(fieldID).
			WillReturnRows(fieldRows)

		// Mock getBaseIDForTable for GetField
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetField (viewer can view)
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getBaseIDForTable for update
		baseRows2 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows2)

		// Mock GetUserRole for update permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		field, err := store.UpdateField(ctx, fieldID, &newName, nil, userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, field)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFieldStore_DeleteField(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes field successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID := uuid.New()
		now := time.Now().UTC()
		options := json.RawMessage(`{}`)

		// Mock GetField
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
			AddRow(fieldID, tableID, "Name", models.FieldTypeText, options, 0, now, now)
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at").
			WithArgs(fieldID).
			WillReturnRows(fieldRows)

		// Mock getBaseIDForTable for GetField
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetField
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getBaseIDForTable for delete
		baseRows2 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows2)

		// Mock GetUserRole for delete permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock delete
		mock.ExpectExec("DELETE FROM fields WHERE id").
			WithArgs(fieldID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteField(ctx, fieldID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFieldStore_ReorderFields(t *testing.T) {
	ctx := context.Background()

	t.Run("reorders fields successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		fieldID1 := uuid.New()
		fieldID2 := uuid.New()

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// Mock transaction
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE fields SET position").
			WithArgs(0, fieldID1, tableID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectExec("UPDATE fields SET position").
			WithArgs(1, fieldID2, tableID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))
		mock.ExpectCommit()

		err = store.ReorderFields(ctx, tableID, []uuid.UUID{fieldID1, fieldID2}, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to reorder", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewFieldStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()

		// Mock getBaseIDForTable
		baseRows := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows)

		// Mock GetUserRole returns viewer
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		err = store.ReorderFields(ctx, tableID, []uuid.UUID{uuid.New()}, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFieldStore_SetHub(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	tableStore := NewTableStore(mock, baseStore)
	store := NewFieldStore(mock, baseStore, tableStore)

	// Initially hub is nil
	assert.Nil(t, store.hub)

	// Import realtime for testing
	hub := &realtime.Hub{}
	store.SetHub(hub)

	assert.Equal(t, hub, store.hub)
}
