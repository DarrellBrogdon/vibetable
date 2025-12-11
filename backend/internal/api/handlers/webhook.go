package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

type WebhookHandler struct {
	webhookStore *store.WebhookStore
	baseStore    *store.BaseStore
}

func NewWebhookHandler(webhookStore *store.WebhookStore, baseStore *store.BaseStore) *WebhookHandler {
	return &WebhookHandler{
		webhookStore: webhookStore,
		baseStore:    baseStore,
	}
}

// ListWebhooks handles GET /bases/{id}/webhooks
func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	baseIDStr := chi.URLParam(r, "id")
	baseID, err := uuid.Parse(baseIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid base ID")
		return
	}

	// Check access
	role, err := h.baseStore.GetUserRole(r.Context(), baseID, user.ID)
	if err != nil || role == "" {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	webhooks, err := h.webhookStore.ListByBase(r.Context(), baseID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list webhooks")
		return
	}

	if webhooks == nil {
		webhooks = []models.Webhook{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"webhooks": webhooks,
	})
}

// CreateWebhook handles POST /bases/{id}/webhooks
func (h *WebhookHandler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	baseIDStr := chi.URLParam(r, "id")
	baseID, err := uuid.Parse(baseIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid base ID")
		return
	}

	// Check access (need at least editor)
	role, err := h.baseStore.GetUserRole(r.Context(), baseID, user.ID)
	if err != nil || (role != "owner" && role != "editor") {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	var req models.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Name is required")
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "URL is required")
		return
	}

	webhook, err := h.webhookStore.Create(r.Context(), baseID, user.ID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, webhook)
}

// GetWebhook handles GET /webhooks/{id}
func (h *WebhookHandler) GetWebhook(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid webhook ID")
		return
	}

	webhook, err := h.webhookStore.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Webhook not found")
		return
	}

	// Check access
	role, err := h.baseStore.GetUserRole(r.Context(), webhook.BaseID, user.ID)
	if err != nil || role == "" {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	writeJSON(w, http.StatusOK, webhook)
}

// UpdateWebhook handles PATCH /webhooks/{id}
func (h *WebhookHandler) UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid webhook ID")
		return
	}

	webhook, err := h.webhookStore.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Webhook not found")
		return
	}

	// Check access (need at least editor)
	role, err := h.baseStore.GetUserRole(r.Context(), webhook.BaseID, user.ID)
	if err != nil || (role != "owner" && role != "editor") {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	var req models.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	updated, err := h.webhookStore.Update(r.Context(), id, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

// DeleteWebhook handles DELETE /webhooks/{id}
func (h *WebhookHandler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid webhook ID")
		return
	}

	webhook, err := h.webhookStore.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Webhook not found")
		return
	}

	// Check access (need at least editor)
	role, err := h.baseStore.GetUserRole(r.Context(), webhook.BaseID, user.ID)
	if err != nil || (role != "owner" && role != "editor") {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	if err := h.webhookStore.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete webhook")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Webhook deleted",
	})
}

// ListDeliveries handles GET /webhooks/{id}/deliveries
func (h *WebhookHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid webhook ID")
		return
	}

	webhook, err := h.webhookStore.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "Webhook not found")
		return
	}

	// Check access
	role, err := h.baseStore.GetUserRole(r.Context(), webhook.BaseID, user.ID)
	if err != nil || role == "" {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	deliveries, err := h.webhookStore.ListDeliveries(r.Context(), id, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list deliveries")
		return
	}

	if deliveries == nil {
		deliveries = []models.WebhookDelivery{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"deliveries": deliveries,
	})
}
