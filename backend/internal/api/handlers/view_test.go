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

func TestViewHandler_ListViews(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewViewHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/123/views", nil)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ListViews(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/tables/not-a-uuid/views", nil)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListViews(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestViewHandler_CreateView(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewViewHandler(nil)

		body := bytes.NewBufferString(`{"name": "My View", "type": "grid"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/views", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.CreateView(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My View", "type": "grid"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/views", body)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/views", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "", "type": "grid"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/views", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "name_required", response.Error)
	})

	t.Run("should return 400 for whitespace-only name", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "   ", "type": "grid"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/views", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid view type", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My View", "type": "invalid_type"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/views", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_view_type", response.Error)
	})

	t.Run("should accept grid view type", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My View", "type": "grid"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/views", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic or return 500 due to nil store, but validates type accepted
		defer func() {
			recover()
		}()
		handler.CreateView(w, req)

		// If we get here with 400, it's a bad type error
		if w.Code == http.StatusBadRequest {
			t.Error("grid view type was rejected but should be valid")
		}
	})

	t.Run("should accept kanban view type", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My View", "type": "kanban"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/views", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		defer func() {
			recover()
		}()
		handler.CreateView(w, req)

		if w.Code == http.StatusBadRequest {
			t.Error("kanban view type was rejected but should be valid")
		}
	})
}

func TestViewHandler_GetView(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewViewHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/views/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetView(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/views/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestViewHandler_UpdateView(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewViewHandler(nil)

		body := bytes.NewBufferString(`{"name": "New Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateView(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "New Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": ""}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "name_required", response.Error)
	})

	t.Run("should return 400 for whitespace-only name", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "   "}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestViewHandler_DeleteView(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewViewHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/views/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteView(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/views/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestViewHandler_SetViewPublic(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewViewHandler(nil)

		body := bytes.NewBufferString(`{"is_public": true}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/123/public", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.SetViewPublic(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"is_public": true}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/not-a-uuid/public", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.SetViewPublic(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewViewHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/views/123/public", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.SetViewPublic(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})
}

func TestViewHandler_GetPublicView(t *testing.T) {
	t.Run("should return 400 for empty token", func(t *testing.T) {
		handler := NewViewHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/public/views/", nil)
		req = withURLParam(req, "token", "")
		w := httptest.NewRecorder()

		handler.GetPublicView(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "token_required", response.Error)
	})
}
