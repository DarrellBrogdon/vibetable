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

type BaseHandler struct {
	store *store.BaseStore
}

func NewBaseHandler(store *store.BaseStore) *BaseHandler {
	return &BaseHandler{store: store}
}

// Request types
type CreateBaseRequest struct {
	Name string `json:"name"`
}

type UpdateBaseRequest struct {
	Name string `json:"name"`
}

type AddCollaboratorRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type UpdateCollaboratorRequest struct {
	Role string `json:"role"`
}

type DuplicateBaseRequest struct {
	IncludeRecords bool `json:"include_records"`
}

// ListBases handles GET /bases
func (h *BaseHandler) ListBases(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	bases, err := h.store.ListBasesForUser(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error listing bases: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list bases")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"bases": bases,
	})
}

// CreateBase handles POST /bases
func (h *BaseHandler) CreateBase(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req CreateBaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "Base name is required")
		return
	}

	base, err := h.store.CreateBase(r.Context(), name, user.ID)
	if err != nil {
		log.Printf("Error creating base: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create base")
		return
	}

	log.Printf("Base created: %s (id=%s) by user %s", base.Name, base.ID, user.Email)
	writeJSON(w, http.StatusCreated, base)
}

// GetBase handles GET /bases/:id
func (h *BaseHandler) GetBase(w http.ResponseWriter, r *http.Request) {
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

	base, err := h.store.GetBase(r.Context(), baseID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found or access denied")
			return
		}
		log.Printf("Error getting base: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get base")
		return
	}

	writeJSON(w, http.StatusOK, base)
}

// UpdateBase handles PATCH /bases/:id
func (h *BaseHandler) UpdateBase(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateBaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "Base name is required")
		return
	}

	base, err := h.store.UpdateBase(r.Context(), baseID, name, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to edit this base")
			return
		}
		log.Printf("Error updating base: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update base")
		return
	}

	writeJSON(w, http.StatusOK, base)
}

// DeleteBase handles DELETE /bases/:id
func (h *BaseHandler) DeleteBase(w http.ResponseWriter, r *http.Request) {
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

	err = h.store.DeleteBase(r.Context(), baseID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "Only the owner can delete this base")
			return
		}
		log.Printf("Error deleting base: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete base")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Base deleted successfully",
	})
}

// DuplicateBase handles POST /bases/:id/duplicate
func (h *BaseHandler) DuplicateBase(w http.ResponseWriter, r *http.Request) {
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

	var req DuplicateBaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Default to not including records if body is empty
		req.IncludeRecords = false
	}

	base, err := h.store.DuplicateBase(r.Context(), baseID, user.ID, req.IncludeRecords)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to duplicate this base")
			return
		}
		log.Printf("Error duplicating base: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to duplicate base")
		return
	}

	log.Printf("Base duplicated: %s (id=%s) by user %s", base.Name, base.ID, user.Email)
	writeJSON(w, http.StatusCreated, base)
}

// --- Collaborator handlers ---

// ListCollaborators handles GET /bases/:id/collaborators
func (h *BaseHandler) ListCollaborators(w http.ResponseWriter, r *http.Request) {
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

	collaborators, err := h.store.ListCollaborators(r.Context(), baseID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found or access denied")
			return
		}
		log.Printf("Error listing collaborators: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list collaborators")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"collaborators": collaborators,
	})
}

// AddCollaborator handles POST /bases/:id/collaborators
func (h *BaseHandler) AddCollaborator(w http.ResponseWriter, r *http.Request) {
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

	var req AddCollaboratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email_required", "Email is required")
		return
	}

	role := models.CollaboratorRole(req.Role)
	if role != models.RoleEditor && role != models.RoleViewer {
		writeError(w, http.StatusBadRequest, "invalid_role", "Role must be 'editor' or 'viewer'")
		return
	}

	collab, err := h.store.AddCollaborator(r.Context(), baseID, email, role, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Base not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "Only the owner can manage collaborators")
			return
		}
		if err.Error() == "user not found" {
			writeError(w, http.StatusBadRequest, "user_not_found", "No user found with that email")
			return
		}
		log.Printf("Error adding collaborator: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to add collaborator")
		return
	}

	writeJSON(w, http.StatusCreated, collab)
}

// UpdateCollaborator handles PATCH /bases/:id/collaborators/:userId
func (h *BaseHandler) UpdateCollaborator(w http.ResponseWriter, r *http.Request) {
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

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	var req UpdateCollaboratorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	role := models.CollaboratorRole(req.Role)
	if role != models.RoleEditor && role != models.RoleViewer {
		writeError(w, http.StatusBadRequest, "invalid_role", "Role must be 'editor' or 'viewer'")
		return
	}

	collab, err := h.store.UpdateCollaboratorRole(r.Context(), baseID, targetUserID, role, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Collaborator not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "Only the owner can manage collaborators")
			return
		}
		if err.Error() == "cannot change owner's role" {
			writeError(w, http.StatusBadRequest, "cannot_change_owner", "Cannot change the owner's role")
			return
		}
		log.Printf("Error updating collaborator: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update collaborator")
		return
	}

	writeJSON(w, http.StatusOK, collab)
}

// RemoveCollaborator handles DELETE /bases/:id/collaborators/:userId
func (h *BaseHandler) RemoveCollaborator(w http.ResponseWriter, r *http.Request) {
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

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	err = h.store.RemoveCollaborator(r.Context(), baseID, targetUserID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Collaborator not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "Only the owner can manage collaborators")
			return
		}
		if err.Error() == "cannot remove owner" {
			writeError(w, http.StatusBadRequest, "cannot_remove_owner", "Cannot remove the owner")
			return
		}
		log.Printf("Error removing collaborator: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to remove collaborator")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Collaborator removed successfully",
	})
}
