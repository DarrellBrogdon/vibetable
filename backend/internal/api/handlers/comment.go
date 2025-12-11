package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/store"
)

type CommentHandler struct {
	store *store.CommentStore
}

func NewCommentHandler(store *store.CommentStore) *CommentHandler {
	return &CommentHandler{store: store}
}

// Request types
type CreateCommentRequest struct {
	Content  string  `json:"content"`
	ParentID *string `json:"parent_id,omitempty"`
}

type UpdateCommentRequest struct {
	Content string `json:"content"`
}

type ResolveCommentRequest struct {
	Resolved bool `json:"resolved"`
}

// ListComments handles GET /records/:recordId/comments
func (h *CommentHandler) ListComments(w http.ResponseWriter, r *http.Request) {
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

	comments, err := h.store.ListComments(r.Context(), recordID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found or access denied")
			return
		}
		log.Printf("Error listing comments: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list comments")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"comments": comments,
	})
}

// CreateComment handles POST /records/:recordId/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
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

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content_required", "Comment content is required")
		return
	}

	// Parse parent ID if provided
	var parentID *uuid.UUID
	if req.ParentID != nil && *req.ParentID != "" {
		parsed, err := uuid.Parse(*req.ParentID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_parent_id", "Invalid parent comment ID")
			return
		}
		parentID = &parsed
	}

	comment, err := h.store.CreateComment(r.Context(), recordID, user.ID, req.Content, parentID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found or access denied")
			return
		}
		log.Printf("Error creating comment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create comment")
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

// GetComment handles GET /comments/:id
func (h *CommentHandler) GetComment(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid comment ID")
		return
	}

	comment, err := h.store.GetComment(r.Context(), commentID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Comment not found or access denied")
			return
		}
		log.Printf("Error getting comment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get comment")
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

// UpdateComment handles PATCH /comments/:id
func (h *CommentHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid comment ID")
		return
	}

	var req UpdateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content_required", "Comment content is required")
		return
	}

	comment, err := h.store.UpdateComment(r.Context(), commentID, user.ID, req.Content)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Comment not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You can only edit your own comments")
			return
		}
		log.Printf("Error updating comment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update comment")
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

// DeleteComment handles DELETE /comments/:id
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid comment ID")
		return
	}

	err = h.store.DeleteComment(r.Context(), commentID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Comment not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to delete this comment")
			return
		}
		log.Printf("Error deleting comment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete comment")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Comment deleted successfully",
	})
}

// ResolveComment handles POST /comments/:id/resolve
func (h *CommentHandler) ResolveComment(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	commentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid comment ID")
		return
	}

	var req ResolveCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	comment, err := h.store.ResolveComment(r.Context(), commentID, user.ID, req.Resolved)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Comment not found")
			return
		}
		log.Printf("Error resolving comment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to resolve comment")
		return
	}

	writeJSON(w, http.StatusOK, comment)
}
