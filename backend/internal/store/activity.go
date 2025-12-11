package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
)

type ActivityStore struct {
	db        DBTX
	baseStore *BaseStore
}

func NewActivityStore(db DBTX, baseStore *BaseStore) *ActivityStore {
	return &ActivityStore{
		db:        db,
		baseStore: baseStore,
	}
}

// LogActivity records an activity event
func (s *ActivityStore) LogActivity(ctx context.Context, activity *models.Activity) error {
	// If BaseID is not set but TableID is, look up the base_id from the table
	if activity.BaseID == uuid.Nil && activity.TableID != nil {
		err := s.db.QueryRow(ctx, `SELECT base_id FROM tables WHERE id = $1`, *activity.TableID).Scan(&activity.BaseID)
		if err != nil {
			return fmt.Errorf("failed to get base_id for table: %w", err)
		}
	}

	_, err := s.db.Exec(ctx, `
		INSERT INTO activities (base_id, table_id, record_id, user_id, action, entity_type, entity_name, changes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, activity.BaseID, activity.TableID, activity.RecordID, activity.UserID,
		activity.Action, activity.EntityType, activity.EntityName, activity.Changes)
	return err
}

// ActivityFilters for querying activities
type ActivityFilters struct {
	UserID     *uuid.UUID
	Action     *string
	EntityType *string
	TableID    *uuid.UUID
}

// ListActivitiesForBase returns activities for a base
func (s *ActivityStore) ListActivitiesForBase(ctx context.Context, baseID uuid.UUID, userID uuid.UUID, filters ActivityFilters, limit, offset int) ([]*models.Activity, error) {
	// Verify user has access
	_, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Build query with filters
	query := `
		SELECT a.id, a.base_id, a.table_id, a.record_id, a.user_id, a.action,
		       a.entity_type, a.entity_name, a.changes, a.created_at,
		       u.id, u.email, u.name, u.created_at, u.updated_at
		FROM activities a
		JOIN users u ON a.user_id = u.id
		WHERE a.base_id = $1
	`
	args := []interface{}{baseID}
	argIndex := 2

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND a.user_id = $%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}
	if filters.Action != nil {
		query += fmt.Sprintf(" AND a.action = $%d", argIndex)
		args = append(args, *filters.Action)
		argIndex++
	}
	if filters.EntityType != nil {
		query += fmt.Sprintf(" AND a.entity_type = $%d", argIndex)
		args = append(args, *filters.EntityType)
		argIndex++
	}
	if filters.TableID != nil {
		query += fmt.Sprintf(" AND a.table_id = $%d", argIndex)
		args = append(args, *filters.TableID)
		argIndex++
	}

	query += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		var a models.Activity
		var user models.User
		if err := rows.Scan(
			&a.ID, &a.BaseID, &a.TableID, &a.RecordID, &a.UserID, &a.Action,
			&a.EntityType, &a.EntityName, &a.Changes, &a.CreatedAt,
			&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		a.User = &user
		activities = append(activities, &a)
	}

	if activities == nil {
		activities = []*models.Activity{}
	}

	return activities, rows.Err()
}

// ListActivitiesForRecord returns activities for a specific record
func (s *ActivityStore) ListActivitiesForRecord(ctx context.Context, recordID uuid.UUID, userID uuid.UUID, limit int) ([]*models.Activity, error) {
	// Get base ID for the record
	var baseID uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT t.base_id
		FROM records r
		JOIN tables t ON r.table_id = t.id
		WHERE r.id = $1
	`, recordID).Scan(&baseID)
	if err != nil {
		return nil, err
	}

	// Verify user has access
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.base_id, a.table_id, a.record_id, a.user_id, a.action,
		       a.entity_type, a.entity_name, a.changes, a.created_at,
		       u.id, u.email, u.name, u.created_at, u.updated_at
		FROM activities a
		JOIN users u ON a.user_id = u.id
		WHERE a.record_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2
	`, recordID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []*models.Activity
	for rows.Next() {
		var a models.Activity
		var user models.User
		if err := rows.Scan(
			&a.ID, &a.BaseID, &a.TableID, &a.RecordID, &a.UserID, &a.Action,
			&a.EntityType, &a.EntityName, &a.Changes, &a.CreatedAt,
			&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		a.User = &user
		activities = append(activities, &a)
	}

	if activities == nil {
		activities = []*models.Activity{}
	}

	return activities, rows.Err()
}

// Helper functions to create activity records

// LogRecordCreate logs a record creation
func (s *ActivityStore) LogRecordCreate(ctx context.Context, baseID, tableID, recordID, userID uuid.UUID) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		RecordID:   &recordID,
		UserID:     userID,
		Action:     models.ActionCreate,
		EntityType: models.EntityTypeRecord,
	})
}

// LogRecordUpdate logs a record update with changes
func (s *ActivityStore) LogRecordUpdate(ctx context.Context, baseID, tableID, recordID, userID uuid.UUID, changes []models.ActivityChanges) error {
	var changesJSON json.RawMessage
	if len(changes) > 0 {
		changesJSON, _ = json.Marshal(changes)
	}
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		RecordID:   &recordID,
		UserID:     userID,
		Action:     models.ActionUpdate,
		EntityType: models.EntityTypeRecord,
		Changes:    changesJSON,
	})
}

// LogRecordDelete logs a record deletion
func (s *ActivityStore) LogRecordDelete(ctx context.Context, baseID, tableID, userID uuid.UUID) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		UserID:     userID,
		Action:     models.ActionDelete,
		EntityType: models.EntityTypeRecord,
	})
}

// LogFieldCreate logs a field creation
func (s *ActivityStore) LogFieldCreate(ctx context.Context, baseID, tableID, userID uuid.UUID, fieldName string) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		UserID:     userID,
		Action:     models.ActionCreate,
		EntityType: models.EntityTypeField,
		EntityName: &fieldName,
	})
}

// LogFieldUpdate logs a field update
func (s *ActivityStore) LogFieldUpdate(ctx context.Context, baseID, tableID, userID uuid.UUID, fieldName string) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		UserID:     userID,
		Action:     models.ActionUpdate,
		EntityType: models.EntityTypeField,
		EntityName: &fieldName,
	})
}

// LogFieldDelete logs a field deletion
func (s *ActivityStore) LogFieldDelete(ctx context.Context, baseID, tableID, userID uuid.UUID, fieldName string) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		UserID:     userID,
		Action:     models.ActionDelete,
		EntityType: models.EntityTypeField,
		EntityName: &fieldName,
	})
}

// LogTableCreate logs a table creation
func (s *ActivityStore) LogTableCreate(ctx context.Context, baseID, tableID, userID uuid.UUID, tableName string) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		UserID:     userID,
		Action:     models.ActionCreate,
		EntityType: models.EntityTypeTable,
		EntityName: &tableName,
	})
}

// LogTableUpdate logs a table update
func (s *ActivityStore) LogTableUpdate(ctx context.Context, baseID, tableID, userID uuid.UUID, tableName string) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		TableID:    &tableID,
		UserID:     userID,
		Action:     models.ActionUpdate,
		EntityType: models.EntityTypeTable,
		EntityName: &tableName,
	})
}

// LogTableDelete logs a table deletion
func (s *ActivityStore) LogTableDelete(ctx context.Context, baseID, userID uuid.UUID, tableName string) error {
	return s.LogActivity(ctx, &models.Activity{
		BaseID:     baseID,
		UserID:     userID,
		Action:     models.ActionDelete,
		EntityType: models.EntityTypeTable,
		EntityName: &tableName,
	})
}
