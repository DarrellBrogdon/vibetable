package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestFieldHandler_ListFields(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/123/fields", nil)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ListFields(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/tables/not-a-uuid/fields", nil)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListFields(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestFieldHandler_CreateField(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		body := bytes.NewBufferString(`{"name": "My Field", "field_type": "text"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/fields", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.CreateField(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My Field", "field_type": "text"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/fields", body)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/fields", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "", "field_type": "text"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/fields", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "name_required", response.Error)
	})

	t.Run("should return 400 for whitespace-only name", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "   ", "field_type": "text"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/fields", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid field type", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My Field", "field_type": "invalid_type"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/fields", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_field_type", response.Error)
	})

	t.Run("should accept valid field types", func(t *testing.T) {
		validTypes := []string{"text", "number", "checkbox", "date", "single_select", "multi_select", "linked_record", "formula", "rollup", "lookup", "attachment"}

		for _, fieldType := range validTypes {
			t.Run(fieldType, func(t *testing.T) {
				handler := NewFieldHandler(nil)

				user := &models.User{ID: uuid.New(), Email: "test@example.com"}
				body := bytes.NewBufferString(`{"name": "My Field", "field_type": "` + fieldType + `"}`)
				req := httptest.NewRequest(http.MethodPost, "/tables/123/fields", body)
				req = withURLParam(req, "tableId", uuid.New().String())
				req = req.WithContext(SetUserInContext(req.Context(), user))
				w := httptest.NewRecorder()

				// Will panic or return 500 due to nil store, but validates field type accepted
				defer func() {
					recover()
				}()
				handler.CreateField(w, req)

				// If we get here with 400, it's a bad field type error
				if w.Code == http.StatusBadRequest {
					t.Errorf("field type %s was rejected but should be valid", fieldType)
				}
			})
		}
	})
}

func TestFieldHandler_GetField(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/fields/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetField(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/fields/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestFieldHandler_UpdateField(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		body := bytes.NewBufferString(`{"name": "New Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/fields/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateField(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "New Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/fields/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/fields/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": ""}`)
		req := httptest.NewRequest(http.MethodPatch, "/fields/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "name_required", response.Error)
	})

	t.Run("should return 400 for whitespace-only name", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "   "}`)
		req := httptest.NewRequest(http.MethodPatch, "/fields/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestFieldHandler_DeleteField(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/fields/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteField(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/fields/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteField(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestFieldHandler_ReorderFields(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		body := bytes.NewBufferString(`{"field_ids": ["` + uuid.New().String() + `"]}`)
		req := httptest.NewRequest(http.MethodPut, "/tables/123/fields/reorder", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ReorderFields(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"field_ids": ["` + uuid.New().String() + `"]}`)
		req := httptest.NewRequest(http.MethodPut, "/tables/not-a-uuid/fields/reorder", body)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ReorderFields(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPut, "/tables/123/fields/reorder", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ReorderFields(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty field_ids", func(t *testing.T) {
		handler := NewFieldHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"field_ids": []}`)
		req := httptest.NewRequest(http.MethodPut, "/tables/123/fields/reorder", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ReorderFields(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "field_ids_required", response.Error)
	})
}
