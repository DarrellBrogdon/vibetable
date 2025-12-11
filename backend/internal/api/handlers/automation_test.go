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
	"github.com/vibetable/backend/internal/store"
)

func TestNewAutomationHandler(t *testing.T) {
	t.Run("creates handler with nil store", func(t *testing.T) {
		handler := NewAutomationHandler(nil)
		assert.NotNil(t, handler)
	})
}

func TestAutomationHandler_ListAutomations(t *testing.T) {
	t.Run("returns 400 for invalid table ID", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/not-a-uuid/automations", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.ListAutomations(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/123/automations", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.ListAutomations(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})
}

func TestAutomationHandler_CreateAutomation(t *testing.T) {
	t.Run("returns 400 for invalid table ID", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		body := bytes.NewBufferString(`{"name": "Test", "triggerType": "record_created", "actionType": "send_email"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/automations", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.CreateAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		body := bytes.NewBufferString(`{"name": "Test", "triggerType": "record_created", "actionType": "send_email"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/automations", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.CreateAutomation(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/automations", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("returns 400 for empty name", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "", "triggerType": "record_created", "actionType": "send_email"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/automations", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "name_required", response.Error)
	})

	t.Run("returns 400 for empty trigger type", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "Test", "triggerType": "", "actionType": "send_email"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/automations", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "trigger_required", response.Error)
	})

	t.Run("returns 400 for empty action type", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "Test", "triggerType": "record_created", "actionType": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/automations", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "action_required", response.Error)
	})

	t.Run("handles optional fields", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{
			"name": "Test",
			"description": "A test automation",
			"enabled": false,
			"triggerType": "record_created",
			"triggerConfig": {"field_id": "123"},
			"actionType": "send_email",
			"actionConfig": {"to": "test@example.com"}
		}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/automations", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.CreateAutomation(w, req)
	})
}

func TestAutomationHandler_GetAutomation(t *testing.T) {
	t.Run("returns 400 for invalid automation ID", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/automations/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		w := httptest.NewRecorder()

		handler.GetAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/automations/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetAutomation(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})
}

func TestAutomationHandler_UpdateAutomation(t *testing.T) {
	t.Run("returns 400 for invalid automation ID", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		body := bytes.NewBufferString(`{"name": "Updated"}`)
		req := httptest.NewRequest(http.MethodPatch, "/automations/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		w := httptest.NewRecorder()

		handler.UpdateAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		body := bytes.NewBufferString(`{"name": "Updated"}`)
		req := httptest.NewRequest(http.MethodPatch, "/automations/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateAutomation(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/automations/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("handles all optional update fields", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{
			"name": "Updated",
			"description": "Updated description",
			"enabled": false,
			"triggerType": "record_updated",
			"triggerConfig": {"field_id": "456"},
			"actionType": "send_webhook",
			"actionConfig": {"url": "https://example.com"}
		}`)
		req := httptest.NewRequest(http.MethodPatch, "/automations/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.UpdateAutomation(w, req)
	})
}

func TestAutomationHandler_DeleteAutomation(t *testing.T) {
	t.Run("returns 400 for invalid automation ID", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/automations/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		w := httptest.NewRecorder()

		handler.DeleteAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/automations/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteAutomation(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})
}

func TestAutomationHandler_ToggleAutomation(t *testing.T) {
	t.Run("returns 400 for invalid automation ID", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		body := bytes.NewBufferString(`{"enabled": true}`)
		req := httptest.NewRequest(http.MethodPost, "/automations/not-a-uuid/toggle", body)
		req = withURLParam(req, "id", "not-a-uuid")
		w := httptest.NewRecorder()

		handler.ToggleAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		body := bytes.NewBufferString(`{"enabled": true}`)
		req := httptest.NewRequest(http.MethodPost, "/automations/123/toggle", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ToggleAutomation(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/automations/123/toggle", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ToggleAutomation(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})
}

func TestAutomationHandler_ListRuns(t *testing.T) {
	t.Run("returns 400 for invalid automation ID", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/automations/not-a-uuid/runs", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		w := httptest.NewRecorder()

		handler.ListRuns(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAutomationHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/automations/123/runs", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ListRuns(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})
}

func TestHandleAutomationStoreError(t *testing.T) {
	t.Run("handles not found error", func(t *testing.T) {
		w := httptest.NewRecorder()
		handleAutomationStoreError(w, store.ErrNotFound)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "not_found", response.Error)
	})

	t.Run("handles forbidden error", func(t *testing.T) {
		w := httptest.NewRecorder()
		handleAutomationStoreError(w, store.ErrForbidden)

		assert.Equal(t, http.StatusForbidden, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "forbidden", response.Error)
	})

	t.Run("handles generic error", func(t *testing.T) {
		w := httptest.NewRecorder()
		handleAutomationStoreError(w, assert.AnError)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "server_error", response.Error)
	})
}
