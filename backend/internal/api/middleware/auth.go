package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/vibetable/backend/internal/api/handlers"
	"github.com/vibetable/backend/internal/store"
)

// AuthMiddleware creates a middleware that validates session tokens
// and adds the authenticated user to the request context
type AuthMiddleware struct {
	store *store.AuthStore
}

func NewAuthMiddleware(store *store.AuthStore) *AuthMiddleware {
	return &AuthMiddleware{store: store}
}

// extractToken gets the bearer token from Authorization header or query param
func extractToken(r *http.Request) string {
	// First check Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// Expect "Bearer <token>"
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	// Fall back to query param (for file downloads)
	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}

	return ""
}

// Optional adds user info to context if authenticated, but doesn't require it
// Use this for routes that work for both guests and authenticated users
func (m *AuthMiddleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			next.ServeHTTP(w, r)
			return
		}

		session, err := m.store.GetSessionByToken(r.Context(), token)
		if err != nil {
			// Invalid token, continue as guest
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.store.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			// User not found, continue as guest
			next.ServeHTTP(w, r)
			return
		}

		// Add user and token to context
		ctx := handlers.SetUserInContext(r.Context(), user)
		ctx = handlers.SetTokenInContext(ctx, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Required ensures the request is authenticated
// Returns 401 if no valid session
func (m *AuthMiddleware) Required(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized","message":"Authentication required"}`, http.StatusUnauthorized)
			return
		}

		session, err := m.store.GetSessionByToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, store.ErrExpired) {
				http.Error(w, `{"error":"session_expired","message":"Your session has expired. Please log in again."}`, http.StatusUnauthorized)
				return
			}
			http.Error(w, `{"error":"unauthorized","message":"Invalid session"}`, http.StatusUnauthorized)
			return
		}

		user, err := m.store.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			http.Error(w, `{"error":"unauthorized","message":"User not found"}`, http.StatusUnauthorized)
			return
		}

		// Add user and token to context
		ctx := handlers.SetUserInContext(r.Context(), user)
		ctx = handlers.SetTokenInContext(ctx, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
