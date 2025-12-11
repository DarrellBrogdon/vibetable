package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrExpired            = errors.New("expired")
	ErrAlreadyUsed        = errors.New("already used")
	ErrInvalidToken       = errors.New("invalid token")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type AuthStore struct {
	db DBTX
}

func NewAuthStore(db DBTX) *AuthStore {
	return &AuthStore{db: db}
}

// GenerateToken creates a cryptographically secure random token
func GenerateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// --- User operations ---

// GetUserByEmail finds a user by email address
func (s *AuthStore) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user by ID
func (s *AuthStore) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user (without password - for backwards compatibility)
func (s *AuthStore) CreateUser(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, `
		INSERT INTO users (email)
		VALUES ($1)
		RETURNING id, email, name, password_hash, created_at, updated_at
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUserWithPassword creates a new user with a hashed password
func (s *AuthStore) CreateUserWithPassword(ctx context.Context, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.QueryRow(ctx, `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, name, password_hash, created_at, updated_at
	`, email, string(hashedPassword)).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetOrCreateUser finds an existing user or creates a new one
func (s *AuthStore) GetOrCreateUser(ctx context.Context, email string) (*models.User, bool, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err == nil {
		return user, false, nil // existing user
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, false, err
	}

	user, err = s.CreateUser(ctx, email)
	if err != nil {
		return nil, false, err
	}
	return user, true, nil // new user
}

// UpdateUserName updates a user's name
func (s *AuthStore) UpdateUserName(ctx context.Context, id uuid.UUID, name string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, `
		UPDATE users SET name = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, name, password_hash, created_at, updated_at
	`, id, name).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// --- Magic Link operations ---

// CreateMagicLink creates a new magic link for an email
func (s *AuthStore) CreateMagicLink(ctx context.Context, email string, expiresIn time.Duration) (*models.MagicLink, error) {
	token, err := GenerateToken(32)
	if err != nil {
		return nil, err
	}

	var link models.MagicLink
	err = s.db.QueryRow(ctx, `
		INSERT INTO magic_links (email, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, email, token, expires_at, used_at, created_at
	`, email, token, time.Now().Add(expiresIn)).Scan(
		&link.ID, &link.Email, &link.Token, &link.ExpiresAt, &link.UsedAt, &link.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &link, nil
}

// VerifyMagicLink validates and consumes a magic link token
func (s *AuthStore) VerifyMagicLink(ctx context.Context, token string) (*models.MagicLink, error) {
	var link models.MagicLink
	err := s.db.QueryRow(ctx, `
		SELECT id, email, token, expires_at, used_at, created_at
		FROM magic_links WHERE token = $1
	`, token).Scan(&link.ID, &link.Email, &link.Token, &link.ExpiresAt, &link.UsedAt, &link.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	if link.UsedAt != nil {
		return nil, ErrAlreadyUsed
	}

	if time.Now().After(link.ExpiresAt) {
		return nil, ErrExpired
	}

	// Mark as used
	_, err = s.db.Exec(ctx, `
		UPDATE magic_links SET used_at = NOW() WHERE id = $1
	`, link.ID)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

// CleanupExpiredMagicLinks removes expired magic links
func (s *AuthStore) CleanupExpiredMagicLinks(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM magic_links WHERE expires_at < NOW()
	`)
	return err
}

// --- Session operations ---

// CreateSession creates a new session for a user
func (s *AuthStore) CreateSession(ctx context.Context, userID uuid.UUID, expiresIn time.Duration) (*models.Session, error) {
	token, err := GenerateToken(32)
	if err != nil {
		return nil, err
	}

	var session models.Session
	err = s.db.QueryRow(ctx, `
		INSERT INTO sessions (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token, expires_at, created_at
	`, userID, token, time.Now().Add(expiresIn)).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionByToken validates a session token and returns the session
func (s *AuthStore) GetSessionByToken(ctx context.Context, token string) (*models.Session, error) {
	var session models.Session
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions WHERE token = $1
	`, token).Scan(&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		s.db.Exec(ctx, "DELETE FROM sessions WHERE id = $1", session.ID)
		return nil, ErrExpired
	}

	return &session, nil
}

// DeleteSession removes a session (logout)
func (s *AuthStore) DeleteSession(ctx context.Context, token string) error {
	result, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteUserSessions removes all sessions for a user
func (s *AuthStore) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

// CleanupExpiredSessions removes expired sessions
func (s *AuthStore) CleanupExpiredSessions(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	return err
}

// --- Password operations ---

// VerifyPassword checks if the provided password matches the user's stored hash
func (s *AuthStore) VerifyPassword(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user has a password set
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

// SetUserPassword sets or updates a user's password
func (s *AuthStore) SetUserPassword(ctx context.Context, userID uuid.UUID, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	result, err := s.db.Exec(ctx, `
		UPDATE users SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`, userID, string(hashedPassword))
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Password Reset Token operations ---

// CreatePasswordResetToken creates a new password reset token for a user
func (s *AuthStore) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, expiresIn time.Duration) (*models.PasswordResetToken, error) {
	token, err := GenerateToken(32)
	if err != nil {
		return nil, err
	}

	var resetToken models.PasswordResetToken
	err = s.db.QueryRow(ctx, `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token, expires_at, used_at, created_at
	`, userID, token, time.Now().Add(expiresIn)).Scan(
		&resetToken.ID, &resetToken.UserID, &resetToken.Token, &resetToken.ExpiresAt, &resetToken.UsedAt, &resetToken.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &resetToken, nil
}

// VerifyPasswordResetToken validates and consumes a password reset token
func (s *AuthStore) VerifyPasswordResetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	var resetToken models.PasswordResetToken
	err := s.db.QueryRow(ctx, `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens WHERE token = $1
	`, token).Scan(&resetToken.ID, &resetToken.UserID, &resetToken.Token, &resetToken.ExpiresAt, &resetToken.UsedAt, &resetToken.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	if resetToken.UsedAt != nil {
		return nil, ErrAlreadyUsed
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return nil, ErrExpired
	}

	// Mark as used
	_, err = s.db.Exec(ctx, `
		UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1
	`, resetToken.ID)
	if err != nil {
		return nil, err
	}

	return &resetToken, nil
}

// CleanupExpiredPasswordResetTokens removes expired password reset tokens
func (s *AuthStore) CleanupExpiredPasswordResetTokens(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM password_reset_tokens WHERE expires_at < NOW()
	`)
	return err
}
