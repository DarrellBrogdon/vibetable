package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/store"
)

type TableHandler struct {
	store *store.TableStore
}

func NewTableHandler(store *store.TableStore) *TableHandler {
	return &TableHandler{store: store}
}

// Request types
type CreateTableRequest struct {
	Name string `json:"name"`
}

type UpdateTableRequest struct {
	Name string `json:"name"`
}

type ReorderTablesRequest struct {
	TableIDs []uuid.UUID `json:"table_ids"`
}

type DuplicateTableRequest struct {
	IncludeRecords bool `json:"include_records"`
}

// ListTables handles GET /bases/:baseId/tables
func (h *TableHandler) ListTables(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	baseID, err := uuid.Parse(chi.URLParam(r, "baseId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid base ID")
		return
	}

	tables, err := h.store.ListTablesForBase(r.Context(), baseID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found or access denied")
			return
		}
		log.Printf("Error listing tables: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list tables")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tables": tables,
	})
}

// CreateTable handles POST /bases/:baseId/tables
func (h *TableHandler) CreateTable(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	baseID, err := uuid.Parse(chi.URLParam(r, "baseId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid base ID")
		return
	}

	var req CreateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "Table name is required")
		return
	}

	table, err := h.store.CreateTable(r.Context(), baseID, name, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to create tables in this base")
			return
		}
		log.Printf("Error creating table: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create table")
		return
	}

	log.Printf("Table created: %s (id=%s) in base %s", table.Name, table.ID, baseID)
	writeJSON(w, http.StatusCreated, table)
}

// GetTable handles GET /tables/:id
func (h *TableHandler) GetTable(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	table, err := h.store.GetTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found or access denied")
			return
		}
		log.Printf("Error getting table: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get table")
		return
	}

	writeJSON(w, http.StatusOK, table)
}

// UpdateTable handles PATCH /tables/:id
func (h *TableHandler) UpdateTable(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	var req UpdateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "Table name is required")
		return
	}

	table, err := h.store.UpdateTable(r.Context(), tableID, name, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to edit this table")
			return
		}
		log.Printf("Error updating table: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update table")
		return
	}

	writeJSON(w, http.StatusOK, table)
}

// DeleteTable handles DELETE /tables/:id
func (h *TableHandler) DeleteTable(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	err = h.store.DeleteTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to delete this table")
			return
		}
		log.Printf("Error deleting table: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete table")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Table deleted successfully",
	})
}

// ReorderTables handles PUT /bases/:baseId/tables/reorder
func (h *TableHandler) ReorderTables(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	baseID, err := uuid.Parse(chi.URLParam(r, "baseId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid base ID")
		return
	}

	var req ReorderTablesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if len(req.TableIDs) == 0 {
		writeError(w, http.StatusBadRequest, "table_ids_required", "Table IDs are required")
		return
	}

	err = h.store.ReorderTables(r.Context(), baseID, req.TableIDs, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to reorder tables")
			return
		}
		log.Printf("Error reordering tables: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to reorder tables")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Tables reordered successfully",
	})
}

// DuplicateTable handles POST /tables/:id/duplicate
func (h *TableHandler) DuplicateTable(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	tableID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid table ID")
		return
	}

	var req DuplicateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to not including records if body is empty
		req.IncludeRecords = false
	}

	table, err := h.store.DuplicateTable(r.Context(), tableID, user.ID, req.IncludeRecords)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to duplicate this table")
			return
		}
		log.Printf("Error duplicating table: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to duplicate table")
		return
	}

	log.Printf("Table duplicated: %s (id=%s)", table.Name, table.ID)
	writeJSON(w, http.StatusCreated, table)
}
