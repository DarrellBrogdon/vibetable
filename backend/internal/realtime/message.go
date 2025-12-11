package realtime

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Message types
const (
	// Presence messages
	MsgTypePresence     = "presence"      // User joined/left
	MsgTypeCursor       = "cursor"        // Cursor position update
	MsgTypeUserJoined   = "user_joined"   // User joined the base
	MsgTypeUserLeft     = "user_left"     // User left the base
	MsgTypePresenceList = "presence_list" // Full list of active users

	// Record messages
	MsgTypeRecordCreated = "record_created"
	MsgTypeRecordUpdated = "record_updated"
	MsgTypeRecordDeleted = "record_deleted"

	// Field messages
	MsgTypeFieldCreated = "field_created"
	MsgTypeFieldUpdated = "field_updated"
	MsgTypeFieldDeleted = "field_deleted"

	// Table messages
	MsgTypeTableCreated = "table_created"
	MsgTypeTableUpdated = "table_updated"
	MsgTypeTableDeleted = "table_deleted"

	// View messages
	MsgTypeViewCreated = "view_created"
	MsgTypeViewUpdated = "view_updated"
	MsgTypeViewDeleted = "view_deleted"
)

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	BaseID    uuid.UUID   `json:"baseId"`
	TableID   *uuid.UUID  `json:"tableId,omitempty"`
	RecordID  *uuid.UUID  `json:"recordId,omitempty"`
	FieldID   *uuid.UUID  `json:"fieldId,omitempty"`
	ViewID    *uuid.UUID  `json:"viewId,omitempty"`
	UserID    uuid.UUID   `json:"userId"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// UserPresence represents a user's presence status
type UserPresence struct {
	UserID    uuid.UUID  `json:"userId"`
	Email     string     `json:"email"`
	Name      *string    `json:"name,omitempty"`
	TableID   *uuid.UUID `json:"tableId,omitempty"` // Which table they're viewing
	ViewID    *uuid.UUID `json:"viewId,omitempty"`  // Which view they're in
	CellRef   *CellRef   `json:"cellRef,omitempty"` // Which cell they have selected
	JoinedAt  time.Time  `json:"joinedAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// CellRef represents a cell reference (for cursor tracking)
type CellRef struct {
	RecordID uuid.UUID `json:"recordId"`
	FieldID  uuid.UUID `json:"fieldId"`
}

// CursorUpdate represents a cursor position update
type CursorUpdate struct {
	TableID  uuid.UUID `json:"tableId"`
	ViewID   uuid.UUID `json:"viewId,omitempty"`
	CellRef  *CellRef  `json:"cellRef,omitempty"`
}

// IncomingMessage represents a message from the client
type IncomingMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// SubscribeMessage is sent by client to subscribe to a base
type SubscribeMessage struct {
	BaseID uuid.UUID `json:"baseId"`
}

// NewMessage creates a new outgoing message
func NewMessage(msgType string, baseID, userID uuid.UUID) *Message {
	return &Message{
		Type:      msgType,
		BaseID:    baseID,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
	}
}

// WithTable adds table context to the message
func (m *Message) WithTable(tableID uuid.UUID) *Message {
	m.TableID = &tableID
	return m
}

// WithRecord adds record context to the message
func (m *Message) WithRecord(recordID uuid.UUID) *Message {
	m.RecordID = &recordID
	return m
}

// WithField adds field context to the message
func (m *Message) WithField(fieldID uuid.UUID) *Message {
	m.FieldID = &fieldID
	return m
}

// WithView adds view context to the message
func (m *Message) WithView(viewID uuid.UUID) *Message {
	m.ViewID = &viewID
	return m
}

// WithPayload adds payload to the message
func (m *Message) WithPayload(payload interface{}) *Message {
	m.Payload = payload
	return m
}
