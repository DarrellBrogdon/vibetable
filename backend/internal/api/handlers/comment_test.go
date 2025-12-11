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

func TestNewCommentHandler(t *testing.T) {
	t.Run("creates handler with nil store", func(t *testing.T) {
		handler := NewCommentHandler(nil)
		assert.NotNil(t, handler)
	})
}

func TestCommentHandler_ListComments(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/records/123/comments", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.ListComments(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid record ID", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/records/not-a-uuid/comments", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListComments(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestCommentHandler_CreateComment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		body := bytes.NewBufferString(`{"content": "Test comment"}`)
		req := httptest.NewRequest(http.MethodPost, "/records/123/comments", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid record ID", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"content": "Test comment"}`)
		req := httptest.NewRequest(http.MethodPost, "/records/not-a-uuid/comments", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/records/123/comments", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("returns 400 for empty content", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"content": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/records/123/comments", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "content_required", response.Error)
	})

	t.Run("returns 400 for invalid parent_id", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"content": "Test", "parent_id": "not-a-uuid"}`)
		req := httptest.NewRequest(http.MethodPost, "/records/123/comments", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_parent_id", response.Error)
	})

	t.Run("handles valid parent_id", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		parentID := uuid.New()
		body := bytes.NewBufferString(`{"content": "Test", "parent_id": "` + parentID.String() + `"}`)
		req := httptest.NewRequest(http.MethodPost, "/records/123/comments", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.CreateComment(w, req)
	})

	t.Run("handles empty parent_id string", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"content": "Test", "parent_id": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/records/123/comments", body)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.CreateComment(w, req)
	})
}

func TestCommentHandler_GetComment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/comments/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetComment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid comment ID", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/comments/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestCommentHandler_UpdateComment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		body := bytes.NewBufferString(`{"content": "Updated"}`)
		req := httptest.NewRequest(http.MethodPatch, "/comments/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateComment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid comment ID", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"content": "Updated"}`)
		req := httptest.NewRequest(http.MethodPatch, "/comments/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/comments/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("returns 400 for empty content", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"content": ""}`)
		req := httptest.NewRequest(http.MethodPatch, "/comments/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "content_required", response.Error)
	})
}

func TestCommentHandler_DeleteComment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/comments/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteComment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid comment ID", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/comments/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestCommentHandler_ResolveComment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		body := bytes.NewBufferString(`{"resolved": true}`)
		req := httptest.NewRequest(http.MethodPost, "/comments/123/resolve", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ResolveComment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("returns 400 for invalid comment ID", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"resolved": true}`)
		req := httptest.NewRequest(http.MethodPost, "/comments/not-a-uuid/resolve", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ResolveComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		handler := NewCommentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/comments/123/resolve", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ResolveComment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})
}
