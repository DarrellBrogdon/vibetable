package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
)

type WebhookStore struct {
	db        DBTX
	baseStore *BaseStore
}

func NewWebhookStore(db DBTX, baseStore *BaseStore) *WebhookStore {
	return &WebhookStore{db: db, baseStore: baseStore}
}

// Create creates a new webhook
func (s *WebhookStore) Create(ctx context.Context, baseID, userID uuid.UUID, req *models.CreateWebhookRequest) (*models.Webhook, error) {
	// Default events if not provided
	events := req.Events
	if len(events) == 0 {
		events = []models.WebhookEvent{
			models.WebhookEventRecordCreated,
			models.WebhookEventRecordUpdated,
			models.WebhookEventRecordDeleted,
		}
	}

	eventsJSON, err := json.Marshal(events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO webhooks (id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, $7, $8, $8)
		RETURNING id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at
	`

	webhook := &models.Webhook{}
	var eventsRaw []byte

	err = s.db.QueryRow(ctx, query,
		id, baseID, req.Name, req.URL, eventsJSON, req.Secret, userID, now,
	).Scan(
		&webhook.ID, &webhook.BaseID, &webhook.Name, &webhook.URL,
		&eventsRaw, &webhook.Secret, &webhook.IsActive, &webhook.CreatedBy,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	if err := json.Unmarshal(eventsRaw, &webhook.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	return webhook, nil
}

// GetByID retrieves a webhook by ID
func (s *WebhookStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Webhook, error) {
	query := `
		SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at
		FROM webhooks WHERE id = $1
	`

	webhook := &models.Webhook{}
	var eventsRaw []byte

	err := s.db.QueryRow(ctx, query, id).Scan(
		&webhook.ID, &webhook.BaseID, &webhook.Name, &webhook.URL,
		&eventsRaw, &webhook.Secret, &webhook.IsActive, &webhook.CreatedBy,
		&webhook.CreatedAt, &webhook.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(eventsRaw, &webhook.Events); err != nil {
		return nil, fmt.Errorf("failed to unmarshal events: %w", err)
	}

	return webhook, nil
}

// ListByBase lists all webhooks for a base
func (s *WebhookStore) ListByBase(ctx context.Context, baseID uuid.UUID) ([]models.Webhook, error) {
	query := `
		SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at
		FROM webhooks WHERE base_id = $1 ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, baseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var webhook models.Webhook
		var eventsRaw []byte

		if err := rows.Scan(
			&webhook.ID, &webhook.BaseID, &webhook.Name, &webhook.URL,
			&eventsRaw, &webhook.Secret, &webhook.IsActive, &webhook.CreatedBy,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(eventsRaw, &webhook.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks, rows.Err()
}

// GetActiveByBaseAndEvent gets active webhooks for a base that listen to a specific event
func (s *WebhookStore) GetActiveByBaseAndEvent(ctx context.Context, baseID uuid.UUID, event models.WebhookEvent) ([]models.Webhook, error) {
	query := `
		SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at
		FROM webhooks
		WHERE base_id = $1 AND is_active = true AND events @> $2
		ORDER BY created_at ASC
	`

	eventJSON, _ := json.Marshal([]models.WebhookEvent{event})

	rows, err := s.db.Query(ctx, query, baseID, eventJSON)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []models.Webhook
	for rows.Next() {
		var webhook models.Webhook
		var eventsRaw []byte

		if err := rows.Scan(
			&webhook.ID, &webhook.BaseID, &webhook.Name, &webhook.URL,
			&eventsRaw, &webhook.Secret, &webhook.IsActive, &webhook.CreatedBy,
			&webhook.CreatedAt, &webhook.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(eventsRaw, &webhook.Events); err != nil {
			return nil, fmt.Errorf("failed to unmarshal events: %w", err)
		}

		webhooks = append(webhooks, webhook)
	}

	return webhooks, rows.Err()
}

// Update updates a webhook
func (s *WebhookStore) Update(ctx context.Context, id uuid.UUID, req *models.UpdateWebhookRequest) (*models.Webhook, error) {
	// Get existing webhook
	webhook, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Name != nil {
		webhook.Name = *req.Name
	}
	if req.URL != nil {
		webhook.URL = *req.URL
	}
	if req.Events != nil {
		webhook.Events = req.Events
	}
	if req.Secret != nil {
		webhook.Secret = req.Secret
	}
	if req.IsActive != nil {
		webhook.IsActive = *req.IsActive
	}

	eventsJSON, err := json.Marshal(webhook.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	query := `
		UPDATE webhooks
		SET name = $2, url = $3, events = $4, secret = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err = s.db.QueryRow(ctx, query,
		id, webhook.Name, webhook.URL, eventsJSON, webhook.Secret, webhook.IsActive,
	).Scan(&webhook.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}

	return webhook, nil
}

// Delete deletes a webhook
func (s *WebhookStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM webhooks WHERE id = $1`
	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("webhook not found")
	}

	return nil
}

// CreateDelivery records a webhook delivery attempt
func (s *WebhookStore) CreateDelivery(ctx context.Context, delivery *models.WebhookDelivery) (*models.WebhookDelivery, error) {
	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO webhook_deliveries (id, webhook_id, event_type, payload, response_status, response_body, error, duration_ms, delivered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, webhook_id, event_type, payload, response_status, response_body, error, duration_ms, delivered_at
	`

	d := &models.WebhookDelivery{}
	err := s.db.QueryRow(ctx, query,
		id, delivery.WebhookID, delivery.EventType, delivery.Payload,
		delivery.ResponseStatus, delivery.ResponseBody, delivery.Error, delivery.DurationMs, now,
	).Scan(
		&d.ID, &d.WebhookID, &d.EventType, &d.Payload,
		&d.ResponseStatus, &d.ResponseBody, &d.Error, &d.DurationMs, &d.DeliveredAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create delivery record: %w", err)
	}

	return d, nil
}

// ListDeliveries lists recent deliveries for a webhook
func (s *WebhookStore) ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit int) ([]models.WebhookDelivery, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, webhook_id, event_type, payload, response_status, response_body, error, duration_ms, delivered_at
		FROM webhook_deliveries WHERE webhook_id = $1
		ORDER BY delivered_at DESC LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, webhookID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []models.WebhookDelivery
	for rows.Next() {
		var d models.WebhookDelivery
		if err := rows.Scan(
			&d.ID, &d.WebhookID, &d.EventType, &d.Payload,
			&d.ResponseStatus, &d.ResponseBody, &d.Error, &d.DurationMs, &d.DeliveredAt,
		); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}

	return deliveries, rows.Err()
}
