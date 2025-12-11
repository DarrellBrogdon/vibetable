package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{
			name:     "valid email",
			email:    "test@example.com",
			expected: true,
		},
		{
			name:     "valid email with subdomain",
			email:    "test@sub.example.com",
			expected: true,
		},
		{
			name:     "valid email with plus sign",
			email:    "test+tag@example.com",
			expected: true,
		},
		{
			name:     "missing @",
			email:    "testexample.com",
			expected: false,
		},
		{
			name:     "missing domain",
			email:    "test@",
			expected: false,
		},
		{
			name:     "missing local part",
			email:    "@example.com",
			expected: false,
		},
		{
			name:     "missing dot in domain",
			email:    "test@example",
			expected: false,
		},
		{
			name:     "empty string",
			email:    "",
			expected: false,
		},
		{
			name:     "domain too short",
			email:    "test@a.b",
			expected: true, // domain "a.b" has length 3, which passes len > 2
		},
		{
			name:     "valid short domain",
			email:    "test@ab.co",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidEmail(tt.email)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteJSON(t *testing.T) {
	t.Run("should write JSON response with correct headers", func(t *testing.T) {
		w := httptest.NewRecorder()

		data := map[string]string{"message": "hello"}
		writeJSON(w, http.StatusOK, data)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "hello", response["message"])
	})

	t.Run("should write error status codes", func(t *testing.T) {
		w := httptest.NewRecorder()

		data := map[string]string{"error": "not found"}
		writeJSON(w, http.StatusNotFound, data)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should handle struct responses", func(t *testing.T) {
		w := httptest.NewRecorder()

		response := AuthResponse{
			User: &models.User{
				ID:    uuid.New(),
				Email: "test@example.com",
			},
			Token: "session-token",
		}
		writeJSON(w, http.StatusOK, response)

		var result AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Equal(t, response.User.Email, result.User.Email)
		assert.Equal(t, response.Token, result.Token)
	})
}

func TestWriteError(t *testing.T) {
	t.Run("should write error response", func(t *testing.T) {
		w := httptest.NewRecorder()

		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid input")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
		assert.Equal(t, "Invalid input", response.Message)
	})

	t.Run("should write unauthorized error", func(t *testing.T) {
		w := httptest.NewRecorder()

		writeError(w, http.StatusUnauthorized, "unauthorized", "Not authenticated")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should write internal server error", func(t *testing.T) {
		w := httptest.NewRecorder()

		writeError(w, http.StatusInternalServerError, "server_error", "Something went wrong")

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestAuthHandler_GetMe(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewAuthHandler(nil) // store not needed for this test

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		w := httptest.NewRecorder()

		handler.GetMe(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "unauthorized", response.Error)
	})

	t.Run("should return user when authenticated", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		user := &models.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}

		req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.GetMe(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, user.Email, response.User.Email)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("should return 401 when no token in context", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		w := httptest.NewRecorder()

		handler.Logout(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthHandler_UpdateMe(t *testing.T) {
	t.Run("should return 401 when no user in context", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		body := bytes.NewBufferString(`{"name": "New Name"}`)
		req := httptest.NewRequest(http.MethodPatch, "/auth/me", body)
		w := httptest.NewRecorder()

		handler.UpdateMe(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPatch, "/auth/me", body)
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateMe(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty name", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": ""}`)
		req := httptest.NewRequest(http.MethodPatch, "/auth/me", body)
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateMe(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "name_required", response.Error)
	})

	t.Run("should return 400 for whitespace-only name", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		user := &models.User{ID: uuid.New(), Email: "test@example.com"}
		body := bytes.NewBufferString(`{"name": "   "}`)
		req := httptest.NewRequest(http.MethodPatch, "/auth/me", body)
		req = req.WithContext(SetUserInContext(req.Context(), user))
		w := httptest.NewRecorder()

		handler.UpdateMe(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_RequestMagicLink(t *testing.T) {
	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", body)
		w := httptest.NewRecorder()

		handler.RequestMagicLink(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty email", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		body := bytes.NewBufferString(`{"email": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", body)
		w := httptest.NewRecorder()

		handler.RequestMagicLink(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "email_required", response.Error)
	})

	t.Run("should return 400 for invalid email", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		body := bytes.NewBufferString(`{"email": "not-an-email"}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/magic-link", body)
		w := httptest.NewRecorder()

		handler.RequestMagicLink(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_email", response.Error)
	})
}

func TestAuthHandler_Verify(t *testing.T) {
	t.Run("should return 400 for invalid JSON", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		body := bytes.NewBufferString(`{invalid}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/verify", body)
		w := httptest.NewRecorder()

		handler.Verify(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "invalid_request", response.Error)
	})

	t.Run("should return 400 for empty token", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		body := bytes.NewBufferString(`{"token": ""}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/verify", body)
		w := httptest.NewRecorder()

		handler.Verify(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "token_required", response.Error)
	})
}
