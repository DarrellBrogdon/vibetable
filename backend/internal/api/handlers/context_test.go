package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/vibetable/backend/internal/models"
)

func TestSetUserInContext(t *testing.T) {
	t.Run("should store user in context", func(t *testing.T) {
		user := &models.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}

		ctx := context.Background()
		ctx = SetUserInContext(ctx, user)

		// Verify user is stored
		retrieved := ctx.Value(userContextKey)
		assert.NotNil(t, retrieved)
		assert.Equal(t, user, retrieved)
	})

	t.Run("should handle nil user", func(t *testing.T) {
		ctx := context.Background()
		ctx = SetUserInContext(ctx, nil)

		retrieved := ctx.Value(userContextKey)
		assert.Nil(t, retrieved)
	})
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("should retrieve user from context", func(t *testing.T) {
		user := &models.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}

		ctx := context.Background()
		ctx = SetUserInContext(ctx, user)

		retrieved := GetUserFromContext(ctx)
		assert.NotNil(t, retrieved)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("should return nil when no user in context", func(t *testing.T) {
		ctx := context.Background()

		retrieved := GetUserFromContext(ctx)
		assert.Nil(t, retrieved)
	})

	t.Run("should return nil when context value is wrong type", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, userContextKey, "not a user")

		retrieved := GetUserFromContext(ctx)
		assert.Nil(t, retrieved)
	})
}

func TestSetTokenInContext(t *testing.T) {
	t.Run("should store token in context", func(t *testing.T) {
		token := "test-session-token"

		ctx := context.Background()
		ctx = SetTokenInContext(ctx, token)

		retrieved := ctx.Value(tokenContextKey)
		assert.Equal(t, token, retrieved)
	})

	t.Run("should handle empty token", func(t *testing.T) {
		ctx := context.Background()
		ctx = SetTokenInContext(ctx, "")

		retrieved := ctx.Value(tokenContextKey)
		assert.Equal(t, "", retrieved)
	})
}

func TestGetTokenFromContext(t *testing.T) {
	t.Run("should retrieve token from context", func(t *testing.T) {
		token := "test-session-token"

		ctx := context.Background()
		ctx = SetTokenInContext(ctx, token)

		retrieved := GetTokenFromContext(ctx)
		assert.Equal(t, token, retrieved)
	})

	t.Run("should return empty string when no token in context", func(t *testing.T) {
		ctx := context.Background()

		retrieved := GetTokenFromContext(ctx)
		assert.Equal(t, "", retrieved)
	})

	t.Run("should return empty string when context value is wrong type", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, tokenContextKey, 12345)

		retrieved := GetTokenFromContext(ctx)
		assert.Equal(t, "", retrieved)
	})
}

func TestContextIntegration(t *testing.T) {
	t.Run("should store both user and token in same context", func(t *testing.T) {
		user := &models.User{
			ID:    uuid.New(),
			Email: "test@example.com",
		}
		token := "session-token"

		ctx := context.Background()
		ctx = SetUserInContext(ctx, user)
		ctx = SetTokenInContext(ctx, token)

		retrievedUser := GetUserFromContext(ctx)
		retrievedToken := GetTokenFromContext(ctx)

		assert.NotNil(t, retrievedUser)
		assert.Equal(t, user.ID, retrievedUser.ID)
		assert.Equal(t, token, retrievedToken)
	})
}
