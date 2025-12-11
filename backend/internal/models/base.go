package models

import (
	"time"

	"github.com/google/uuid"
)

type CollaboratorRole string

const (
	RoleOwner  CollaboratorRole = "owner"
	RoleEditor CollaboratorRole = "editor"
	RoleViewer CollaboratorRole = "viewer"
)

type Base struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Populated when fetching with user context
	Role *CollaboratorRole `json:"role,omitempty"`
}

type BaseCollaborator struct {
	ID        uuid.UUID        `json:"id"`
	BaseID    uuid.UUID        `json:"base_id"`
	UserID    uuid.UUID        `json:"user_id"`
	Role      CollaboratorRole `json:"role"`
	CreatedAt time.Time        `json:"created_at"`

	// Populated when fetching collaborators list
	User *User `json:"user,omitempty"`
}

// CanEdit returns true if the role has edit permissions
func (r CollaboratorRole) CanEdit() bool {
	return r == RoleOwner || r == RoleEditor
}

// CanDelete returns true if the role can delete the base
func (r CollaboratorRole) CanDelete() bool {
	return r == RoleOwner
}

// CanManageCollaborators returns true if the role can manage collaborators
func (r CollaboratorRole) CanManageCollaborators() bool {
	return r == RoleOwner
}
