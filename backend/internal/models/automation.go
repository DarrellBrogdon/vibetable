package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TriggerType defines the type of event that triggers an automation
type TriggerType string

const (
	TriggerRecordCreated     TriggerType = "record_created"
	TriggerRecordUpdated     TriggerType = "record_updated"
	TriggerRecordDeleted     TriggerType = "record_deleted"
	TriggerFieldValueChanged TriggerType = "field_value_changed"
	TriggerScheduled         TriggerType = "scheduled" // Future: cron-based triggers
)

// ActionType defines the type of action to perform
type ActionType string

const (
	ActionSendEmail    ActionType = "send_email"
	ActionUpdateRecord ActionType = "update_record"
	ActionCreateRecord ActionType = "create_record"
	ActionSendWebhook  ActionType = "send_webhook"
)

// RunStatus represents the status of an automation run
type RunStatus string

const (
	RunStatusPending RunStatus = "pending"
	RunStatusRunning RunStatus = "running"
	RunStatusSuccess RunStatus = "success"
	RunStatusFailed  RunStatus = "failed"
)

// Automation represents an automation configuration
type Automation struct {
	ID              uuid.UUID       `json:"id"`
	BaseID          uuid.UUID       `json:"baseId"`
	TableID         uuid.UUID       `json:"tableId"`
	Name            string          `json:"name"`
	Description     *string         `json:"description,omitempty"`
	Enabled         bool            `json:"enabled"`
	TriggerType     TriggerType     `json:"triggerType"`
	TriggerConfig   json.RawMessage `json:"triggerConfig"`
	ActionType      ActionType      `json:"actionType"`
	ActionConfig    json.RawMessage `json:"actionConfig"`
	CreatedBy       uuid.UUID       `json:"createdBy"`
	LastTriggeredAt *time.Time      `json:"lastTriggeredAt,omitempty"`
	RunCount        int             `json:"runCount"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

// AutomationRun represents a single execution of an automation
type AutomationRun struct {
	ID              uuid.UUID       `json:"id"`
	AutomationID    uuid.UUID       `json:"automationId"`
	Status          RunStatus       `json:"status"`
	TriggerRecordID *uuid.UUID      `json:"triggerRecordId,omitempty"`
	TriggerData     json.RawMessage `json:"triggerData,omitempty"`
	Result          json.RawMessage `json:"result,omitempty"`
	Error           *string         `json:"error,omitempty"`
	StartedAt       time.Time       `json:"startedAt"`
	CompletedAt     *time.Time      `json:"completedAt,omitempty"`
}

// TriggerConfig types for each trigger type

// FieldValueChangedConfig specifies which field to watch
type FieldValueChangedConfig struct {
	FieldID  uuid.UUID `json:"fieldId"`
	Operator string    `json:"operator,omitempty"` // equals, not_equals, contains, is_empty, is_not_empty
	Value    any       `json:"value,omitempty"`    // Optional: only trigger when value matches
}

// ScheduledConfig for scheduled triggers (future)
type ScheduledConfig struct {
	CronExpression string `json:"cronExpression"`
	Timezone       string `json:"timezone"`
}

// ActionConfig types for each action type

// SendEmailConfig configures email sending
type SendEmailConfig struct {
	To      string `json:"to"`      // Can be static email or field reference like {{field:email_field_id}}
	Subject string `json:"subject"` // Can include field references
	Body    string `json:"body"`    // Can include field references
}

// UpdateRecordConfig configures record updates
type UpdateRecordConfig struct {
	Updates []FieldUpdate `json:"updates"`
}

// FieldUpdate specifies a field value update
type FieldUpdate struct {
	FieldID uuid.UUID `json:"fieldId"`
	Value   any       `json:"value"` // Can be static value or expression
}

// CreateRecordConfig configures new record creation
type CreateRecordConfig struct {
	TargetTableID uuid.UUID     `json:"targetTableId"`
	Values        []FieldUpdate `json:"values"`
}

// SendWebhookConfig configures webhook calls
type SendWebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"` // POST, PUT, PATCH
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"` // JSON template with field references
}

// AutomationWithRuns includes recent run history
type AutomationWithRuns struct {
	Automation
	RecentRuns []AutomationRun `json:"recentRuns,omitempty"`
}
