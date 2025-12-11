package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

// DeliveryEngine handles webhook delivery
type DeliveryEngine struct {
	webhookStore *store.WebhookStore
	tableStore   *store.TableStore
	httpClient   *http.Client
}

// NewDeliveryEngine creates a new webhook delivery engine
func NewDeliveryEngine(webhookStore *store.WebhookStore, tableStore *store.TableStore) *DeliveryEngine {
	return &DeliveryEngine{
		webhookStore: webhookStore,
		tableStore:   tableStore,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// DeliveryContext contains the context for a webhook delivery
type DeliveryContext struct {
	BaseID    uuid.UUID
	TableID   uuid.UUID
	RecordID  *uuid.UUID
	Record    *models.Record
	OldRecord *models.Record
	Event     models.WebhookEvent
	UserID    uuid.UUID
}

// ProcessEvent processes a webhook event and delivers to all registered webhooks
func (e *DeliveryEngine) ProcessEvent(ctx context.Context, deliveryCtx *DeliveryContext) {
	// Get webhooks that should receive this event
	webhooks, err := e.webhookStore.GetActiveByBaseAndEvent(ctx, deliveryCtx.BaseID, deliveryCtx.Event)
	if err != nil {
		log.Printf("Failed to get webhooks for event %s: %v", deliveryCtx.Event, err)
		return
	}

	if len(webhooks) == 0 {
		return
	}

	// Build the payload
	payload := &models.WebhookPayload{
		Event:     deliveryCtx.Event,
		Timestamp: time.Now(),
		BaseID:    deliveryCtx.BaseID,
		TableID:   deliveryCtx.TableID,
		RecordID:  deliveryCtx.RecordID,
		Record:    deliveryCtx.Record,
		OldRecord: deliveryCtx.OldRecord,
		UserID:    deliveryCtx.UserID,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal webhook payload: %v", err)
		return
	}

	// Deliver to each webhook asynchronously
	for _, webhook := range webhooks {
		go e.deliverToWebhook(ctx, webhook, string(deliveryCtx.Event), payloadJSON)
	}
}

// deliverToWebhook delivers a payload to a single webhook
func (e *DeliveryEngine) deliverToWebhook(ctx context.Context, webhook models.Webhook, eventType string, payloadJSON []byte) {
	startTime := time.Now()

	delivery := &models.WebhookDelivery{
		WebhookID: webhook.ID,
		EventType: eventType,
		Payload:   string(payloadJSON),
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.URL, bytes.NewReader(payloadJSON))
	if err != nil {
		errStr := fmt.Sprintf("failed to create request: %v", err)
		delivery.Error = &errStr
		e.recordDelivery(ctx, delivery, startTime)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "VibeTable-Webhook/1.0")
	req.Header.Set("X-Webhook-Event", eventType)
	req.Header.Set("X-Webhook-ID", webhook.ID.String())
	req.Header.Set("X-Webhook-Timestamp", fmt.Sprintf("%d", time.Now().Unix()))

	// Add HMAC signature if secret is configured
	if webhook.Secret != nil && *webhook.Secret != "" {
		signature := computeHMAC(payloadJSON, *webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Send the request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		errStr := fmt.Sprintf("request failed: %v", err)
		delivery.Error = &errStr
		e.recordDelivery(ctx, delivery, startTime)
		log.Printf("Webhook delivery failed for %s: %v", webhook.URL, err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 10240)) // Limit to 10KB
	bodyStr := string(body)

	delivery.ResponseStatus = &resp.StatusCode
	delivery.ResponseBody = &bodyStr

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("Webhook delivered successfully to %s (status: %d)", webhook.URL, resp.StatusCode)
	} else {
		errStr := fmt.Sprintf("non-success status code: %d", resp.StatusCode)
		delivery.Error = &errStr
		log.Printf("Webhook delivery failed for %s: status %d", webhook.URL, resp.StatusCode)
	}

	e.recordDelivery(ctx, delivery, startTime)
}

// recordDelivery records a delivery attempt to the database
func (e *DeliveryEngine) recordDelivery(ctx context.Context, delivery *models.WebhookDelivery, startTime time.Time) {
	durationMs := int(time.Since(startTime).Milliseconds())
	delivery.DurationMs = &durationMs

	if _, err := e.webhookStore.CreateDelivery(ctx, delivery); err != nil {
		log.Printf("Failed to record webhook delivery: %v", err)
	}
}

// computeHMAC computes an HMAC-SHA256 signature
func computeHMAC(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return "sha256=" + hex.EncodeToString(h.Sum(nil))
}
