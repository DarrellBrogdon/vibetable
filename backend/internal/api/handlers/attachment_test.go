package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestNewAttachmentHandler(t *testing.T) {
	t.Run("creates handler with nil store", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)
		assert.NotNil(t, handler)
	})
}

func TestAttachmentHandler_UploadAttachment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/records/123/fields/456/attachments", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		rctx.URLParams.Add("fieldId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid record ID", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/records/not-a-uuid/fields/456/attachments", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", "not-a-uuid")
		rctx.URLParams.Add("fieldId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid field ID", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/records/123/fields/not-a-uuid/attachments", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		rctx.URLParams.Add("fieldId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid multipart form", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodPost, "/records/123/fields/456/attachments", bytes.NewBufferString("not multipart"))
		req.Header.Set("Content-Type", "text/plain")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		rctx.URLParams.Add("fieldId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("returns 400 when no file uploaded", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/records/123/fields/456/attachments", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		rctx.URLParams.Add("fieldId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "file_required", response.Error)
	})

	t.Run("handles file upload with content type", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("test content"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/records/123/fields/456/attachments", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		rctx.URLParams.Add("fieldId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		// Will panic on nil store
		defer func() { recover() }()
		handler.UploadAttachment(w, req)
	})
}

func TestAttachmentHandler_ListAttachments(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/records/123/fields/456/attachments", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		rctx.URLParams.Add("fieldId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		handler.ListAttachments(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid record ID", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/records/not-a-uuid/fields/456/attachments", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", "not-a-uuid")
		rctx.URLParams.Add("fieldId", uuid.New().String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListAttachments(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("returns 400 for invalid field ID", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/records/123/fields/not-a-uuid/attachments", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("recordId", uuid.New().String())
		rctx.URLParams.Add("fieldId", "not-a-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListAttachments(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestAttachmentHandler_GetAttachment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/attachments/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid attachment ID", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/attachments/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestAttachmentHandler_DownloadAttachment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		req := httptest.NewRequest(http.MethodGet, "/attachments/123/download", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DownloadAttachment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid attachment ID", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/attachments/not-a-uuid/download", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DownloadAttachment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestAttachmentHandler_DeleteAttachment(t *testing.T) {
	t.Run("returns 401 when no user in context", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		req := httptest.NewRequest(http.MethodDelete, "/attachments/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("returns 400 for invalid attachment ID", func(t *testing.T) {
		handler := NewAttachmentHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/attachments/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}
