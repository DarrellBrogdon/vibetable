package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vibetable/backend/internal/realtime"
	"github.com/vibetable/backend/internal/store"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Get allowed origins from environment
		allowedOrigins := []string{"http://localhost:5173", "http://localhost:3000"}
		if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
			allowedOrigins = strings.Split(origins, ",")
			for i := range allowedOrigins {
				allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
			}
		}

		origin := r.Header.Get("Origin")
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return origin == ""
	},
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub          *realtime.Hub
	authStore    *store.AuthStore
	baseStore    *store.BaseStore
	ticketStore  *store.WSTicketStore
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *realtime.Hub, authStore *store.AuthStore, baseStore *store.BaseStore) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		authStore:   authStore,
		baseStore:   baseStore,
		ticketStore: store.NewWSTicketStore(),
	}
}

// GetTicket generates a short-lived WebSocket ticket
// POST /api/v1/ws/ticket
func (h *WebSocketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userID, ok := r.Context().Value("user_id").(uuid.UUID)
	if !ok {
		writeWSError(w, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	// Parse request body
	var req struct {
		BaseID string `json:"base_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWSError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	baseID, err := uuid.Parse(req.BaseID)
	if err != nil {
		writeWSError(w, http.StatusBadRequest, "invalid_base_id", "Invalid base ID")
		return
	}

	// Verify user has access to the base
	ctx := r.Context()
	_, err = h.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		writeWSError(w, http.StatusForbidden, "forbidden", "You don't have access to this base")
		return
	}

	// Generate ticket
	ticket, err := h.ticketStore.GenerateTicket(ctx, userID, baseID)
	if err != nil {
		log.Printf("Failed to generate WebSocket ticket: %v", err)
		writeWSError(w, http.StatusInternalServerError, "internal_error", "Failed to generate ticket")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ticket": ticket,
	})
}

// ServeWS handles WebSocket upgrade requests
// GET /ws?baseId=xxx&ticket=xxx
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

	// Get ticket from query (short-lived, one-time use token)
	ticket := r.URL.Query().Get("ticket")
	if ticket == "" {
		http.Error(w, "ticket query parameter required", http.StatusUnauthorized)
		return
	}

	// Validate ticket and get user ID
	ctx := context.Background()
	userID, err := h.ticketStore.ValidateTicket(ctx, ticket, baseID)
	if err != nil {
		log.Printf("Invalid WebSocket ticket: %v", err)
		http.Error(w, "invalid or expired ticket", http.StatusUnauthorized)
		return
	}

	// Get user details
	user, err := h.authStore.GetUserByID(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user for WebSocket: %v", err)
		http.Error(w, "user not found", http.StatusUnauthorized)
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

// writeWSError writes a JSON error response
func writeWSError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}
