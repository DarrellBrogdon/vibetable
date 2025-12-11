package store

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestNewActivityStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)
	assert.NotNil(t, store)
}

func TestActivityStore_LogActivity(t *testing.T) {
	ctx := context.Background()

	t.Run("logs activity with base ID", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		userID := uuid.New()

		activity := &models.Activity{
			BaseID:     baseID,
			TableID:    &tableID,
			RecordID:   &recordID,
			UserID:     userID,
			Action:     models.ActionCreate,
			EntityType: models.EntityTypeRecord,
		}

		mock.ExpectExec("INSERT INTO activities").
			WithArgs(baseID, &tableID, &recordID, userID, models.ActionCreate, models.EntityTypeRecord, (*string)(nil), (json.RawMessage)(nil)).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = store.LogActivity(ctx, activity)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("looks up base ID from table when not provided", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		tableID := uuid.New()
		userID := uuid.New()
		baseID := uuid.New()

		activity := &models.Activity{
			BaseID:     uuid.Nil,
			TableID:    &tableID,
			UserID:     userID,
			Action:     models.ActionCreate,
			EntityType: models.EntityTypeTable,
		}

		// Expect lookup of base_id from table
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectExec("INSERT INTO activities").
			WithArgs(baseID, &tableID, (*uuid.UUID)(nil), userID, models.ActionCreate, models.EntityTypeTable, (*string)(nil), (json.RawMessage)(nil)).
			WillReturnResult(pgxmock.NewResult("INSERT", 1))

		err = store.LogActivity(ctx, activity)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when table lookup fails", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		tableID := uuid.New()

		activity := &models.Activity{
			BaseID:     uuid.Nil,
			TableID:    &tableID,
			UserID:     uuid.New(),
			Action:     models.ActionCreate,
			EntityType: models.EntityTypeTable,
		}

		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnError(errors.New("table not found"))

		err = store.LogActivity(ctx, activity)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get base_id for table")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestActivityStore_ListActivitiesForBase(t *testing.T) {
	ctx := context.Background()

	t.Run("returns activities for base", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		userID := uuid.New()
		baseID := uuid.New()
		activityID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// List activities - use pointers for nullable columns
		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "record_id", "user_id", "action",
			"entity_type", "entity_name", "changes", "created_at",
			"u_id", "email", "name", "u_created_at", "u_updated_at",
		}).AddRow(
			activityID, baseID, &tableID, &recordID, userID, models.ActionCreate,
			models.EntityTypeRecord, nil, nil, now,
			userID, "test@example.com", name, now, now,
		)

		mock.ExpectQuery("SELECT a.id, a.base_id, a.table_id, a.record_id, a.user_id, a.action").
			WithArgs(baseID, 50, 0).
			WillReturnRows(rows)

		filters := ActivityFilters{}
		activities, err := store.ListActivitiesForBase(ctx, baseID, userID, filters, 0, 0)
		require.NoError(t, err)
		assert.Len(t, activities, 1)
		assert.Equal(t, activityID, activities[0].ID)
		assert.NotNil(t, activities[0].User)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when user has no access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		userID := uuid.New()
		baseID := uuid.New()

		// GetUserRole returns no rows
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}))

		filters := ActivityFilters{}
		activities, err := store.ListActivitiesForBase(ctx, baseID, userID, filters, 50, 0)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, activities)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no activities", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		userID := uuid.New()
		baseID := uuid.New()

		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		mock.ExpectQuery("SELECT a.id, a.base_id, a.table_id, a.record_id, a.user_id, a.action").
			WithArgs(baseID, 50, 0).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "base_id", "table_id", "record_id", "user_id", "action",
				"entity_type", "entity_name", "changes", "created_at",
				"u_id", "email", "name", "u_created_at", "u_updated_at",
			}))

		filters := ActivityFilters{}
		activities, err := store.ListActivitiesForBase(ctx, baseID, userID, filters, 0, 0)
		require.NoError(t, err)
		assert.NotNil(t, activities)
		assert.Empty(t, activities)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestActivityStore_ListActivitiesForRecord(t *testing.T) {
	ctx := context.Background()

	t.Run("returns activities for record", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		userID := uuid.New()
		recordID := uuid.New()
		baseID := uuid.New()
		activityID := uuid.New()
		tableID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		// Get base ID for record
		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		roleRows := pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner)
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(roleRows)

		// List activities - use pointers for nullable columns
		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "record_id", "user_id", "action",
			"entity_type", "entity_name", "changes", "created_at",
			"u_id", "email", "name", "u_created_at", "u_updated_at",
		}).AddRow(
			activityID, baseID, &tableID, &recordID, userID, models.ActionUpdate,
			models.EntityTypeRecord, nil, nil, now,
			userID, "test@example.com", name, now, now,
		)

		mock.ExpectQuery("SELECT a.id, a.base_id, a.table_id, a.record_id, a.user_id, a.action").
			WithArgs(recordID, 20).
			WillReturnRows(rows)

		activities, err := store.ListActivitiesForRecord(ctx, recordID, userID, 0)
		require.NoError(t, err)
		assert.Len(t, activities, 1)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when record not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewActivityStore(mock, baseStore)

		recordID := uuid.New()

		mock.ExpectQuery("SELECT t.base_id").
			WithArgs(recordID).
			WillReturnError(errors.New("no rows"))

		activities, err := store.ListActivitiesForRecord(ctx, recordID, uuid.New(), 20)
		assert.Error(t, err)
		assert.Nil(t, activities)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestActivityStore_LogRecordCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	recordID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(baseID, &tableID, &recordID, userID, models.ActionCreate, models.EntityTypeRecord, (*string)(nil), (json.RawMessage)(nil)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogRecordCreate(context.Background(), baseID, tableID, recordID, userID)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogRecordUpdate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	recordID := uuid.New()
	userID := uuid.New()

	changes := []models.ActivityChanges{
		{FieldID: uuid.New().String(), FieldName: "Name", OldValue: "Old", NewValue: "New"},
	}

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogRecordUpdate(context.Background(), baseID, tableID, recordID, userID, changes)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogRecordDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	userID := uuid.New()

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(baseID, &tableID, (*uuid.UUID)(nil), userID, models.ActionDelete, models.EntityTypeRecord, (*string)(nil), (json.RawMessage)(nil)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogRecordDelete(context.Background(), baseID, tableID, userID)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogFieldCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	userID := uuid.New()
	fieldName := "New Field"

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogFieldCreate(context.Background(), baseID, tableID, userID, fieldName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogFieldUpdate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	userID := uuid.New()
	fieldName := "Updated Field"

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogFieldUpdate(context.Background(), baseID, tableID, userID, fieldName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogFieldDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	userID := uuid.New()
	fieldName := "Deleted Field"

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogFieldDelete(context.Background(), baseID, tableID, userID, fieldName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogTableCreate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	userID := uuid.New()
	tableName := "New Table"

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogTableCreate(context.Background(), baseID, tableID, userID, tableName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogTableUpdate(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	tableID := uuid.New()
	userID := uuid.New()
	tableName := "Updated Table"

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogTableUpdate(context.Background(), baseID, tableID, userID, tableName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestActivityStore_LogTableDelete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewActivityStore(mock, baseStore)

	baseID := uuid.New()
	userID := uuid.New()
	tableName := "Deleted Table"

	mock.ExpectExec("INSERT INTO activities").
		WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err = store.LogTableDelete(context.Background(), baseID, userID, tableName)
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
