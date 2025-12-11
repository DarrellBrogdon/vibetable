package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/realtime"
)

// AutomationCallback is called when records change to trigger automations
type AutomationCallback func(tableID uuid.UUID, recordID *uuid.UUID, record *models.Record, oldRecord *models.Record, triggerType string, userID uuid.UUID)

type RecordStore struct {
	db                 DBTX
	baseStore          *BaseStore
	tableStore         *TableStore
	computedService    *ComputedFieldService
	hub                *realtime.Hub
	automationCallback AutomationCallback
}

func NewRecordStore(db DBTX, baseStore *BaseStore, tableStore *TableStore) *RecordStore {
	return &RecordStore{
		db:              db,
		baseStore:       baseStore,
		tableStore:      tableStore,
		computedService: NewComputedFieldService(db),
	}
}

// SetHub sets the realtime hub for broadcasting changes
func (s *RecordStore) SetHub(hub *realtime.Hub) {
	s.hub = hub
}

// SetAutomationCallback sets the callback for triggering automations
func (s *RecordStore) SetAutomationCallback(cb AutomationCallback) {
	s.automationCallback = cb
}

// getBaseIDForTable returns the base ID for a table
func (s *RecordStore) getBaseIDForTable(ctx context.Context, tableID uuid.UUID) (uuid.UUID, error) {
	var baseID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT base_id FROM tables WHERE id = $1`, tableID).Scan(&baseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return baseID, err
}

// ListRecordsForTable returns all records in a table
func (s *RecordStore) ListRecordsForTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID) ([]models.Record, error) {
	// Verify user has access
	baseID, err := s.getBaseIDForTable(ctx, tableID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, table_id, values, position, color, created_at, updated_at
		FROM records
		WHERE table_id = $1
		ORDER BY position, created_at
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var r models.Record
		if err := rows.Scan(&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	if records == nil {
		records = []models.Record{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Fetch fields to compute formula/rollup/lookup values
	fields, err := s.getFieldsForTable(ctx, tableID)
	if err != nil {
		return nil, err
	}

	// Compute formula, rollup, and lookup fields
	records, err = s.computedService.ComputeFieldsForRecords(ctx, records, fields)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// getFieldsForTable returns all fields for a table (internal use, no auth check)
func (s *RecordStore) getFieldsForTable(ctx context.Context, tableID uuid.UUID) ([]models.Field, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, table_id, name, field_type, options, position, created_at, updated_at
		FROM fields
		WHERE table_id = $1
		ORDER BY position, created_at
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []models.Field
	for rows.Next() {
		var f models.Field
		if err := rows.Scan(&f.ID, &f.TableID, &f.Name, &f.FieldType, &f.Options, &f.Position, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}

	return fields, rows.Err()
}

// CreateRecord creates a new record in a table
func (s *RecordStore) CreateRecord(ctx context.Context, tableID uuid.UUID, values json.RawMessage, userID uuid.UUID) (*models.Record, error) {
	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, tableID)
	if err != nil {
		return nil, err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	// Default empty values if not provided
	if values == nil {
		values = json.RawMessage(`{}`)
	}

	// Get next position
	var maxPosition int
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM records WHERE table_id = $1
	`, tableID).Scan(&maxPosition)
	if err != nil {
		return nil, err
	}

	var r models.Record
	err = s.db.QueryRow(ctx, `
		INSERT INTO records (table_id, values, position)
		VALUES ($1, $2, $3)
		RETURNING id, table_id, values, position, color, created_at, updated_at
	`, tableID, values, maxPosition+1).Scan(
		&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Broadcast record created
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeRecordCreated, baseID, userID).
			WithTable(tableID).
			WithRecord(r.ID).
			WithPayload(r)
		s.hub.Broadcast(msg)
	}

	// Trigger automations
	if s.automationCallback != nil {
		s.automationCallback(tableID, &r.ID, &r, nil, "record_created", userID)
	}

	return &r, nil
}

// GetRecord returns a record by ID
func (s *RecordStore) GetRecord(ctx context.Context, recordID uuid.UUID, userID uuid.UUID) (*models.Record, error) {
	var r models.Record
	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, values, position, color, created_at, updated_at
		FROM records WHERE id = $1
	`, recordID).Scan(&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Verify user has access
	baseID, err := s.getBaseIDForTable(ctx, r.TableID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	// Compute formula, rollup, and lookup fields
	fields, err := s.getFieldsForTable(ctx, r.TableID)
	if err != nil {
		return nil, err
	}

	records, err := s.computedService.ComputeFieldsForRecords(ctx, []models.Record{r}, fields)
	if err != nil {
		return nil, err
	}
	if len(records) > 0 {
		return &records[0], nil
	}

	return &r, nil
}

// UpdateRecord updates a record's values
func (s *RecordStore) UpdateRecord(ctx context.Context, recordID uuid.UUID, values json.RawMessage, userID uuid.UUID) (*models.Record, error) {
	// Get record to check access
	r, err := s.GetRecord(ctx, recordID, userID)
	if err != nil {
		return nil, err
	}

	// Store old record for automation comparison
	oldRecord := *r

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, r.TableID)
	if err != nil {
		return nil, err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	err = s.db.QueryRow(ctx, `
		UPDATE records SET values = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, table_id, values, position, color, created_at, updated_at
	`, recordID, values).Scan(
		&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Broadcast record updated
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeRecordUpdated, baseID, userID).
			WithTable(r.TableID).
			WithRecord(r.ID).
			WithPayload(r)
		s.hub.Broadcast(msg)
	}

	// Trigger automations
	if s.automationCallback != nil {
		s.automationCallback(r.TableID, &r.ID, r, &oldRecord, "record_updated", userID)
	}

	return r, nil
}

// PatchRecord merges new values into existing record values
func (s *RecordStore) PatchRecord(ctx context.Context, recordID uuid.UUID, newValues map[string]interface{}, userID uuid.UUID) (*models.Record, error) {
	// Get record to check access and get current values
	r, err := s.GetRecord(ctx, recordID, userID)
	if err != nil {
		return nil, err
	}

	// Store old record for automation comparison
	oldRecord := *r

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, r.TableID)
	if err != nil {
		return nil, err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	// Parse existing values
	var existingValues map[string]interface{}
	if err := json.Unmarshal(r.Values, &existingValues); err != nil {
		existingValues = make(map[string]interface{})
	}

	// Merge new values
	for k, v := range newValues {
		if v == nil {
			delete(existingValues, k)
		} else {
			existingValues[k] = v
		}
	}

	// Convert back to JSON
	mergedValues, err := json.Marshal(existingValues)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(ctx, `
		UPDATE records SET values = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, table_id, values, position, color, created_at, updated_at
	`, recordID, mergedValues).Scan(
		&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Broadcast record updated
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeRecordUpdated, baseID, userID).
			WithTable(r.TableID).
			WithRecord(r.ID).
			WithPayload(r)
		s.hub.Broadcast(msg)
	}

	// Trigger automations
	if s.automationCallback != nil {
		s.automationCallback(r.TableID, &r.ID, r, &oldRecord, "record_updated", userID)
	}

	return r, nil
}

// UpdateRecordColor updates only the color of a record
func (s *RecordStore) UpdateRecordColor(ctx context.Context, recordID uuid.UUID, color *string, userID uuid.UUID) (*models.Record, error) {
	// Get record to check access
	r, err := s.GetRecord(ctx, recordID, userID)
	if err != nil {
		return nil, err
	}

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, r.TableID)
	if err != nil {
		return nil, err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	err = s.db.QueryRow(ctx, `
		UPDATE records SET color = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, table_id, values, position, color, created_at, updated_at
	`, recordID, color).Scan(
		&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// DeleteRecord deletes a record
func (s *RecordStore) DeleteRecord(ctx context.Context, recordID uuid.UUID, userID uuid.UUID) error {
	// Get record to check access
	r, err := s.GetRecord(ctx, recordID, userID)
	if err != nil {
		return err
	}

	// Store record for automation
	deletedRecord := *r
	tableID := r.TableID

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, r.TableID)
	if err != nil {
		return err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return err
	}
	if !role.CanEdit() {
		return ErrForbidden
	}

	result, err := s.db.Exec(ctx, `DELETE FROM records WHERE id = $1`, recordID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Broadcast record deleted
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeRecordDeleted, baseID, userID).
			WithTable(r.TableID).
			WithRecord(recordID).
			WithPayload(map[string]interface{}{
				"id":      recordID,
				"tableId": r.TableID,
			})
		s.hub.Broadcast(msg)
	}

	// Trigger automations
	if s.automationCallback != nil {
		s.automationCallback(tableID, &recordID, &deletedRecord, nil, "record_deleted", userID)
	}

	return nil
}

// BulkCreateRecords creates multiple records at once
func (s *RecordStore) BulkCreateRecords(ctx context.Context, tableID uuid.UUID, recordValues []json.RawMessage, userID uuid.UUID) ([]models.Record, error) {
	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, tableID)
	if err != nil {
		return nil, err
	}
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Get current max position
	var maxPosition int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM records WHERE table_id = $1
	`, tableID).Scan(&maxPosition)
	if err != nil {
		return nil, err
	}

	var records []models.Record
	for i, values := range recordValues {
		if values == nil {
			values = json.RawMessage(`{}`)
		}

		var r models.Record
		err = tx.QueryRow(ctx, `
			INSERT INTO records (table_id, values, position)
			VALUES ($1, $2, $3)
			RETURNING id, table_id, values, position, color, created_at, updated_at
		`, tableID, values, maxPosition+1+i).Scan(
			&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return records, nil
}

// PatchRecordValues updates record values (used by automation engine, does not trigger automations)
func (s *RecordStore) PatchRecordValues(ctx context.Context, recordID uuid.UUID, newValues map[string]interface{}, userID uuid.UUID) (*models.Record, error) {
	// Get current record values
	var r models.Record
	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, values, position, color, created_at, updated_at
		FROM records WHERE id = $1
	`, recordID).Scan(&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Parse existing values
	var existingValues map[string]interface{}
	if err := json.Unmarshal(r.Values, &existingValues); err != nil {
		existingValues = make(map[string]interface{})
	}

	// Merge new values
	for k, v := range newValues {
		if v == nil {
			delete(existingValues, k)
		} else {
			existingValues[k] = v
		}
	}

	// Convert back to JSON
	mergedValues, err := json.Marshal(existingValues)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(ctx, `
		UPDATE records SET values = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, table_id, values, position, color, created_at, updated_at
	`, recordID, mergedValues).Scan(
		&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Broadcast record updated (but don't trigger automations to avoid loops)
	if s.hub != nil {
		baseID, _ := s.getBaseIDForTable(ctx, r.TableID)
		msg := realtime.NewMessage(realtime.MsgTypeRecordUpdated, baseID, userID).
			WithTable(r.TableID).
			WithRecord(r.ID).
			WithPayload(r)
		s.hub.Broadcast(msg)
	}

	return &r, nil
}
