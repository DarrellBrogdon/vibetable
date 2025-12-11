package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
)

var (
	ErrForbidden = errors.New("forbidden")
)

type BaseStore struct {
	db DBTX
}

func NewBaseStore(db DBTX) *BaseStore {
	return &BaseStore{db: db}
}

// ListBasesForUser returns all bases the user has access to
func (s *BaseStore) ListBasesForUser(ctx context.Context, userID uuid.UUID) ([]models.Base, error) {
	rows, err := s.db.Query(ctx, `
		SELECT b.id, b.name, b.created_by, b.created_at, b.updated_at, bc.role
		FROM bases b
		JOIN base_collaborators bc ON b.id = bc.base_id
		WHERE bc.user_id = $1
		ORDER BY b.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bases []models.Base
	for rows.Next() {
		var base models.Base
		var role models.CollaboratorRole
		if err := rows.Scan(&base.ID, &base.Name, &base.CreatedBy, &base.CreatedAt, &base.UpdatedAt, &role); err != nil {
			return nil, err
		}
		base.Role = &role
		bases = append(bases, base)
	}

	if bases == nil {
		bases = []models.Base{} // Return empty array, not null
	}

	return bases, rows.Err()
}

// CreateBase creates a new base and adds the creator as owner
func (s *BaseStore) CreateBase(ctx context.Context, name string, userID uuid.UUID) (*models.Base, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create the base
	var base models.Base
	err = tx.QueryRow(ctx, `
		INSERT INTO bases (name, created_by)
		VALUES ($1, $2)
		RETURNING id, name, created_by, created_at, updated_at
	`, name, userID).Scan(&base.ID, &base.Name, &base.CreatedBy, &base.CreatedAt, &base.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Add creator as owner
	_, err = tx.Exec(ctx, `
		INSERT INTO base_collaborators (base_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, base.ID, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	role := models.RoleOwner
	base.Role = &role
	return &base, nil
}

// GetBase returns a base by ID, checking user access
func (s *BaseStore) GetBase(ctx context.Context, baseID uuid.UUID, userID uuid.UUID) (*models.Base, error) {
	var base models.Base
	var role models.CollaboratorRole

	err := s.db.QueryRow(ctx, `
		SELECT b.id, b.name, b.created_by, b.created_at, b.updated_at, bc.role
		FROM bases b
		JOIN base_collaborators bc ON b.id = bc.base_id
		WHERE b.id = $1 AND bc.user_id = $2
	`, baseID, userID).Scan(&base.ID, &base.Name, &base.CreatedBy, &base.CreatedAt, &base.UpdatedAt, &role)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	base.Role = &role
	return &base, nil
}

// GetUserRole returns the user's role for a base (or error if no access)
func (s *BaseStore) GetUserRole(ctx context.Context, baseID uuid.UUID, userID uuid.UUID) (models.CollaboratorRole, error) {
	var role models.CollaboratorRole
	err := s.db.QueryRow(ctx, `
		SELECT role FROM base_collaborators
		WHERE base_id = $1 AND user_id = $2
	`, baseID, userID).Scan(&role)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return role, nil
}

// UpdateBase updates a base's name
func (s *BaseStore) UpdateBase(ctx context.Context, baseID uuid.UUID, name string, userID uuid.UUID) (*models.Base, error) {
	// Check user has edit permission
	role, err := s.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	var base models.Base
	err = s.db.QueryRow(ctx, `
		UPDATE bases SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, created_by, created_at, updated_at
	`, baseID, name).Scan(&base.ID, &base.Name, &base.CreatedBy, &base.CreatedAt, &base.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	base.Role = &role
	return &base, nil
}

// DeleteBase deletes a base (owner only)
func (s *BaseStore) DeleteBase(ctx context.Context, baseID uuid.UUID, userID uuid.UUID) error {
	// Check user is owner
	role, err := s.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return err
	}
	if !role.CanDelete() {
		return ErrForbidden
	}

	result, err := s.db.Exec(ctx, `DELETE FROM bases WHERE id = $1`, baseID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DuplicateBase duplicates a base with all tables, fields, views, and optionally records
func (s *BaseStore) DuplicateBase(ctx context.Context, baseID uuid.UUID, userID uuid.UUID, includeRecords bool) (*models.Base, error) {
	// Check user has access to base
	role, err := s.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}
	if !role.CanEdit() {
		return nil, ErrForbidden
	}

	// Get original base
	origBase, err := s.GetBase(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Create new base
	newName := fmt.Sprintf("%s (Copy)", origBase.Name)
	var newBase models.Base
	err = tx.QueryRow(ctx, `
		INSERT INTO bases (name, created_by)
		VALUES ($1, $2)
		RETURNING id, name, created_by, created_at, updated_at
	`, newName, userID).Scan(&newBase.ID, &newBase.Name, &newBase.CreatedBy, &newBase.CreatedAt, &newBase.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Add creator as owner
	_, err = tx.Exec(ctx, `
		INSERT INTO base_collaborators (base_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, newBase.ID, userID)
	if err != nil {
		return nil, err
	}

	// Get all tables from original base
	tableRows, err := tx.Query(ctx, `
		SELECT id, name, position
		FROM tables
		WHERE base_id = $1
		ORDER BY position
	`, baseID)
	if err != nil {
		return nil, err
	}
	defer tableRows.Close()

	// Map old table IDs to new table IDs (for linked records)
	tableIDMap := make(map[uuid.UUID]uuid.UUID)
	type tableInfo struct {
		oldID    uuid.UUID
		name     string
		position int
	}
	var originalTables []tableInfo

	for tableRows.Next() {
		var t tableInfo
		if err := tableRows.Scan(&t.oldID, &t.name, &t.position); err != nil {
			return nil, err
		}
		originalTables = append(originalTables, t)
	}
	tableRows.Close()

	// Create all tables first (so we have IDs for linked records)
	for _, origTable := range originalTables {
		var newTableID uuid.UUID
		err = tx.QueryRow(ctx, `
			INSERT INTO tables (base_id, name, position)
			VALUES ($1, $2, $3)
			RETURNING id
		`, newBase.ID, origTable.name, origTable.position).Scan(&newTableID)
		if err != nil {
			return nil, err
		}
		tableIDMap[origTable.oldID] = newTableID
	}

	// Now duplicate each table's contents
	for _, origTable := range originalTables {
		newTableID := tableIDMap[origTable.oldID]
		fieldIDMap := make(map[uuid.UUID]uuid.UUID)

		// Copy fields
		fieldRows, err := tx.Query(ctx, `
			SELECT id, name, field_type, options, position
			FROM fields
			WHERE table_id = $1
			ORDER BY position
		`, origTable.oldID)
		if err != nil {
			return nil, err
		}

		for fieldRows.Next() {
			var oldFieldID uuid.UUID
			var name string
			var fieldType models.FieldType
			var options json.RawMessage
			var position int

			if err := fieldRows.Scan(&oldFieldID, &name, &fieldType, &options, &position); err != nil {
				fieldRows.Close()
				return nil, err
			}

			// Update linked_table_id in options if it's a linked_record field
			if fieldType == models.FieldTypeLinkedRecord {
				var optMap map[string]interface{}
				if err := json.Unmarshal(options, &optMap); err == nil {
					if linkedTableIDStr, ok := optMap["linked_table_id"].(string); ok {
						oldLinkedID, _ := uuid.Parse(linkedTableIDStr)
						if newLinkedID, exists := tableIDMap[oldLinkedID]; exists {
							optMap["linked_table_id"] = newLinkedID.String()
							options, _ = json.Marshal(optMap)
						}
					}
				}
			}

			var newFieldID uuid.UUID
			err = tx.QueryRow(ctx, `
				INSERT INTO fields (table_id, name, field_type, options, position)
				VALUES ($1, $2, $3, $4, $5)
				RETURNING id
			`, newTableID, name, fieldType, options, position).Scan(&newFieldID)
			if err != nil {
				fieldRows.Close()
				return nil, err
			}

			fieldIDMap[oldFieldID] = newFieldID
		}
		fieldRows.Close()

		// Copy views
		viewRows, err := tx.Query(ctx, `
			SELECT name, view_type, config, position
			FROM views
			WHERE table_id = $1
			ORDER BY position
		`, origTable.oldID)
		if err != nil {
			return nil, err
		}

		for viewRows.Next() {
			var name string
			var viewType models.ViewType
			var config json.RawMessage
			var position int

			if err := viewRows.Scan(&name, &viewType, &config, &position); err != nil {
				viewRows.Close()
				return nil, err
			}

			// Update field IDs in config
			var configMap map[string]interface{}
			if err := json.Unmarshal(config, &configMap); err == nil {
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
			`, newTableID, name, viewType, config, position)
			if err != nil {
				viewRows.Close()
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
			`, origTable.oldID)
			if err != nil {
				return nil, err
			}

			for recordRows.Next() {
				var values json.RawMessage
				var position int

				if err := recordRows.Scan(&values, &position); err != nil {
					recordRows.Close()
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
				`, newTableID, values, position)
				if err != nil {
					recordRows.Close()
					return nil, err
				}
			}
			recordRows.Close()
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	newRole := models.RoleOwner
	newBase.Role = &newRole
	return &newBase, nil
}

// --- Collaborator operations ---

// ListCollaborators returns all collaborators for a base
func (s *BaseStore) ListCollaborators(ctx context.Context, baseID uuid.UUID, userID uuid.UUID) ([]models.BaseCollaborator, error) {
	// Verify user has access
	_, err := s.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT bc.id, bc.base_id, bc.user_id, bc.role, bc.created_at,
		       u.id, u.email, u.name, u.created_at, u.updated_at
		FROM base_collaborators bc
		JOIN users u ON bc.user_id = u.id
		WHERE bc.base_id = $1
		ORDER BY bc.created_at
	`, baseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var collaborators []models.BaseCollaborator
	for rows.Next() {
		var collab models.BaseCollaborator
		var user models.User
		if err := rows.Scan(
			&collab.ID, &collab.BaseID, &collab.UserID, &collab.Role, &collab.CreatedAt,
			&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		collab.User = &user
		collaborators = append(collaborators, collab)
	}

	if collaborators == nil {
		collaborators = []models.BaseCollaborator{}
	}

	return collaborators, rows.Err()
}

// AddCollaborator adds a user as a collaborator (by email)
func (s *BaseStore) AddCollaborator(ctx context.Context, baseID uuid.UUID, email string, role models.CollaboratorRole, requestingUserID uuid.UUID) (*models.BaseCollaborator, error) {
	// Verify requester is owner
	requesterRole, err := s.GetUserRole(ctx, baseID, requestingUserID)
	if err != nil {
		return nil, err
	}
	if !requesterRole.CanManageCollaborators() {
		return nil, ErrForbidden
	}

	// Can't add someone as owner
	if role == models.RoleOwner {
		return nil, errors.New("cannot add someone as owner")
	}

	// Find user by email
	var targetUserID uuid.UUID
	err = s.db.QueryRow(ctx, `SELECT id FROM users WHERE email = $1`, email).Scan(&targetUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	// Add collaborator (or update if exists)
	var collab models.BaseCollaborator
	err = s.db.QueryRow(ctx, `
		INSERT INTO base_collaborators (base_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (base_id, user_id) DO UPDATE SET role = $3
		RETURNING id, base_id, user_id, role, created_at
	`, baseID, targetUserID, role).Scan(&collab.ID, &collab.BaseID, &collab.UserID, &collab.Role, &collab.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &collab, nil
}

// UpdateCollaboratorRole changes a collaborator's role
func (s *BaseStore) UpdateCollaboratorRole(ctx context.Context, baseID uuid.UUID, targetUserID uuid.UUID, newRole models.CollaboratorRole, requestingUserID uuid.UUID) (*models.BaseCollaborator, error) {
	// Verify requester is owner
	requesterRole, err := s.GetUserRole(ctx, baseID, requestingUserID)
	if err != nil {
		return nil, err
	}
	if !requesterRole.CanManageCollaborators() {
		return nil, ErrForbidden
	}

	// Can't change to/from owner
	if newRole == models.RoleOwner {
		return nil, errors.New("cannot change role to owner")
	}

	// Check target's current role
	targetRole, err := s.GetUserRole(ctx, baseID, targetUserID)
	if err != nil {
		return nil, err
	}
	if targetRole == models.RoleOwner {
		return nil, errors.New("cannot change owner's role")
	}

	var collab models.BaseCollaborator
	err = s.db.QueryRow(ctx, `
		UPDATE base_collaborators SET role = $3
		WHERE base_id = $1 AND user_id = $2
		RETURNING id, base_id, user_id, role, created_at
	`, baseID, targetUserID, newRole).Scan(&collab.ID, &collab.BaseID, &collab.UserID, &collab.Role, &collab.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &collab, nil
}

// RemoveCollaborator removes a collaborator from a base
func (s *BaseStore) RemoveCollaborator(ctx context.Context, baseID uuid.UUID, targetUserID uuid.UUID, requestingUserID uuid.UUID) error {
	// Verify requester is owner
	requesterRole, err := s.GetUserRole(ctx, baseID, requestingUserID)
	if err != nil {
		return err
	}
	if !requesterRole.CanManageCollaborators() {
		return ErrForbidden
	}

	// Can't remove owner
	targetRole, err := s.GetUserRole(ctx, baseID, targetUserID)
	if err != nil {
		return err
	}
	if targetRole == models.RoleOwner {
		return errors.New("cannot remove owner")
	}

	result, err := s.db.Exec(ctx, `
		DELETE FROM base_collaborators
		WHERE base_id = $1 AND user_id = $2
	`, baseID, targetUserID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
