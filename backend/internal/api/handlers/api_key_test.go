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

func TestAPIKeyHandler_ListAPIKeys(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/api-keys", nil)
		w := httptest.NewRecorder()

		handler.ListAPIKeys(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})
}

func TestAPIKeyHandler_CreateAPIKey(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		body := bytes.NewBufferString(`{"name": "My API Key"}`)
		req := httptest.NewRequest(http.MethodPost, "/api-keys", body)
		w := httptest.NewRecorder()

		handler.CreateAPIKey(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/api-keys", body)
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateAPIKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/api-keys", body)
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateAPIKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})
}

func TestAPIKeyHandler_GetAPIKey(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/api-keys/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetAPIKey(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/api-keys/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetAPIKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestAPIKeyHandler_DeleteAPIKey(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/api-keys/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteAPIKey(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewAPIKeyHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/api-keys/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteAPIKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
