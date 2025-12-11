package realtime

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessageTypes(t *testing.T) {
	t.Run("presence message types are defined", func(t *testing.T) {
		assert.Equal(t, "presence", MsgTypePresence)
		assert.Equal(t, "cursor", MsgTypeCursor)
		assert.Equal(t, "user_joined", MsgTypeUserJoined)
		assert.Equal(t, "user_left", MsgTypeUserLeft)
		assert.Equal(t, "presence_list", MsgTypePresenceList)
	})

	t.Run("record message types are defined", func(t *testing.T) {
		assert.Equal(t, "record_created", MsgTypeRecordCreated)
		assert.Equal(t, "record_updated", MsgTypeRecordUpdated)
		assert.Equal(t, "record_deleted", MsgTypeRecordDeleted)
	})

	t.Run("field message types are defined", func(t *testing.T) {
		assert.Equal(t, "field_created", MsgTypeFieldCreated)
		assert.Equal(t, "field_updated", MsgTypeFieldUpdated)
		assert.Equal(t, "field_deleted", MsgTypeFieldDeleted)
	})

	t.Run("table message types are defined", func(t *testing.T) {
		assert.Equal(t, "table_created", MsgTypeTableCreated)
		assert.Equal(t, "table_updated", MsgTypeTableUpdated)
		assert.Equal(t, "table_deleted", MsgTypeTableDeleted)
	})

	t.Run("view message types are defined", func(t *testing.T) {
		assert.Equal(t, "view_created", MsgTypeViewCreated)
		assert.Equal(t, "view_updated", MsgTypeViewUpdated)
		assert.Equal(t, "view_deleted", MsgTypeViewDeleted)
	})
}

func TestNewMessage(t *testing.T) {
	t.Run("creates message with required fields", func(t *testing.T) {
		baseID := uuid.New()
		userID := uuid.New()

		msg := NewMessage(MsgTypeRecordCreated, baseID, userID)

		assert.Equal(t, MsgTypeRecordCreated, msg.Type)
		assert.Equal(t, baseID, msg.BaseID)
		assert.Equal(t, userID, msg.UserID)
		assert.NotZero(t, msg.Timestamp)
		assert.Nil(t, msg.TableID)
		assert.Nil(t, msg.RecordID)
		assert.Nil(t, msg.FieldID)
		assert.Nil(t, msg.ViewID)
		assert.Nil(t, msg.Payload)
	})

	t.Run("timestamp is in UTC", func(t *testing.T) {
		beforeTest := time.Now().UTC()
		msg := NewMessage(MsgTypePresence, uuid.New(), uuid.New())
		afterTest := time.Now().UTC()

		assert.True(t, msg.Timestamp.Equal(beforeTest) || msg.Timestamp.After(beforeTest))
		assert.True(t, msg.Timestamp.Equal(afterTest) || msg.Timestamp.Before(afterTest))
	})
}

func TestMessageWithTable(t *testing.T) {
	t.Run("adds table ID to message", func(t *testing.T) {
		msg := NewMessage(MsgTypeRecordCreated, uuid.New(), uuid.New())
		tableID := uuid.New()

		result := msg.WithTable(tableID)

		assert.Same(t, msg, result)
		require.NotNil(t, msg.TableID)
		assert.Equal(t, tableID, *msg.TableID)
	})
}

func TestMessageWithRecord(t *testing.T) {
	t.Run("adds record ID to message", func(t *testing.T) {
		msg := NewMessage(MsgTypeRecordUpdated, uuid.New(), uuid.New())
		recordID := uuid.New()

		result := msg.WithRecord(recordID)

		assert.Same(t, msg, result)
		require.NotNil(t, msg.RecordID)
		assert.Equal(t, recordID, *msg.RecordID)
	})
}

func TestMessageWithField(t *testing.T) {
	t.Run("adds field ID to message", func(t *testing.T) {
		msg := NewMessage(MsgTypeFieldCreated, uuid.New(), uuid.New())
		fieldID := uuid.New()

		result := msg.WithField(fieldID)

		assert.Same(t, msg, result)
		require.NotNil(t, msg.FieldID)
		assert.Equal(t, fieldID, *msg.FieldID)
	})
}

func TestMessageWithView(t *testing.T) {
	t.Run("adds view ID to message", func(t *testing.T) {
		msg := NewMessage(MsgTypeViewCreated, uuid.New(), uuid.New())
		viewID := uuid.New()

		result := msg.WithView(viewID)

		assert.Same(t, msg, result)
		require.NotNil(t, msg.ViewID)
		assert.Equal(t, viewID, *msg.ViewID)
	})
}

func TestMessageWithPayload(t *testing.T) {
	t.Run("adds payload to message", func(t *testing.T) {
		msg := NewMessage(MsgTypeRecordCreated, uuid.New(), uuid.New())
		payload := map[string]string{"key": "value"}

		result := msg.WithPayload(payload)

		assert.Same(t, msg, result)
		assert.Equal(t, payload, msg.Payload)
	})

	t.Run("accepts nil payload", func(t *testing.T) {
		msg := NewMessage(MsgTypeRecordDeleted, uuid.New(), uuid.New())

		result := msg.WithPayload(nil)

		assert.Same(t, msg, result)
		assert.Nil(t, msg.Payload)
	})
}

func TestMessageChaining(t *testing.T) {
	t.Run("supports method chaining", func(t *testing.T) {
		baseID := uuid.New()
		userID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		payload := map[string]interface{}{"data": "test"}

		msg := NewMessage(MsgTypeRecordUpdated, baseID, userID).
			WithTable(tableID).
			WithRecord(recordID).
			WithPayload(payload)

		assert.Equal(t, MsgTypeRecordUpdated, msg.Type)
		assert.Equal(t, baseID, msg.BaseID)
		assert.Equal(t, userID, msg.UserID)
		require.NotNil(t, msg.TableID)
		assert.Equal(t, tableID, *msg.TableID)
		require.NotNil(t, msg.RecordID)
		assert.Equal(t, recordID, *msg.RecordID)
		assert.Equal(t, payload, msg.Payload)
	})
}

func TestMessageJSON(t *testing.T) {
	t.Run("serializes to JSON correctly", func(t *testing.T) {
		baseID := uuid.New()
		userID := uuid.New()
		tableID := uuid.New()

		msg := NewMessage(MsgTypeRecordCreated, baseID, userID).
			WithTable(tableID).
			WithPayload(map[string]string{"name": "Test"})

		jsonData, err := json.Marshal(msg)
		require.NoError(t, err)

		assert.Contains(t, string(jsonData), `"type":"record_created"`)
		assert.Contains(t, string(jsonData), baseID.String())
		assert.Contains(t, string(jsonData), userID.String())
		assert.Contains(t, string(jsonData), tableID.String())
	})

	t.Run("omits empty optional fields", func(t *testing.T) {
		msg := NewMessage(MsgTypeUserJoined, uuid.New(), uuid.New())

		jsonData, err := json.Marshal(msg)
		require.NoError(t, err)

		assert.NotContains(t, string(jsonData), `"tableId"`)
		assert.NotContains(t, string(jsonData), `"recordId"`)
		assert.NotContains(t, string(jsonData), `"fieldId"`)
		assert.NotContains(t, string(jsonData), `"viewId"`)
		assert.NotContains(t, string(jsonData), `"payload"`)
	})
}

func TestUserPresence(t *testing.T) {
	t.Run("creates valid user presence", func(t *testing.T) {
		userID := uuid.New()
		tableID := uuid.New()
		viewID := uuid.New()
		name := "Test User"

		presence := &UserPresence{
			UserID:    userID,
			Email:     "test@example.com",
			Name:      &name,
			TableID:   &tableID,
			ViewID:    &viewID,
			JoinedAt:  time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		assert.Equal(t, userID, presence.UserID)
		assert.Equal(t, "test@example.com", presence.Email)
		require.NotNil(t, presence.Name)
		assert.Equal(t, "Test User", *presence.Name)
		require.NotNil(t, presence.TableID)
		assert.Equal(t, tableID, *presence.TableID)
	})

	t.Run("supports cell reference", func(t *testing.T) {
		presence := &UserPresence{
			UserID: uuid.New(),
			Email:  "test@example.com",
			CellRef: &CellRef{
				RecordID: uuid.New(),
				FieldID:  uuid.New(),
			},
			JoinedAt:  time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}

		require.NotNil(t, presence.CellRef)
		assert.NotEqual(t, uuid.Nil, presence.CellRef.RecordID)
		assert.NotEqual(t, uuid.Nil, presence.CellRef.FieldID)
	})
}

func TestCellRef(t *testing.T) {
	t.Run("creates valid cell reference", func(t *testing.T) {
		recordID := uuid.New()
		fieldID := uuid.New()

		cellRef := &CellRef{
			RecordID: recordID,
			FieldID:  fieldID,
		}

		assert.Equal(t, recordID, cellRef.RecordID)
		assert.Equal(t, fieldID, cellRef.FieldID)
	})

	t.Run("serializes to JSON correctly", func(t *testing.T) {
		recordID := uuid.New()
		fieldID := uuid.New()

		cellRef := &CellRef{
			RecordID: recordID,
			FieldID:  fieldID,
		}

		jsonData, err := json.Marshal(cellRef)
		require.NoError(t, err)

		assert.Contains(t, string(jsonData), recordID.String())
		assert.Contains(t, string(jsonData), fieldID.String())
	})
}

func TestCursorUpdate(t *testing.T) {
	t.Run("creates valid cursor update", func(t *testing.T) {
		tableID := uuid.New()
		viewID := uuid.New()

		cursor := &CursorUpdate{
			TableID: tableID,
			ViewID:  viewID,
			CellRef: &CellRef{
				RecordID: uuid.New(),
				FieldID:  uuid.New(),
			},
		}

		assert.Equal(t, tableID, cursor.TableID)
		assert.Equal(t, viewID, cursor.ViewID)
		require.NotNil(t, cursor.CellRef)
	})

	t.Run("allows nil cell reference", func(t *testing.T) {
		cursor := &CursorUpdate{
			TableID: uuid.New(),
		}

		assert.Nil(t, cursor.CellRef)
	})
}

func TestIncomingMessage(t *testing.T) {
	t.Run("deserializes from JSON", func(t *testing.T) {
		jsonData := []byte(`{"type":"cursor","payload":{"tableId":"123"}}`)

		var msg IncomingMessage
		err := json.Unmarshal(jsonData, &msg)
		require.NoError(t, err)

		assert.Equal(t, "cursor", msg.Type)
		assert.NotNil(t, msg.Payload)
	})
}

func TestSubscribeMessage(t *testing.T) {
	t.Run("deserializes from JSON", func(t *testing.T) {
		baseID := uuid.New()
		jsonData := []byte(`{"baseId":"` + baseID.String() + `"}`)

		var msg SubscribeMessage
		err := json.Unmarshal(jsonData, &msg)
		require.NoError(t, err)

		assert.Equal(t, baseID, msg.BaseID)
	})
}
