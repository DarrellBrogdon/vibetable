package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidScopes(t *testing.T) {
	scopes := ValidScopes()

	assert.Len(t, scopes, 6)
	assert.Contains(t, scopes, ScopeReadAll)
	assert.Contains(t, scopes, ScopeWriteAll)
	assert.Contains(t, scopes, ScopeReadBases)
	assert.Contains(t, scopes, ScopeWriteBases)
	assert.Contains(t, scopes, ScopeReadData)
	assert.Contains(t, scopes, ScopeWriteData)
}

func TestIsValidScope(t *testing.T) {
	t.Run("returns true for valid scopes", func(t *testing.T) {
		assert.True(t, IsValidScope("read:all"))
		assert.True(t, IsValidScope("write:all"))
		assert.True(t, IsValidScope("read:bases"))
		assert.True(t, IsValidScope("write:bases"))
		assert.True(t, IsValidScope("read:data"))
		assert.True(t, IsValidScope("write:data"))
	})

	t.Run("returns false for invalid scopes", func(t *testing.T) {
		assert.False(t, IsValidScope("invalid"))
		assert.False(t, IsValidScope(""))
		assert.False(t, IsValidScope("read:everything"))
		assert.False(t, IsValidScope("admin"))
	})
}
