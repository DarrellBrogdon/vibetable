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
)

func TestNewAutomationStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewAutomationStore(mock, nil, nil)
	assert.NotNil(t, store)
}

func TestAutomationStore_GetAutomation(t *testing.T) {
	ctx := context.Background()

	t.Run("returns automation when found and user has access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAutomationStore(mock, baseStore, nil)

		automationID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnRows(rows)

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		automation, err := store.GetAutomation(ctx, automationID, userID)
		require.NoError(t, err)
		assert.Equal(t, automationID, automation.ID)
		assert.Equal(t, "Test Automation", automation.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when automation not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAutomationStore(mock, nil, nil)
		automationID := uuid.New()

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnError(pgx.ErrNoRows)

		automation, err := store.GetAutomation(ctx, automationID, uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, automation)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when user has no access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAutomationStore(mock, baseStore, nil)

		automationID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnRows(rows)

		// GetUserRole returns no rows
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}))

		automation, err := store.GetAutomation(ctx, automationID, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, automation)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_ListAutomationsForTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns automations for table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewAutomationStore(mock, baseStore, tableStore)

		tableID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		automationID := uuid.New()
		now := time.Now().UTC()

		// GetTable query first (without description column)
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// GetTable - then check role
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		// List automations
		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(tableID).
			WillReturnRows(rows)

		automations, err := store.ListAutomationsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.Len(t, automations, 1)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no automations", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		tableStore := NewTableStore(mock, baseStore)
		store := NewAutomationStore(mock, baseStore, tableStore)

		tableID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		// GetTable query first (without description column)
		tableRows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
			AddRow(tableID, baseID, "Test Table", 0, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, position, created_at, updated_at FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(tableRows)

		// Then check role
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "base_id", "table_id", "name", "description", "enabled",
				"trigger_type", "trigger_config", "action_type", "action_config",
				"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
			}))

		automations, err := store.ListAutomationsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.NotNil(t, automations)
		assert.Empty(t, automations)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_DeleteAutomation(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes automation when user has edit access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAutomationStore(mock, baseStore, nil)

		automationID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		// GetAutomation
		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnRows(rows)

		// GetUserRole check for GetAutomation
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// GetUserRole check for delete
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Delete
		mock.ExpectExec("DELETE FROM automations WHERE id").
			WithArgs(automationID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteAutomation(ctx, automationID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAutomationStore(mock, baseStore, nil)

		automationID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnRows(rows)

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		err = store.DeleteAutomation(ctx, automationID, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_ToggleAutomation(t *testing.T) {
	ctx := context.Background()

	t.Run("toggles automation enabled status", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAutomationStore(mock, baseStore, nil)

		automationID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		// GetAutomation
		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnRows(rows)

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Update enabled
		updateRows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, false,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("UPDATE automations SET enabled").
			WithArgs(false, automationID).
			WillReturnRows(updateRows)

		automation, err := store.ToggleAutomation(ctx, automationID, false, userID)
		require.NoError(t, err)
		assert.False(t, automation.Enabled)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_GetAutomationsByTrigger(t *testing.T) {
	ctx := context.Background()

	t.Run("returns enabled automations by trigger type", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAutomationStore(mock, nil, nil)

		tableID := uuid.New()
		automationID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(tableID, models.TriggerRecordCreated).
			WillReturnRows(rows)

		automations, err := store.GetAutomationsByTrigger(ctx, tableID, models.TriggerRecordCreated)
		require.NoError(t, err)
		assert.Len(t, automations, 1)
		assert.Equal(t, automationID, automations[0].ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no matching automations", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAutomationStore(mock, nil, nil)
		tableID := uuid.New()

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(tableID, models.TriggerRecordDeleted).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "base_id", "table_id", "name", "description", "enabled",
				"trigger_type", "trigger_config", "action_type", "action_config",
				"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
			}))

		automations, err := store.GetAutomationsByTrigger(ctx, tableID, models.TriggerRecordDeleted)
		require.NoError(t, err)
		assert.Nil(t, automations)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_CreateRun(t *testing.T) {
	ctx := context.Background()

	t.Run("creates automation run", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAutomationStore(mock, nil, nil)

		automationID := uuid.New()
		runID := uuid.New()
		recordID := uuid.New()
		now := time.Now().UTC()

		run := &models.AutomationRun{
			AutomationID:    automationID,
			Status:          models.RunStatusPending,
			TriggerRecordID: &recordID,
			TriggerData:     json.RawMessage(`{"test": true}`),
		}

		mock.ExpectQuery("INSERT INTO automation_runs").
			WithArgs(automationID, models.RunStatusPending, &recordID, json.RawMessage(`{"test": true}`)).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "automation_id", "status", "trigger_record_id", "trigger_data", "result", "error", "started_at", "completed_at",
			}).AddRow(runID, automationID, models.RunStatusPending, &recordID, json.RawMessage(`{"test": true}`), nil, nil, now, nil))

		result, err := store.CreateRun(ctx, run)
		require.NoError(t, err)
		assert.Equal(t, runID, result.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_UpdateRun(t *testing.T) {
	ctx := context.Background()

	t.Run("updates run status", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAutomationStore(mock, nil, nil)
		runID := uuid.New()
		errMsg := "something failed"

		mock.ExpectExec("UPDATE automation_runs SET status").
			WithArgs(models.RunStatusFailed, ([]byte)(nil), &errMsg, pgxmock.AnyArg(), runID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = store.UpdateRun(ctx, runID, models.RunStatusFailed, nil, &errMsg)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_UpdateAutomationStats(t *testing.T) {
	ctx := context.Background()

	t.Run("updates automation stats", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAutomationStore(mock, nil, nil)
		automationID := uuid.New()

		mock.ExpectExec("UPDATE automations SET last_triggered_at").
			WithArgs(automationID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = store.UpdateAutomationStats(ctx, automationID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAutomationStore_ListRunsForAutomation(t *testing.T) {
	ctx := context.Background()

	t.Run("returns runs for automation", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAutomationStore(mock, baseStore, nil)

		automationID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		runID := uuid.New()
		now := time.Now().UTC()

		// GetAutomation
		automationRows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnRows(automationRows)

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		// List runs
		runRows := pgxmock.NewRows([]string{
			"id", "automation_id", "status", "trigger_record_id", "trigger_data", "result", "error", "started_at", "completed_at",
		}).AddRow(runID, automationID, models.RunStatusSuccess, nil, nil, nil, nil, now, &now)

		mock.ExpectQuery("SELECT id, automation_id, status, trigger_record_id, trigger_data, result, error, started_at, completed_at").
			WithArgs(automationID, 10).
			WillReturnRows(runRows)

		runs, err := store.ListRunsForAutomation(ctx, automationID, 10, userID)
		require.NoError(t, err)
		assert.Len(t, runs, 1)
		assert.Equal(t, models.RunStatusSuccess, runs[0].Status)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no runs", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewAutomationStore(mock, baseStore, nil)

		automationID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		automationRows := pgxmock.NewRows([]string{
			"id", "base_id", "table_id", "name", "description", "enabled",
			"trigger_type", "trigger_config", "action_type", "action_config",
			"created_by", "last_triggered_at", "run_count", "created_at", "updated_at",
		}).AddRow(
			automationID, baseID, tableID, "Test Automation", nil, true,
			models.TriggerRecordCreated, json.RawMessage(`{}`), models.ActionSendWebhook, json.RawMessage(`{}`),
			userID, nil, 0, now, now,
		)

		mock.ExpectQuery("SELECT id, base_id, table_id, name, description, enabled").
			WithArgs(automationID).
			WillReturnRows(automationRows)

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		mock.ExpectQuery("SELECT id, automation_id, status, trigger_record_id, trigger_data, result, error, started_at, completed_at").
			WithArgs(automationID, 10).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "automation_id", "status", "trigger_record_id", "trigger_data", "result", "error", "started_at", "completed_at",
			}))

		runs, err := store.ListRunsForAutomation(ctx, automationID, 10, userID)
		require.NoError(t, err)
		assert.NotNil(t, runs)
		assert.Empty(t, runs)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
