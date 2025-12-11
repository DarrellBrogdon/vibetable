package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestNewCSVHandler(t *testing.T) {
	t.Run("creates handler with nil stores", func(t *testing.T) {
		handler := NewCSVHandler(nil, nil, nil)
		assert.NotNil(t, handler)
	})
}

func TestCSVHandler_Preview(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCSVHandler(nil, nil, nil)

		body := bytes.NewBufferString(`{"data": "col1,col2\nval1,val2"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/import/preview", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.Preview(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid table ID", func(t *testing.T) {
		handler := NewCSVHandler(nil, nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"data": "col1,col2\nval1,val2"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/import/preview", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.Preview(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	// Note: These tests require a non-nil fieldStore to pass validation
	// The request format validation happens after table access check which needs a store
}

func TestCSVHandler_Import(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCSVHandler(nil, nil, nil)

		body := bytes.NewBufferString(`{"data": "col1,col2\nval1,val2", "mappings": {}}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/import", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.Import(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid table ID", func(t *testing.T) {
		handler := NewCSVHandler(nil, nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"data": "col1,col2\nval1,val2", "mappings": {}}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/import", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.Import(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestCSVHandler_Export(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCSVHandler(nil, nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/123/export", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.Export(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid table ID", func(t *testing.T) {
		handler := NewCSVHandler(nil, nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/tables/not-a-uuid/export", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.Export(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestCSVHandler_parseCSVPreview(t *testing.T) {
	handler := NewCSVHandler(nil, nil, nil)

	t.Run("parses valid CSV", func(t *testing.T) {
		preview, err := handler.parseCSVPreview("col1,col2\nval1,val2\nval3,val4", 5)
		require.NoError(t, err)

		assert.Equal(t, []string{"col1", "col2"}, preview.Columns)
		assert.Len(t, preview.Rows, 2)
		assert.Equal(t, 2, preview.Total)
		assert.Equal(t, "val1", preview.Rows[0]["col1"])
		assert.Equal(t, "val2", preview.Rows[0]["col2"])
	})

	t.Run("limits preview rows", func(t *testing.T) {
		csv := "col1\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10"
		preview, err := handler.parseCSVPreview(csv, 3)
		require.NoError(t, err)

		assert.Len(t, preview.Rows, 3)
		assert.Equal(t, 10, preview.Total)
	})

	t.Run("returns error for empty CSV", func(t *testing.T) {
		_, err := handler.parseCSVPreview("", 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("returns error for invalid CSV", func(t *testing.T) {
		// Invalid CSV with mismatched quotes
		_, err := handler.parseCSVPreview("col1,\"col2\ncol3", 5)
		assert.Error(t, err)
	})

	// Note: Go's CSV parser doesn't handle rows with fewer columns by default
	// It returns an error for inconsistent field counts
}

func TestCSVHandler_convertCellValue(t *testing.T) {
	handler := NewCSVHandler(nil, nil, nil)

	t.Run("converts text field", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeText}
		result := handler.convertCellValue("hello", field)
		assert.Equal(t, "hello", result)
	})

	t.Run("converts number field", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeNumber}

		result := handler.convertCellValue("123.45", field)
		assert.Equal(t, 123.45, result)

		// Invalid number returns string
		result = handler.convertCellValue("not a number", field)
		assert.Equal(t, "not a number", result)
	})

	t.Run("converts checkbox field", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeCheckbox}

		assert.True(t, handler.convertCellValue("true", field).(bool))
		assert.True(t, handler.convertCellValue("1", field).(bool))
		assert.True(t, handler.convertCellValue("yes", field).(bool))
		assert.True(t, handler.convertCellValue("TRUE", field).(bool))
		assert.True(t, handler.convertCellValue("Yes", field).(bool))
		assert.True(t, handler.convertCellValue("Y", field).(bool))
		assert.True(t, handler.convertCellValue("y", field).(bool))
		assert.False(t, handler.convertCellValue("false", field).(bool))
		assert.False(t, handler.convertCellValue("0", field).(bool))
		assert.False(t, handler.convertCellValue("no", field).(bool))
	})

	t.Run("converts date field", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeDate}
		result := handler.convertCellValue("2024-01-15", field)
		assert.Equal(t, "2024-01-15", result)
	})

	t.Run("converts linked record field", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeLinkedRecord}

		// JSON array
		result := handler.convertCellValue(`["id1", "id2"]`, field)
		assert.Equal(t, []string{"id1", "id2"}, result)

		// Single ID
		result = handler.convertCellValue("single-id", field)
		assert.Equal(t, []string{"single-id"}, result)

		// Empty string
		result = handler.convertCellValue("", field)
		assert.Equal(t, []string{}, result)
	})
}

func TestCSVHandler_formatCellValue(t *testing.T) {
	handler := NewCSVHandler(nil, nil, nil)

	t.Run("formats nil value", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeText}
		result := handler.formatCellValue(nil, field)
		assert.Equal(t, "", result)
	})

	t.Run("formats checkbox value", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeCheckbox}

		assert.Equal(t, "true", handler.formatCellValue(true, field))
		assert.Equal(t, "false", handler.formatCellValue(false, field))
	})

	t.Run("formats number value", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeNumber}

		assert.Equal(t, "123.45", handler.formatCellValue(123.45, field))
		assert.Equal(t, "42", handler.formatCellValue(42.0, field))
	})

	t.Run("formats linked record value", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeLinkedRecord}

		result := handler.formatCellValue([]interface{}{"id1", "id2"}, field)
		assert.Equal(t, `["id1","id2"]`, result)
	})

	t.Run("formats text value", func(t *testing.T) {
		field := models.Field{FieldType: models.FieldTypeText}

		assert.Equal(t, "hello", handler.formatCellValue("hello", field))
		assert.Equal(t, "123", handler.formatCellValue(123, field))
	})
}
