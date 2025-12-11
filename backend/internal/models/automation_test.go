package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTriggerTypeConstants(t *testing.T) {
	assert.Equal(t, TriggerType("record_created"), TriggerRecordCreated)
	assert.Equal(t, TriggerType("record_updated"), TriggerRecordUpdated)
	assert.Equal(t, TriggerType("record_deleted"), TriggerRecordDeleted)
	assert.Equal(t, TriggerType("field_value_changed"), TriggerFieldValueChanged)
	assert.Equal(t, TriggerType("scheduled"), TriggerScheduled)
}

func TestActionTypeConstants(t *testing.T) {
	assert.Equal(t, ActionType("send_email"), ActionSendEmail)
	assert.Equal(t, ActionType("update_record"), ActionUpdateRecord)
	assert.Equal(t, ActionType("create_record"), ActionCreateRecord)
	assert.Equal(t, ActionType("send_webhook"), ActionSendWebhook)
}

func TestRunStatusConstants(t *testing.T) {
	assert.Equal(t, RunStatus("pending"), RunStatusPending)
	assert.Equal(t, RunStatus("running"), RunStatusRunning)
	assert.Equal(t, RunStatus("success"), RunStatusSuccess)
	assert.Equal(t, RunStatus("failed"), RunStatusFailed)
}
