package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestNewWebhookStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	baseStore := NewBaseStore(mock)
	store := NewWebhookStore(mock, baseStore)
	assert.NotNil(t, store)
}

func TestWebhookStore_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("creates webhook successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		baseID := uuid.New()
		userID := uuid.New()
		webhookID := uuid.New()
		now := time.Now().UTC()
		secret := "test-secret"
		events := []models.WebhookEvent{models.WebhookEventRecordCreated}
		eventsJSON, _ := json.Marshal(events)

		req := &models.CreateWebhookRequest{
			Name:   "Test Webhook",
			URL:    "https://example.com/webhook",
			Events: events,
			Secret: &secret,
		}

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID, baseID, req.Name, req.URL, eventsJSON, &secret, true, userID, now, now)
		mock.ExpectQuery("INSERT INTO webhooks").
			WithArgs(pgxmock.AnyArg(), baseID, req.Name, req.URL, pgxmock.AnyArg(), &secret, userID, pgxmock.AnyArg()).
			WillReturnRows(insertRows)

		webhook, err := store.Create(ctx, baseID, userID, req)
		require.NoError(t, err)
		assert.NotNil(t, webhook)
		assert.Equal(t, req.Name, webhook.Name)
		assert.Equal(t, req.URL, webhook.URL)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creates webhook with default events", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		baseID := uuid.New()
		userID := uuid.New()
		webhookID := uuid.New()
		now := time.Now().UTC()
		defaultEvents := []models.WebhookEvent{
			models.WebhookEventRecordCreated,
			models.WebhookEventRecordUpdated,
			models.WebhookEventRecordDeleted,
		}
		eventsJSON, _ := json.Marshal(defaultEvents)

		req := &models.CreateWebhookRequest{
			Name: "Test Webhook",
			URL:  "https://example.com/webhook",
			// No events - should default
		}

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID, baseID, req.Name, req.URL, eventsJSON, nil, true, userID, now, now)
		mock.ExpectQuery("INSERT INTO webhooks").
			WithArgs(pgxmock.AnyArg(), baseID, req.Name, req.URL, pgxmock.AnyArg(), (*string)(nil), userID, pgxmock.AnyArg()).
			WillReturnRows(insertRows)

		webhook, err := store.Create(ctx, baseID, userID, req)
		require.NoError(t, err)
		assert.NotNil(t, webhook)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookStore_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns webhook when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()
		events := []models.WebhookEvent{models.WebhookEventRecordCreated}
		eventsJSON, _ := json.Marshal(events)

		rows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID, baseID, "Test Webhook", "https://example.com/webhook", eventsJSON, nil, true, userID, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at").
			WithArgs(webhookID).
			WillReturnRows(rows)

		webhook, err := store.GetByID(ctx, webhookID)
		require.NoError(t, err)
		assert.Equal(t, webhookID, webhook.ID)
		assert.Equal(t, baseID, webhook.BaseID)
		assert.Equal(t, "Test Webhook", webhook.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()

		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at").
			WithArgs(webhookID).
			WillReturnError(pgx.ErrNoRows)

		webhook, err := store.GetByID(ctx, webhookID)
		assert.Error(t, err)
		assert.Nil(t, webhook)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookStore_ListByBase(t *testing.T) {
	ctx := context.Background()

	t.Run("returns webhooks for base", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		baseID := uuid.New()
		userID := uuid.New()
		webhookID1 := uuid.New()
		webhookID2 := uuid.New()
		now := time.Now().UTC()
		events := []models.WebhookEvent{models.WebhookEventRecordCreated}
		eventsJSON, _ := json.Marshal(events)

		rows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID1, baseID, "Webhook 1", "https://example.com/1", eventsJSON, nil, true, userID, now, now).
			AddRow(webhookID2, baseID, "Webhook 2", "https://example.com/2", eventsJSON, nil, true, userID, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at").
			WithArgs(baseID).
			WillReturnRows(rows)

		webhooks, err := store.ListByBase(ctx, baseID)
		require.NoError(t, err)
		assert.Len(t, webhooks, 2)
		assert.Equal(t, webhookID1, webhooks[0].ID)
		assert.Equal(t, webhookID2, webhooks[1].ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no webhooks", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		baseID := uuid.New()

		rows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"})
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at").
			WithArgs(baseID).
			WillReturnRows(rows)

		webhooks, err := store.ListByBase(ctx, baseID)
		require.NoError(t, err)
		assert.Empty(t, webhooks)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookStore_GetActiveByBaseAndEvent(t *testing.T) {
	ctx := context.Background()

	t.Run("returns active webhooks matching event", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		baseID := uuid.New()
		userID := uuid.New()
		webhookID := uuid.New()
		now := time.Now().UTC()
		events := []models.WebhookEvent{models.WebhookEventRecordCreated}
		eventsJSON, _ := json.Marshal(events)
		eventQuery, _ := json.Marshal([]models.WebhookEvent{models.WebhookEventRecordCreated})

		rows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID, baseID, "Test Webhook", "https://example.com/webhook", eventsJSON, nil, true, userID, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at").
			WithArgs(baseID, eventQuery).
			WillReturnRows(rows)

		webhooks, err := store.GetActiveByBaseAndEvent(ctx, baseID, models.WebhookEventRecordCreated)
		require.NoError(t, err)
		assert.Len(t, webhooks, 1)
		assert.Equal(t, webhookID, webhooks[0].ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookStore_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("updates webhook successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()
		baseID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()
		events := []models.WebhookEvent{models.WebhookEventRecordCreated}
		eventsJSON, _ := json.Marshal(events)
		newName := "Updated Webhook"

		// Mock GetByID
		getRows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID, baseID, "Old Name", "https://example.com/webhook", eventsJSON, nil, true, userID, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at").
			WithArgs(webhookID).
			WillReturnRows(getRows)

		// Mock Update
		updateRows := pgxmock.NewRows([]string{"updated_at"}).AddRow(now)
		mock.ExpectQuery("UPDATE webhooks").
			WithArgs(webhookID, newName, "https://example.com/webhook", pgxmock.AnyArg(), (*string)(nil), true).
			WillReturnRows(updateRows)

		req := &models.UpdateWebhookRequest{
			Name: &newName,
		}

		webhook, err := store.Update(ctx, webhookID, req)
		require.NoError(t, err)
		assert.Equal(t, newName, webhook.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookStore_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes webhook successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()

		mock.ExpectExec("DELETE FROM webhooks WHERE id").
			WithArgs(webhookID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.Delete(ctx, webhookID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when webhook not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()

		mock.ExpectExec("DELETE FROM webhooks WHERE id").
			WithArgs(webhookID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = store.Delete(ctx, webhookID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookStore_CreateDelivery(t *testing.T) {
	ctx := context.Background()

	t.Run("creates delivery record successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()
		deliveryID := uuid.New()
		now := time.Now().UTC()
		status := 200
		responseBody := "OK"
		durationMs := 150
		payload := `{"test": "data"}`

		delivery := &models.WebhookDelivery{
			WebhookID:      webhookID,
			EventType:      "record_created",
			Payload:        payload,
			ResponseStatus: &status,
			ResponseBody:   &responseBody,
			DurationMs:     &durationMs,
		}

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record_created", payload, &status, &responseBody, nil, &durationMs, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record_created", payload, &status, &responseBody, (*string)(nil), &durationMs, pgxmock.AnyArg()).
			WillReturnRows(insertRows)

		result, err := store.CreateDelivery(ctx, delivery)
		require.NoError(t, err)
		assert.Equal(t, webhookID, result.WebhookID)
		assert.Equal(t, "record_created", result.EventType)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creates delivery record with error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()
		deliveryID := uuid.New()
		now := time.Now().UTC()
		errorMsg := "connection refused"
		durationMs := 5000
		payload := `{"test": "data"}`

		delivery := &models.WebhookDelivery{
			WebhookID:  webhookID,
			EventType:  "record_created",
			Payload:    payload,
			Error:      &errorMsg,
			DurationMs: &durationMs,
		}

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record_created", payload, nil, nil, &errorMsg, &durationMs, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record_created", payload, (*int)(nil), (*string)(nil), &errorMsg, &durationMs, pgxmock.AnyArg()).
			WillReturnRows(insertRows)

		result, err := store.CreateDelivery(ctx, delivery)
		require.NoError(t, err)
		assert.Equal(t, &errorMsg, result.Error)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestWebhookStore_ListDeliveries(t *testing.T) {
	ctx := context.Background()

	t.Run("returns deliveries for webhook", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()
		deliveryID1 := uuid.New()
		deliveryID2 := uuid.New()
		now := time.Now().UTC()
		status := 200
		responseBody := "OK"
		durationMs1 := 100
		durationMs2 := 150
		payload := `{"test": "data"}`

		rows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID1, webhookID, "record_created", payload, &status, &responseBody, nil, &durationMs1, now).
			AddRow(deliveryID2, webhookID, "record_updated", payload, &status, &responseBody, nil, &durationMs2, now)
		mock.ExpectQuery("SELECT id, webhook_id, event_type, payload, response_status, response_body, error, duration_ms, delivered_at").
			WithArgs(webhookID, 50).
			WillReturnRows(rows)

		deliveries, err := store.ListDeliveries(ctx, webhookID, 50)
		require.NoError(t, err)
		assert.Len(t, deliveries, 2)
		assert.Equal(t, deliveryID1, deliveries[0].ID)
		assert.Equal(t, deliveryID2, deliveries[1].ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("uses default limit when zero", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := NewBaseStore(mock)
		store := NewWebhookStore(mock, baseStore)
		webhookID := uuid.New()

		rows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"})
		mock.ExpectQuery("SELECT id, webhook_id, event_type, payload, response_status, response_body, error, duration_ms, delivered_at").
			WithArgs(webhookID, 50).
			WillReturnRows(rows)

		deliveries, err := store.ListDeliveries(ctx, webhookID, 0)
		require.NoError(t, err)
		assert.Empty(t, deliveries)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
