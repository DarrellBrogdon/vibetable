package models

import (
	"time"

	"github.com/google/uuid"
)

// APIKey represents a user's API key for programmatic access
type APIKey struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	Name       string     `json:"name"`
	KeyHash    string     `json:"-"` // Never expose the hash
	KeyPrefix  string     `json:"key_prefix"`
	Scopes     []string   `json:"scopes"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// APIKeyWithToken is returned only when creating a new key
type APIKeyWithToken struct {
	APIKey
	Token string `json:"token"` // The actual API key, only shown once
}

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	Scopes    []string   `json:"scopes,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// APIKeyScope represents available scopes
type APIKeyScope string

const (
	ScopeReadAll    APIKeyScope = "read:all"
	ScopeWriteAll   APIKeyScope = "write:all"
	ScopeReadBases  APIKeyScope = "read:bases"
	ScopeWriteBases APIKeyScope = "write:bases"
	ScopeReadData   APIKeyScope = "read:data"
	ScopeWriteData  APIKeyScope = "write:data"
)

// ValidScopes returns all valid API key scopes
func ValidScopes() []APIKeyScope {
	return []APIKeyScope{
		ScopeReadAll,
		ScopeWriteAll,
		ScopeReadBases,
		ScopeWriteBases,
		ScopeReadData,
		ScopeWriteData,
	}
}

// IsValidScope checks if a scope is valid
func IsValidScope(scope string) bool {
	for _, s := range ValidScopes() {
		if string(s) == scope {
			return true
		}
	}
	return false
}
