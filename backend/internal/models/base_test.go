package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollaboratorRole_CanEdit(t *testing.T) {
	t.Run("owner can edit", func(t *testing.T) {
		assert.True(t, RoleOwner.CanEdit())
	})

	t.Run("editor can edit", func(t *testing.T) {
		assert.True(t, RoleEditor.CanEdit())
	})

	t.Run("viewer cannot edit", func(t *testing.T) {
		assert.False(t, RoleViewer.CanEdit())
	})

	t.Run("empty role cannot edit", func(t *testing.T) {
		var emptyRole CollaboratorRole
		assert.False(t, emptyRole.CanEdit())
	})
}

func TestCollaboratorRole_CanDelete(t *testing.T) {
	t.Run("owner can delete", func(t *testing.T) {
		assert.True(t, RoleOwner.CanDelete())
	})

	t.Run("editor cannot delete", func(t *testing.T) {
		assert.False(t, RoleEditor.CanDelete())
	})

	t.Run("viewer cannot delete", func(t *testing.T) {
		assert.False(t, RoleViewer.CanDelete())
	})
}

func TestCollaboratorRole_CanManageCollaborators(t *testing.T) {
	t.Run("owner can manage collaborators", func(t *testing.T) {
		assert.True(t, RoleOwner.CanManageCollaborators())
	})

	t.Run("editor cannot manage collaborators", func(t *testing.T) {
		assert.False(t, RoleEditor.CanManageCollaborators())
	})

	t.Run("viewer cannot manage collaborators", func(t *testing.T) {
		assert.False(t, RoleViewer.CanManageCollaborators())
	})
}
