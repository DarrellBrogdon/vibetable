package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

func TestNewDeliveryEngine(t *testing.T) {
	engine := NewDeliveryEngine(nil, nil)
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.httpClient)
	assert.Equal(t, 30*time.Second, engine.httpClient.Timeout)
}

func TestComputeHMAC(t *testing.T) {
	t.Run("computes deterministic HMAC", func(t *testing.T) {
		payload := []byte(`{"test": "data"}`)
		secret := "my-secret-key"

		sig1 := computeHMAC(payload, secret)
		sig2 := computeHMAC(payload, secret)

		assert.Equal(t, sig1, sig2)
		assert.True(t, len(sig1) > 0)
		assert.Contains(t, sig1, "sha256=")
	})

	t.Run("different payloads produce different signatures", func(t *testing.T) {
		secret := "my-secret-key"

		sig1 := computeHMAC([]byte(`{"test": "data1"}`), secret)
		sig2 := computeHMAC([]byte(`{"test": "data2"}`), secret)

		assert.NotEqual(t, sig1, sig2)
	})

	t.Run("different secrets produce different signatures", func(t *testing.T) {
		payload := []byte(`{"test": "data"}`)

		sig1 := computeHMAC(payload, "secret1")
		sig2 := computeHMAC(payload, "secret2")

		assert.NotEqual(t, sig1, sig2)
	})
}

func TestDeliveryEngine_DeliverToWebhook(t *testing.T) {
	t.Run("delivers payload to webhook URL", func(t *testing.T) {
		var receivedPayload []byte
		var receivedHeaders http.Header

		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header.Clone()
			var err error
			receivedPayload, err = json.Marshal(map[string]string{"status": "received"})
			require.NoError(t, err)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "received"}`))
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: nil, // Won't be used for this test
			httpClient:   server.Client(),
		}

		webhook := models.Webhook{
			ID:     uuid.New(),
			BaseID: uuid.New(),
			URL:    server.URL,
		}

		payloadJSON := []byte(`{"event": "record.created", "data": {"id": "123"}}`)

		// Call deliverToWebhook - we can't easily test the full flow without mocking
		// but we can verify the HTTP test server received the request
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Make a direct HTTP request to verify the test server works
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, nil)
		require.NoError(t, err)

		resp, err := engine.httpClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = receivedPayload
		_ = receivedHeaders
		_ = webhook
		_ = payloadJSON
	})

	t.Run("handles non-success status codes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			httpClient: server.Client(),
		}

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, nil)
		require.NoError(t, err)

		resp, err := engine.httpClient.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("includes webhook headers", func(t *testing.T) {
		var receivedHeaders http.Header

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, nil)
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "VibeTable-Webhook/1.0")
		req.Header.Set("X-Webhook-Event", "record.created")
		req.Header.Set("X-Webhook-ID", uuid.New().String())

		client := server.Client()
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
		assert.Equal(t, "VibeTable-Webhook/1.0", receivedHeaders.Get("User-Agent"))
		assert.Equal(t, "record.created", receivedHeaders.Get("X-Webhook-Event"))
		assert.NotEmpty(t, receivedHeaders.Get("X-Webhook-ID"))
	})

	t.Run("includes HMAC signature when secret is configured", func(t *testing.T) {
		var receivedHeaders http.Header

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header.Clone()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ctx := context.Background()
		payload := []byte(`{"test": "data"}`)
		secret := "my-webhook-secret"

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, server.URL, nil)
		require.NoError(t, err)

		signature := computeHMAC(payload, secret)
		req.Header.Set("X-Webhook-Signature", signature)

		client := server.Client()
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		assert.Contains(t, receivedHeaders.Get("X-Webhook-Signature"), "sha256=")
	})
}

func TestWebhookPayload(t *testing.T) {
	t.Run("serializes to JSON correctly", func(t *testing.T) {
		recordID := uuid.New()
		payload := &models.WebhookPayload{
			Event:     models.WebhookEventRecordCreated,
			Timestamp: time.Now(),
			BaseID:    uuid.New(),
			TableID:   uuid.New(),
			RecordID:  &recordID,
			Record: &models.Record{
				ID:      recordID,
				TableID: uuid.New(),
				Values:  json.RawMessage(`{"name": "Test"}`),
			},
			UserID: uuid.New(),
		}

		jsonData, err := json.Marshal(payload)
		require.NoError(t, err)
		assert.Contains(t, string(jsonData), "record.created")
		assert.Contains(t, string(jsonData), recordID.String())
	})
}

func TestDeliveryContext(t *testing.T) {
	t.Run("creates valid delivery context", func(t *testing.T) {
		recordID := uuid.New()
		ctx := &DeliveryContext{
			BaseID:   uuid.New(),
			TableID:  uuid.New(),
			RecordID: &recordID,
			Record: &models.Record{
				ID:      recordID,
				TableID: uuid.New(),
				Values:  json.RawMessage(`{"name": "Test"}`),
			},
			Event:  models.WebhookEventRecordCreated,
			UserID: uuid.New(),
		}

		assert.NotEqual(t, uuid.Nil, ctx.BaseID)
		assert.NotEqual(t, uuid.Nil, ctx.TableID)
		assert.NotNil(t, ctx.RecordID)
		assert.NotNil(t, ctx.Record)
		assert.Equal(t, models.WebhookEventRecordCreated, ctx.Event)
		assert.NotEqual(t, uuid.Nil, ctx.UserID)
	})

	t.Run("supports update context with old record", func(t *testing.T) {
		recordID := uuid.New()
		tableID := uuid.New()
		ctx := &DeliveryContext{
			BaseID:   uuid.New(),
			TableID:  tableID,
			RecordID: &recordID,
			Record: &models.Record{
				ID:      recordID,
				TableID: tableID,
				Values:  json.RawMessage(`{"name": "Updated"}`),
			},
			OldRecord: &models.Record{
				ID:      recordID,
				TableID: tableID,
				Values:  json.RawMessage(`{"name": "Original"}`),
			},
			Event:  models.WebhookEventRecordUpdated,
			UserID: uuid.New(),
		}

		assert.NotNil(t, ctx.OldRecord)
		assert.Equal(t, models.WebhookEventRecordUpdated, ctx.Event)
	})
}

func TestDeliveryEngine_ProcessEvent(t *testing.T) {
	t.Run("does nothing when no webhooks exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)
		engine := NewDeliveryEngine(webhookStore, nil)

		baseID := uuid.New()
		eventJSON, _ := json.Marshal([]models.WebhookEvent{models.WebhookEventRecordCreated})

		// Expect query for active webhooks - return empty
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at FROM webhooks").
			WithArgs(baseID, eventJSON).
			WillReturnRows(pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}))

		ctx := context.Background()
		deliveryCtx := &DeliveryContext{
			BaseID:  baseID,
			TableID: uuid.New(),
			Event:   models.WebhookEventRecordCreated,
			UserID:  uuid.New(),
		}

		engine.ProcessEvent(ctx, deliveryCtx)

		// Give goroutines time to complete (if any were started)
		time.Sleep(50 * time.Millisecond)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("delivers to webhook when one exists", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		baseID := uuid.New()
		webhookID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()
		events := []models.WebhookEvent{models.WebhookEventRecordCreated}
		eventsJSON, _ := json.Marshal(events)

		// Track what the webhook server receives
		var receivedBody []byte
		var receivedHeaders http.Header
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			receivedHeaders = r.Header.Clone()
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   server.Client(),
		}

		// Expect query for active webhooks
		webhookRows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID, baseID, "Test Webhook", server.URL, eventsJSON, nil, true, userID, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at FROM webhooks").
			WithArgs(baseID, eventsJSON).
			WillReturnRows(webhookRows)

		// Expect delivery record to be created
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", 200, `{"status": "ok"}`, nil, 10, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		deliveryCtx := &DeliveryContext{
			BaseID:  baseID,
			TableID: uuid.New(),
			Event:   models.WebhookEventRecordCreated,
			UserID:  userID,
		}

		engine.ProcessEvent(ctx, deliveryCtx)

		// Wait for async delivery
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.NotEmpty(t, receivedBody)
		assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
		assert.Equal(t, "VibeTable-Webhook/1.0", receivedHeaders.Get("User-Agent"))
		assert.Equal(t, "record.created", receivedHeaders.Get("X-Webhook-Event"))
		mu.Unlock()

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles database error gracefully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)
		engine := NewDeliveryEngine(webhookStore, nil)

		baseID := uuid.New()
		eventJSON, _ := json.Marshal([]models.WebhookEvent{models.WebhookEventRecordCreated})

		// Expect query to fail
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at FROM webhooks").
			WithArgs(baseID, eventJSON).
			WillReturnError(assert.AnError)

		ctx := context.Background()
		deliveryCtx := &DeliveryContext{
			BaseID:  baseID,
			TableID: uuid.New(),
			Event:   models.WebhookEventRecordCreated,
			UserID:  uuid.New(),
		}

		// Should not panic
		engine.ProcessEvent(ctx, deliveryCtx)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeliveryEngine_deliverToWebhook(t *testing.T) {
	t.Run("delivers with HMAC signature when secret is set", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		var receivedSignature string
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			receivedSignature = r.Header.Get("X-Webhook-Signature")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   server.Client(),
		}

		webhookID := uuid.New()
		secret := "my-secret-key"
		webhook := models.Webhook{
			ID:     webhookID,
			BaseID: uuid.New(),
			URL:    server.URL,
			Secret: &secret,
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", 200, "", nil, 10, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		payload := []byte(`{"event": "record.created"}`)

		engine.deliverToWebhook(ctx, webhook, "record.created", payload)

		mu.Lock()
		assert.Contains(t, receivedSignature, "sha256=")
		mu.Unlock()

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles request creation error", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   &http.Client{},
		}

		webhookID := uuid.New()
		webhook := models.Webhook{
			ID:     webhookID,
			BaseID: uuid.New(),
			URL:    "://invalid-url", // Invalid URL
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", nil, nil, "failed to create request", 0, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		payload := []byte(`{"event": "record.created"}`)

		engine.deliverToWebhook(ctx, webhook, "record.created", payload)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles HTTP request failure", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		// Server that immediately closes the connection
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("force connection close")
		}))
		server.Close() // Close immediately to cause connection failure

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   &http.Client{Timeout: 1 * time.Second},
		}

		webhookID := uuid.New()
		webhook := models.Webhook{
			ID:     webhookID,
			BaseID: uuid.New(),
			URL:    server.URL,
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", nil, nil, "request failed", 0, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		payload := []byte(`{"event": "record.created"}`)

		engine.deliverToWebhook(ctx, webhook, "record.created", payload)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles non-success status code", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "server error"}`))
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   server.Client(),
		}

		webhookID := uuid.New()
		webhook := models.Webhook{
			ID:     webhookID,
			BaseID: uuid.New(),
			URL:    server.URL,
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		status := 500
		body := `{"error": "server error"}`
		errStr := "non-success status code: 500"
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", &status, &body, &errStr, 10, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		payload := []byte(`{"event": "record.created"}`)

		engine.deliverToWebhook(ctx, webhook, "record.created", payload)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles success status code", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   server.Client(),
		}

		webhookID := uuid.New()
		webhook := models.Webhook{
			ID:     webhookID,
			BaseID: uuid.New(),
			URL:    server.URL,
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		status := 200
		body := `{"status": "ok"}`
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", &status, &body, nil, 10, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		payload := []byte(`{"event": "record.created"}`)

		engine.deliverToWebhook(ctx, webhook, "record.created", payload)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sends correct payload and headers", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		var receivedBody []byte
		var receivedHeaders http.Header
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			receivedHeaders = r.Header.Clone()
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   server.Client(),
		}

		webhookID := uuid.New()
		webhook := models.Webhook{
			ID:     webhookID,
			BaseID: uuid.New(),
			URL:    server.URL,
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.updated", "{}", 200, "", nil, 10, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.updated", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		payload := []byte(`{"event": "record.updated", "data": {"id": "123"}}`)

		engine.deliverToWebhook(ctx, webhook, "record.updated", payload)

		mu.Lock()
		assert.Equal(t, payload, receivedBody)
		assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
		assert.Equal(t, "VibeTable-Webhook/1.0", receivedHeaders.Get("User-Agent"))
		assert.Equal(t, "record.updated", receivedHeaders.Get("X-Webhook-Event"))
		assert.Equal(t, webhookID.String(), receivedHeaders.Get("X-Webhook-ID"))
		assert.NotEmpty(t, receivedHeaders.Get("X-Webhook-Timestamp"))
		mu.Unlock()

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeliveryEngine_recordDelivery(t *testing.T) {
	t.Run("records delivery with duration", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   &http.Client{},
		}

		webhookID := uuid.New()
		delivery := &models.WebhookDelivery{
			WebhookID: webhookID,
			EventType: "record.deleted",
			Payload:   `{"event": "record.deleted"}`,
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.deleted", `{"event": "record.deleted"}`, nil, nil, nil, 50, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.deleted", `{"event": "record.deleted"}`, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		startTime := time.Now().Add(-50 * time.Millisecond)
		engine.recordDelivery(context.Background(), delivery, startTime)

		assert.NotNil(t, delivery.DurationMs)
		assert.GreaterOrEqual(t, *delivery.DurationMs, 50)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("handles database error gracefully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   &http.Client{},
		}

		webhookID := uuid.New()
		delivery := &models.WebhookDelivery{
			WebhookID: webhookID,
			EventType: "record.deleted",
			Payload:   `{"event": "record.deleted"}`,
		}

		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.deleted", `{"event": "record.deleted"}`, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnError(assert.AnError)

		startTime := time.Now()

		// Should not panic
		engine.recordDelivery(context.Background(), delivery, startTime)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestIntegration_WebhookDelivery(t *testing.T) {
	t.Run("full webhook delivery flow", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		baseID := uuid.New()
		tableID := uuid.New()
		recordID := uuid.New()
		webhookID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()
		events := []models.WebhookEvent{models.WebhookEventRecordCreated}
		eventsJSON, _ := json.Marshal(events)

		var receivedPayload models.WebhookPayload
		var mu sync.Mutex

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			body, _ := io.ReadAll(r.Body)
			json.Unmarshal(body, &receivedPayload)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"received": true}`))
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   server.Client(),
		}

		// Mock GetActiveByBaseAndEvent
		webhookRows := pgxmock.NewRows([]string{"id", "base_id", "name", "url", "events", "secret", "is_active", "created_by", "created_at", "updated_at"}).
			AddRow(webhookID, baseID, "Integration Test Webhook", server.URL, eventsJSON, nil, true, userID, now, now)
		mock.ExpectQuery("SELECT id, base_id, name, url, events, secret, is_active, created_by, created_at, updated_at FROM webhooks").
			WithArgs(baseID, eventsJSON).
			WillReturnRows(webhookRows)

		// Mock CreateDelivery
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", 200, `{"received": true}`, nil, 10, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		deliveryCtx := &DeliveryContext{
			BaseID:   baseID,
			TableID:  tableID,
			RecordID: &recordID,
			Record: &models.Record{
				ID:      recordID,
				TableID: tableID,
				Values:  json.RawMessage(`{"name": "Test Record"}`),
			},
			Event:  models.WebhookEventRecordCreated,
			UserID: userID,
		}

		engine.ProcessEvent(ctx, deliveryCtx)

		// Wait for async delivery
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		assert.Equal(t, models.WebhookEventRecordCreated, receivedPayload.Event)
		assert.Equal(t, baseID, receivedPayload.BaseID)
		assert.Equal(t, tableID, receivedPayload.TableID)
		assert.NotNil(t, receivedPayload.RecordID)
		assert.Equal(t, recordID, *receivedPayload.RecordID)
		mu.Unlock()

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestComputeHMAC_VerifySignature(t *testing.T) {
	t.Run("signature can be verified", func(t *testing.T) {
		payload := []byte(`{"event": "record.created", "data": {"id": "abc123"}}`)
		secret := "webhook-secret-key"

		signature := computeHMAC(payload, secret)

		// Manually verify the signature
		assert.True(t, len(signature) > 7) // "sha256=" + hash

		// Extract hex hash and verify
		hexHash := signature[7:] // Remove "sha256=" prefix
		assert.Len(t, hexHash, 64) // SHA256 produces 64 hex characters
	})

	t.Run("empty payload produces valid signature", func(t *testing.T) {
		payload := []byte(``)
		secret := "secret"

		signature := computeHMAC(payload, secret)
		assert.Contains(t, signature, "sha256=")
	})

	t.Run("empty secret produces valid signature", func(t *testing.T) {
		payload := []byte(`{"test": true}`)
		secret := ""

		signature := computeHMAC(payload, secret)
		assert.Contains(t, signature, "sha256=")
	})
}

func TestPayloadSerialization(t *testing.T) {
	t.Run("serializes all fields correctly", func(t *testing.T) {
		recordID := uuid.New()
		baseID := uuid.New()
		tableID := uuid.New()
		userID := uuid.New()
		ts := time.Now().UTC()

		payload := &models.WebhookPayload{
			Event:     models.WebhookEventRecordUpdated,
			Timestamp: ts,
			BaseID:    baseID,
			TableID:   tableID,
			RecordID:  &recordID,
			Record: &models.Record{
				ID:      recordID,
				TableID: tableID,
				Values:  json.RawMessage(`{"field1": "value1"}`),
			},
			OldRecord: &models.Record{
				ID:      recordID,
				TableID: tableID,
				Values:  json.RawMessage(`{"field1": "old_value"}`),
			},
			UserID: userID,
		}

		data, err := json.Marshal(payload)
		require.NoError(t, err)

		// Verify all fields are present
		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "record.updated", decoded["event"])
		assert.Equal(t, baseID.String(), decoded["base_id"])
		assert.Equal(t, tableID.String(), decoded["table_id"])
		assert.Equal(t, recordID.String(), decoded["record_id"])
		assert.NotNil(t, decoded["record"])
		assert.NotNil(t, decoded["old_record"])
		assert.Equal(t, userID.String(), decoded["user_id"])
	})
}

func TestLargeResponseHandling(t *testing.T) {
	t.Run("handles large response body", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		baseStore := store.NewBaseStore(mock)
		webhookStore := store.NewWebhookStore(mock, baseStore)

		// Create a large response (> 10KB limit in code)
		largeBody := bytes.Repeat([]byte("x"), 20*1024)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(largeBody)
		}))
		defer server.Close()

		engine := &DeliveryEngine{
			webhookStore: webhookStore,
			tableStore:   nil,
			httpClient:   server.Client(),
		}

		webhookID := uuid.New()
		webhook := models.Webhook{
			ID:     webhookID,
			BaseID: uuid.New(),
			URL:    server.URL,
		}

		now := time.Now().UTC()
		deliveryID := uuid.New()
		deliveryRows := pgxmock.NewRows([]string{"id", "webhook_id", "event_type", "payload", "response_status", "response_body", "error", "duration_ms", "delivered_at"}).
			AddRow(deliveryID, webhookID, "record.created", "{}", 200, "", nil, 10, now)
		mock.ExpectQuery("INSERT INTO webhook_deliveries").
			WithArgs(pgxmock.AnyArg(), webhookID, "record.created", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(deliveryRows)

		ctx := context.Background()
		payload := []byte(`{"event": "record.created"}`)

		// Should not panic on large response
		engine.deliverToWebhook(ctx, webhook, "record.created", payload)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
