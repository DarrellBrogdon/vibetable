package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/api/handlers"
	"github.com/vibetable/backend/internal/models"
	"github.com/vibetable/backend/internal/store"
)

func TestExtractToken(t *testing.T) {
	t.Run("extracts Bearer token from Authorization header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer my-token-123")

		token := extractToken(req)
		assert.Equal(t, "my-token-123", token)
	})

	t.Run("handles lowercase bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "bearer my-token-123")

		token := extractToken(req)
		assert.Equal(t, "my-token-123", token)
	})

	t.Run("trims whitespace from token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer   my-token-123  ")

		token := extractToken(req)
		assert.Equal(t, "my-token-123", token)
	})

	t.Run("extracts token from query parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test?token=query-token-456", nil)

		token := extractToken(req)
		assert.Equal(t, "query-token-456", token)
	})

	t.Run("prefers Authorization header over query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test?token=query-token", nil)
		req.Header.Set("Authorization", "Bearer header-token")

		token := extractToken(req)
		assert.Equal(t, "header-token", token)
	})

	t.Run("returns empty string when no token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)

		token := extractToken(req)
		assert.Equal(t, "", token)
	})

	t.Run("returns empty string for invalid Authorization format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

		token := extractToken(req)
		assert.Equal(t, "", token)
	})

	t.Run("returns empty string for Authorization with no space", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "BearerNoSpace")

		token := extractToken(req)
		assert.Equal(t, "", token)
	})
}

func TestNewAuthMiddleware(t *testing.T) {
	middleware := NewAuthMiddleware(nil)
	assert.NotNil(t, middleware)
}

func TestAuthMiddleware_Required(t *testing.T) {
	t.Run("returns 401 when no token provided", func(t *testing.T) {
		middleware := NewAuthMiddleware(nil)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.Required(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, nextCalled)
		assert.Contains(t, w.Body.String(), "unauthorized")
	})
}

func TestAuthMiddleware_Optional(t *testing.T) {
	t.Run("continues without user when no token", func(t *testing.T) {
		middleware := NewAuthMiddleware(nil)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		middleware.Optional(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)
	})

	t.Run("continues as guest when session lookup fails", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		authStore := store.NewAuthStore(mock)
		middleware := NewAuthMiddleware(authStore)

		// Expect session lookup to fail
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("invalid-token").
			WillReturnError(assert.AnError)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			// Check that no user is in context
			user := handlers.GetUserFromContext(r.Context())
			assert.Nil(t, user)
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		middleware.Optional(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("continues as guest when user lookup fails", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		authStore := store.NewAuthStore(mock)
		middleware := NewAuthMiddleware(authStore)

		userID := uuid.New()
		sessionID := uuid.New()
		now := time.Now().UTC()
		expiresAt := now.Add(24 * time.Hour)

		// Expect session lookup to succeed
		sessionRows := pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(sessionID, userID, "valid-token", expiresAt, now)
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("valid-token").
			WillReturnRows(sessionRows)

		// Expect user lookup to fail
		mock.ExpectQuery("SELECT (.+) FROM users").
			WithArgs(userID).
			WillReturnError(assert.AnError)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		middleware.Optional(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("sets user in context when authenticated", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		authStore := store.NewAuthStore(mock)
		middleware := NewAuthMiddleware(authStore)

		userID := uuid.New()
		sessionID := uuid.New()
		now := time.Now().UTC()
		expiresAt := now.Add(24 * time.Hour)

		// Expect session lookup to succeed
		sessionRows := pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(sessionID, userID, "valid-token", expiresAt, now)
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("valid-token").
			WillReturnRows(sessionRows)

		// Expect user lookup to succeed
		userRows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", "Test User", now, now)
		mock.ExpectQuery("SELECT (.+) FROM users").
			WithArgs(userID).
			WillReturnRows(userRows)

		var receivedUser *models.User
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			receivedUser = handlers.GetUserFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		middleware.Optional(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)
		assert.NotNil(t, receivedUser)
		assert.Equal(t, "test@example.com", receivedUser.Email)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthMiddleware_Required_WithMock(t *testing.T) {
	t.Run("returns 401 when session lookup fails", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		authStore := store.NewAuthStore(mock)
		middleware := NewAuthMiddleware(authStore)

		// Expect session lookup to fail
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("invalid-token").
			WillReturnError(assert.AnError)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		middleware.Required(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, nextCalled)
		assert.Contains(t, w.Body.String(), "Invalid session")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns 401 with expired message when session is expired", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		authStore := store.NewAuthStore(mock)
		middleware := NewAuthMiddleware(authStore)

		// Expect session lookup to return expired error
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("expired-token").
			WillReturnError(store.ErrExpired)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		w := httptest.NewRecorder()

		middleware.Required(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, nextCalled)
		assert.Contains(t, w.Body.String(), "session_expired")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns 401 when user lookup fails", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		authStore := store.NewAuthStore(mock)
		middleware := NewAuthMiddleware(authStore)

		userID := uuid.New()
		sessionID := uuid.New()
		now := time.Now().UTC()
		expiresAt := now.Add(24 * time.Hour)

		// Expect session lookup to succeed
		sessionRows := pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(sessionID, userID, "valid-token", expiresAt, now)
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("valid-token").
			WillReturnRows(sessionRows)

		// Expect user lookup to fail
		mock.ExpectQuery("SELECT (.+) FROM users").
			WithArgs(userID).
			WillReturnError(assert.AnError)

		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		middleware.Required(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.False(t, nextCalled)
		assert.Contains(t, w.Body.String(), "User not found")

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("calls next handler when authenticated", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		authStore := store.NewAuthStore(mock)
		middleware := NewAuthMiddleware(authStore)

		userID := uuid.New()
		sessionID := uuid.New()
		now := time.Now().UTC()
		expiresAt := now.Add(24 * time.Hour)

		// Expect session lookup to succeed
		sessionRows := pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(sessionID, userID, "valid-token", expiresAt, now)
		mock.ExpectQuery("SELECT (.+) FROM sessions").
			WithArgs("valid-token").
			WillReturnRows(sessionRows)

		// Expect user lookup to succeed
		userRows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", "Test User", now, now)
		mock.ExpectQuery("SELECT (.+) FROM users").
			WithArgs(userID).
			WillReturnRows(userRows)

		var receivedUser *models.User
		var receivedToken string
		nextCalled := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
			receivedUser = handlers.GetUserFromContext(r.Context())
			receivedToken = handlers.GetTokenFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		middleware.Required(next).ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, nextCalled)
		assert.NotNil(t, receivedUser)
		assert.Equal(t, "test@example.com", receivedUser.Email)
		assert.Equal(t, "valid-token", receivedToken)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}
