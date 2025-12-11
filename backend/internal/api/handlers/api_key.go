package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

type APIKeyHandler struct {
	apiKeyStore *store.APIKeyStore
}

func NewAPIKeyHandler(apiKeyStore *store.APIKeyStore) *APIKeyHandler {
	return &APIKeyHandler{apiKeyStore: apiKeyStore}
}

// ListAPIKeys handles GET /api-keys
func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	keys, err := h.apiKeyStore.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list API keys")
		return
	}

	if keys == nil {
		keys = []models.APIKey{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"api_keys": keys,
	})
}

// CreateAPIKey handles POST /api-keys
func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	var req models.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Name is required")
		return
	}

	apiKey, err := h.apiKeyStore.Create(r.Context(), user.ID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Return the API key with the token (only shown once!)
	writeJSON(w, http.StatusCreated, apiKey)
}

// GetAPIKey handles GET /api-keys/{id}
func (h *APIKeyHandler) GetAPIKey(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid API key ID")
		return
	}

	apiKey, err := h.apiKeyStore.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}

	// Verify ownership
	if apiKey.UserID != user.ID {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	writeJSON(w, http.StatusOK, apiKey)
}

// DeleteAPIKey handles DELETE /api-keys/{id}
func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid API key ID")
		return
	}

	// Verify ownership
	apiKey, err := h.apiKeyStore.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "API key not found")
		return
	}
	if apiKey.UserID != user.ID {
		writeError(w, http.StatusForbidden, "forbidden", "Access denied")
		return
	}

	if err := h.apiKeyStore.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to delete API key")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "API key deleted",
	})
}
