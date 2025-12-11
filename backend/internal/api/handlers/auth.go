package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

const (
	SessionExpiry        = 7 * 24 * time.Hour // 7 days
	PasswordResetExpiry  = 1 * time.Hour      // 1 hour
	MinPasswordLength    = 8
)

type AuthHandler struct {
	store *store.AuthStore
}

func NewAuthHandler(store *store.AuthStore) *AuthHandler {
	return &AuthHandler{store: store}
}

// Request types
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"password"`
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// Response types
type AuthResponse struct {
	User    *models.User `json:"user"`
	Token   string       `json:"token,omitempty"`
	Message string       `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Helper functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err string, message string) {
	writeJSON(w, status, ErrorResponse{Error: err, Message: message})
}

// isValidEmail does basic email validation
func isValidEmail(email string) bool {
	// Simple validation: contains @ and at least one dot after @
	atIndex := strings.Index(email, "@")
	if atIndex < 1 {
		return false
	}
	domain := email[atIndex+1:]
	return strings.Contains(domain, ".") && len(domain) > 2
}

// Login handles POST /auth/login
// Authenticates with email/password, auto-creates account if email doesn't exist
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email_required", "Email is required")
		return
	}

	if !isValidEmail(email) {
		writeError(w, http.StatusBadRequest, "invalid_email", "Please enter a valid email address")
		return
	}

	if len(req.Password) < MinPasswordLength {
		writeError(w, http.StatusBadRequest, "password_too_short", "Password must be at least 8 characters")
		return
	}

	// Check if user exists
	existingUser, err := h.store.GetUserByEmail(r.Context(), email)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		log.Printf("Error checking user: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "An error occurred")
		return
	}

	var user *models.User

	if existingUser == nil {
		// New user - create account with password
		user, err = h.store.CreateUserWithPassword(r.Context(), email, req.Password)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			writeError(w, http.StatusInternalServerError, "server_error", "Failed to create account")
			return
		}
		log.Printf("New user registered: %s", user.Email)
	} else {
		// Existing user - verify password
		user, err = h.store.VerifyPassword(r.Context(), email, req.Password)
		if err != nil {
			if errors.Is(err, store.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
				return
			}
			log.Printf("Error verifying password: %v", err)
			writeError(w, http.StatusInternalServerError, "server_error", "An error occurred")
			return
		}
	}

	// Create session
	session, err := h.store.CreateSession(r.Context(), user.ID, SessionExpiry)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create session")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{
		User:  user,
		Token: session.Token,
	})
}

// ForgotPassword handles POST /auth/forgot-password
// Creates a password reset token and logs it (dev) or sends email (prod)
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" {
		writeError(w, http.StatusBadRequest, "email_required", "Email is required")
		return
	}

	if !isValidEmail(email) {
		writeError(w, http.StatusBadRequest, "invalid_email", "Please enter a valid email address")
		return
	}

	// Find user by email
	user, err := h.store.GetUserByEmail(r.Context(), email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// Don't reveal if email exists - return success anyway
			writeJSON(w, http.StatusOK, map[string]string{
				"message": "If an account exists with this email, you will receive a password reset link.",
			})
			return
		}
		log.Printf("Error finding user: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "An error occurred")
		return
	}

	// Create password reset token
	resetToken, err := h.store.CreatePasswordResetToken(r.Context(), user.ID, PasswordResetExpiry)
	if err != nil {
		log.Printf("Error creating password reset token: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to create reset token")
		return
	}

	// In development, log the reset URL
	// In production, this would send an email via Resend/SendGrid/etc.
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	resetURL := frontendURL + "/auth/reset-password?token=" + resetToken.Token

	log.Printf("========================================")
	log.Printf("PASSWORD RESET for %s", email)
	log.Printf("%s", resetURL)
	log.Printf("Expires: %s", resetToken.ExpiresAt.Format(time.RFC3339))
	log.Printf("========================================")

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "If an account exists with this email, you will receive a password reset link.",
	})
}

// ResetPassword handles POST /auth/reset-password
// Validates reset token and updates user's password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	if req.Token == "" {
		writeError(w, http.StatusBadRequest, "token_required", "Token is required")
		return
	}

	if len(req.NewPassword) < MinPasswordLength {
		writeError(w, http.StatusBadRequest, "password_too_short", "Password must be at least 8 characters")
		return
	}

	// Verify reset token
	resetToken, err := h.store.VerifyPasswordResetToken(r.Context(), req.Token)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrInvalidToken):
			writeError(w, http.StatusBadRequest, "invalid_token", "Invalid or expired reset link")
		case errors.Is(err, store.ErrExpired):
			writeError(w, http.StatusBadRequest, "expired_token", "This reset link has expired")
		case errors.Is(err, store.ErrAlreadyUsed):
			writeError(w, http.StatusBadRequest, "used_token", "This reset link has already been used")
		default:
			log.Printf("Error verifying reset token: %v", err)
			writeError(w, http.StatusInternalServerError, "server_error", "Failed to verify reset token")
		}
		return
	}

	// Update user's password
	err = h.store.SetUserPassword(r.Context(), resetToken.UserID, req.NewPassword)
	if err != nil {
		log.Printf("Error setting password: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update password")
		return
	}

	// Invalidate all existing sessions for this user (security measure)
	if err := h.store.DeleteUserSessions(r.Context(), resetToken.UserID); err != nil {
		log.Printf("Error deleting user sessions: %v", err)
		// Don't fail the request - password was already updated
	}

	log.Printf("Password reset completed for user %s", resetToken.UserID)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Password has been reset successfully. Please log in with your new password.",
	})
}

// GetMe handles GET /auth/me
// Returns the current authenticated user
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{User: user})
}

// UpdateMe handles PATCH /auth/me
// Updates the current user's profile
func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user := GetUserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name_required", "Name is required")
		return
	}

	updatedUser, err := h.store.UpdateUserName(r.Context(), user.ID, name)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		writeError(w, http.StatusInternalServerError, "server_error", "Failed to update profile")
		return
	}

	writeJSON(w, http.StatusOK, AuthResponse{User: updatedUser})
}

// Logout handles POST /auth/logout
// Invalidates the current session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := GetTokenFromContext(r.Context())
	if token == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")
		return
	}

	if err := h.store.DeleteSession(r.Context(), token); err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			log.Printf("Error deleting session: %v", err)
		}
		// Still return success even if session not found (idempotent)
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}
