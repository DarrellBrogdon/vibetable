package realtime

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.bases)
	assert.NotNil(t, hub.presence)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

func TestHub_GetActiveUsers(t *testing.T) {
	t.Run("returns 0 for base with no users", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()

		count := hub.GetActiveUsers(baseID)
		assert.Equal(t, 0, count)
	})

	t.Run("returns correct count after adding users", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()

		// Manually add clients
		hub.bases[baseID] = make(map[*Client]bool)
		client1 := &Client{userID: uuid.New()}
		client2 := &Client{userID: uuid.New()}
		hub.bases[baseID][client1] = true
		hub.bases[baseID][client2] = true

		count := hub.GetActiveUsers(baseID)
		assert.Equal(t, 2, count)
	})
}

func TestHub_GetPresence(t *testing.T) {
	t.Run("returns empty slice for base with no users", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()

		presence := hub.GetPresence(baseID)
		assert.NotNil(t, presence)
		assert.Empty(t, presence)
	})

	t.Run("returns presence list for base with users", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()

		// Manually add presence
		hub.presence[baseID] = make(map[uuid.UUID]*UserPresence)
		hub.presence[baseID][userID] = &UserPresence{
			UserID:    userID,
			Email:     "test@example.com",
			JoinedAt:  time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		presence := hub.GetPresence(baseID)
		assert.Len(t, presence, 1)
		assert.Equal(t, userID, presence[0].UserID)
	})
}

func TestHub_UpdatePresence(t *testing.T) {
	t.Run("updates presence for existing user", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()

		// Setup presence
		hub.bases[baseID] = make(map[*Client]bool)
		hub.presence[baseID] = make(map[uuid.UUID]*UserPresence)
		hub.presence[baseID][userID] = &UserPresence{
			UserID:    userID,
			Email:     "test@example.com",
			JoinedAt:  time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		tableID := uuid.New()
		hub.UpdatePresence(baseID, userID, func(p *UserPresence) {
			p.TableID = &tableID
		})

		require.NotNil(t, hub.presence[baseID][userID].TableID)
		assert.Equal(t, tableID, *hub.presence[baseID][userID].TableID)
	})

	t.Run("does nothing for non-existent base", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()

		// Should not panic
		hub.UpdatePresence(baseID, userID, func(p *UserPresence) {
			p.Email = "updated@example.com"
		})
	})

	t.Run("does nothing for non-existent user", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()

		hub.presence[baseID] = make(map[uuid.UUID]*UserPresence)

		// Should not panic
		hub.UpdatePresence(baseID, userID, func(p *UserPresence) {
			p.Email = "updated@example.com"
		})
	})
}

func TestHub_Broadcast(t *testing.T) {
	t.Run("sends message to broadcast channel", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()

		msg := NewMessage(MsgTypeRecordCreated, baseID, userID)

		// Non-blocking send to broadcast channel
		hub.Broadcast(msg)

		// Message should be in the channel
		select {
		case received := <-hub.broadcast:
			assert.Equal(t, msg, received)
		default:
			t.Error("Expected message in broadcast channel")
		}
	})
}

func TestHub_registerClient(t *testing.T) {
	t.Run("registers client to base", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()
		name := "Test User"

		client := &Client{
			hub:    hub,
			userID: userID,
			email:  "test@example.com",
			name:   &name,
			baseID: baseID,
			send:   make(chan *Message, 256),
		}

		hub.registerClient(client)

		// Client should be registered
		assert.True(t, hub.bases[baseID][client])

		// Presence should be added
		require.NotNil(t, hub.presence[baseID][userID])
		assert.Equal(t, "test@example.com", hub.presence[baseID][userID].Email)
	})

	t.Run("does nothing for nil base ID", func(t *testing.T) {
		hub := NewHub()

		client := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: uuid.Nil,
			send:   make(chan *Message, 256),
		}

		hub.registerClient(client)

		// Nothing should be registered
		assert.Empty(t, hub.bases)
	})
}

func TestHub_unregisterClient(t *testing.T) {
	t.Run("unregisters client from base", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()

		client := &Client{
			hub:    hub,
			userID: userID,
			email:  "test@example.com",
			baseID: baseID,
			send:   make(chan *Message, 256),
		}

		// First register
		hub.bases[baseID] = make(map[*Client]bool)
		hub.bases[baseID][client] = true
		hub.presence[baseID] = make(map[uuid.UUID]*UserPresence)
		hub.presence[baseID][userID] = &UserPresence{
			UserID: userID,
			Email:  "test@example.com",
		}

		hub.unregisterClient(client)

		// Client should be unregistered
		assert.Empty(t, hub.bases[baseID])
		// Base should be cleaned up when empty
		_, exists := hub.bases[baseID]
		assert.False(t, exists)
	})

	t.Run("does nothing for nil base ID", func(t *testing.T) {
		hub := NewHub()

		client := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: uuid.Nil,
			send:   make(chan *Message, 256),
		}

		// Should not panic
		hub.unregisterClient(client)
	})

	t.Run("does nothing for unregistered client", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()

		client := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: baseID,
			send:   make(chan *Message, 256),
		}

		// Should not panic
		hub.unregisterClient(client)
	})
}

func TestHub_getPresenceList(t *testing.T) {
	t.Run("returns empty list for unknown base", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()

		list := hub.getPresenceList(baseID)
		assert.NotNil(t, list)
		assert.Empty(t, list)
	})

	t.Run("returns all users in base", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		user1ID := uuid.New()
		user2ID := uuid.New()

		hub.presence[baseID] = make(map[uuid.UUID]*UserPresence)
		hub.presence[baseID][user1ID] = &UserPresence{
			UserID: user1ID,
			Email:  "user1@example.com",
		}
		hub.presence[baseID][user2ID] = &UserPresence{
			UserID: user2ID,
			Email:  "user2@example.com",
		}

		list := hub.getPresenceList(baseID)
		assert.Len(t, list, 2)
	})
}

func TestHub_broadcastToBase(t *testing.T) {
	t.Run("sends message to all clients in base", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		userID := uuid.New()

		client1 := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: baseID,
			send:   make(chan *Message, 256),
		}
		client2 := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: baseID,
			send:   make(chan *Message, 256),
		}

		hub.bases[baseID] = make(map[*Client]bool)
		hub.bases[baseID][client1] = true
		hub.bases[baseID][client2] = true

		msg := NewMessage(MsgTypeRecordCreated, baseID, userID)
		hub.broadcastToBase(baseID, msg, uuid.Nil)

		// Both clients should receive the message
		select {
		case received := <-client1.send:
			assert.Equal(t, msg, received)
		default:
			t.Error("Client1 should have received message")
		}

		select {
		case received := <-client2.send:
			assert.Equal(t, msg, received)
		default:
			t.Error("Client2 should have received message")
		}
	})

	t.Run("excludes specified user", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()
		excludeUserID := uuid.New()

		client1 := &Client{
			hub:    hub,
			userID: excludeUserID,
			baseID: baseID,
			send:   make(chan *Message, 256),
		}
		client2 := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: baseID,
			send:   make(chan *Message, 256),
		}

		hub.bases[baseID] = make(map[*Client]bool)
		hub.bases[baseID][client1] = true
		hub.bases[baseID][client2] = true

		msg := NewMessage(MsgTypeRecordCreated, baseID, excludeUserID)
		hub.broadcastToBase(baseID, msg, excludeUserID)

		// Client1 (excluded) should NOT receive the message
		select {
		case <-client1.send:
			t.Error("Excluded client should not have received message")
		default:
			// Good - no message
		}

		// Client2 should receive the message
		select {
		case received := <-client2.send:
			assert.Equal(t, msg, received)
		default:
			t.Error("Client2 should have received message")
		}
	})

	t.Run("does nothing for unknown base", func(t *testing.T) {
		hub := NewHub()
		baseID := uuid.New()

		msg := NewMessage(MsgTypeRecordCreated, baseID, uuid.New())

		// Should not panic
		hub.broadcastToBase(baseID, msg, uuid.Nil)
	})
}

func TestHub_Register(t *testing.T) {
	t.Run("sends client to register channel", func(t *testing.T) {
		hub := NewHub()

		client := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: uuid.New(),
			send:   make(chan *Message, 256),
		}

		go func() {
			hub.Register(client)
		}()

		select {
		case received := <-hub.register:
			assert.Equal(t, client, received)
		case <-time.After(time.Second):
			t.Error("Timeout waiting for client registration")
		}
	})
}

func TestHub_Unregister(t *testing.T) {
	t.Run("sends client to unregister channel", func(t *testing.T) {
		hub := NewHub()

		client := &Client{
			hub:    hub,
			userID: uuid.New(),
			baseID: uuid.New(),
			send:   make(chan *Message, 256),
		}

		go func() {
			hub.Unregister(client)
		}()

		select {
		case received := <-hub.unregister:
			assert.Equal(t, client, received)
		case <-time.After(time.Second):
			t.Error("Timeout waiting for client unregistration")
		}
	})
}
