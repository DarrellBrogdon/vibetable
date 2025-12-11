package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
)

type FormStore struct {
	db          DBTX
	baseStore   *BaseStore
	tableStore  *TableStore
	recordStore *RecordStore
}

func NewFormStore(db DBTX, baseStore *BaseStore, tableStore *TableStore, recordStore *RecordStore) *FormStore {
	return &FormStore{
		db:          db,
		baseStore:   baseStore,
		tableStore:  tableStore,
		recordStore: recordStore,
	}
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:32]
}

// getBaseIDForTable returns the base ID for a table
func (s *FormStore) getBaseIDForTable(ctx context.Context, tableID uuid.UUID) (uuid.UUID, error) {
	var baseID uuid.UUID
	err := s.db.QueryRow(ctx, `SELECT base_id FROM tables WHERE id = $1`, tableID).Scan(&baseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return baseID, err
}

// ListFormsForTable returns all forms for a table
func (s *FormStore) ListFormsForTable(ctx context.Context, tableID uuid.UUID, userID uuid.UUID) ([]models.Form, error) {
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
		SELECT id, table_id, name, description, public_token, is_active,
		       success_message, redirect_url, submit_button_text, created_by, created_at, updated_at
		FROM forms
		WHERE table_id = $1
		ORDER BY created_at DESC
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forms []models.Form
	for rows.Next() {
		var f models.Form
		if err := rows.Scan(
			&f.ID, &f.TableID, &f.Name, &f.Description, &f.PublicToken,
			&f.IsActive, &f.SuccessMessage, &f.RedirectURL, &f.SubmitButtonText,
			&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt,
		); err != nil {
			return nil, err
		}
		forms = append(forms, f)
	}

	if forms == nil {
		forms = []models.Form{}
	}

	return forms, rows.Err()
}

// CreateForm creates a new form for a table
func (s *FormStore) CreateForm(ctx context.Context, tableID uuid.UUID, name string, userID uuid.UUID) (*models.Form, error) {
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

	// Generate public token
	token := generateToken()

	var f models.Form
	err = s.db.QueryRow(ctx, `
		INSERT INTO forms (table_id, name, public_token, created_by)
		VALUES ($1, $2, $3, $4)
		RETURNING id, table_id, name, description, public_token, is_active,
		          success_message, redirect_url, submit_button_text, created_by, created_at, updated_at
	`, tableID, name, token, userID).Scan(
		&f.ID, &f.TableID, &f.Name, &f.Description, &f.PublicToken,
		&f.IsActive, &f.SuccessMessage, &f.RedirectURL, &f.SubmitButtonText,
		&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Create form fields for all table fields
	fieldRows, err := s.db.Query(ctx, `
		SELECT id, name, field_type, options, position
		FROM fields
		WHERE table_id = $1
		ORDER BY position
	`, tableID)
	if err != nil {
		return nil, err
	}
	defer fieldRows.Close()

	var formFields []models.FormField
	position := 0
	for fieldRows.Next() {
		var fieldID uuid.UUID
		var fieldName, fieldType string
		var options json.RawMessage
		var fieldPos int
		if err := fieldRows.Scan(&fieldID, &fieldName, &fieldType, &options, &fieldPos); err != nil {
			return nil, err
		}

		var ff models.FormField
		err = s.db.QueryRow(ctx, `
			INSERT INTO form_fields (form_id, field_id, label, is_required, is_visible, position)
			VALUES ($1, $2, $3, false, true, $4)
			RETURNING id, form_id, field_id, label, help_text, is_required, is_visible, position
		`, f.ID, fieldID, fieldName, position).Scan(
			&ff.ID, &ff.FormID, &ff.FieldID, &ff.Label, &ff.HelpText,
			&ff.IsRequired, &ff.IsVisible, &ff.Position,
		)
		if err != nil {
			return nil, err
		}
		ff.FieldName = fieldName
		ff.FieldType = fieldType
		formFields = append(formFields, ff)
		position++
	}

	f.Fields = formFields
	return &f, nil
}

// GetForm returns a form by ID with fields
func (s *FormStore) GetForm(ctx context.Context, formID uuid.UUID, userID uuid.UUID) (*models.Form, error) {
	var f models.Form
	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, name, description, public_token, is_active,
		       success_message, redirect_url, submit_button_text, created_by, created_at, updated_at
		FROM forms WHERE id = $1
	`, formID).Scan(
		&f.ID, &f.TableID, &f.Name, &f.Description, &f.PublicToken,
		&f.IsActive, &f.SuccessMessage, &f.RedirectURL, &f.SubmitButtonText,
		&f.CreatedBy, &f.CreatedAt, &f.UpdatedAt,
	)
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

	// Get form fields with field info
	f.Fields, err = s.getFormFields(ctx, formID)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (s *FormStore) getFormFields(ctx context.Context, formID uuid.UUID) ([]models.FormField, error) {
	rows, err := s.db.Query(ctx, `
		SELECT ff.id, ff.form_id, ff.field_id, ff.label, ff.help_text,
		       ff.is_required, ff.is_visible, ff.position,
		       f.name, f.field_type, f.options
		FROM form_fields ff
		JOIN fields f ON f.id = ff.field_id
		WHERE ff.form_id = $1
		ORDER BY ff.position
	`, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []models.FormField
	for rows.Next() {
		var ff models.FormField
		var options json.RawMessage
		if err := rows.Scan(
			&ff.ID, &ff.FormID, &ff.FieldID, &ff.Label, &ff.HelpText,
			&ff.IsRequired, &ff.IsVisible, &ff.Position,
			&ff.FieldName, &ff.FieldType, &options,
		); err != nil {
			return nil, err
		}
		if len(options) > 0 {
			json.Unmarshal(options, &ff.FieldOptions)
		}
		fields = append(fields, ff)
	}

	if fields == nil {
		fields = []models.FormField{}
	}

	return fields, rows.Err()
}

// UpdateForm updates form settings
func (s *FormStore) UpdateForm(ctx context.Context, formID uuid.UUID, userID uuid.UUID, updates map[string]interface{}) (*models.Form, error) {
	// Get form to verify access
	form, err := s.GetForm(ctx, formID, userID)
	if err != nil {
		return nil, err
	}

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, form.TableID)
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

	// Build update query dynamically
	if name, ok := updates["name"].(string); ok && name != "" {
		form.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		form.Description = &desc
	}
	if active, ok := updates["is_active"].(bool); ok {
		form.IsActive = active
	}
	if msg, ok := updates["success_message"].(string); ok {
		form.SuccessMessage = msg
	}
	if url, ok := updates["redirect_url"].(string); ok {
		if url == "" {
			form.RedirectURL = nil
		} else {
			form.RedirectURL = &url
		}
	}
	if btnText, ok := updates["submit_button_text"].(string); ok && btnText != "" {
		form.SubmitButtonText = btnText
	}

	err = s.db.QueryRow(ctx, `
		UPDATE forms
		SET name = $2, description = $3, is_active = $4, success_message = $5,
		    redirect_url = $6, submit_button_text = $7, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`, formID, form.Name, form.Description, form.IsActive, form.SuccessMessage,
		form.RedirectURL, form.SubmitButtonText).Scan(&form.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return form, nil
}

// UpdateFormFields updates the form field configuration
func (s *FormStore) UpdateFormFields(ctx context.Context, formID uuid.UUID, userID uuid.UUID, fields []models.FormField) error {
	// Get form to verify access
	form, err := s.GetForm(ctx, formID, userID)
	if err != nil {
		return err
	}

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, form.TableID)
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

	// Update each field
	for _, ff := range fields {
		_, err := s.db.Exec(ctx, `
			UPDATE form_fields
			SET label = $3, help_text = $4, is_required = $5, is_visible = $6, position = $7
			WHERE form_id = $1 AND field_id = $2
		`, formID, ff.FieldID, ff.Label, ff.HelpText, ff.IsRequired, ff.IsVisible, ff.Position)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteForm deletes a form
func (s *FormStore) DeleteForm(ctx context.Context, formID uuid.UUID, userID uuid.UUID) error {
	// Get form to verify access
	form, err := s.GetForm(ctx, formID, userID)
	if err != nil {
		return err
	}

	// Verify user has edit access
	baseID, err := s.getBaseIDForTable(ctx, form.TableID)
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

	result, err := s.db.Exec(ctx, `DELETE FROM forms WHERE id = $1`, formID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetPublicForm returns a form by public token (no auth required)
func (s *FormStore) GetPublicForm(ctx context.Context, token string) (*models.PublicForm, error) {
	var f models.Form
	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, name, description, is_active, success_message, redirect_url, submit_button_text
		FROM forms
		WHERE public_token = $1
	`, token).Scan(
		&f.ID, &f.TableID, &f.Name, &f.Description, &f.IsActive,
		&f.SuccessMessage, &f.RedirectURL, &f.SubmitButtonText,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if !f.IsActive {
		return nil, ErrNotFound
	}

	// Get visible form fields
	rows, err := s.db.Query(ctx, `
		SELECT ff.field_id, COALESCE(ff.label, f.name), ff.help_text,
		       ff.is_required, f.field_type, f.options, ff.position
		FROM form_fields ff
		JOIN fields f ON f.id = ff.field_id
		WHERE ff.form_id = $1 AND ff.is_visible = true
		ORDER BY ff.position
	`, f.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []models.PublicFormField
	for rows.Next() {
		var pf models.PublicFormField
		var options json.RawMessage
		if err := rows.Scan(
			&pf.FieldID, &pf.Label, &pf.HelpText,
			&pf.IsRequired, &pf.FieldType, &options, &pf.Position,
		); err != nil {
			return nil, err
		}
		if len(options) > 0 {
			json.Unmarshal(options, &pf.FieldOptions)
		}
		fields = append(fields, pf)
	}

	return &models.PublicForm{
		ID:               f.ID,
		Name:             f.Name,
		Description:      f.Description,
		SuccessMessage:   f.SuccessMessage,
		RedirectURL:      f.RedirectURL,
		SubmitButtonText: f.SubmitButtonText,
		Fields:           fields,
	}, nil
}

// SubmitPublicForm creates a record from a public form submission
func (s *FormStore) SubmitPublicForm(ctx context.Context, token string, values map[string]interface{}) (*models.Record, error) {
	// Get form
	var formID, tableID uuid.UUID
	var isActive bool
	err := s.db.QueryRow(ctx, `
		SELECT id, table_id, is_active FROM forms WHERE public_token = $1
	`, token).Scan(&formID, &tableID, &isActive)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if !isActive {
		return nil, errors.New("form is not active")
	}

	// Validate required fields
	rows, err := s.db.Query(ctx, `
		SELECT field_id, is_required FROM form_fields
		WHERE form_id = $1 AND is_visible = true
	`, formID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var fieldID uuid.UUID
		var isRequired bool
		if err := rows.Scan(&fieldID, &isRequired); err != nil {
			return nil, err
		}
		if isRequired {
			val, exists := values[fieldID.String()]
			if !exists || val == nil || val == "" {
				return nil, errors.New("required field missing: " + fieldID.String())
			}
		}
	}

	// Create the record (we need to bypass normal auth for public submissions)
	valuesJSON, err := json.Marshal(values)
	if err != nil {
		return nil, err
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
	`, tableID, valuesJSON, maxPosition+1).Scan(
		&r.ID, &r.TableID, &r.Values, &r.Position, &r.Color, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
