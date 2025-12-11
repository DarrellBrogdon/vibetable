package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
)

type AutomationStore struct {
	db         DBTX
	baseStore  *BaseStore
	tableStore *TableStore
}

func NewAutomationStore(db DBTX, baseStore *BaseStore, tableStore *TableStore) *AutomationStore {
	return &AutomationStore{
		db:         db,
		baseStore:  baseStore,
		tableStore: tableStore,
	}
}

// ListAutomationsForTable returns all automations for a table
func (s *AutomationStore) ListAutomationsForTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID) ([]models.Automation, error) {
	// Verify user has access to this table
	table, err := s.tableStore.GetTable(ctx, tableID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, base_id, table_id, name, description, enabled,
			   trigger_type, trigger_config, action_type, action_config,
			   created_by, last_triggered_at, run_count, created_at, updated_at
		FROM automations
		WHERE table_id = $1
		ORDER BY created_at DESC
	`, table.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var automations []models.Automation
	for rows.Next() {
		var a models.Automation
		if err := rows.Scan(
			&a.ID, &a.BaseID, &a.TableID, &a.Name, &a.Description, &a.Enabled,
			&a.TriggerType, &a.TriggerConfig, &a.ActionType, &a.ActionConfig,
			&a.CreatedBy, &a.LastTriggeredAt, &a.RunCount, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		automations = append(automations, a)
	}

	if automations == nil {
		automations = []models.Automation{}
	}

	return automations, rows.Err()
}

// CreateAutomation creates a new automation
func (s *AutomationStore) CreateAutomation(ctx context.Context, a *models.Automation, userID uuid.UUID) (*models.Automation, error) {
	// Verify user has edit access to this table's base
	table, err := s.tableStore.GetTable(ctx, a.TableID, userID)
	if err != nil {
		return nil, err
	}

	role, err := s.baseStore.GetUserRole(ctx, table.BaseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	a.BaseID = table.BaseID
	a.CreatedBy = userID

	err = s.db.QueryRow(ctx, `
		INSERT INTO automations (base_id, table_id, name, description, enabled,
								 trigger_type, trigger_config, action_type, action_config, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, base_id, table_id, name, description, enabled,
				  trigger_type, trigger_config, action_type, action_config,
				  created_by, last_triggered_at, run_count, created_at, updated_at
	`, a.BaseID, a.TableID, a.Name, a.Description, a.Enabled,
		a.TriggerType, a.TriggerConfig, a.ActionType, a.ActionConfig, a.CreatedBy,
	).Scan(
		&a.ID, &a.BaseID, &a.TableID, &a.Name, &a.Description, &a.Enabled,
		&a.TriggerType, &a.TriggerConfig, &a.ActionType, &a.ActionConfig,
		&a.CreatedBy, &a.LastTriggeredAt, &a.RunCount, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// GetAutomation returns an automation by ID
func (s *AutomationStore) GetAutomation(ctx context.Context, automationID uuid.UUID, userID uuid.UUID) (*models.Automation, error) {
	var a models.Automation

	err := s.db.QueryRow(ctx, `
		SELECT id, base_id, table_id, name, description, enabled,
			   trigger_type, trigger_config, action_type, action_config,
			   created_by, last_triggered_at, run_count, created_at, updated_at
		FROM automations
		WHERE id = $1
	`, automationID).Scan(
		&a.ID, &a.BaseID, &a.TableID, &a.Name, &a.Description, &a.Enabled,
		&a.TriggerType, &a.TriggerConfig, &a.ActionType, &a.ActionConfig,
		&a.CreatedBy, &a.LastTriggeredAt, &a.RunCount, &a.CreatedAt, &a.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Verify user has access to this automation's base
	_, err = s.baseStore.GetUserRole(ctx, a.BaseID, userID)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// UpdateAutomation updates an automation
func (s *AutomationStore) UpdateAutomation(ctx context.Context, automationID uuid.UUID, updates map[string]interface{}, userID uuid.UUID) (*models.Automation, error) {
	// Get current automation and verify access
	a, err := s.GetAutomation(ctx, automationID, userID)
	if err != nil {
		return nil, err
	}

	// Check edit permission
	role, err := s.baseStore.GetUserRole(ctx, a.BaseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		a.Name = name
	}
	if desc, ok := updates["description"].(*string); ok {
		a.Description = desc
	}
	if enabled, ok := updates["enabled"].(bool); ok {
		a.Enabled = enabled
	}
	if triggerType, ok := updates["triggerType"].(models.TriggerType); ok {
		a.TriggerType = triggerType
	}
	if triggerConfig, ok := updates["triggerConfig"]; ok {
		a.TriggerConfig, _ = triggerConfig.([]byte)
	}
	if actionType, ok := updates["actionType"].(models.ActionType); ok {
		a.ActionType = actionType
	}
	if actionConfig, ok := updates["actionConfig"]; ok {
		a.ActionConfig, _ = actionConfig.([]byte)
	}

	err = s.db.QueryRow(ctx, `
		UPDATE automations
		SET name = $1, description = $2, enabled = $3,
			trigger_type = $4, trigger_config = $5,
			action_type = $6, action_config = $7,
			updated_at = NOW()
		WHERE id = $8
		RETURNING id, base_id, table_id, name, description, enabled,
				  trigger_type, trigger_config, action_type, action_config,
				  created_by, last_triggered_at, run_count, created_at, updated_at
	`, a.Name, a.Description, a.Enabled,
		a.TriggerType, a.TriggerConfig, a.ActionType, a.ActionConfig,
		automationID,
	).Scan(
		&a.ID, &a.BaseID, &a.TableID, &a.Name, &a.Description, &a.Enabled,
		&a.TriggerType, &a.TriggerConfig, &a.ActionType, &a.ActionConfig,
		&a.CreatedBy, &a.LastTriggeredAt, &a.RunCount, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// DeleteAutomation deletes an automation
func (s *AutomationStore) DeleteAutomation(ctx context.Context, automationID uuid.UUID, userID uuid.UUID) error {
	// Get automation and verify access
	a, err := s.GetAutomation(ctx, automationID, userID)
	if err != nil {
		return err
	}

	// Check edit permission
	role, err := s.baseStore.GetUserRole(ctx, a.BaseID, userID)
	if err != nil {
		return err
	}
	if !role.CanEdit() {
		return ErrForbidden
	}

	_, err = s.db.Exec(ctx, `DELETE FROM automations WHERE id = $1`, automationID)
	return err
}

// ToggleAutomation enables or disables an automation
func (s *AutomationStore) ToggleAutomation(ctx context.Context, automationID uuid.UUID, enabled bool, userID uuid.UUID) (*models.Automation, error) {
	// Get automation and verify access
	a, err := s.GetAutomation(ctx, automationID, userID)
	if err != nil {
		return nil, err
	}

	// Check edit permission
	role, err := s.baseStore.GetUserRole(ctx, a.BaseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	err = s.db.QueryRow(ctx, `
		UPDATE automations
		SET enabled = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, base_id, table_id, name, description, enabled,
				  trigger_type, trigger_config, action_type, action_config,
				  created_by, last_triggered_at, run_count, created_at, updated_at
	`, enabled, automationID).Scan(
		&a.ID, &a.BaseID, &a.TableID, &a.Name, &a.Description, &a.Enabled,
		&a.TriggerType, &a.TriggerConfig, &a.ActionType, &a.ActionConfig,
		&a.CreatedBy, &a.LastTriggeredAt, &a.RunCount, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// GetAutomationsByTrigger returns all enabled automations for a given trigger type and table
func (s *AutomationStore) GetAutomationsByTrigger(ctx context.Context, tableID uuid.UUID, triggerType models.TriggerType) ([]models.Automation, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, base_id, table_id, name, description, enabled,
			   trigger_type, trigger_config, action_type, action_config,
			   created_by, last_triggered_at, run_count, created_at, updated_at
		FROM automations
		WHERE table_id = $1 AND trigger_type = $2 AND enabled = true
	`, tableID, triggerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var automations []models.Automation
	for rows.Next() {
		var a models.Automation
		if err := rows.Scan(
			&a.ID, &a.BaseID, &a.TableID, &a.Name, &a.Description, &a.Enabled,
			&a.TriggerType, &a.TriggerConfig, &a.ActionType, &a.ActionConfig,
			&a.CreatedBy, &a.LastTriggeredAt, &a.RunCount, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		automations = append(automations, a)
	}

	return automations, rows.Err()
}

// CreateRun creates a new automation run record
func (s *AutomationStore) CreateRun(ctx context.Context, run *models.AutomationRun) (*models.AutomationRun, error) {
	err := s.db.QueryRow(ctx, `
		INSERT INTO automation_runs (automation_id, status, trigger_record_id, trigger_data)
		VALUES ($1, $2, $3, $4)
		RETURNING id, automation_id, status, trigger_record_id, trigger_data, result, error, started_at, completed_at
	`, run.AutomationID, run.Status, run.TriggerRecordID, run.TriggerData,
	).Scan(
		&run.ID, &run.AutomationID, &run.Status, &run.TriggerRecordID, &run.TriggerData,
		&run.Result, &run.Error, &run.StartedAt, &run.CompletedAt,
	)
	if err != nil {
		return nil, err
	}

	return run, nil
}

// UpdateRun updates an automation run
func (s *AutomationStore) UpdateRun(ctx context.Context, runID uuid.UUID, status models.RunStatus, result []byte, errMsg *string) error {
	completedAt := time.Now()
	_, err := s.db.Exec(ctx, `
		UPDATE automation_runs
		SET status = $1, result = $2, error = $3, completed_at = $4
		WHERE id = $5
	`, status, result, errMsg, completedAt, runID)
	return err
}

// UpdateAutomationStats updates the last_triggered_at and run_count
func (s *AutomationStore) UpdateAutomationStats(ctx context.Context, automationID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE automations
		SET last_triggered_at = NOW(), run_count = run_count + 1
		WHERE id = $1
	`, automationID)
	return err
}

// ListRunsForAutomation returns recent runs for an automation
func (s *AutomationStore) ListRunsForAutomation(ctx context.Context, automationID uuid.UUID, limit int, userID uuid.UUID) ([]models.AutomationRun, error) {
	// Verify access
	_, err := s.GetAutomation(ctx, automationID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, automation_id, status, trigger_record_id, trigger_data, result, error, started_at, completed_at
		FROM automation_runs
		WHERE automation_id = $1
		ORDER BY started_at DESC
		LIMIT $2
	`, automationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.AutomationRun
	for rows.Next() {
		var r models.AutomationRun
		if err := rows.Scan(
			&r.ID, &r.AutomationID, &r.Status, &r.TriggerRecordID, &r.TriggerData,
			&r.Result, &r.Error, &r.StartedAt, &r.CompletedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}

	if runs == nil {
		runs = []models.AutomationRun{}
	}

	return runs, rows.Err()
}
