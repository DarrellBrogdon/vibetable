package handlers

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/store"
)

type AttachmentHandler struct {
	store *store.AttachmentStore
}

func NewAttachmentHandler(store *store.AttachmentStore) *AttachmentHandler {
	return &AttachmentHandler{store: store}
}

// UploadAttachment handles POST /records/:recordId/fields/:fieldId/attachments
func (h *AttachmentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
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

	fieldID, err := uuid.Parse(chi.URLParam(r, "fieldId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid field ID")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file_required", "File is required")
		return
	}
	defer file.Close()

	// Get content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	attachment, err := h.store.CreateAttachment(r.Context(), recordID, fieldID, user.ID, header.Filename, contentType, header.Size, file)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record or field not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to add attachments")
			return
		}
		log.Printf("Error creating attachment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to upload attachment")
		return
	}

	writeJSON(w, http.StatusCreated, attachment)
}

// ListAttachments handles GET /records/:recordId/fields/:fieldId/attachments
func (h *AttachmentHandler) ListAttachments(w http.ResponseWriter, r *http.Request) {
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

	fieldID, err := uuid.Parse(chi.URLParam(r, "fieldId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid field ID")
		return
	}

	attachments, err := h.store.ListAttachmentsForField(r.Context(), recordID, fieldID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Record not found or access denied")
			return
		}
		log.Printf("Error listing attachments: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list attachments")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"attachments": attachments,
	})
}

// GetAttachment handles GET /attachments/:id
func (h *AttachmentHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	attachmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid attachment ID")
		return
	}

	attachment, err := h.store.GetAttachment(r.Context(), attachmentID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Attachment not found or access denied")
			return
		}
		log.Printf("Error getting attachment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get attachment")
		return
	}

	writeJSON(w, http.StatusOK, attachment)
}

// DownloadAttachment handles GET /attachments/:id/download
func (h *AttachmentHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	attachmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid attachment ID")
		return
	}

	reader, attachment, err := h.store.DownloadAttachment(r.Context(), attachmentID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Attachment not found or access denied")
			return
		}
		log.Printf("Error downloading attachment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to download attachment")
		return
	}
	defer reader.Close()

	// Set headers for download
	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+attachment.Filename+"\"")
	w.Header().Set("Content-Length", strconv.FormatInt(attachment.SizeBytes, 10))

	// Stream the file
	if _, err := io.Copy(w, reader); err != nil {
		log.Printf("Error streaming attachment: %v", err)
	}
}

// DeleteAttachment handles DELETE /attachments/:id
func (h *AttachmentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	attachmentID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid attachment ID")
		return
	}

	err = h.store.DeleteAttachment(r.Context(), attachmentID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Attachment not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to delete this attachment")
			return
		}
		log.Printf("Error deleting attachment: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete attachment")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Attachment deleted successfully",
	})
}
