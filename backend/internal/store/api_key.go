package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vibetable/backend/internal/models"
)

type APIKeyStore struct {
	db DBTX
}

func NewAPIKeyStore(db DBTX) *APIKeyStore {
	return &APIKeyStore{db: db}
}

// generateAPIKey generates a random API key
func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// hashAPIKey creates a SHA-256 hash of the API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// Create creates a new API key and returns it with the token (only shown once)
func (s *APIKeyStore) Create(ctx context.Context, userID uuid.UUID, req *models.CreateAPIKeyRequest) (*models.APIKeyWithToken, error) {
	// Generate the API key
	token, err := generateAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	// Hash it for storage
	keyHash := hashAPIKey(token)

	// Get prefix (first 10 chars for identification)
	keyPrefix := token[:10]

	// Default scopes if not provided
	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = []string{string(models.ScopeReadAll), string(models.ScopeWriteAll)}
	}

	// Validate scopes
	for _, scope := range scopes {
		if !models.IsValidScope(scope) {
			return nil, fmt.Errorf("invalid scope: %s", scope)
		}
	}

	scopesJSON, err := json.Marshal(scopes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scopes: %w", err)
	}

	id := uuid.New()
	now := time.Now()

	query := `
		INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, scopes, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
	`

	apiKey := &models.APIKey{}
	var scopesRaw []byte

	err = s.db.QueryRow(ctx, query,
		id, userID, req.Name, keyHash, keyPrefix, scopesJSON, req.ExpiresAt, now,
	).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyPrefix,
		&scopesRaw, &apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	if err := json.Unmarshal(scopesRaw, &apiKey.Scopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
	}

	return &models.APIKeyWithToken{
		APIKey: *apiKey,
		Token:  token,
	}, nil
}

// GetByID retrieves an API key by ID
func (s *APIKeyStore) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
		FROM api_keys WHERE id = $1
	`

	apiKey := &models.APIKey{}
	var scopesRaw []byte

	err := s.db.QueryRow(ctx, query, id).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyPrefix,
		&scopesRaw, &apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(scopesRaw, &apiKey.Scopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
	}

	return apiKey, nil
}

// GetByToken retrieves an API key by its token (validates the hash)
func (s *APIKeyStore) GetByToken(ctx context.Context, token string) (*models.APIKey, error) {
	if len(token) < 10 {
		return nil, fmt.Errorf("invalid API key format")
	}

	keyPrefix := token[:10]
	keyHash := hashAPIKey(token)

	query := `
		SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
		FROM api_keys WHERE key_prefix = $1 AND key_hash = $2
	`

	apiKey := &models.APIKey{}
	var scopesRaw []byte

	err := s.db.QueryRow(ctx, query, keyPrefix, keyHash).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyPrefix,
		&scopesRaw, &apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(scopesRaw, &apiKey.Scopes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("API key has expired")
	}

	return apiKey, nil
}

// ListByUser lists all API keys for a user
func (s *APIKeyStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
		FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []models.APIKey
	for rows.Next() {
		var key models.APIKey
		var scopesRaw []byte

		if err := rows.Scan(
			&key.ID, &key.UserID, &key.Name, &key.KeyPrefix,
			&scopesRaw, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(scopesRaw, &key.Scopes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal scopes: %w", err)
		}

		keys = append(keys, key)
	}

	return keys, rows.Err()
}

// UpdateLastUsed updates the last_used_at timestamp
func (s *APIKeyStore) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := s.db.Exec(ctx, query, id)
	return err
}

// Delete deletes an API key
func (s *APIKeyStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	result, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	rows := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

// HasScope checks if an API key has a specific scope
func (s *APIKeyStore) HasScope(apiKey *models.APIKey, requiredScope string) bool {
	for _, scope := range apiKey.Scopes {
		// Check for wildcard scopes
		if scope == string(models.ScopeReadAll) && (requiredScope == string(models.ScopeReadBases) || requiredScope == string(models.ScopeReadData)) {
			return true
		}
		if scope == string(models.ScopeWriteAll) && (requiredScope == string(models.ScopeWriteBases) || requiredScope == string(models.ScopeWriteData)) {
			return true
		}
		if scope == requiredScope {
			return true
		}
	}
	return false
}
