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

func TestTableHandler_ListTables(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewTableHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/bases/123/tables", nil)
		req = withURLParam(req, "baseId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ListTables(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/bases/not-a-uuid/tables", nil)
		req = withURLParam(req, "baseId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListTables(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestTableHandler_CreateTable(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewTableHandler(nil)

		body := bytes.NewBufferString(`{"name": "My Table"}`)
		req := httptest.NewRequest(http.MethodPost, "/bases/123/tables", body)
		req = withURLParam(req, "baseId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.CreateTable(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My Table"}`)
		req := httptest.NewRequest(http.MethodPost, "/bases/not-a-uuid/tables", body)
		req = withURLParam(req, "baseId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/bases/123/tables", body)
		req = withURLParam(req, "baseId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/bases/123/tables", body)
		req = withURLParam(req, "baseId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "name_required", response.Error)
	})

	t.Run("should return 400 for whitespace-only name", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "   "}`)
		req := httptest.NewRequest(http.MethodPost, "/bases/123/tables", body)
		req = withURLParam(req, "baseId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTableHandler_GetTable(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewTableHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetTable(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/tables/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestTableHandler_UpdateTable(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewTableHandler(nil)

		body := bytes.NewBufferString(`{"name": "New Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/tables/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateTable(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "New Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/tables/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/tables/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": ""}`)
		req := httptest.NewRequest(http.MethodPatch, "/tables/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTableHandler_DeleteTable(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewTableHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/tables/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteTable(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/tables/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTableHandler_ReorderTables(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewTableHandler(nil)

		body := bytes.NewBufferString(`{"table_ids": ["` + uuid.New().String() + `"]}`)
		req := httptest.NewRequest(http.MethodPut, "/bases/123/tables/reorder", body)
		req = withURLParam(req, "baseId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ReorderTables(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid base UUID", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"table_ids": ["` + uuid.New().String() + `"]}`)
		req := httptest.NewRequest(http.MethodPut, "/bases/not-a-uuid/tables/reorder", body)
		req = withURLParam(req, "baseId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ReorderTables(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPut, "/bases/123/tables/reorder", body)
		req = withURLParam(req, "baseId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ReorderTables(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty table_ids", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"table_ids": []}`)
		req := httptest.NewRequest(http.MethodPut, "/bases/123/tables/reorder", body)
		req = withURLParam(req, "baseId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ReorderTables(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "table_ids_required", response.Error)
	})
}

func TestTableHandler_DuplicateTable(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewTableHandler(nil)

		body := bytes.NewBufferString(`{"include_records": true}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/duplicate", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DuplicateTable(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewTableHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"include_records": true}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/duplicate", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DuplicateTable(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}
