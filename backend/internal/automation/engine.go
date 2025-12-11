package automation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

// Engine handles automation execution
type Engine struct {
	automationStore *store.AutomationStore
	recordStore     *store.RecordStore
	fieldStore      *store.FieldStore
	httpClient      *http.Client
}

// NewEngine creates a new automation engine
func NewEngine(automationStore *store.AutomationStore, recordStore *store.RecordStore, fieldStore *store.FieldStore) *Engine {
	return &Engine{
		automationStore: automationStore,
		recordStore:     recordStore,
		fieldStore:      fieldStore,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TriggerContext contains information about what triggered the automation
type TriggerContext struct {
	TableID     uuid.UUID
	RecordID    *uuid.UUID
	Record      *models.Record
	OldRecord   *models.Record // For updates
	TriggerType models.TriggerType
	UserID      uuid.UUID
}

// ProcessTrigger finds and executes all matching automations
func (e *Engine) ProcessTrigger(ctx context.Context, triggerCtx *TriggerContext) {
	// Get all enabled automations for this trigger type and table
	automations, err := e.automationStore.GetAutomationsByTrigger(ctx, triggerCtx.TableID, triggerCtx.TriggerType)
	if err != nil {
		log.Printf("[Automation] Error fetching automations: %v", err)
		return
	}

	for _, automation := range automations {
		go e.executeAutomation(ctx, automation, triggerCtx)
	}
}

// executeAutomation runs a single automation
func (e *Engine) executeAutomation(ctx context.Context, automation models.Automation, triggerCtx *TriggerContext) {
	log.Printf("[Automation] Executing: %s (trigger: %s)", automation.Name, automation.TriggerType)

	// Check trigger conditions if applicable
	if !e.checkTriggerConditions(automation, triggerCtx) {
		log.Printf("[Automation] Trigger conditions not met for: %s", automation.Name)
		return
	}

	// Create run record
	triggerData, _ := json.Marshal(map[string]interface{}{
		"recordId": triggerCtx.RecordID,
		"userId":   triggerCtx.UserID,
	})

	run := &models.AutomationRun{
		AutomationID:    automation.ID,
		Status:          models.RunStatusRunning,
		TriggerRecordID: triggerCtx.RecordID,
		TriggerData:     triggerData,
	}

	run, err := e.automationStore.CreateRun(ctx, run)
	if err != nil {
		log.Printf("[Automation] Error creating run record: %v", err)
		return
	}

	// Execute the action
	result, execErr := e.executeAction(ctx, automation, triggerCtx)

	// Update run status
	var errMsg *string
	status := models.RunStatusSuccess
	if execErr != nil {
		status = models.RunStatusFailed
		msg := execErr.Error()
		errMsg = &msg
		log.Printf("[Automation] Execution failed for %s: %v", automation.Name, execErr)
	} else {
		log.Printf("[Automation] Execution succeeded for: %s", automation.Name)
	}

	resultJSON, _ := json.Marshal(result)
	e.automationStore.UpdateRun(ctx, run.ID, status, resultJSON, errMsg)
	e.automationStore.UpdateAutomationStats(ctx, automation.ID)
}

// checkTriggerConditions checks if trigger-specific conditions are met
func (e *Engine) checkTriggerConditions(automation models.Automation, triggerCtx *TriggerContext) bool {
	if automation.TriggerType != models.TriggerFieldValueChanged {
		return true // No conditions for other trigger types
	}

	var config models.FieldValueChangedConfig
	if err := json.Unmarshal(automation.TriggerConfig, &config); err != nil {
		return false
	}

	if triggerCtx.Record == nil {
		return false
	}

	// Get the field value from the record
	var recordValues map[string]interface{}
	if err := json.Unmarshal(triggerCtx.Record.Values, &recordValues); err != nil {
		return false
	}

	fieldValue, exists := recordValues[config.FieldID.String()]
	if !exists {
		fieldValue = nil
	}

	// Check operator conditions
	switch config.Operator {
	case "equals":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", config.Value)
	case "not_equals":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", config.Value)
	case "is_empty":
		return fieldValue == nil || fieldValue == ""
	case "is_not_empty":
		return fieldValue != nil && fieldValue != ""
	case "contains":
		return strings.Contains(fmt.Sprintf("%v", fieldValue), fmt.Sprintf("%v", config.Value))
	default:
		// If no operator specified, just check if the field changed
		if triggerCtx.OldRecord != nil {
			var oldValues map[string]interface{}
			if err := json.Unmarshal(triggerCtx.OldRecord.Values, &oldValues); err != nil {
				return true
			}
			oldValue := oldValues[config.FieldID.String()]
			return fmt.Sprintf("%v", oldValue) != fmt.Sprintf("%v", fieldValue)
		}
		return true
	}
}

// executeAction executes the automation's action
func (e *Engine) executeAction(ctx context.Context, automation models.Automation, triggerCtx *TriggerContext) (interface{}, error) {
	switch automation.ActionType {
	case models.ActionSendEmail:
		return e.executeSendEmail(ctx, automation, triggerCtx)
	case models.ActionUpdateRecord:
		return e.executeUpdateRecord(ctx, automation, triggerCtx)
	case models.ActionCreateRecord:
		return e.executeCreateRecord(ctx, automation, triggerCtx)
	case models.ActionSendWebhook:
		return e.executeSendWebhook(ctx, automation, triggerCtx)
	default:
		return nil, fmt.Errorf("unknown action type: %s", automation.ActionType)
	}
}

// executeSendEmail sends an email (for now, just logs it)
func (e *Engine) executeSendEmail(ctx context.Context, automation models.Automation, triggerCtx *TriggerContext) (interface{}, error) {
	var config models.SendEmailConfig
	if err := json.Unmarshal(automation.ActionConfig, &config); err != nil {
		return nil, fmt.Errorf("invalid email config: %w", err)
	}

	// Resolve field references in the config
	to := e.resolveFieldReferences(config.To, triggerCtx)
	subject := e.resolveFieldReferences(config.Subject, triggerCtx)
	body := e.resolveFieldReferences(config.Body, triggerCtx)

	// For now, just log the email (in production, use Resend/SendGrid/etc.)
	log.Printf("[Automation] Would send email: to=%s, subject=%s, body=%s", to, subject, body)

	return map[string]string{
		"to":      to,
		"subject": subject,
		"status":  "logged",
	}, nil
}

// executeUpdateRecord updates the triggering record
func (e *Engine) executeUpdateRecord(ctx context.Context, automation models.Automation, triggerCtx *TriggerContext) (interface{}, error) {
	if triggerCtx.RecordID == nil {
		return nil, fmt.Errorf("no record to update")
	}

	var config models.UpdateRecordConfig
	if err := json.Unmarshal(automation.ActionConfig, &config); err != nil {
		return nil, fmt.Errorf("invalid update config: %w", err)
	}

	// Build the values map
	values := make(map[string]interface{})
	for _, update := range config.Updates {
		value := update.Value
		// If value is a string, resolve field references
		if strVal, ok := value.(string); ok {
			value = e.resolveFieldReferences(strVal, triggerCtx)
		}
		values[update.FieldID.String()] = value
	}

	// Use the automation creator as the user for the update
	record, err := e.recordStore.PatchRecordValues(ctx, *triggerCtx.RecordID, values, automation.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	return record, nil
}

// executeCreateRecord creates a new record
func (e *Engine) executeCreateRecord(ctx context.Context, automation models.Automation, triggerCtx *TriggerContext) (interface{}, error) {
	var config models.CreateRecordConfig
	if err := json.Unmarshal(automation.ActionConfig, &config); err != nil {
		return nil, fmt.Errorf("invalid create config: %w", err)
	}

	// Build the values map
	values := make(map[string]interface{})
	for _, update := range config.Values {
		value := update.Value
		// If value is a string, resolve field references
		if strVal, ok := value.(string); ok {
			value = e.resolveFieldReferences(strVal, triggerCtx)
		}
		values[update.FieldID.String()] = value
	}

	valuesJSON, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	// Create the record in the target table
	record, err := e.recordStore.CreateRecord(ctx, config.TargetTableID, valuesJSON, automation.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	return record, nil
}

// executeSendWebhook sends an HTTP request to a webhook URL
func (e *Engine) executeSendWebhook(ctx context.Context, automation models.Automation, triggerCtx *TriggerContext) (interface{}, error) {
	var config models.SendWebhookConfig
	if err := json.Unmarshal(automation.ActionConfig, &config); err != nil {
		return nil, fmt.Errorf("invalid webhook config: %w", err)
	}

	// Resolve field references in the body
	body := e.resolveFieldReferences(config.Body, triggerCtx)

	// If no custom body, use the record as JSON
	if body == "" && triggerCtx.Record != nil {
		bodyBytes, _ := json.Marshal(triggerCtx.Record)
		body = string(bodyBytes)
	}

	method := config.Method
	if method == "" {
		method = "POST"
	}

	req, err := http.NewRequestWithContext(ctx, method, config.URL, bytes.NewBufferString(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	return map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"body":       string(respBody),
	}, nil
}

// resolveFieldReferences replaces {{field:fieldId}} with actual values
func (e *Engine) resolveFieldReferences(template string, triggerCtx *TriggerContext) string {
	if triggerCtx.Record == nil {
		return template
	}

	var recordValues map[string]interface{}
	if err := json.Unmarshal(triggerCtx.Record.Values, &recordValues); err != nil {
		return template
	}

	// Match {{field:uuid}} pattern
	re := regexp.MustCompile(`\{\{field:([a-f0-9-]+)\}\}`)
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		fieldID := re.FindStringSubmatch(match)[1]
		if value, exists := recordValues[fieldID]; exists {
			return fmt.Sprintf("%v", value)
		}
		return ""
	})

	// Also support {{recordId}}
	if triggerCtx.RecordID != nil {
		result = strings.ReplaceAll(result, "{{recordId}}", triggerCtx.RecordID.String())
	}

	return result
}
