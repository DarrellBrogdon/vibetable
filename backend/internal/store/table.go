package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/realtime"
)

type TableStore struct {
	db        DBTX
	baseStore *BaseStore
	hub       *realtime.Hub
}

func NewTableStore(db DBTX, baseStore *BaseStore) *TableStore {
	return &TableStore{db: db, baseStore: baseStore}
}

// SetHub sets the realtime hub for broadcasting changes
func (s *TableStore) SetHub(hub *realtime.Hub) {
	s.hub = hub
}

// ListTablesForBase returns all tables in a base
func (s *TableStore) ListTablesForBase(ctx context.Context, baseID uuid.UUID, userID uuid.UUID) ([]models.Table, error) {
	// Verify user has access to base
	_, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, base_id, name, position, created_at, updated_at
		FROM tables
		WHERE base_id = $1
		ORDER BY position, created_at
	`, baseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []models.Table
	for rows.Next() {
		var t models.Table
		if err := rows.Scan(&t.ID, &t.BaseID, &t.Name, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}

	if tables == nil {
		tables = []models.Table{}
	}

	return tables, rows.Err()
}

// CreateTable creates a new table in a base
func (s *TableStore) CreateTable(ctx context.Context, baseID uuid.UUID, name string, userID uuid.UUID) (*models.Table, error) {
	// Verify user has edit access
	role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	// Get next position
	var maxPosition int
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM tables WHERE base_id = $1
	`, baseID).Scan(&maxPosition)
	if err != nil {
		return nil, err
	}

	var t models.Table
	err = s.db.QueryRow(ctx, `
		INSERT INTO tables (base_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, base_id, name, position, created_at, updated_at
	`, baseID, name, maxPosition+1).Scan(&t.ID, &t.BaseID, &t.Name, &t.Position, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Broadcast table created
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeTableCreated, baseID, userID).
			WithTable(t.ID).
			WithPayload(t)
		s.hub.Broadcast(msg)
	}

	return &t, nil
}

// GetTable returns a table by ID
func (s *TableStore) GetTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID) (*models.Table, error) {
	var t models.Table
	err := s.db.QueryRow(ctx, `
		SELECT id, base_id, name, position, created_at, updated_at
		FROM tables WHERE id = $1
	`, tableID).Scan(&t.ID, &t.BaseID, &t.Name, &t.Position, &t.CreatedAt, &t.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Verify user has access to the base
	_, err = s.baseStore.GetUserRole(ctx, t.BaseID, userID)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// UpdateTable updates a table's name
func (s *TableStore) UpdateTable(ctx context.Context, tableID uuid.UUID, name string, userID uuid.UUID) (*models.Table, error) {
	// Get table to check base access
	t, err := s.GetTable(ctx, tableID, userID)
	if err != nil {
		return nil, err
	}

	// Verify user has edit access
	role, err := s.baseStore.GetUserRole(ctx, t.BaseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	err = s.db.QueryRow(ctx, `
		UPDATE tables SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, base_id, name, position, created_at, updated_at
	`, tableID, name).Scan(&t.ID, &t.BaseID, &t.Name, &t.Position, &t.CreatedAt, &t.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Broadcast table updated
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeTableUpdated, t.BaseID, userID).
			WithTable(t.ID).
			WithPayload(t)
		s.hub.Broadcast(msg)
	}

	return t, nil
}

// DeleteTable deletes a table
func (s *TableStore) DeleteTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID) error {
	// Get table to check base access
	t, err := s.GetTable(ctx, tableID, userID)
	if err != nil {
		return err
	}

	// Verify user has edit access
	role, err := s.baseStore.GetUserRole(ctx, t.BaseID, userID)
	if err != nil {
		return err
	}
	if !role.CanEdit() {
		return ErrForbidden
	}

	result, err := s.db.Exec(ctx, `DELETE FROM tables WHERE id = $1`, tableID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	// Broadcast table deleted
	if s.hub != nil {
		msg := realtime.NewMessage(realtime.MsgTypeTableDeleted, t.BaseID, userID).
			WithTable(tableID).
			WithPayload(map[string]interface{}{
				"id":     tableID,
				"baseId": t.BaseID,
			})
		s.hub.Broadcast(msg)
	}

	return nil
}

// ReorderTables updates the position of tables
func (s *TableStore) ReorderTables(ctx context.Context, baseID uuid.UUID, tableIDs []uuid.UUID, userID uuid.UUID) error {
	// Verify user has edit access
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

	for i, id := range tableIDs {
		_, err := tx.Exec(ctx, `
			UPDATE tables SET position = $1, updated_at = NOW()
			WHERE id = $2 AND base_id = $3
		`, i, id, baseID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// DuplicateTable duplicates a table with all its fields, views, and optionally records
func (s *TableStore) DuplicateTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID, includeRecords bool) (*models.Table, error) {
	// Get original table
	origTable, err := s.GetTable(ctx, tableID, userID)
	if err != nil {
		return nil, err
	}

	// Verify user has edit access
	role, err := s.baseStore.GetUserRole(ctx, origTable.BaseID, userID)
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

	// Get next position
	var maxPosition int
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM tables WHERE base_id = $1
	`, origTable.BaseID).Scan(&maxPosition)
	if err != nil {
		return nil, err
	}

	// Create new table
	newName := fmt.Sprintf("%s (Copy)", origTable.Name)
	var newTable models.Table
	err = tx.QueryRow(ctx, `
		INSERT INTO tables (base_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, base_id, name, position, created_at, updated_at
	`, origTable.BaseID, newName, maxPosition+1).Scan(
		&newTable.ID, &newTable.BaseID, &newTable.Name, &newTable.Position, &newTable.CreatedAt, &newTable.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Copy fields and build ID mapping
	fieldIDMap := make(map[uuid.UUID]uuid.UUID) // old ID -> new ID
	rows, err := tx.Query(ctx, `
		SELECT id, name, field_type, options, position
		FROM fields
		WHERE table_id = $1
		ORDER BY position
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var oldID uuid.UUID
		var name string
		var fieldType models.FieldType
		var options json.RawMessage
		var position int

		if err := rows.Scan(&oldID, &name, &fieldType, &options, &position); err != nil {
			return nil, err
		}

		var newFieldID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO fields (table_id, name, field_type, options, position)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id
		`, newTable.ID, name, fieldType, options, position).Scan(&newFieldID)
		if err != nil {
			return nil, err
		}

		fieldIDMap[oldID] = newFieldID
	}
	rows.Close()

	// Copy views
	viewRows, err := tx.Query(ctx, `
		SELECT name, view_type, config, position
		FROM views
		WHERE table_id = $1
		ORDER BY position
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer viewRows.Close()

	for viewRows.Next() {
		var name string
		var viewType models.ViewType
		var config json.RawMessage
		var position int

		if err := viewRows.Scan(&name, &viewType, &config, &position); err != nil {
			return nil, err
		}

		// Update field IDs in config
		var configMap map[string]interface{}
		if err := json.Unmarshal(config, &configMap); err == nil {
			// Update filter field IDs
			if filters, ok := configMap["filters"].([]interface{}); ok {
				for _, f := range filters {
					if filter, ok := f.(map[string]interface{}); ok {
						if oldFieldID, ok := filter["field_id"].(string); ok {
							oldUUID, _ := uuid.Parse(oldFieldID)
							if newUUID, exists := fieldIDMap[oldUUID]; exists {
								filter["field_id"] = newUUID.String()
							}
						}
					}
				}
			}
			// Update sort field IDs
			if sorts, ok := configMap["sorts"].([]interface{}); ok {
				for _, s := range sorts {
					if sort, ok := s.(map[string]interface{}); ok {
						if oldFieldID, ok := sort["field_id"].(string); ok {
							oldUUID, _ := uuid.Parse(oldFieldID)
							if newUUID, exists := fieldIDMap[oldUUID]; exists {
								sort["field_id"] = newUUID.String()
							}
						}
					}
				}
			}
			config, _ = json.Marshal(configMap)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO views (table_id, name, view_type, config, position)
			VALUES ($1, $2, $3, $4, $5)
		`, newTable.ID, name, viewType, config, position)
		if err != nil {
			return nil, err
		}
	}
	viewRows.Close()

	// Copy records if requested
	if includeRecords {
		recordRows, err := tx.Query(ctx, `
			SELECT values, position
			FROM records
			WHERE table_id = $1
			ORDER BY position
		`, tableID)
		if err != nil {
			return nil, err
		}
		defer recordRows.Close()

		for recordRows.Next() {
			var values json.RawMessage
			var position int

			if err := recordRows.Scan(&values, &position); err != nil {
				return nil, err
			}

			// Remap field IDs in values
			var valuesMap map[string]interface{}
			if err := json.Unmarshal(values, &valuesMap); err == nil {
				newValuesMap := make(map[string]interface{})
				for oldFieldIDStr, val := range valuesMap {
					oldUUID, err := uuid.Parse(oldFieldIDStr)
					if err != nil {
						continue
					}
					if newUUID, exists := fieldIDMap[oldUUID]; exists {
						newValuesMap[newUUID.String()] = val
					}
				}
				values, _ = json.Marshal(newValuesMap)
			}

			_, err = tx.Exec(ctx, `
				INSERT INTO records (table_id, values, position)
				VALUES ($1, $2, $3)
			`, newTable.ID, values, position)
			if err != nil {
				return nil, err
			}
		}
		recordRows.Close()
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &newTable, nil
}
