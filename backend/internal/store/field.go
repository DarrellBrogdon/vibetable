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

type FieldStore struct {
	db         DBTX
	baseStore  *BaseStore
	tableStore *TableStore
	hub        *realtime.Hub
}

func NewFieldStore(db DBTX, baseStore *BaseStore, tableStore *TableStore) *FieldStore {
	return &FieldStore{db: db, baseStore: baseStore, tableStore: tableStore}
}

// SetHub sets the realtime hub for broadcasting changes
func (s *FieldStore) SetHub(hub *realtime.Hub) {
	s.hub = hub
}

// getBaseIDForTable returns the base ID for a table
func (s *FieldStore) getBaseIDForTable(ctx context.Context, tableID uuid.UUID) (uuid.UUID, error) {
	var baseID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT base_id FROM tables WHERE id = $1`, tableID).Scan(&baseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return baseID, err
}

// ListFieldsForTable returns all fields in a table
func (s *FieldStore) ListFieldsForTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID) ([]models.Field, error) {
	// Verify user has access via table -> base
	baseID, err := s.getBaseIDForTable(ctx, tableID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

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

	if fields == nil {
		fields = []models.Field{}
	}

	return fields, rows.Err()
}

// CreateField creates a new field in a table
func (s *FieldStore) CreateField(ctx context.Context, tableID uuid.UUID, name string, fieldType models.FieldType, options json.RawMessage, userID uuid.UUID) (*models.Field, error) {
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

	// Default empty options if not provided
	if options == nil {
		options = json.RawMessage(`{}`)
	}

	// Get next position
	var maxPosition int
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM fields WHERE table_id = $1
	`, tableID).Scan(&maxPosition)
	if err != nil {
		return nil, err
	}

	var f models.Field
	err = s.db.QueryRow(ctx, `
		INSERT INTO fields (table_id, name, field_type, options, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, table_id, name, field_type, options, position, created_at, updated_at
	`, tableID, name, fieldType, options, maxPosition+1).Scan(
		&f.ID, &f.TableID, &f.Name, &f.FieldType, &f.Options, &f.Position, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Broadcast field created
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeFieldCreated, baseID, userID).
			WithTable(tableID).
			WithField(f.ID).
			WithPayload(f)
		s.hub.Broadcast(msg)
	}

	return &f, nil
}

// GetField returns a field by ID
func (s *FieldStore) GetField(ctx context.Context, fieldID uuid.UUID, userID uuid.UUID) (*models.Field, error) {
	var f models.Field
	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, name, field_type, options, position, created_at, updated_at
		FROM fields WHERE id = $1
	`, fieldID).Scan(&f.ID, &f.TableID, &f.Name, &f.FieldType, &f.Options, &f.Position, &f.CreatedAt, &f.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Verify user has access
	baseID, err := s.getBaseIDForTable(ctx, f.TableID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// UpdateField updates a field's name and/or options
func (s *FieldStore) UpdateField(ctx context.Context, fieldID uuid.UUID, name *string, options *json.RawMessage, userID uuid.UUID) (*models.Field, error) {
	// Get field to check access
	f, err := s.GetField(ctx, fieldID, userID)
	if err != nil {
		return nil, err
	}

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, f.TableID)
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

	// Build update query
	if name != nil {
		f.Name = *name
	}
	if options != nil {
		f.Options = *options
	}

	err = s.db.QueryRow(ctx, `
		UPDATE fields SET name = $2, options = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, table_id, name, field_type, options, position, created_at, updated_at
	`, fieldID, f.Name, f.Options).Scan(
		&f.ID, &f.TableID, &f.Name, &f.FieldType, &f.Options, &f.Position, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Broadcast field updated
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeFieldUpdated, baseID, userID).
			WithTable(f.TableID).
			WithField(f.ID).
			WithPayload(f)
		s.hub.Broadcast(msg)
	}

	return f, nil
}

// DeleteField deletes a field
func (s *FieldStore) DeleteField(ctx context.Context, fieldID uuid.UUID, userID uuid.UUID) error {
	// Get field to check access
	f, err := s.GetField(ctx, fieldID, userID)
	if err != nil {
		return err
	}

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, f.TableID)
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

	result, err := s.db.Exec(ctx, `DELETE FROM fields WHERE id = $1`, fieldID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Broadcast field deleted
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeFieldDeleted, baseID, userID).
			WithTable(f.TableID).
			WithField(fieldID).
			WithPayload(map[string]interface{}{
				"id":      fieldID,
				"tableId": f.TableID,
			})
		s.hub.Broadcast(msg)
	}

	return nil
}

// ReorderFields updates the position of fields
func (s *FieldStore) ReorderFields(ctx context.Context, tableID uuid.UUID, fieldIDs []uuid.UUID, userID uuid.UUID) error {
	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, tableID)
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

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, id := range fieldIDs {
		_, err := tx.Exec(ctx, `
			UPDATE fields SET position = $1, updated_at = NOW()
			WHERE id = $2 AND table_id = $3
		`, i, id, tableID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
