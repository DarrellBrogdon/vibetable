package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/store"
)

type ActivityHandler struct {
	store *store.ActivityStore
}

func NewActivityHandler(store *store.ActivityStore) *ActivityHandler {
	return &ActivityHandler{store: store}
}

// ListActivitiesForBase handles GET /bases/:baseId/activity
func (h *ActivityHandler) ListActivitiesForBase(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	baseID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid base ID")
		return
	}

	// Parse query parameters for filtering
	var filters store.ActivityFilters

	if userIDStr := r.URL.Query().Get("userId"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filters.UserID = &userID
		}
	}

	if action := r.URL.Query().Get("action"); action != "" {
		filters.Action = &action
	}

	if entityType := r.URL.Query().Get("entityType"); entityType != "" {
		filters.EntityType = &entityType
	}

	if tableIDStr := r.URL.Query().Get("tableId"); tableIDStr != "" {
		tableID, err := uuid.Parse(tableIDStr)
		if err == nil {
			filters.TableID = &tableID
		}
	}

	// Parse pagination
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	activities, err := h.store.ListActivitiesForBase(r.Context(), baseID, user.ID, filters, limit, offset)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found or access denied")
			return
		}
		log.Printf("Error listing activities: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list activities")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
	})
}

// ListActivitiesForRecord handles GET /records/:recordId/activity
func (h *ActivityHandler) ListActivitiesForRecord(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	recordID, err := uuid.Parse(chi.URLParam(r, "recordId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid record ID")
		return
	}

	// Parse limit
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	activities, err := h.store.ListActivitiesForRecord(r.Context(), recordID, user.ID, limit)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found or access denied")
			return
		}
		log.Printf("Error listing record activities: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list activities")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"activities": activities,
	})
}
