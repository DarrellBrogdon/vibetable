package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

type ViewHandler struct {
	store *store.ViewStore
}

func NewViewHandler(store *store.ViewStore) *ViewHandler {
	return &ViewHandler{store: store}
}

// Request types
type CreateViewRequest struct {
	Name   string          `json:"name"`
	Type   string          `json:"type"`
	Config json.RawMessage `json:"config,omitempty"`
}

type UpdateViewRequest struct {
	Name   *string          `json:"name,omitempty"`
	Config *json.RawMessage `json:"config,omitempty"`
}

// ListViews handles GET /tables/:tableId/views
func (h *ViewHandler) ListViews(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "tableId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	views, err := h.store.ListViewsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found or access denied")
			return
		}
		log.Printf("Error listing views: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list views")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"views": views,
	})
}

// CreateView handles POST /tables/:tableId/views
func (h *ViewHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "tableId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	var req CreateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "View name is required")
		return
	}

	viewType := models.ViewType(req.Type)
	if !models.IsValidViewType(viewType) {
		writeError(w, http.StatusBadRequest, "invalid_view_type", "Invalid view type. Valid types: grid, kanban")
		return
	}

	view, err := h.store.CreateView(r.Context(), tableID, name, viewType, req.Config, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to create views in this table")
			return
		}
		log.Printf("Error creating view: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create view")
		return
	}

	log.Printf("View created: %s (id=%s) in table %s", view.Name, view.ID, tableID)
	writeJSON(w, http.StatusCreated, view)
}

// GetView handles GET /views/:id
func (h *ViewHandler) GetView(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	viewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid view ID")
		return
	}

	view, err := h.store.GetView(r.Context(), viewID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "View not found or access denied")
			return
		}
		log.Printf("Error getting view: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get view")
		return
	}

	writeJSON(w, http.StatusOK, view)
}

// UpdateView handles PATCH /views/:id
func (h *ViewHandler) UpdateView(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	viewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid view ID")
		return
	}

	var req UpdateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Validate name if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "name_required", "View name cannot be empty")
			return
		}
		req.Name = &name
	}

	view, err := h.store.UpdateView(r.Context(), viewID, req.Name, req.Config, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "View not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to edit this view")
			return
		}
		log.Printf("Error updating view: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update view")
		return
	}

	writeJSON(w, http.StatusOK, view)
}

// DeleteView handles DELETE /views/:id
func (h *ViewHandler) DeleteView(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	viewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid view ID")
		return
	}

	err = h.store.DeleteView(r.Context(), viewID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "View not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to delete this view")
			return
		}
		log.Printf("Error deleting view: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete view")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "View deleted successfully",
	})
}

// SetViewPublicRequest is the request body for setting view public status
type SetViewPublicRequest struct {
	IsPublic bool `json:"is_public"`
}

// SetViewPublic handles PATCH /views/:id/public
func (h *ViewHandler) SetViewPublic(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	viewID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid view ID")
		return
	}

	var req SetViewPublicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	view, err := h.store.SetViewPublic(r.Context(), viewID, req.IsPublic, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "View not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "Only base owners can share views publicly")
			return
		}
		log.Printf("Error setting view public: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update view sharing")
		return
	}

	writeJSON(w, http.StatusOK, view)
}

// GetPublicView handles GET /public/views/:token
func (h *ViewHandler) GetPublicView(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "token_required", "Token is required")
		return
	}

	publicView, err := h.store.GetPublicView(r.Context(), token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "View not found or sharing disabled")
			return
		}
		log.Printf("Error getting public view: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get view")
		return
	}

	writeJSON(w, http.StatusOK, publicView)
}
