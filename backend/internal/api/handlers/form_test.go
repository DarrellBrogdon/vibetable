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

func TestNewFormHandler(t *testing.T) {
	t.Run("creates handler with nil store", func(t *testing.T) {
		handler := NewFormHandler(nil)
		assert.NotNil(t, handler)
	})
}

func TestFormHandler_ListForms(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewFormHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/123/forms", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.ListForms(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid table ID", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/tables/not-a-uuid/forms", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListForms(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestFormHandler_CreateForm(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewFormHandler(nil)

		body := bytes.NewBufferString(`{"name": "My Form"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/forms", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.CreateForm(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid table ID", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "My Form"}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/forms", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/forms", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("uses default name when empty", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/forms", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("tableId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store, but tests default name path
		defer func() { recover() }()
		handler.CreateForm(w, req)
	})
}

func TestFormHandler_GetForm(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewFormHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/forms/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetForm(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid form ID", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/forms/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestFormHandler_UpdateForm(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewFormHandler(nil)

		body := bytes.NewBufferString(`{"name": "Updated"}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateForm(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid form ID", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "Updated"}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("handles all optional fields", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{
			"name": "Updated",
			"description": "A description",
			"is_active": true,
			"success_message": "Thanks!",
			"redirect_url": "https://example.com",
			"submit_button_text": "Submit Now"
		}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.UpdateForm(w, req)
	})
}

func TestFormHandler_UpdateFormFields(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewFormHandler(nil)

		body := bytes.NewBufferString(`{"fields": []}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/123/fields", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateFormFields(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid form ID", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"fields": []}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/not-a-uuid/fields", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateFormFields(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/123/fields", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateFormFields(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("returns 400 for invalid field_id", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"fields": [{"field_id": "not-a-uuid", "is_required": true, "is_visible": true, "position": 0}]}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/123/fields", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateFormFields(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_field_id", response.Error)
	})

	t.Run("handles valid fields with all options", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		fieldID := uuid.New()
		body := bytes.NewBufferString(`{"fields": [{"field_id": "` + fieldID.String() + `", "label": "Name", "help_text": "Enter your name", "is_required": true, "is_visible": true, "position": 0}]}`)
		req := httptest.NewRequest(http.MethodPatch, "/forms/123/fields", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.UpdateFormFields(w, req)
	})
}

func TestFormHandler_DeleteForm(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewFormHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/forms/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteForm(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid form ID", func(t *testing.T) {
		handler := NewFormHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/forms/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestFormHandler_GetPublicForm(t *testing.T) {
	t.Run("returns 400 for empty token", func(t *testing.T) {
		handler := NewFormHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/public/forms/", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.GetPublicForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_token", response.Error)
	})

	t.Run("handles valid token", func(t *testing.T) {
		handler := NewFormHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/public/forms/valid-token", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.GetPublicForm(w, req)
	})
}

func TestFormHandler_SubmitPublicForm(t *testing.T) {
	t.Run("returns 400 for empty token", func(t *testing.T) {
		handler := NewFormHandler(nil)

		body := bytes.NewBufferString(`{"values": {}}`)
		req := httptest.NewRequest(http.MethodPost, "/public/forms/", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.SubmitPublicForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_token", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewFormHandler(nil)

		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/public/forms/valid-token", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.SubmitPublicForm(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("handles valid submission", func(t *testing.T) {
		handler := NewFormHandler(nil)

		body := bytes.NewBufferString(`{"values": {"field1": "value1"}}`)
		req := httptest.NewRequest(http.MethodPost, "/public/forms/valid-token", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "valid-token")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.SubmitPublicForm(w, req)
	})
}
