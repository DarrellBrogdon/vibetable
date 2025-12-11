package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vibetable/backend/internal/realtime"
)

func TestNewWebSocketHandler(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub, nil, nil)

	assert.NotNil(t, handler)
	assert.Equal(t, hub, handler.GetHub())
}

func TestWebSocketHandler_GetHub(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub, nil, nil)

	assert.Equal(t, hub, handler.GetHub())
}

func TestWebSocketHandler_ServeWS_MissingBaseId(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	w := httptest.NewRecorder()

	handler.ServeWS(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "baseId query parameter required")
}

func TestWebSocketHandler_ServeWS_InvalidBaseId(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/ws?baseId=not-a-uuid", nil)
	w := httptest.NewRecorder()

	handler.ServeWS(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid baseId")
}

func TestWebSocketHandler_ServeWS_MissingToken(t *testing.T) {
	hub := realtime.NewHub()
	handler := NewWebSocketHandler(hub, nil, nil)

	baseID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/ws?baseId="+baseID.String(), nil)
	w := httptest.NewRecorder()

	handler.ServeWS(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "token query parameter required")
}

func TestWriteWSError(t *testing.T) {
	w := httptest.NewRecorder()

	writeWSError(w, http.StatusBadRequest, "test_error", "Test error message")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Body.String(), "test_error")
	assert.Contains(t, w.Body.String(), "Test error message")
}
