package realtime

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192
)

// Client represents a WebSocket client
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan *Message
	userID uuid.UUID
	email  string
	name   *string
	baseID uuid.UUID
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, email string, name *string, baseID uuid.UUID) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan *Message, 256),
		userID: userID,
		email:  email,
		name:   name,
		baseID: baseID,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle different message types
		c.handleMessage(&incoming)
	}
}

// handleMessage processes an incoming message from the client
func (c *Client) handleMessage(msg *IncomingMessage) {
	switch msg.Type {
	case MsgTypeCursor:
		// Handle cursor position update
		var update CursorUpdate
		if err := json.Unmarshal(msg.Payload, &update); err != nil {
			log.Printf("Error parsing cursor update: %v", err)
			return
		}

		c.hub.UpdatePresence(c.baseID, c.userID, func(p *UserPresence) {
			p.TableID = &update.TableID
			if update.ViewID != uuid.Nil {
				p.ViewID = &update.ViewID
			}
			p.CellRef = update.CellRef
		})

	case "ping":
		// Respond with pong
		pong := NewMessage("pong", c.baseID, c.userID)
		c.send <- pong

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("Error writing message: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
