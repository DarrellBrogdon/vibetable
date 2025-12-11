package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	csrfTokenHeader = "X-CSRF-Token"
	csrfCookieName  = "csrf_token"
	csrfTokenLength = 32
	csrfTokenMaxAge = 3600 // 1 hour
)

type CSRFMiddleware struct {
	secret []byte
}

func NewCSRFMiddleware() *CSRFMiddleware {
	secret := os.Getenv("CSRF_SECRET")
	if secret == "" {
		// Generate a random secret if not configured (not recommended for production)
		b := make([]byte, 32)
		rand.Read(b)
		secret = base64.StdEncoding.EncodeToString(b)
	}
	return &CSRFMiddleware{
		secret: []byte(secret),
	}
}

// generateToken creates a new CSRF token
func (m *CSRFMiddleware) generateToken() (string, error) {
	// Generate random bytes
	tokenBytes := make([]byte, csrfTokenLength)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	// Create timestamp
	timestamp := time.Now().Unix()

	// Create HMAC signature
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(tokenBytes)
	mac.Write([]byte{byte(timestamp >> 56), byte(timestamp >> 48), byte(timestamp >> 40), byte(timestamp >> 32),
		byte(timestamp >> 24), byte(timestamp >> 16), byte(timestamp >> 8), byte(timestamp)})
	signature := mac.Sum(nil)

	// Combine: token + timestamp (8 bytes) + signature (32 bytes)
	combined := make([]byte, csrfTokenLength+8+32)
	copy(combined, tokenBytes)
	combined[csrfTokenLength] = byte(timestamp >> 56)
	combined[csrfTokenLength+1] = byte(timestamp >> 48)
	combined[csrfTokenLength+2] = byte(timestamp >> 40)
	combined[csrfTokenLength+3] = byte(timestamp >> 32)
	combined[csrfTokenLength+4] = byte(timestamp >> 24)
	combined[csrfTokenLength+5] = byte(timestamp >> 16)
	combined[csrfTokenLength+6] = byte(timestamp >> 8)
	combined[csrfTokenLength+7] = byte(timestamp)
	copy(combined[csrfTokenLength+8:], signature)

	return base64.URLEncoding.EncodeToString(combined), nil
}

// validateToken verifies a CSRF token
func (m *CSRFMiddleware) validateToken(token string) bool {
	combined, err := base64.URLEncoding.DecodeString(token)
	if err != nil || len(combined) != csrfTokenLength+8+32 {
		return false
	}

	tokenBytes := combined[:csrfTokenLength]
	timestampBytes := combined[csrfTokenLength : csrfTokenLength+8]
	providedSignature := combined[csrfTokenLength+8:]

	// Verify signature
	mac := hmac.New(sha256.New, m.secret)
	mac.Write(tokenBytes)
	mac.Write(timestampBytes)
	expectedSignature := mac.Sum(nil)

	if !hmac.Equal(providedSignature, expectedSignature) {
		return false
	}

	// Check timestamp (token should not be older than max age)
	timestamp := int64(timestampBytes[0])<<56 | int64(timestampBytes[1])<<48 |
		int64(timestampBytes[2])<<40 | int64(timestampBytes[3])<<32 |
		int64(timestampBytes[4])<<24 | int64(timestampBytes[5])<<16 |
		int64(timestampBytes[6])<<8 | int64(timestampBytes[7])

	if time.Now().Unix()-timestamp > csrfTokenMaxAge {
		return false
	}

	return true
}

// Protect is middleware that enforces CSRF protection on state-changing methods
func (m *CSRFMiddleware) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF for API key authenticated requests
		if r.Header.Get("Authorization") != "" && strings.HasPrefix(r.Header.Get("Authorization"), "Bearer vt_") {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from header
		token := r.Header.Get(csrfTokenHeader)
		if token == "" {
			http.Error(w, `{"error": "csrf_required", "message": "CSRF token required"}`, http.StatusForbidden)
			return
		}

		// Validate token
		if !m.validateToken(token) {
			http.Error(w, `{"error": "csrf_invalid", "message": "Invalid or expired CSRF token"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// TokenHandler returns an endpoint that provides a CSRF token
func (m *CSRFMiddleware) TokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := m.generateToken()
	if err != nil {
		http.Error(w, `{"error": "internal_error", "message": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// Set as cookie for double-submit pattern
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   csrfTokenMaxAge,
		HttpOnly: false, // JavaScript needs to read this
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})

	// Also return in response body and header
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set(csrfTokenHeader, token)
	w.Write([]byte(`{"csrf_token": "` + token + `"}`))
}
