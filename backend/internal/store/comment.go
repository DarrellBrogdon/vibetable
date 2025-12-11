package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/vibetable/backend/internal/models"
)

type CommentStore struct {
	db          DBTX
	baseStore   *BaseStore
	tableStore  *TableStore
	recordStore *RecordStore
}

func NewCommentStore(db DBTX, baseStore *BaseStore, tableStore *TableStore, recordStore *RecordStore) *CommentStore {
	return &CommentStore{
		db:          db,
		baseStore:   baseStore,
		tableStore:  tableStore,
		recordStore: recordStore,
	}
}

// getBaseIDForRecord returns the base ID for a record
func (s *CommentStore) getBaseIDForRecord(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	var baseID uuid.UUID
	err := s.db.QueryRow(ctx, `
		SELECT t.base_id
		FROM records r
		JOIN tables t ON r.table_id = t.id
		WHERE r.id = $1
	`, recordID).Scan(&baseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrNotFound
	}
	return baseID, err
}

// ListComments returns all comments for a record with user info
func (s *CommentStore) ListComments(ctx context.Context, recordID uuid.UUID, userID uuid.UUID) ([]*models.Comment, error) {
	// Verify user has access to the record's base
	baseID, err := s.getBaseIDForRecord(ctx, recordID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id, c.is_resolved, c.created_at, c.updated_at,
		       u.id, u.email, u.name, u.created_at, u.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.record_id = $1
		ORDER BY c.created_at ASC
	`, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	commentMap := make(map[uuid.UUID]*models.Comment)
	var rootComments []*models.Comment

	for rows.Next() {
		var c models.Comment
		var user models.User
		if err := rows.Scan(
			&c.ID, &c.RecordID, &c.UserID, &c.Content, &c.ParentID, &c.IsResolved, &c.CreatedAt, &c.UpdatedAt,
			&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		c.User = &user
		c.Replies = []*models.Comment{}
		commentMap[c.ID] = &c
	}

	// Build tree structure
	for _, comment := range commentMap {
		if comment.ParentID == nil {
			rootComments = append(rootComments, comment)
		} else {
			if parent, exists := commentMap[*comment.ParentID]; exists {
				parent.Replies = append(parent.Replies, comment)
			}
		}
	}

	if rootComments == nil {
		rootComments = []*models.Comment{}
	}

	return rootComments, rows.Err()
}

// CreateComment creates a new comment on a record
func (s *CommentStore) CreateComment(ctx context.Context, recordID uuid.UUID, userID uuid.UUID, content string, parentID *uuid.UUID) (*models.Comment, error) {
	// Verify user has at least view access (anyone can comment if they can see the record)
	baseID, err := s.getBaseIDForRecord(ctx, recordID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	// If parentID is provided, verify it exists and belongs to same record
	if parentID != nil {
		var parentRecordID uuid.UUID
		err = s.db.QueryRow(ctx, `SELECT record_id FROM comments WHERE id = $1`, parentID).Scan(&parentRecordID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("parent comment not found")
		}
		if err != nil {
			return nil, err
		}
		if parentRecordID != recordID {
			return nil, errors.New("parent comment belongs to a different record")
		}
	}

	var c models.Comment
	err = s.db.QueryRow(ctx, `
		INSERT INTO comments (record_id, user_id, content, parent_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, record_id, user_id, content, parent_id, is_resolved, created_at, updated_at
	`, recordID, userID, content, parentID).Scan(
		&c.ID, &c.RecordID, &c.UserID, &c.Content, &c.ParentID, &c.IsResolved, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Fetch user info
	var user models.User
	err = s.db.QueryRow(ctx, `
		SELECT id, email, name, created_at, updated_at FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	c.User = &user
	c.Replies = []*models.Comment{}

	return &c, nil
}

// GetComment returns a comment by ID
func (s *CommentStore) GetComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) (*models.Comment, error) {
	var c models.Comment
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT c.id, c.record_id, c.user_id, c.content, c.parent_id, c.is_resolved, c.created_at, c.updated_at,
		       u.id, u.email, u.name, u.created_at, u.updated_at
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.id = $1
	`, commentID).Scan(
		&c.ID, &c.RecordID, &c.UserID, &c.Content, &c.ParentID, &c.IsResolved, &c.CreatedAt, &c.UpdatedAt,
		&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Verify user has access
	baseID, err := s.getBaseIDForRecord(ctx, c.RecordID)
	if err != nil {
		return nil, err
	}
	_, err = s.baseStore.GetUserRole(ctx, baseID, userID)
	if err != nil {
		return nil, err
	}

	c.User = &user
	return &c, nil
}

// UpdateComment updates a comment's content (only by the comment author)
func (s *CommentStore) UpdateComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, content string) (*models.Comment, error) {
	// Get comment to verify ownership
	comment, err := s.GetComment(ctx, commentID, userID)
	if err != nil {
		return nil, err
	}

	// Only the author can edit their comment
	if comment.UserID != userID {
		return nil, ErrForbidden
	}

	err = s.db.QueryRow(ctx, `
		UPDATE comments SET content = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, record_id, user_id, content, parent_id, is_resolved, created_at, updated_at
	`, commentID, content).Scan(
		&comment.ID, &comment.RecordID, &comment.UserID, &comment.Content, &comment.ParentID, &comment.IsResolved, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// DeleteComment deletes a comment (only by the comment author or base owner/editor)
func (s *CommentStore) DeleteComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID) error {
	// Get comment first
	comment, err := s.GetComment(ctx, commentID, userID)
	if err != nil {
		return err
	}

	// Check if user is the author
	if comment.UserID != userID {
		// Check if user has edit access to the base
		baseID, err := s.getBaseIDForRecord(ctx, comment.RecordID)
		if err != nil {
			return err
		}
		role, err := s.baseStore.GetUserRole(ctx, baseID, userID)
		if err != nil {
			return err
		}
		if !role.CanEdit() {
			return ErrForbidden
		}
	}

	result, err := s.db.Exec(ctx, `DELETE FROM comments WHERE id = $1`, commentID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ResolveComment toggles the resolved status of a comment
func (s *CommentStore) ResolveComment(ctx context.Context, commentID uuid.UUID, userID uuid.UUID, resolved bool) (*models.Comment, error) {
	// Get comment to verify access
	comment, err := s.GetComment(ctx, commentID, userID)
	if err != nil {
		return nil, err
	}

	// Anyone with access can resolve/unresolve comments
	err = s.db.QueryRow(ctx, `
		UPDATE comments SET is_resolved = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, record_id, user_id, content, parent_id, is_resolved, created_at, updated_at
	`, commentID, resolved).Scan(
		&comment.ID, &comment.RecordID, &comment.UserID, &comment.Content, &comment.ParentID, &comment.IsResolved, &comment.CreatedAt, &comment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// GetCommentCount returns the number of comments on a record
func (s *CommentStore) GetCommentCount(ctx context.Context, recordID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM comments WHERE record_id = $1`, recordID).Scan(&count)
	return count, err
}
