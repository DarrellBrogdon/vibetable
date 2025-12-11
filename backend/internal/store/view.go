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

type ViewStore struct {
	db         DBTX
	baseStore  *BaseStore
	tableStore *TableStore
	hub        *realtime.Hub
}

func NewViewStore(db DBTX, baseStore *BaseStore, tableStore *TableStore) *ViewStore {
	return &ViewStore{
		db:         db,
		baseStore:  baseStore,
		tableStore: tableStore,
	}
}

// SetHub sets the realtime hub for broadcasting changes
func (s *ViewStore) SetHub(hub *realtime.Hub) {
	s.hub = hub
}

// ListViewsForTable returns all views for a table
func (s *ViewStore) ListViewsForTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID) ([]models.View, error) {
	// First verify user has access to this table's base
	table, err := s.tableStore.GetTable(ctx, tableID, userID)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, ErrNotFound
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at
		FROM views
		WHERE table_id = $1
		ORDER BY position ASC, created_at ASC
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []models.View
	for rows.Next() {
		var v models.View
		if err := rows.Scan(&v.ID, &v.TableID, &v.Name, &v.Type, &v.Config, &v.Position, &v.PublicToken, &v.IsPublic, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		views = append(views, v)
	}

	if views == nil {
		views = []models.View{}
	}

	return views, rows.Err()
}

// CreateView creates a new view
func (s *ViewStore) CreateView(ctx context.Context, tableID uuid.UUID, name string, viewType models.ViewType, config json.RawMessage, userID uuid.UUID) (*models.View, error) {
	// Verify user has edit access to this table's base
	table, err := s.tableStore.GetTable(ctx, tableID, userID)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, ErrNotFound
	}

	// Check permission through base
	role, err := s.baseStore.GetUserRole(ctx, table.BaseID, userID)
	if err != nil {
		return nil, err
	}
	if role != "owner" && role != "editor" {
		return nil, ErrForbidden
	}

	// Get next position
	var maxPos int
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM views WHERE table_id = $1
	`, tableID).Scan(&maxPos)
	if err != nil {
		return nil, err
	}

	if config == nil {
		config = json.RawMessage("{}")
	}

	view := &models.View{}
	err = s.db.QueryRow(ctx, `
		INSERT INTO views (table_id, name, view_type, config, position)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at
	`, tableID, name, viewType, config, maxPos+1).Scan(
		&view.ID, &view.TableID, &view.Name, &view.Type, &view.Config,
		&view.Position, &view.PublicToken, &view.IsPublic, &view.CreatedAt, &view.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return view, nil
}

// GetView returns a view by ID
func (s *ViewStore) GetView(ctx context.Context, viewID uuid.UUID, userID uuid.UUID) (*models.View, error) {
	var view models.View
	var tableID uuid.UUID

	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at
		FROM views
		WHERE id = $1
	`, viewID).Scan(&view.ID, &tableID, &view.Name, &view.Type, &view.Config, &view.Position, &view.PublicToken, &view.IsPublic, &view.CreatedAt, &view.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	view.TableID = tableID

	// Verify user has access to this table
	table, err := s.tableStore.GetTable(ctx, tableID, userID)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, ErrNotFound
	}

	return &view, nil
}

// UpdateView updates a view's name and/or config
func (s *ViewStore) UpdateView(ctx context.Context, viewID uuid.UUID, name *string, config *json.RawMessage, userID uuid.UUID) (*models.View, error) {
	// Get current view and verify access
	view, err := s.GetView(ctx, viewID, userID)
	if err != nil {
		return nil, err
	}

	// Check edit permission
	table, err := s.tableStore.GetTable(ctx, view.TableID, userID)
	if err != nil {
		return nil, err
	}

	role, err := s.baseStore.GetUserRole(ctx, table.BaseID, userID)
	if err != nil {
		return nil, err
	}
	if role != "owner" && role != "editor" {
		return nil, ErrForbidden
	}

	// Build update query
	if name != nil {
		view.Name = *name
	}
	if config != nil {
		view.Config = *config
	}

	err = s.db.QueryRow(ctx, `
		UPDATE views
		SET name = $1, config = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at
	`, view.Name, view.Config, viewID).Scan(
		&view.ID, &view.TableID, &view.Name, &view.Type, &view.Config,
		&view.Position, &view.PublicToken, &view.IsPublic, &view.CreatedAt, &view.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return view, nil
}

// DeleteView deletes a view
func (s *ViewStore) DeleteView(ctx context.Context, viewID uuid.UUID, userID uuid.UUID) error {
	// Get view and verify access
	view, err := s.GetView(ctx, viewID, userID)
	if err != nil {
		return err
	}

	// Check edit permission
	table, err := s.tableStore.GetTable(ctx, view.TableID, userID)
	if err != nil {
		return err
	}

	role, err := s.baseStore.GetUserRole(ctx, table.BaseID, userID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "editor" {
		return ErrForbidden
	}

	_, err = s.db.Exec(ctx, `DELETE FROM views WHERE id = $1`, viewID)
	return err
}

// SetViewPublic enables or disables public sharing for a view
func (s *ViewStore) SetViewPublic(ctx context.Context, viewID uuid.UUID, isPublic bool, userID uuid.UUID) (*models.View, error) {
	// Get current view and verify access
	view, err := s.GetView(ctx, viewID, userID)
	if err != nil {
		return nil, err
	}

	// Check owner permission (only owners can share views publicly)
	table, err := s.tableStore.GetTable(ctx, view.TableID, userID)
	if err != nil {
		return nil, err
	}

	role, err := s.baseStore.GetUserRole(ctx, table.BaseID, userID)
	if err != nil {
		return nil, err
	}
	if role != "owner" {
		return nil, ErrForbidden
	}

	var publicToken *string
	if isPublic {
		// Generate a new token if enabling public access
		if view.PublicToken == nil {
			token := generateToken()
			publicToken = &token
		} else {
			publicToken = view.PublicToken
		}
	} else {
		// Clear token when disabling
		publicToken = nil
	}

	err = s.db.QueryRow(ctx, `
		UPDATE views
		SET public_token = $1, is_public = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at
	`, publicToken, isPublic, viewID).Scan(
		&view.ID, &view.TableID, &view.Name, &view.Type, &view.Config,
		&view.Position, &view.PublicToken, &view.IsPublic, &view.CreatedAt, &view.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return view, nil
}

// GetPublicView returns a view and its data by public token (no auth required)
func (s *ViewStore) GetPublicView(ctx context.Context, token string) (*models.PublicView, error) {
	var view models.View

	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, name, view_type, config, position, public_token, is_public, created_at, updated_at
		FROM views
		WHERE public_token = $1 AND is_public = true
	`, token).Scan(&view.ID, &view.TableID, &view.Name, &view.Type, &view.Config, &view.Position, &view.PublicToken, &view.IsPublic, &view.CreatedAt, &view.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// Get table info
	var table models.Table
	err = s.db.QueryRow(ctx, `
		SELECT id, base_id, name, position, created_at, updated_at
		FROM tables
		WHERE id = $1
	`, view.TableID).Scan(&table.ID, &table.BaseID, &table.Name, &table.Position, &table.CreatedAt, &table.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get fields
	fieldRows, err := s.db.Query(ctx, `
		SELECT id, table_id, name, field_type, options, position, created_at, updated_at
		FROM fields
		WHERE table_id = $1
		ORDER BY position ASC
	`, view.TableID)
	if err != nil {
		return nil, err
	}
	defer fieldRows.Close()

	var fields []*models.Field
	for fieldRows.Next() {
		var f models.Field
		if err := fieldRows.Scan(&f.ID, &f.TableID, &f.Name, &f.FieldType, &f.Options, &f.Position, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fields = append(fields, &f)
	}

	// Get records
	recordRows, err := s.db.Query(ctx, `
		SELECT id, table_id, values, position, color, created_at, updated_at
		FROM records
		WHERE table_id = $1
		ORDER BY position ASC
	`, view.TableID)
	if err != nil {
		return nil, err
	}
	defer recordRows.Close()

	var records []*models.Record
	for recordRows.Next() {
		var r models.Record
		if err := recordRows.Scan(&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, &r)
	}

	return &models.PublicView{
		View:    &view,
		Table:   &table,
		Fields:  fields,
		Records: records,
	}, nil
}
