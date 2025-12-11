package realtime

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients by base ID
	bases map[uuid.UUID]map[*Client]bool

	// User presence by base ID
	presence map[uuid.UUID]map[uuid.UUID]*UserPresence

	// Channel for broadcasting messages
	broadcast chan *Message

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	h := &Hub{
		bases:      make(map[uuid.UUID]map[*Client]bool),
		presence:   make(map[uuid.UUID]map[uuid.UUID]*UserPresence),
		broadcast:  make(chan *Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	return h
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	baseID := client.baseID
	if baseID == uuid.Nil {
		return
	}

	// Create base map if it doesn't exist
	if h.bases[baseID] == nil {
		h.bases[baseID] = make(map[*Client]bool)
	}
	h.bases[baseID][client] = true

	// Add to presence
	if h.presence[baseID] == nil {
		h.presence[baseID] = make(map[uuid.UUID]*UserPresence)
	}

	presence := &UserPresence{
		UserID:    client.userID,
		Email:     client.email,
		Name:      client.name,
		JoinedAt:  time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	h.presence[baseID][client.userID] = presence

	log.Printf("Client registered: user=%s base=%s", client.userID, baseID)

	// Notify other clients that this user joined
	joinMsg := NewMessage(MsgTypeUserJoined, baseID, client.userID).
		WithPayload(presence)
	h.broadcastToBase(baseID, joinMsg, client.userID)

	// Send presence list to the new client
	presenceList := h.getPresenceList(baseID)
	listMsg := NewMessage(MsgTypePresenceList, baseID, client.userID).
		WithPayload(presenceList)
	client.send <- listMsg
}

// unregisterClient removes a client
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	baseID := client.baseID
	if baseID == uuid.Nil {
		return
	}

	if _, ok := h.bases[baseID]; ok {
		if _, ok := h.bases[baseID][client]; ok {
			delete(h.bases[baseID], client)
			close(client.send)

			// Remove from presence
			delete(h.presence[baseID], client.userID)

			log.Printf("Client unregistered: user=%s base=%s", client.userID, baseID)

			// Notify other clients that this user left
			leftMsg := NewMessage(MsgTypeUserLeft, baseID, client.userID).
				WithPayload(map[string]interface{}{
					"userId": client.userID,
				})
			h.broadcastToBase(baseID, leftMsg, client.userID)

			// Clean up empty base
			if len(h.bases[baseID]) == 0 {
				delete(h.bases, baseID)
				delete(h.presence, baseID)
			}
		}
	}
}

// broadcastMessage sends a message to all clients in the relevant base
func (h *Hub) broadcastMessage(message *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.broadcastToBase(message.BaseID, message, uuid.Nil)
}

// broadcastToBase sends a message to all clients in a base (optionally excluding a user)
func (h *Hub) broadcastToBase(baseID uuid.UUID, message *Message, excludeUserID uuid.UUID) {
	clients, ok := h.bases[baseID]
	if !ok {
		return
	}

	for client := range clients {
		// Optionally exclude the sender
		if excludeUserID != uuid.Nil && client.userID == excludeUserID {
			continue
		}

		select {
		case client.send <- message:
		default:
			// Client's buffer is full, close and remove
			close(client.send)
			delete(h.bases[baseID], client)
		}
	}
}

// getPresenceList returns all users present in a base
func (h *Hub) getPresenceList(baseID uuid.UUID) []*UserPresence {
	presenceMap, ok := h.presence[baseID]
	if !ok {
		return []*UserPresence{}
	}

	list := make([]*UserPresence, 0, len(presenceMap))
	for _, p := range presenceMap {
		list = append(list, p)
	}
	return list
}

// UpdatePresence updates a user's presence info
func (h *Hub) UpdatePresence(baseID, userID uuid.UUID, update func(*UserPresence)) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.presence[baseID] == nil {
		return
	}

	presence, ok := h.presence[baseID][userID]
	if !ok {
		return
	}

	update(presence)
	presence.UpdatedAt = time.Now().UTC()

	// Broadcast presence update
	msg := NewMessage(MsgTypePresence, baseID, userID).WithPayload(presence)
	h.broadcastToBase(baseID, msg, uuid.Nil)
}

// Broadcast sends a message to the broadcast channel
func (h *Hub) Broadcast(message *Message) {
	select {
	case h.broadcast <- message:
	default:
		log.Printf("Broadcast channel full, dropping message: %s", message.Type)
	}
}

// GetActiveUsers returns the count of active users in a base
func (h *Hub) GetActiveUsers(baseID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.bases[baseID]; ok {
		return len(clients)
	}
	return 0
}

// GetPresence returns the presence list for a base
func (h *Hub) GetPresence(baseID uuid.UUID) []*UserPresence {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return h.getPresenceList(baseID)
}

// Register adds a client to the register channel
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister adds a client to the unregister channel
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
