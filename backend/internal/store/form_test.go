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

func TestNewFormStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewFormStore(mock, nil, nil, nil)
	assert.NotNil(t, store)
}

func TestFormStore_GetForm(t *testing.T) {
	ctx := context.Background()

	t.Run("returns form when found and user has access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewFormStore(mock, baseStore, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		// Query for form
		formRows := pgxmock.NewRows([]string{
			"id", "table_id", "name", "description", "public_token", "is_active",
			"success_message", "redirect_url", "submit_button_text", "created_by", "created_at", "updated_at",
		}).AddRow(
			formID, tableID, "Test Form", nil, "token123", true,
			"Thanks!", nil, "Submit", userID, now, now,
		)

		mock.ExpectQuery("SELECT id, table_id, name, description, public_token, is_active").
			WithArgs(formID).
			WillReturnRows(formRows)

		// Get base ID for table
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleOwner))

		// Get form fields
		mock.ExpectQuery("SELECT ff.id, ff.form_id, ff.field_id, ff.label, ff.help_text").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "form_id", "field_id", "label", "help_text",
				"is_required", "is_visible", "position",
				"name", "field_type", "options",
			}))

		form, err := store.GetForm(ctx, formID, userID)
		require.NoError(t, err)
		assert.Equal(t, formID, form.ID)
		assert.Equal(t, "Test Form", form.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when form not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)
		formID := uuid.New()

		mock.ExpectQuery("SELECT id, table_id, name, description, public_token, is_active").
			WithArgs(formID).
			WillReturnError(pgx.ErrNoRows)

		form, err := store.GetForm(ctx, formID, uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, form)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFormStore_ListFormsForTable(t *testing.T) {
	ctx := context.Background()

	t.Run("returns forms for table", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewFormStore(mock, baseStore, nil, nil)

		tableID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		formID := uuid.New()
		now := time.Now().UTC()

		// Get base ID for table
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		// GetUserRole check
		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		// List forms - use string pointer for public_token since it's a *string
		token := "token123"
		successMsg := "Thanks!"
		formRows := pgxmock.NewRows([]string{
			"id", "table_id", "name", "description", "public_token", "is_active",
			"success_message", "redirect_url", "submit_button_text", "created_by", "created_at", "updated_at",
		}).AddRow(
			formID, tableID, "Test Form", nil, &token, true,
			successMsg, nil, "Submit", userID, now, now,
		)

		mock.ExpectQuery("SELECT id, table_id, name, description, public_token, is_active").
			WithArgs(tableID).
			WillReturnRows(formRows)

		forms, err := store.ListFormsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.Len(t, forms, 1)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no forms", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewFormStore(mock, baseStore, nil, nil)

		tableID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()

		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		mock.ExpectQuery("SELECT id, table_id, name, description, public_token, is_active").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "table_id", "name", "description", "public_token", "is_active",
				"success_message", "redirect_url", "submit_button_text", "created_by", "created_at", "updated_at",
			}))

		forms, err := store.ListFormsForTable(ctx, tableID, userID)
		require.NoError(t, err)
		assert.NotNil(t, forms)
		assert.Empty(t, forms)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when table not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)
		tableID := uuid.New()

		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnError(pgx.ErrNoRows)

		forms, err := store.ListFormsForTable(ctx, tableID, uuid.New())
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, forms)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFormStore_DeleteForm(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes form when user has edit access", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewFormStore(mock, baseStore, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		// GetForm query
		mock.ExpectQuery("SELECT id, table_id, name, description, public_token, is_active").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "table_id", "name", "description", "public_token", "is_active",
				"success_message", "redirect_url", "submit_button_text", "created_by", "created_at", "updated_at",
			}).AddRow(formID, tableID, "Test Form", nil, "token", true, "Thanks!", nil, "Submit", userID, now, now))

		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Get form fields (in GetForm)
		mock.ExpectQuery("SELECT ff.id, ff.form_id, ff.field_id, ff.label, ff.help_text").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "form_id", "field_id", "label", "help_text",
				"is_required", "is_visible", "position",
				"name", "field_type", "options",
			}))

		// Delete checks - getBaseIDForTable again
		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleEditor))

		// Delete
		mock.ExpectExec("DELETE FROM forms WHERE id").
			WithArgs(formID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteForm(ctx, formID, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrForbidden when viewer tries to delete", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewFormStore(mock, baseStore, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()

		mock.ExpectQuery("SELECT id, table_id, name, description, public_token, is_active").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "table_id", "name", "description", "public_token", "is_active",
				"success_message", "redirect_url", "submit_button_text", "created_by", "created_at", "updated_at",
			}).AddRow(formID, tableID, "Test Form", nil, "token", true, "Thanks!", nil, "Submit", userID, now, now))

		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		mock.ExpectQuery("SELECT ff.id, ff.form_id, ff.field_id, ff.label, ff.help_text").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "form_id", "field_id", "label", "help_text",
				"is_required", "is_visible", "position",
				"name", "field_type", "options",
			}))

		mock.ExpectQuery("SELECT base_id FROM tables WHERE id").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"base_id"}).AddRow(baseID))

		mock.ExpectQuery("SELECT role FROM base_collaborators").
			WithArgs(baseID, userID).
			WillReturnRows(pgxmock.NewRows([]string{"role"}).AddRow(models.RoleViewer))

		err = store.DeleteForm(ctx, formID, userID)
		assert.ErrorIs(t, err, ErrForbidden)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFormStore_GetPublicForm(t *testing.T) {
	ctx := context.Background()

	t.Run("returns public form when active", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()
		token := "public-token-123"

		// Query for form
		mock.ExpectQuery("SELECT id, table_id, name, description, is_active").
			WithArgs(token).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "table_id", "name", "description", "is_active",
				"success_message", "redirect_url", "submit_button_text",
			}).AddRow(formID, tableID, "Test Form", nil, true, "Thanks!", nil, "Submit"))

		// Get visible form fields
		mock.ExpectQuery("SELECT ff.field_id, COALESCE").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{
				"field_id", "label", "help_text", "is_required", "field_type", "options", "position",
			}))

		publicForm, err := store.GetPublicForm(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, formID, publicForm.ID)
		assert.Equal(t, "Test Form", publicForm.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when form not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)

		mock.ExpectQuery("SELECT id, table_id, name, description, is_active").
			WithArgs("invalid-token").
			WillReturnError(pgx.ErrNoRows)

		publicForm, err := store.GetPublicForm(ctx, "invalid-token")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, publicForm)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when form is inactive", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()

		mock.ExpectQuery("SELECT id, table_id, name, description, is_active").
			WithArgs("inactive-token").
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "table_id", "name", "description", "is_active",
				"success_message", "redirect_url", "submit_button_text",
			}).AddRow(formID, tableID, "Test Form", nil, false, "Thanks!", nil, "Submit"))

		publicForm, err := store.GetPublicForm(ctx, "inactive-token")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, publicForm)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFormStore_SubmitPublicForm(t *testing.T) {
	ctx := context.Background()

	t.Run("creates record from form submission", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		fieldID := uuid.New()
		token := "submit-token"
		now := time.Now().UTC()

		values := map[string]interface{}{
			fieldID.String(): "test value",
		}

		// Query for form
		mock.ExpectQuery("SELECT id, table_id, is_active FROM forms WHERE public_token").
			WithArgs(token).
			WillReturnRows(pgxmock.NewRows([]string{"id", "table_id", "is_active"}).
				AddRow(formID, tableID, true))

		// Validate required fields
		mock.ExpectQuery("SELECT field_id, is_required FROM form_fields").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{"field_id", "is_required"}).
				AddRow(fieldID, false))

		// Get max position
		mock.ExpectQuery("SELECT COALESCE").
			WithArgs(tableID).
			WillReturnRows(pgxmock.NewRows([]string{"max"}).AddRow(0))

		// Insert record
		valuesJSON, _ := json.Marshal(values)
		mock.ExpectQuery("INSERT INTO records").
			WithArgs(tableID, valuesJSON, 1).
			WillReturnRows(pgxmock.NewRows([]string{
				"id", "table_id", "values", "position", "color", "created_at", "updated_at",
			}).AddRow(recordID, tableID, valuesJSON, 1, nil, now, now))

		record, err := store.SubmitPublicForm(ctx, token, values)
		require.NoError(t, err)
		assert.Equal(t, recordID, record.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when form is inactive", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()

		mock.ExpectQuery("SELECT id, table_id, is_active FROM forms WHERE public_token").
			WithArgs("inactive-token").
			WillReturnRows(pgxmock.NewRows([]string{"id", "table_id", "is_active"}).
				AddRow(formID, tableID, false))

		record, err := store.SubmitPublicForm(ctx, "inactive-token", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "form is not active")
		assert.Nil(t, record)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when required field is missing", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewFormStore(mock, nil, nil, nil)

		formID := uuid.New()
		tableID := uuid.New()
		requiredFieldID := uuid.New()

		mock.ExpectQuery("SELECT id, table_id, is_active FROM forms WHERE public_token").
			WithArgs("token").
			WillReturnRows(pgxmock.NewRows([]string{"id", "table_id", "is_active"}).
				AddRow(formID, tableID, true))

		mock.ExpectQuery("SELECT field_id, is_required FROM form_fields").
			WithArgs(formID).
			WillReturnRows(pgxmock.NewRows([]string{"field_id", "is_required"}).
				AddRow(requiredFieldID, true))

		// Submit without required field
		record, err := store.SubmitPublicForm(ctx, "token", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required field missing")
		assert.Nil(t, record)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestFormStore_GenerateToken(t *testing.T) {
	token1 := generateToken()
	token2 := generateToken()

	assert.Len(t, token1, 32)
	assert.Len(t, token2, 32)
	assert.NotEqual(t, token1, token2)
}
