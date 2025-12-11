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

func TestNewRecordStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	tableStore := NewTableStore(mock, baseStore)
	store := NewRecordStore(mock, baseStore, tableStore)
	assert.NotNil(t, store)
}

func TestRecordStore_getBaseIDForTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns base ID when table exists", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
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
		store := NewRecordStore(mock, baseStore, tableStore)
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

func TestRecordStore_ListRecordsForTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns records for table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		values := json.RawMessage(`{"field1": "value1"}`)

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

		// Mock list records
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, values, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(recordRows)

		// Mock getFieldsForTable for computed fields
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows)

		records, err := store.ListRecordsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.Len(t, records, 1)
		assert.Equal(t, recordID, records[0].ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no records", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
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
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(tableID).
			WillReturnRows(recordRows)

		// Mock getFieldsForTable for computed fields
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows)

		records, err := store.ListRecordsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.NotNil(t, records)
		assert.Empty(t, records)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecordStore_CreateRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("creates record successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		values := json.RawMessage(`{"field1": "value1"}`)

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
		insertRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, values, 0, nil, now, now)
		mock.ExpectQuery("INSERT INTO records").
			WithArgs(tableID, values, 0).
			WillReturnRows(insertRows)

		record, err := store.CreateRecord(ctx, tableID, values, userID)
		require.NoError(t, err)
		assert.Equal(t, recordID, record.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creates record with nil values defaults to empty object", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		defaultValues := json.RawMessage(`{}`)

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
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(5)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(tableID).
			WillReturnRows(posRows)

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, defaultValues, 6, nil, now, now)
		mock.ExpectQuery("INSERT INTO records").
			WithArgs(tableID, defaultValues, 6).
			WillReturnRows(insertRows)

		record, err := store.CreateRecord(ctx, tableID, nil, userID)
		require.NoError(t, err)
		assert.Equal(t, recordID, record.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
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

		record, err := store.CreateRecord(ctx, tableID, nil, userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, record)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecordStore_GetRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("returns record when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		values := json.RawMessage(`{"field1": "value1"}`)

		// Mock get record
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, values, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnRows(recordRows)

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

		// Mock getFieldsForTable for computed fields
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows)

		record, err := store.GetRecord(ctx, recordID, userID)
		require.NoError(t, err)
		assert.Equal(t, recordID, record.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when record doesn't exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		recordID := uuid.New()

		// Mock get record returns no rows
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnError(pgx.ErrNoRows)

		record, err := store.GetRecord(ctx, recordID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, record)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecordStore_UpdateRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("updates record successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		oldValues := json.RawMessage(`{"field1": "old"}`)
		newValues := json.RawMessage(`{"field1": "new"}`)

		// Mock GetRecord
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, oldValues, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnRows(recordRows)

		// Mock getBaseIDForTable for GetRecord
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetRecord
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getFieldsForTable for computed fields (GetRecord)
		fieldRows1 := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows1)

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
		updateRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, newValues, 0, nil, now, now)
		mock.ExpectQuery("UPDATE records SET values").
			WithArgs(recordID, newValues).
			WillReturnRows(updateRows)

		record, err := store.UpdateRecord(ctx, recordID, newValues, userID)
		require.NoError(t, err)
		assert.Equal(t, recordID, record.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to update", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		values := json.RawMessage(`{"field1": "value"}`)

		// Mock GetRecord
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, values, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnRows(recordRows)

		// Mock getBaseIDForTable for GetRecord
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetRecord
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getFieldsForTable for computed fields (GetRecord)
		fieldRows1 := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows1)

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

		record, err := store.UpdateRecord(ctx, recordID, values, userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, record)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecordStore_PatchRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("patches record by merging values", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		existingValues := json.RawMessage(`{"field1": "old", "field2": "keep"}`)
		mergedValues := json.RawMessage(`{"field1":"new","field2":"keep"}`)

		// Mock GetRecord
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, existingValues, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnRows(recordRows)

		// Mock getBaseIDForTable for GetRecord
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetRecord
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getFieldsForTable for computed fields (GetRecord)
		fieldRows1 := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows1)

		// Mock getBaseIDForTable for patch
		baseRows2 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows2)

		// Mock GetUserRole for patch permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock update - use AnyArg for the merged values since JSON ordering may vary
		updateRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, mergedValues, 0, nil, now, now)
		mock.ExpectQuery("UPDATE records SET values").
			WithArgs(recordID, pgxmock.AnyArg()).
			WillReturnRows(updateRows)

		newValues := map[string]interface{}{"field1": "new"}
		record, err := store.PatchRecord(ctx, recordID, newValues, userID)
		require.NoError(t, err)
		assert.Equal(t, recordID, record.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("patches record by deleting nil values", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		existingValues := json.RawMessage(`{"field1": "value", "field2": "delete"}`)
		mergedValues := json.RawMessage(`{"field1":"value"}`)

		// Mock GetRecord
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, existingValues, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnRows(recordRows)

		// Mock getBaseIDForTable for GetRecord
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetRecord
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getFieldsForTable for computed fields (GetRecord)
		fieldRows1 := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows1)

		// Mock getBaseIDForTable for patch
		baseRows2 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows2)

		// Mock GetUserRole for patch permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		// Mock update
		updateRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, mergedValues, 0, nil, now, now)
		mock.ExpectQuery("UPDATE records SET values").
			WithArgs(recordID, pgxmock.AnyArg()).
			WillReturnRows(updateRows)

		newValues := map[string]interface{}{"field2": nil}
		record, err := store.PatchRecord(ctx, recordID, newValues, userID)
		require.NoError(t, err)
		assert.Equal(t, recordID, record.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecordStore_DeleteRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes record successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		values := json.RawMessage(`{}`)

		// Mock GetRecord
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, values, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnRows(recordRows)

		// Mock getBaseIDForTable for GetRecord
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetRecord
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getFieldsForTable for computed fields
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows)

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
		mock.ExpectExec("DELETE FROM records WHERE id").
			WithArgs(recordID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteRecord(ctx, recordID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		values := json.RawMessage(`{}`)

		// Mock GetRecord
		recordRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, values, 0, nil, now, now)
		mock.ExpectQuery("SELECT id, table_id, values, position, color, created_at, updated_at").
			WithArgs(recordID).
			WillReturnRows(recordRows)

		// Mock getBaseIDForTable for GetRecord
		baseRows1 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows1)

		// Mock GetUserRole for GetRecord
		roleRows1 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows1)

		// Mock getFieldsForTable for computed fields
		fieldRows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, table_id, name, field_type, options, position, created_at, updated_at FROM fields").
			WithArgs(tableID).
			WillReturnRows(fieldRows)

		// Mock getBaseIDForTable for delete
		baseRows2 := pgxmock.NewRows([]string{"base_id"}).AddRow(baseID)
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(baseRows2)

		// Mock GetUserRole for delete permission
		roleRows2 := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows2)

		err = store.DeleteRecord(ctx, recordID, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecordStore_BulkCreateRecords(t *testing.T) {
	ctx := context.Background()

	t.Run("creates multiple records successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID1 := uuid.New()
		recordID2 := uuid.New()
		now := time.Now().UTC()
		values1 := json.RawMessage(`{"field1": "value1"}`)
		values2 := json.RawMessage(`{"field1": "value2"}`)

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

		// Mock max position
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(5)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(tableID).
			WillReturnRows(posRows)

		// Mock insert record 1
		insertRows1 := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID1, tableID, values1, 6, nil, now, now)
		mock.ExpectQuery("INSERT INTO records").
			WithArgs(tableID, values1, 6).
			WillReturnRows(insertRows1)

		// Mock insert record 2
		insertRows2 := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID2, tableID, values2, 7, nil, now, now)
		mock.ExpectQuery("INSERT INTO records").
			WithArgs(tableID, values2, 7).
			WillReturnRows(insertRows2)

		mock.ExpectCommit()

		records, err := store.BulkCreateRecords(ctx, tableID, []json.RawMessage{values1, values2}, userID)
		require.NoError(t, err)
		assert.Len(t, records, 2)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles nil values in bulk create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
		userID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		defaultValues := json.RawMessage(`{}`)

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

		// Mock transaction
		mock.ExpectBegin()

		// Mock max position
		posRows := pgxmock.NewRows([]string{"coalesce"}).AddRow(-1)
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(tableID).
			WillReturnRows(posRows)

		// Mock insert record with nil becoming {}
		insertRows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "color", "created_at", "updated_at"}).
			AddRow(recordID, tableID, defaultValues, 0, nil, now, now)
		mock.ExpectQuery("INSERT INTO records").
			WithArgs(tableID, defaultValues, 0).
			WillReturnRows(insertRows)

		mock.ExpectCommit()

		records, err := store.BulkCreateRecords(ctx, tableID, []json.RawMessage{nil}, userID)
		require.NoError(t, err)
		assert.Len(t, records, 1)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries bulk create", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewRecordStore(mock, baseStore, tableStore)
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

		records, err := store.BulkCreateRecords(ctx, tableID, []json.RawMessage{nil}, userID)
		assert.ErrorIs(t, err, ErrForbidden)
		assert.Nil(t, records)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRecordStore_SetHub(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	tableStore := NewTableStore(mock, baseStore)
	store := NewRecordStore(mock, baseStore, tableStore)

	// Initially hub is nil
	assert.Nil(t, store.hub)

	hub := &realtime.Hub{}
	store.SetHub(hub)

	assert.Equal(t, hub, store.hub)
}

func TestRecordStore_SetAutomationCallback(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	tableStore := NewTableStore(mock, baseStore)
	store := NewRecordStore(mock, baseStore, tableStore)

	// Initially callback is nil
	assert.Nil(t, store.automationCallback)

	callback := func(tableID uuid.UUID, recordID *uuid.UUID, record *models.Record, oldRecord *models.Record, triggerType string, userID uuid.UUID) {}
	store.SetAutomationCallback(callback)

	// Can't directly compare funcs, just check it's not nil
	assert.NotNil(t, store.automationCallback)
}
