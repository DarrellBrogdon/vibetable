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

type FieldHandler struct {
	store *store.FieldStore
}

func NewFieldHandler(store *store.FieldStore) *FieldHandler {
	return &FieldHandler{store: store}
}

// Request types
type CreateFieldRequest struct {
	Name      string          `json:"name"`
	FieldType string          `json:"field_type"`
	Options   json.RawMessage `json:"options,omitempty"`
}

type UpdateFieldRequest struct {
	Name    *string          `json:"name,omitempty"`
	Options *json.RawMessage `json:"options,omitempty"`
}

type ReorderFieldsRequest struct {
	FieldIDs []uuid.UUID `json:"field_ids"`
}

// ListFields handles GET /tables/:tableId/fields
func (h *FieldHandler) ListFields(w http.ResponseWriter, r *http.Request) {
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

	fields, err := h.store.ListFieldsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found or access denied")
			return
		}
		log.Printf("Error listing fields: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list fields")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"fields": fields,
	})
}

// CreateField handles POST /tables/:tableId/fields
func (h *FieldHandler) CreateField(w http.ResponseWriter, r *http.Request) {
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

	var req CreateFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "Field name is required")
		return
	}

	fieldType := models.FieldType(req.FieldType)
	if !models.IsValidFieldType(fieldType) {
		writeError(w, http.StatusBadRequest, "invalid_field_type", "Invalid field type. Valid types: text, number, checkbox, date, single_select, multi_select, linked_record")
		return
	}

	field, err := h.store.CreateField(r.Context(), tableID, name, fieldType, req.Options, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to create fields in this table")
			return
		}
		log.Printf("Error creating field: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create field")
		return
	}

	log.Printf("Field created: %s (id=%s) in table %s", field.Name, field.ID, tableID)
	writeJSON(w, http.StatusCreated, field)
}

// GetField handles GET /fields/:id
func (h *FieldHandler) GetField(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	fieldID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid field ID")
		return
	}

	field, err := h.store.GetField(r.Context(), fieldID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Field not found or access denied")
			return
		}
		log.Printf("Error getting field: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get field")
		return
	}

	writeJSON(w, http.StatusOK, field)
}

// UpdateField handles PATCH /fields/:id
func (h *FieldHandler) UpdateField(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	fieldID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid field ID")
		return
	}

	var req UpdateFieldRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Validate name if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			writeError(w, http.StatusBadRequest, "name_required", "Field name cannot be empty")
			return
		}
		req.Name = &name
	}

	field, err := h.store.UpdateField(r.Context(), fieldID, req.Name, req.Options, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Field not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to edit this field")
			return
		}
		log.Printf("Error updating field: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update field")
		return
	}

	writeJSON(w, http.StatusOK, field)
}

// DeleteField handles DELETE /fields/:id
func (h *FieldHandler) DeleteField(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	fieldID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid field ID")
		return
	}

	err = h.store.DeleteField(r.Context(), fieldID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Field not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to delete this field")
			return
		}
		log.Printf("Error deleting field: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete field")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Field deleted successfully",
	})
}

// ReorderFields handles PUT /tables/:tableId/fields/reorder
func (h *FieldHandler) ReorderFields(w http.ResponseWriter, r *http.Request) {
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

	var req ReorderFieldsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if len(req.FieldIDs) == 0 {
		writeError(w, http.StatusBadRequest, "field_ids_required", "Field IDs are required")
		return
	}

	err = h.store.ReorderFields(r.Context(), tableID, req.FieldIDs, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to reorder fields")
			return
		}
		log.Printf("Error reordering fields: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to reorder fields")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Fields reordered successfully",
	})
}
