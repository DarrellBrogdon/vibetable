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

func TestRecordHandler_ListRecords(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/tables/123/records", nil)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.ListRecords(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/tables/not-a-uuid/records", nil)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.ListRecords(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestRecordHandler_CreateRecord(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		body := bytes.NewBufferString(`{"values": {"field1": "value1"}}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/records", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.CreateRecord(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"values": {"field1": "value1"}}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/records", body)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/records", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.CreateRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})
}

func TestRecordHandler_BulkCreateRecords(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		body := bytes.NewBufferString(`{"records": [{"field1": "value1"}]}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/records/bulk", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		w := httptest.NewRecorder()

		handler.BulkCreateRecords(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"records": [{"field1": "value1"}]}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/not-a-uuid/records/bulk", body)
		req = withURLParam(req, "tableId", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.BulkCreateRecords(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/records/bulk", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.BulkCreateRecords(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty records", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"records": []}`)
		req := httptest.NewRequest(http.MethodPost, "/tables/123/records/bulk", body)
		req = withURLParam(req, "tableId", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.BulkCreateRecords(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "records_required", response.Error)
	})
}

func TestRecordHandler_GetRecord(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		req := httptest.NewRequest(http.MethodGet, "/records/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.GetRecord(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodGet, "/records/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_id", response.Error)
	})
}

func TestRecordHandler_UpdateRecord(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		body := bytes.NewBufferString(`{"values": {"field1": "value1"}}`)
		req := httptest.NewRequest(http.MethodPut, "/records/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateRecord(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"values": {"field1": "value1"}}`)
		req := httptest.NewRequest(http.MethodPut, "/records/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPut, "/records/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})
}

func TestRecordHandler_PatchRecord(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		body := bytes.NewBufferString(`{"values": {"field1": "value1"}}`)
		req := httptest.NewRequest(http.MethodPatch, "/records/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.PatchRecord(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"values": {"field1": "value1"}}`)
		req := httptest.NewRequest(http.MethodPatch, "/records/not-a-uuid", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.PatchRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/records/123", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.PatchRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})
}

func TestRecordHandler_DeleteRecord(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		req := httptest.NewRequest(http.MethodDelete, "/records/123", nil)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.DeleteRecord(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		req := httptest.NewRequest(http.MethodDelete, "/records/not-a-uuid", nil)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.DeleteRecord(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRecordHandler_UpdateRecordColor(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		body := bytes.NewBufferString(`{"color": "red"}`)
		req := httptest.NewRequest(http.MethodPatch, "/records/123/color", body)
		req = withURLParam(req, "id", uuid.New().String())
		w := httptest.NewRecorder()

		handler.UpdateRecordColor(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid UUID", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"color": "red"}`)
		req := httptest.NewRequest(http.MethodPatch, "/records/not-a-uuid/color", body)
		req = withURLParam(req, "id", "not-a-uuid")
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateRecordColor(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPatch, "/records/123/color", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateRecordColor(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for invalid color", func(t *testing.T) {
		handler := NewRecordHandler(nil, nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"color": "magenta"}`)
		req := httptest.NewRequest(http.MethodPatch, "/records/123/color", body)
		req = withURLParam(req, "id", uuid.New().String())
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateRecordColor(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_color", response.Error)
	})

	t.Run("should accept valid colors", func(t *testing.T) {
		validColors := []string{"red", "orange", "yellow", "green", "blue", "purple", "pink", "gray"}

		for _, color := range validColors {
			t.Run(color, func(t *testing.T) {
				handler := NewRecordHandler(nil, nil)

				user := &models.User{ID: uuid.New(), Email: "test@example.com"}
				body := bytes.NewBufferString(`{"color": "` + color + `"}`)
				req := httptest.NewRequest(http.MethodPatch, "/records/123/color", body)
				req = withURLParam(req, "id", uuid.New().String())
				req = req.WithContext(SetUserInContext(req.Context(), user))
				w := httptest.NewRecorder()

				// Will panic or return 500 due to nil store, but validates color accepted
				defer func() {
					recover()
				}()
				handler.UpdateRecordColor(w, req)

				// If we get here with 400, it's a bad color error
				if w.Code == http.StatusBadRequest {
					t.Errorf("color %s was rejected but should be valid", color)
				}
			})
		}
	})
}
