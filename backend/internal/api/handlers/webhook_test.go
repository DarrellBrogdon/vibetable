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

func TestWebhookHandler_ListWebhooks(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/bases/123/webhooks", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ListWebhooks(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/bases/not-a-uuid/webhooks", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListWebhooks(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestWebhookHandler_CreateWebhook(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		body := bytes.NewBufferString(`{"name": "My Webhook", "url": "https://example.com/webhook"}`)
		req := httptest.NewRequest(http.MethodPost, "/bases/123/webhooks", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.CreateWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My Webhook", "url": "https://example.com/webhook"}`)
		req := httptest.NewRequest(http.MethodPost, "/bases/not-a-uuid/webhooks", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateWebhook(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestWebhookHandler_GetWebhook(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/webhooks/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/webhooks/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetWebhook(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestWebhookHandler_UpdateWebhook(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		body := bytes.NewBufferString(`{"name": "Updated Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/webhooks/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "Updated Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/webhooks/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateWebhook(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWebhookHandler_DeleteWebhook(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/webhooks/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteWebhook(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/webhooks/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteWebhook(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestWebhookHandler_ListDeliveries(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/webhooks/123/deliveries", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ListDeliveries(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewWebhookHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/webhooks/not-a-uuid/deliveries", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListDeliveries(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
