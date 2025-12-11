package models

import (
	"time"

	"github.com/google/uuid"
)

// WebhookEvent represents event types that can trigger webhooks
type WebhookEvent string

const (
	WebhookEventRecordCreated WebhookEvent = "record.created"
	WebhookEventRecordUpdated WebhookEvent = "record.updated"
	WebhookEventRecordDeleted WebhookEvent = "record.deleted"
)

// ValidWebhookEvents returns all valid webhook events
func ValidWebhookEvents() []WebhookEvent {
	return []WebhookEvent{
		WebhookEventRecordCreated,
		WebhookEventRecordUpdated,
		WebhookEventRecordDeleted,
	}
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID        uuid.UUID      `json:"id"`
	BaseID    uuid.UUID      `json:"base_id"`
	Name      string         `json:"name"`
	URL       string         `json:"url"`
	Events    []WebhookEvent `json:"events"`
	Secret    *string        `json:"secret,omitempty"` // For HMAC verification
	IsActive  bool           `json:"is_active"`
	CreatedBy uuid.UUID      `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// CreateWebhookRequest represents a request to create a webhook
type CreateWebhookRequest struct {
	Name   string         `json:"name"`
	URL    string         `json:"url"`
	Events []WebhookEvent `json:"events,omitempty"`
	Secret *string        `json:"secret,omitempty"`
}

// UpdateWebhookRequest represents a request to update a webhook
type UpdateWebhookRequest struct {
	Name     *string        `json:"name,omitempty"`
	URL      *string        `json:"url,omitempty"`
	Events   []WebhookEvent `json:"events,omitempty"`
	Secret   *string        `json:"secret,omitempty"`
	IsActive *bool          `json:"is_active,omitempty"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	ID             uuid.UUID  `json:"id"`
	WebhookID      uuid.UUID  `json:"webhook_id"`
	EventType      string     `json:"event_type"`
	Payload        string     `json:"payload"`
	ResponseStatus *int       `json:"response_status,omitempty"`
	ResponseBody   *string    `json:"response_body,omitempty"`
	Error          *string    `json:"error,omitempty"`
	DurationMs     *int       `json:"duration_ms,omitempty"`
	DeliveredAt    time.Time  `json:"delivered_at"`
}

// WebhookPayload represents the payload sent to webhook endpoints
type WebhookPayload struct {
	Event     WebhookEvent   `json:"event"`
	Timestamp time.Time      `json:"timestamp"`
	BaseID    uuid.UUID      `json:"base_id"`
	TableID   uuid.UUID      `json:"table_id"`
	RecordID  *uuid.UUID     `json:"record_id,omitempty"`
	Record    *Record        `json:"record,omitempty"`
	OldRecord *Record        `json:"old_record,omitempty"`
	UserID    uuid.UUID      `json:"user_id"`
}
