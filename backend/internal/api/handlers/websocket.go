package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vibetable/backend/internal/realtime"
	"github.com/vibetable/backend/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost during development
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:5173" ||
			origin == "http://localhost:3000" ||
			origin == ""
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub       *realtime.Hub
	authStore *store.AuthStore
	baseStore *store.BaseStore
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *realtime.Hub, authStore *store.AuthStore, baseStore *store.BaseStore) *WebSocketHandler {
	return &WebSocketHandler{
		hub:       hub,
		authStore: authStore,
		baseStore: baseStore,
	}
}

// ServeWS handles WebSocket upgrade requests
// GET /ws?baseId=xxx&token=xxx
func (h *WebSocketHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get base ID from query
	baseIDStr := r.URL.Query().Get("baseId")
	if baseIDStr == "" {
		http.Error(w, "baseId query parameter required", http.StatusBadRequest)
		return
	}

	baseID, err := uuid.Parse(baseIDStr)
	if err != nil {
		http.Error(w, "invalid baseId", http.StatusBadRequest)
		return
	}

	// Get token from query (WebSocket doesn't support Authorization header easily)
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token query parameter required", http.StatusUnauthorized)
		return
	}

	// Validate token and get user
	ctx := context.Background()
	session, err := h.authStore.GetSessionByToken(ctx, token)
	if err != nil {
		log.Printf("Invalid WebSocket token: %v", err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	user, err := h.authStore.GetUserByID(ctx, session.UserID)
	if err != nil {
		log.Printf("Failed to get user for WebSocket: %v", err)
		http.Error(w, "user not found", http.StatusUnauthorized)
		return
	}

	// Verify user has access to the base
	_, err = h.baseStore.GetUserRole(ctx, baseID, user.ID)
	if err != nil {
		log.Printf("User %s does not have access to base %s", user.ID, baseID)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := realtime.NewClient(h.hub, conn, user.ID, user.Email, user.Name, baseID)

	// Register client
	h.hub.Register(client)

	// Start read/write pumps
	go client.WritePump()
	go client.ReadPump()
}

// GetHub returns the hub instance for broadcasting
func (h *WebSocketHandler) GetHub() *realtime.Hub {
	return h.hub
}

// writeError writes a JSON error response (utility for future use)
func writeWSError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}
