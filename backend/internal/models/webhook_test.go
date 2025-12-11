package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidWebhookEvents(t *testing.T) {
	events := ValidWebhookEvents()

	assert.Len(t, events, 3)
	assert.Contains(t, events, WebhookEventRecordCreated)
	assert.Contains(t, events, WebhookEventRecordUpdated)
	assert.Contains(t, events, WebhookEventRecordDeleted)
}

func TestWebhookEventConstants(t *testing.T) {
	assert.Equal(t, WebhookEvent("record.created"), WebhookEventRecordCreated)
	assert.Equal(t, WebhookEvent("record.updated"), WebhookEventRecordUpdated)
	assert.Equal(t, WebhookEvent("record.deleted"), WebhookEventRecordDeleted)
}
