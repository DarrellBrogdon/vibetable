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

type FormHandler struct {
	store *store.FormStore
}

func NewFormHandler(store *store.FormStore) *FormHandler {
	return &FormHandler{store: store}
}

// Request types
type CreateFormRequest struct {
	Name string `json:"name"`
}

type UpdateFormRequest struct {
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
	SuccessMessage   *string `json:"success_message,omitempty"`
	RedirectURL      *string `json:"redirect_url,omitempty"`
	SubmitButtonText *string `json:"submit_button_text,omitempty"`
}

type UpdateFormFieldsRequest struct {
	Fields []FormFieldUpdate `json:"fields"`
}

type FormFieldUpdate struct {
	FieldID    string  `json:"field_id"`
	Label      *string `json:"label,omitempty"`
	HelpText   *string `json:"help_text,omitempty"`
	IsRequired bool    `json:"is_required"`
	IsVisible  bool    `json:"is_visible"`
	Position   int     `json:"position"`
}

type SubmitFormRequest struct {
	Values map[string]interface{} `json:"values"`
}

// ListForms handles GET /tables/:tableId/forms
func (h *FormHandler) ListForms(w http.ResponseWriter, r *http.Request) {
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

	forms, err := h.store.ListFormsForTable(r.Context(), tableID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found or access denied")
			return
		}
		log.Printf("Error listing forms: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to list forms")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"forms": forms,
	})
}

// CreateForm handles POST /tables/:tableId/forms
func (h *FormHandler) CreateForm(w http.ResponseWriter, r *http.Request) {
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

	var req CreateFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Name == "" {
		req.Name = "New Form"
	}

	form, err := h.store.CreateForm(r.Context(), tableID, req.Name, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Table not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to create forms")
			return
		}
		log.Printf("Error creating form: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create form")
		return
	}

	log.Printf("Form created: %s (id=%s)", form.Name, form.ID)
	writeJSON(w, http.StatusCreated, form)
}

// GetForm handles GET /forms/:id
func (h *FormHandler) GetForm(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid form ID")
		return
	}

	form, err := h.store.GetForm(r.Context(), formID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Form not found or access denied")
			return
		}
		log.Printf("Error getting form: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get form")
		return
	}

	writeJSON(w, http.StatusOK, form)
}

// UpdateForm handles PATCH /forms/:id
func (h *FormHandler) UpdateForm(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid form ID")
		return
	}

	var req UpdateFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SuccessMessage != nil {
		updates["success_message"] = *req.SuccessMessage
	}
	if req.RedirectURL != nil {
		updates["redirect_url"] = *req.RedirectURL
	}
	if req.SubmitButtonText != nil {
		updates["submit_button_text"] = *req.SubmitButtonText
	}

	form, err := h.store.UpdateForm(r.Context(), formID, user.ID, updates)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Form not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to update this form")
			return
		}
		log.Printf("Error updating form: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update form")
		return
	}

	writeJSON(w, http.StatusOK, form)
}

// UpdateFormFields handles PATCH /forms/:id/fields
func (h *FormHandler) UpdateFormFields(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid form ID")
		return
	}

	var req UpdateFormFieldsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	// Convert to models
	var fields []models.FormField
	for _, f := range req.Fields {
		fieldID, err := uuid.Parse(f.FieldID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_field_id", "Invalid field ID: "+f.FieldID)
			return
		}
		fields = append(fields, models.FormField{
			FieldID:    fieldID,
			Label:      f.Label,
			HelpText:   f.HelpText,
			IsRequired: f.IsRequired,
			IsVisible:  f.IsVisible,
			Position:   f.Position,
		})
	}

	err = h.store.UpdateFormFields(r.Context(), formID, user.ID, fields)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Form not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to update this form")
			return
		}
		log.Printf("Error updating form fields: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update form fields")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Form fields updated successfully",
	})
}

// DeleteForm handles DELETE /forms/:id
func (h *FormHandler) DeleteForm(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	formID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_id", "Invalid form ID")
		return
	}

	err = h.store.DeleteForm(r.Context(), formID, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Form not found")
			return
		}
		if errors.Is(err, store.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", "You don't have permission to delete this form")
			return
		}
		log.Printf("Error deleting form: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to delete form")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Form deleted successfully",
	})
}

// GetPublicForm handles GET /public/forms/:token (no auth required)
func (h *FormHandler) GetPublicForm(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "Token is required")
		return
	}

	form, err := h.store.GetPublicForm(r.Context(), token)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Form not found or inactive")
			return
		}
		log.Printf("Error getting public form: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to get form")
		return
	}

	writeJSON(w, http.StatusOK, form)
}

// SubmitPublicForm handles POST /public/forms/:token (no auth required)
func (h *FormHandler) SubmitPublicForm(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "invalid_token", "Token is required")
		return
	}

	var req SubmitFormRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	record, err := h.store.SubmitPublicForm(r.Context(), token, req.Values)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Form not found or inactive")
			return
		}
		if err.Error() == "form is not active" {
			writeError(w, http.StatusBadRequest, "form_inactive", "This form is no longer accepting submissions")
			return
		}
		if len(err.Error()) > 15 && err.Error()[:15] == "required field " {
			writeError(w, http.StatusBadRequest, "validation_error", err.Error())
			return
		}
		log.Printf("Error submitting form: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to submit form")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":   "Form submitted successfully",
		"record_id": record.ID,
	})
}
