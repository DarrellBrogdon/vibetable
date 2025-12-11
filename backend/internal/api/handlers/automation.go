package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

type AutomationHandler struct {
	store *store.AutomationStore
}

func NewAutomationHandler(store *store.AutomationStore) *AutomationHandler {
	return &AutomationHandler{store: store}
}

// ListAutomations GET /tables/:tableId/automations
func (h *AutomationHandler) ListAutomations(w http.ResponseWriter, r *http.Request) {
	tableID, err := uuid.Parse(chi.URLParam(r, "tableId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	automations, err := h.store.ListAutomationsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		handleAutomationStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"automations": automations,
	})
}

// CreateAutomation POST /tables/:tableId/automations
func (h *AutomationHandler) CreateAutomation(w http.ResponseWriter, r *http.Request) {
	tableID, err := uuid.Parse(chi.URLParam(r, "tableId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req struct {
		Name          string          `json:"name"`
		Description   *string         `json:"description"`
		Enabled       *bool           `json:"enabled"`
		TriggerType   string          `json:"triggerType"`
		TriggerConfig json.RawMessage `json:"triggerConfig"`
		ActionType    string          `json:"actionType"`
		ActionConfig  json.RawMessage `json:"actionConfig"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "Name is required")
		return
	}
	if req.TriggerType == "" {
		writeError(w, http.StatusBadRequest, "trigger_required", "Trigger type is required")
		return
	}
	if req.ActionType == "" {
		writeError(w, http.StatusBadRequest, "action_required", "Action type is required")
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	if req.TriggerConfig == nil {
		req.TriggerConfig = json.RawMessage("{}")
	}
	if req.ActionConfig == nil {
		req.ActionConfig = json.RawMessage("{}")
	}

	automation := &models.Automation{
		TableID:       tableID,
		Name:          req.Name,
		Description:   req.Description,
		Enabled:       enabled,
		TriggerType:   models.TriggerType(req.TriggerType),
		TriggerConfig: req.TriggerConfig,
		ActionType:    models.ActionType(req.ActionType),
		ActionConfig:  req.ActionConfig,
	}

	automation, err = h.store.CreateAutomation(r.Context(), automation, user.ID)
	if err != nil {
		handleAutomationStoreError(w, err)
		return
	}

	log.Printf("Automation created: %s (id=%s) by user %s", automation.Name, automation.ID, user.Email)
	writeJSON(w, http.StatusCreated, automation)
}

// GetAutomation GET /automations/:id
func (h *AutomationHandler) GetAutomation(w http.ResponseWriter, r *http.Request) {
	automationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid automation ID")
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	automation, err := h.store.GetAutomation(r.Context(), automationID, user.ID)
	if err != nil {
		handleAutomationStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, automation)
}

// UpdateAutomation PATCH /automations/:id
func (h *AutomationHandler) UpdateAutomation(w http.ResponseWriter, r *http.Request) {
	automationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid automation ID")
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req struct {
		Name          *string         `json:"name"`
		Description   *string         `json:"description"`
		Enabled       *bool           `json:"enabled"`
		TriggerType   *string         `json:"triggerType"`
		TriggerConfig json.RawMessage `json:"triggerConfig"`
		ActionType    *string         `json:"actionType"`
		ActionConfig  json.RawMessage `json:"actionConfig"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.TriggerType != nil {
		updates["triggerType"] = models.TriggerType(*req.TriggerType)
	}
	if req.TriggerConfig != nil {
		updates["triggerConfig"] = []byte(req.TriggerConfig)
	}
	if req.ActionType != nil {
		updates["actionType"] = models.ActionType(*req.ActionType)
	}
	if req.ActionConfig != nil {
		updates["actionConfig"] = []byte(req.ActionConfig)
	}

	automation, err := h.store.UpdateAutomation(r.Context(), automationID, updates, user.ID)
	if err != nil {
		handleAutomationStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, automation)
}

// DeleteAutomation DELETE /automations/:id
func (h *AutomationHandler) DeleteAutomation(w http.ResponseWriter, r *http.Request) {
	automationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid automation ID")
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	if err := h.store.DeleteAutomation(r.Context(), automationID, user.ID); err != nil {
		handleAutomationStoreError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ToggleAutomation POST /automations/:id/toggle
func (h *AutomationHandler) ToggleAutomation(w http.ResponseWriter, r *http.Request) {
	automationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid automation ID")
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	automation, err := h.store.ToggleAutomation(r.Context(), automationID, req.Enabled, user.ID)
	if err != nil {
		handleAutomationStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, automation)
}

// ListRuns GET /automations/:id/runs
func (h *AutomationHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	automationID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid automation ID")
		return
	}

	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	runs, err := h.store.ListRunsForAutomation(r.Context(), automationID, 50, user.ID)
	if err != nil {
		handleAutomationStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"runs": runs,
	})
}

// handleAutomationStoreError converts store errors to HTTP responses
func handleAutomationStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "not_found", "Automation not found")
		return
	}
	if errors.Is(err, store.ErrForbidden) {
		writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to access this resource")
		return
	}
	log.Printf("Automation store error: %v", err)
	writeError(w, http.StatusInternalServerError, "server_error", "An error occurred")
}
