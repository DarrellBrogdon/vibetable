package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vibetable/backend/internal/models"
)

func TestNewAPIKeyStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewAPIKeyStore(mock)
	assert.NotNil(t, store)
}

func TestGenerateAPIKey(t *testing.T) {
	key1, err := generateAPIKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key1)

	key2, err := generateAPIKey()
	require.NoError(t, err)
	assert.NotEmpty(t, key2)

	// Keys should be different
	assert.NotEqual(t, key1, key2)

	// Keys should be at least 10 chars (for prefix)
	assert.GreaterOrEqual(t, len(key1), 10)
}

func TestHashAPIKey(t *testing.T) {
	key := "test-api-key-12345"
	hash := hashAPIKey(key)

	// Hash should be deterministic
	assert.Equal(t, hash, hashAPIKey(key))

	// Different keys should have different hashes
	hash2 := hashAPIKey("different-key-12345")
	assert.NotEqual(t, hash, hash2)

	// Hash should be hex encoded SHA-256 (64 chars)
	assert.Len(t, hash, 64)
}

func TestAPIKeyStore_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("creates API key successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		userID := uuid.New()
		apiKeyID := uuid.New()
		now := time.Now().UTC()
		scopes := []string{"read:all", "write:all"}
		scopesJSON, _ := json.Marshal(scopes)

		req := &models.CreateAPIKeyRequest{
			Name:   "Test API Key",
			Scopes: scopes,
		}

		// Mock insert - use AnyArg for dynamic values
		insertRows := pgxmock.NewRows([]string{"id", "user_id", "name", "key_prefix", "scopes", "last_used_at", "expires_at", "created_at"}).
			AddRow(apiKeyID, userID, req.Name, "testprefix", scopesJSON, nil, nil, now)
		mock.ExpectQuery("INSERT INTO api_keys").
			WithArgs(pgxmock.AnyArg(), userID, req.Name, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), (*time.Time)(nil), pgxmock.AnyArg()).
			WillReturnRows(insertRows)

		result, err := store.Create(ctx, userID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Token)
		assert.Equal(t, req.Name, result.APIKey.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creates API key with default scopes", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		userID := uuid.New()
		apiKeyID := uuid.New()
		now := time.Now().UTC()
		defaultScopes := []string{"read:all", "write:all"}
		scopesJSON, _ := json.Marshal(defaultScopes)

		req := &models.CreateAPIKeyRequest{
			Name: "Test API Key",
			// No scopes provided - should default
		}

		// Mock insert
		insertRows := pgxmock.NewRows([]string{"id", "user_id", "name", "key_prefix", "scopes", "last_used_at", "expires_at", "created_at"}).
			AddRow(apiKeyID, userID, req.Name, "testprefix", scopesJSON, nil, nil, now)
		mock.ExpectQuery("INSERT INTO api_keys").
			WithArgs(pgxmock.AnyArg(), userID, req.Name, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), (*time.Time)(nil), pgxmock.AnyArg()).
			WillReturnRows(insertRows)

		result, err := store.Create(ctx, userID, req)
		require.NoError(t, err)
		assert.NotNil(t, result)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error for invalid scope", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		userID := uuid.New()

		req := &models.CreateAPIKeyRequest{
			Name:   "Test API Key",
			Scopes: []string{"invalid:scope"},
		}

		result, err := store.Create(ctx, userID, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid scope")
	})
}

func TestAPIKeyStore_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns API key when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		apiKeyID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()
		scopes := []string{"read:all"}
		scopesJSON, _ := json.Marshal(scopes)

		rows := pgxmock.NewRows([]string{"id", "user_id", "name", "key_prefix", "scopes", "last_used_at", "expires_at", "created_at"}).
			AddRow(apiKeyID, userID, "Test Key", "testprefix", scopesJSON, nil, nil, now)
		mock.ExpectQuery("SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at").
			WithArgs(apiKeyID).
			WillReturnRows(rows)

		apiKey, err := store.GetByID(ctx, apiKeyID)
		require.NoError(t, err)
		assert.Equal(t, apiKeyID, apiKey.ID)
		assert.Equal(t, userID, apiKey.UserID)
		assert.Equal(t, "Test Key", apiKey.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		apiKeyID := uuid.New()

		mock.ExpectQuery("SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at").
			WithArgs(apiKeyID).
			WillReturnError(pgx.ErrNoRows)

		apiKey, err := store.GetByID(ctx, apiKeyID)
		assert.Error(t, err)
		assert.Nil(t, apiKey)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAPIKeyStore_GetByToken(t *testing.T) {
	ctx := context.Background()

	t.Run("returns API key when token valid", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		apiKeyID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()
		token := "abcdefghij1234567890abcdefghij123456789012"
		keyPrefix := token[:10]
		keyHash := hashAPIKey(token)
		scopes := []string{"read:all"}
		scopesJSON, _ := json.Marshal(scopes)

		rows := pgxmock.NewRows([]string{"id", "user_id", "name", "key_prefix", "scopes", "last_used_at", "expires_at", "created_at"}).
			AddRow(apiKeyID, userID, "Test Key", keyPrefix, scopesJSON, nil, nil, now)
		mock.ExpectQuery("SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at").
			WithArgs(keyPrefix, keyHash).
			WillReturnRows(rows)

		apiKey, err := store.GetByToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, apiKeyID, apiKey.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error for short token", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)

		apiKey, err := store.GetByToken(ctx, "short")
		assert.Error(t, err)
		assert.Nil(t, apiKey)
		assert.Contains(t, err.Error(), "invalid API key format")
	})

	t.Run("returns error for expired key", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		apiKeyID := uuid.New()
		userID := uuid.New()
		now := time.Now().UTC()
		expired := time.Now().Add(-24 * time.Hour)
		token := "abcdefghij1234567890abcdefghij123456789012"
		keyPrefix := token[:10]
		keyHash := hashAPIKey(token)
		scopes := []string{"read:all"}
		scopesJSON, _ := json.Marshal(scopes)

		rows := pgxmock.NewRows([]string{"id", "user_id", "name", "key_prefix", "scopes", "last_used_at", "expires_at", "created_at"}).
			AddRow(apiKeyID, userID, "Test Key", keyPrefix, scopesJSON, nil, &expired, now)
		mock.ExpectQuery("SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at").
			WithArgs(keyPrefix, keyHash).
			WillReturnRows(rows)

		apiKey, err := store.GetByToken(ctx, token)
		assert.Error(t, err)
		assert.Nil(t, apiKey)
		assert.Contains(t, err.Error(), "expired")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAPIKeyStore_ListByUser(t *testing.T) {
	ctx := context.Background()

	t.Run("returns API keys for user", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		userID := uuid.New()
		apiKeyID1 := uuid.New()
		apiKeyID2 := uuid.New()
		now := time.Now().UTC()
		scopes := []string{"read:all"}
		scopesJSON, _ := json.Marshal(scopes)

		rows := pgxmock.NewRows([]string{"id", "user_id", "name", "key_prefix", "scopes", "last_used_at", "expires_at", "created_at"}).
			AddRow(apiKeyID1, userID, "Key 1", "prefix0001", scopesJSON, nil, nil, now).
			AddRow(apiKeyID2, userID, "Key 2", "prefix0002", scopesJSON, nil, nil, now)
		mock.ExpectQuery("SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at").
			WithArgs(userID).
			WillReturnRows(rows)

		keys, err := store.ListByUser(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, keys, 2)
		assert.Equal(t, apiKeyID1, keys[0].ID)
		assert.Equal(t, apiKeyID2, keys[1].ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns empty slice when no keys", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		userID := uuid.New()

		rows := pgxmock.NewRows([]string{"id", "user_id", "name", "key_prefix", "scopes", "last_used_at", "expires_at", "created_at"})
		mock.ExpectQuery("SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at").
			WithArgs(userID).
			WillReturnRows(rows)

		keys, err := store.ListByUser(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, keys)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAPIKeyStore_UpdateLastUsed(t *testing.T) {
	ctx := context.Background()

	t.Run("updates last used timestamp", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		apiKeyID := uuid.New()

		mock.ExpectExec("UPDATE api_keys SET last_used_at").
			WithArgs(apiKeyID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		err = store.UpdateLastUsed(ctx, apiKeyID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAPIKeyStore_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes API key successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		apiKeyID := uuid.New()

		mock.ExpectExec("DELETE FROM api_keys WHERE id").
			WithArgs(apiKeyID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.Delete(ctx, apiKeyID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when key not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAPIKeyStore(mock)
		apiKeyID := uuid.New()

		mock.ExpectExec("DELETE FROM api_keys WHERE id").
			WithArgs(apiKeyID).
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = store.Delete(ctx, apiKeyID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAPIKeyStore_HasScope(t *testing.T) {
	store := &APIKeyStore{}

	t.Run("returns true for exact scope match", func(t *testing.T) {
		apiKey := &models.APIKey{
			Scopes: []string{"read:bases"},
		}
		assert.True(t, store.HasScope(apiKey, "read:bases"))
		assert.False(t, store.HasScope(apiKey, "write:bases"))
	})

	t.Run("returns true for read:all wildcard", func(t *testing.T) {
		apiKey := &models.APIKey{
			Scopes: []string{"read:all"},
		}
		assert.True(t, store.HasScope(apiKey, "read:bases"))
		assert.True(t, store.HasScope(apiKey, "read:data"))
		assert.False(t, store.HasScope(apiKey, "write:bases"))
	})

	t.Run("returns true for write:all wildcard", func(t *testing.T) {
		apiKey := &models.APIKey{
			Scopes: []string{"write:all"},
		}
		assert.True(t, store.HasScope(apiKey, "write:bases"))
		assert.True(t, store.HasScope(apiKey, "write:data"))
		assert.False(t, store.HasScope(apiKey, "read:bases"))
	})

	t.Run("returns false for no matching scope", func(t *testing.T) {
		apiKey := &models.APIKey{
			Scopes: []string{"read:bases"},
		}
		assert.False(t, store.HasScope(apiKey, "write:data"))
		assert.False(t, store.HasScope(apiKey, "invalid:scope"))
	})
}
