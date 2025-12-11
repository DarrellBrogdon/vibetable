package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/vibetable/backend/internal/models"
)

// Test fixtures
func newTestUUID() uuid.UUID {
	return uuid.New()
}

func newTestTime() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func mustJSON(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// Mock row helpers for common patterns
func mockUserRow(mock pgxmock.PgxPoolIface, user *models.User) {
	rows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
		AddRow(user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
}

func mockBaseRow(mock pgxmock.PgxPoolIface, base *models.Base, role string) {
	rows := pgxmock.NewRows([]string{"id", "name", "created_by", "created_at", "updated_at", "role"}).
		AddRow(base.ID, base.Name, base.CreatedBy, base.CreatedAt, base.UpdatedAt, role)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
}

func mockTableRow(mock pgxmock.PgxPoolIface, table *models.Table) {
	rows := pgxmock.NewRows([]string{"id", "base_id", "name", "position", "created_at", "updated_at"}).
		AddRow(table.ID, table.BaseID, table.Name, table.Position, table.CreatedAt, table.UpdatedAt)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
}

func mockFieldRow(mock pgxmock.PgxPoolIface, field *models.Field) {
	rows := pgxmock.NewRows([]string{"id", "table_id", "name", "field_type", "options", "position", "created_at", "updated_at"}).
		AddRow(field.ID, field.TableID, field.Name, field.FieldType, field.Options, field.Position, field.CreatedAt, field.UpdatedAt)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
}

func mockRecordRow(mock pgxmock.PgxPoolIface, record *models.Record) {
	rows := pgxmock.NewRows([]string{"id", "table_id", "values", "position", "created_at", "updated_at"}).
		AddRow(record.ID, record.TableID, record.Values, record.Position, record.CreatedAt, record.UpdatedAt)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
}

func mockViewRow(mock pgxmock.PgxPoolIface, view *models.View) {
	rows := pgxmock.NewRows([]string{"id", "table_id", "name", "view_type", "config", "position", "created_at", "updated_at"}).
		AddRow(view.ID, view.TableID, view.Name, view.Type, view.Config, view.Position, view.CreatedAt, view.UpdatedAt)
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
}

// Test context helper
func testContext() context.Context {
	return context.Background()
}

// createMockPool creates a new pgxmock pool for testing
func createMockPool() (pgxmock.PgxPoolIface, error) {
	return pgxmock.NewPool()
}
