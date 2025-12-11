package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

type RecordHandler struct {
	store         *store.RecordStore
	activityStore *store.ActivityStore
}

func NewRecordHandler(store *store.RecordStore, activityStore *store.ActivityStore) *RecordHandler {
	return &RecordHandler{store: store, activityStore: activityStore}
}

// Request types
type CreateRecordRequest struct {
	Values map[string]interface{} `json:"values"`
}

type UpdateRecordRequest struct {
	Values map[string]interface{} `json:"values"`
}

type BulkCreateRecordsRequest struct {
	Records []map[string]interface{} `json:"records"`
}

// ListRecords handles GET /tables/:tableId/records
func (h *RecordHandler) ListRecords(w http.ResponseWriter, r *http.Request) {
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

	records, err := h.store.ListRecordsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found or access denied")
			return
		}
		log.Printf("Error listing records: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list records")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"records": records,
	})
}

// CreateRecord handles POST /tables/:tableId/records
func (h *RecordHandler) CreateRecord(w http.ResponseWriter, r *http.Request) {
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

	var req CreateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Convert values to JSON
	var values json.RawMessage
	if req.Values != nil {
		values, err = json.Marshal(req.Values)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_values", "Invalid values format")
			return
		}
	}

	record, err := h.store.CreateRecord(r.Context(), tableID, values, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to create records in this table")
			return
		}
		log.Printf("Error creating record: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create record")
		return
	}

	// Log activity (use background context since request context may close)
	if h.activityStore != nil {
		tableIDCopy := tableID
		recordIDCopy := record.ID
		userIDCopy := user.ID
		go func() {
			if err := h.activityStore.LogActivity(context.Background(), &models.Activity{
				TableID:    &tableIDCopy,
				RecordID:   &recordIDCopy,
				UserID:     userIDCopy,
				Action:     "create",
				EntityType: "record",
			}); err != nil {
				log.Printf("Failed to log activity: %v", err)
			}
		}()
	}

	writeJSON(w, http.StatusCreated, record)
}

// BulkCreateRecords handles POST /tables/:tableId/records/bulk
func (h *RecordHandler) BulkCreateRecords(w http.ResponseWriter, r *http.Request) {
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

	var req BulkCreateRecordsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if len(req.Records) == 0 {
		writeError(w, http.StatusBadRequest, "records_required", "At least one record is required")
		return
	}

	// Convert each record's values to JSON
	var recordValues []json.RawMessage
	for _, vals := range req.Records {
		jsonVals, err := json.Marshal(vals)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_values", "Invalid values format")
			return
		}
		recordValues = append(recordValues, jsonVals)
	}

	records, err := h.store.BulkCreateRecords(r.Context(), tableID, recordValues, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to create records in this table")
			return
		}
		log.Printf("Error bulk creating records: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create records")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"records": records,
	})
}

// GetRecord handles GET /records/:id
func (h *RecordHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	recordID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid record ID")
		return
	}

	record, err := h.store.GetRecord(r.Context(), recordID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found or access denied")
			return
		}
		log.Printf("Error getting record: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get record")
		return
	}

	writeJSON(w, http.StatusOK, record)
}

// UpdateRecord handles PUT /records/:id (full replace)
func (h *RecordHandler) UpdateRecord(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	recordID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid record ID")
		return
	}

	var req UpdateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Convert values to JSON
	values, err := json.Marshal(req.Values)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_values", "Invalid values format")
		return
	}

	record, err := h.store.UpdateRecord(r.Context(), recordID, values, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to edit this record")
			return
		}
		log.Printf("Error updating record: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update record")
		return
	}

	writeJSON(w, http.StatusOK, record)
}

// PatchRecord handles PATCH /records/:id (partial update)
func (h *RecordHandler) PatchRecord(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	recordID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid record ID")
		return
	}

	var req UpdateRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	record, err := h.store.PatchRecord(r.Context(), recordID, req.Values, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to edit this record")
			return
		}
		log.Printf("Error patching record: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update record")
		return
	}

	// Log activity with changes (use background context since request context may close)
	if h.activityStore != nil {
		tableIDCopy := record.TableID
		recordIDCopy := record.ID
		userIDCopy := user.ID
		valuesCopy := req.Values
		go func() {
			changes, _ := json.Marshal(valuesCopy)
			if err := h.activityStore.LogActivity(context.Background(), &models.Activity{
				TableID:    &tableIDCopy,
				RecordID:   &recordIDCopy,
				UserID:     userIDCopy,
				Action:     "update",
				EntityType: "record",
				Changes:    changes,
			}); err != nil {
				log.Printf("Failed to log activity: %v", err)
			}
		}()
	}

	writeJSON(w, http.StatusOK, record)
}

// DeleteRecord handles DELETE /records/:id
func (h *RecordHandler) DeleteRecord(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	recordID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid record ID")
		return
	}

	// Get the record first to have table ID for activity logging
	record, err := h.store.GetRecord(r.Context(), recordID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found")
			return
		}
		log.Printf("Error getting record for delete: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete record")
		return
	}

	err = h.store.DeleteRecord(r.Context(), recordID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to delete this record")
			return
		}
		log.Printf("Error deleting record: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete record")
		return
	}

	// Log activity (use background context since request context may close)
	if h.activityStore != nil {
		tableIDCopy := record.TableID
		recordIDCopy := recordID
		userIDCopy := user.ID
		go func() {
			if err := h.activityStore.LogActivity(context.Background(), &models.Activity{
				TableID:    &tableIDCopy,
				RecordID:   &recordIDCopy,
				UserID:     userIDCopy,
				Action:     "delete",
				EntityType: "record",
			}); err != nil {
				log.Printf("Failed to log activity: %v", err)
			}
		}()
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Record deleted successfully",
	})
}

// UpdateRecordColorRequest for updating record color
type UpdateRecordColorRequest struct {
	Color *string `json:"color"`
}

// UpdateRecordColor handles PATCH /records/:id/color
func (h *RecordHandler) UpdateRecordColor(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	recordID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid record ID")
		return
	}

	var req UpdateRecordColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Validate color value if provided
	validColors := map[string]bool{
		"red": true, "orange": true, "yellow": true, "green": true,
		"blue": true, "purple": true, "pink": true, "gray": true,
	}
	if req.Color != nil && *req.Color != "" && !validColors[*req.Color] {
		writeError(w, http.StatusBadRequest, "invalid_color", "Color must be one of: red, orange, yellow, green, blue, purple, pink, gray")
		return
	}

	// Convert empty string to nil to clear the color
	var color *string
	if req.Color != nil && *req.Color != "" {
		color = req.Color
	}

	record, err := h.store.UpdateRecordColor(r.Context(), recordID, color, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to edit this record")
			return
		}
		log.Printf("Error updating record color: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update record color")
		return
	}

	writeJSON(w, http.StatusOK, record)
}
