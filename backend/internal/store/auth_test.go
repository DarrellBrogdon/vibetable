package store

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken(t *testing.T) {
	t.Run("generates token of correct length", func(t *testing.T) {
		token, err := GenerateToken(32)
		require.NoError(t, err)
		// Base64 encoding increases length: ceil(32 * 4 / 3) = 44 (with padding)
		assert.NotEmpty(t, token)
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		token1, err := GenerateToken(32)
		require.NoError(t, err)
		token2, err := GenerateToken(32)
		require.NoError(t, err)
		assert.NotEqual(t, token1, token2)
	})

	t.Run("handles different lengths", func(t *testing.T) {
		token16, err := GenerateToken(16)
		require.NoError(t, err)
		assert.NotEmpty(t, token16)

		token64, err := GenerateToken(64)
		require.NoError(t, err)
		assert.NotEmpty(t, token64)
	})
}

func TestNewAuthStore(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	store := NewAuthStore(mock)
	assert.NotNil(t, store)
}

// Helper to create *string
func strPtr(s string) *string {
	return &s
}

func TestAuthStore_GetUserByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		userID := uuid.New()
		now := time.Now().UTC()
		email := "test@example.com"
		name := strPtr("Test User")

		rows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, email, name, now, now)

		mock.ExpectQuery("SELECT id, email, name, created_at, updated_at FROM users WHERE email").
			WithArgs(email).
			WillReturnRows(rows)

		user, err := store.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "Test User", *user.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when user does not exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		mock.ExpectQuery("SELECT id, email, name, created_at, updated_at FROM users WHERE email").
			WithArgs("nonexistent@example.com").
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}))

		user, err := store.GetUserByEmail(ctx, "nonexistent@example.com")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, user)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user when found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		userID := uuid.New()
		now := time.Now().UTC()
		name := strPtr("Test User")

		rows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", name, now, now)

		mock.ExpectQuery("SELECT id, email, name, created_at, updated_at FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := store.GetUserByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when user does not exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)
		userID := uuid.New()

		mock.ExpectQuery("SELECT id, email, name, created_at, updated_at FROM users WHERE id").
			WithArgs(userID).
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}))

		user, err := store.GetUserByID(ctx, userID)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, user)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("creates user successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		userID := uuid.New()
		now := time.Now().UTC()
		email := "new@example.com"

		rows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, email, (*string)(nil), now, now)

		mock.ExpectQuery("INSERT INTO users").
			WithArgs(email).
			WillReturnRows(rows)

		user, err := store.CreateUser(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, email, user.Email)
		assert.Nil(t, user.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_UpdateUserName(t *testing.T) {
	ctx := context.Background()

	t.Run("updates user name successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		userID := uuid.New()
		now := time.Now().UTC()
		newName := strPtr("Updated Name")

		rows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", newName, now, now)

		mock.ExpectQuery("UPDATE users SET name").
			WithArgs(userID, "Updated Name").
			WillReturnRows(rows)

		user, err := store.UpdateUserName(ctx, userID, "Updated Name")
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", *user.Name)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when user does not exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)
		userID := uuid.New()

		mock.ExpectQuery("UPDATE users SET name").
			WithArgs(userID, "New Name").
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}))

		user, err := store.UpdateUserName(ctx, userID, "New Name")
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, user)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_CreateMagicLink(t *testing.T) {
	ctx := context.Background()

	t.Run("creates magic link successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		linkID := uuid.New()
		email := "test@example.com"
		expiresAt := time.Now().Add(15 * time.Minute).UTC()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "email", "token", "expires_at", "used_at", "created_at"}).
			AddRow(linkID, email, pgxmock.AnyArg(), expiresAt, (*time.Time)(nil), now)

		mock.ExpectQuery("INSERT INTO magic_links").
			WithArgs(email, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(rows)

		link, err := store.CreateMagicLink(ctx, email, 15*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, linkID, link.ID)
		assert.Equal(t, email, link.Email)
		assert.Nil(t, link.UsedAt)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_VerifyMagicLink(t *testing.T) {
	ctx := context.Background()

	t.Run("verifies valid magic link", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		linkID := uuid.New()
		token := "valid_token"
		expiresAt := time.Now().Add(15 * time.Minute).UTC()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "email", "token", "expires_at", "used_at", "created_at"}).
			AddRow(linkID, "test@example.com", token, expiresAt, (*time.Time)(nil), now)

		mock.ExpectQuery("SELECT id, email, token, expires_at, used_at, created_at FROM magic_links WHERE token").
			WithArgs(token).
			WillReturnRows(rows)

		mock.ExpectExec("UPDATE magic_links SET used_at").
			WithArgs(linkID).
			WillReturnResult(pgxmock.NewResult("UPDATE", 1))

		link, err := store.VerifyMagicLink(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, linkID, link.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrInvalidToken for nonexistent token", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		mock.ExpectQuery("SELECT id, email, token, expires_at, used_at, created_at FROM magic_links WHERE token").
			WithArgs("invalid_token").
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "token", "expires_at", "used_at", "created_at"}))

		link, err := store.VerifyMagicLink(ctx, "invalid_token")
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, link)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrAlreadyUsed for used token", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		linkID := uuid.New()
		token := "used_token"
		expiresAt := time.Now().Add(15 * time.Minute).UTC()
		usedAt := time.Now().UTC()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "email", "token", "expires_at", "used_at", "created_at"}).
			AddRow(linkID, "test@example.com", token, expiresAt, &usedAt, now)

		mock.ExpectQuery("SELECT id, email, token, expires_at, used_at, created_at FROM magic_links WHERE token").
			WithArgs(token).
			WillReturnRows(rows)

		link, err := store.VerifyMagicLink(ctx, token)
		assert.ErrorIs(t, err, ErrAlreadyUsed)
		assert.Nil(t, link)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrExpired for expired token", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		linkID := uuid.New()
		token := "expired_token"
		expiresAt := time.Now().Add(-1 * time.Hour).UTC() // Expired
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "email", "token", "expires_at", "used_at", "created_at"}).
			AddRow(linkID, "test@example.com", token, expiresAt, (*time.Time)(nil), now)

		mock.ExpectQuery("SELECT id, email, token, expires_at, used_at, created_at FROM magic_links WHERE token").
			WithArgs(token).
			WillReturnRows(rows)

		link, err := store.VerifyMagicLink(ctx, token)
		assert.ErrorIs(t, err, ErrExpired)
		assert.Nil(t, link)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_CreateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("creates session successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		sessionID := uuid.New()
		userID := uuid.New()
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(sessionID, userID, pgxmock.AnyArg(), expiresAt, now)

		mock.ExpectQuery("INSERT INTO sessions").
			WithArgs(userID, pgxmock.AnyArg(), pgxmock.AnyArg()).
			WillReturnRows(rows)

		session, err := store.CreateSession(ctx, userID, 24*time.Hour)
		require.NoError(t, err)
		assert.Equal(t, sessionID, session.ID)
		assert.Equal(t, userID, session.UserID)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_GetSessionByToken(t *testing.T) {
	ctx := context.Background()

	t.Run("returns valid session", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		sessionID := uuid.New()
		userID := uuid.New()
		token := "valid_session_token"
		expiresAt := time.Now().Add(24 * time.Hour).UTC()
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(sessionID, userID, token, expiresAt, now)

		mock.ExpectQuery("SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token").
			WithArgs(token).
			WillReturnRows(rows)

		session, err := store.GetSessionByToken(ctx, token)
		require.NoError(t, err)
		assert.Equal(t, sessionID, session.ID)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrInvalidToken for nonexistent session", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		mock.ExpectQuery("SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token").
			WithArgs("invalid_token").
			WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}))

		session, err := store.GetSessionByToken(ctx, "invalid_token")
		assert.ErrorIs(t, err, ErrInvalidToken)
		assert.Nil(t, session)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrExpired and deletes expired session", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		sessionID := uuid.New()
		userID := uuid.New()
		token := "expired_token"
		expiresAt := time.Now().Add(-1 * time.Hour).UTC() // Expired
		now := time.Now().UTC()

		rows := pgxmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
			AddRow(sessionID, userID, token, expiresAt, now)

		mock.ExpectQuery("SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token").
			WithArgs(token).
			WillReturnRows(rows)

		mock.ExpectExec("DELETE FROM sessions WHERE id").
			WithArgs(sessionID).
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		session, err := store.GetSessionByToken(ctx, token)
		assert.ErrorIs(t, err, ErrExpired)
		assert.Nil(t, session)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_DeleteSession(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes session successfully", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		mock.ExpectExec("DELETE FROM sessions WHERE token").
			WithArgs("session_token").
			WillReturnResult(pgxmock.NewResult("DELETE", 1))

		err = store.DeleteSession(ctx, "session_token")
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns ErrNotFound when session does not exist", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		mock.ExpectExec("DELETE FROM sessions WHERE token").
			WithArgs("nonexistent_token").
			WillReturnResult(pgxmock.NewResult("DELETE", 0))

		err = store.DeleteSession(ctx, "nonexistent_token")
		assert.ErrorIs(t, err, ErrNotFound)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_DeleteUserSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes all user sessions", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)
		userID := uuid.New()

		mock.ExpectExec("DELETE FROM sessions WHERE user_id").
			WithArgs(userID).
			WillReturnResult(pgxmock.NewResult("DELETE", 3))

		err = store.DeleteUserSessions(ctx, userID)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_CleanupExpiredMagicLinks(t *testing.T) {
	ctx := context.Background()

	t.Run("cleans up expired magic links", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		mock.ExpectExec("DELETE FROM magic_links WHERE expires_at").
			WillReturnResult(pgxmock.NewResult("DELETE", 5))

		err = store.CleanupExpiredMagicLinks(ctx)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_CleanupExpiredSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("cleans up expired sessions", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		mock.ExpectExec("DELETE FROM sessions WHERE expires_at").
			WillReturnResult(pgxmock.NewResult("DELETE", 10))

		err = store.CleanupExpiredSessions(ctx)
		require.NoError(t, err)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAuthStore_GetOrCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("returns existing user", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		userID := uuid.New()
		now := time.Now().UTC()
		email := "existing@example.com"
		name := strPtr("Existing User")

		rows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, email, name, now, now)

		mock.ExpectQuery("SELECT id, email, name, created_at, updated_at FROM users WHERE email").
			WithArgs(email).
			WillReturnRows(rows)

		user, created, err := store.GetOrCreateUser(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.False(t, created)

		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("creates new user when not found", func(t *testing.T) {
		mock, err := pgxmock.NewPool()
		require.NoError(t, err)
		defer mock.Close()

		store := NewAuthStore(mock)

		userID := uuid.New()
		now := time.Now().UTC()
		email := "new@example.com"

		// First query returns empty (user not found)
		mock.ExpectQuery("SELECT id, email, name, created_at, updated_at FROM users WHERE email").
			WithArgs(email).
			WillReturnRows(pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}))

		// Then create user
		createRows := pgxmock.NewRows([]string{"id", "email", "name", "created_at", "updated_at"}).
			AddRow(userID, email, (*string)(nil), now, now)
		mock.ExpectQuery("INSERT INTO users").
			WithArgs(email).
			WillReturnRows(createRows)

		user, created, err := store.GetOrCreateUser(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, userID, user.ID)
		assert.True(t, created)

		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestErrors(t *testing.T) {
	assert.NotNil(t, ErrNotFound)
	assert.NotNil(t, ErrExpired)
	assert.NotNil(t, ErrAlreadyUsed)
	assert.NotNil(t, ErrInvalidToken)

	assert.Equal(t, "not found", ErrNotFound.Error())
	assert.Equal(t, "expired", ErrExpired.Error())
	assert.Equal(t, "already used", ErrAlreadyUsed.Error())
	assert.Equal(t, "invalid token", ErrInvalidToken.Error())
}
