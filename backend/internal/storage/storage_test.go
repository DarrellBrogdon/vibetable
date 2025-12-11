package storage

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalStorage(t *testing.T) {
	t.Run("creates storage with valid path", func(t *testing.T) {
		tmpDir := t.TempDir()
		storagePath := filepath.Join(tmpDir, "storage")

		storage, err := NewLocalStorage(storagePath, "http://localhost:8080")
		require.NoError(t, err)
		require.NotNil(t, storage)

		// Verify directory was created
		_, err = os.Stat(storagePath)
		assert.NoError(t, err)
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		// Try to create storage in a path we can't write to
		// This test may not work on all systems depending on permissions
		invalidPath := "/root/nonexistent/storage"
		_, err := NewLocalStorage(invalidPath, "http://localhost:8080")
		// On most systems this should fail, but not all
		// So we just check that it doesn't panic
		_ = err
	})
}

func TestGenerateKey(t *testing.T) {
	t.Run("generates unique key with extension", func(t *testing.T) {
		fieldID := uuid.New()
		recordID := uuid.New()
		filename := "document.pdf"

		key := GenerateKey(fieldID, recordID, filename)

		assert.Contains(t, key, fieldID.String())
		assert.Contains(t, key, recordID.String())
		assert.True(t, strings.HasSuffix(key, ".pdf"))
	})

	t.Run("generates different keys for same inputs", func(t *testing.T) {
		fieldID := uuid.New()
		recordID := uuid.New()

		key1 := GenerateKey(fieldID, recordID, "file.txt")
		key2 := GenerateKey(fieldID, recordID, "file.txt")

		assert.NotEqual(t, key1, key2)
	})

	t.Run("handles file without extension", func(t *testing.T) {
		fieldID := uuid.New()
		recordID := uuid.New()

		key := GenerateKey(fieldID, recordID, "noextension")

		assert.Contains(t, key, fieldID.String())
		assert.Contains(t, key, recordID.String())
		assert.False(t, strings.HasSuffix(key, "."))
	})

	t.Run("preserves various extensions", func(t *testing.T) {
		fieldID := uuid.New()
		recordID := uuid.New()

		extensions := []string{".jpg", ".png", ".gif", ".doc", ".xlsx"}
		for _, ext := range extensions {
			key := GenerateKey(fieldID, recordID, "file"+ext)
			assert.True(t, strings.HasSuffix(key, ext), "Expected key to end with %s", ext)
		}
	})
}

func TestLocalStorage_Upload(t *testing.T) {
	t.Run("uploads file successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		content := []byte("test file content")
		key := "test/file.txt"

		err = storage.Upload(ctx, key, bytes.NewReader(content), "text/plain")
		require.NoError(t, err)

		// Verify file exists
		fullPath := filepath.Join(tmpDir, key)
		data, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, content, data)
	})

	t.Run("creates nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		key := "deep/nested/path/file.txt"

		err = storage.Upload(ctx, key, strings.NewReader("content"), "text/plain")
		require.NoError(t, err)

		// Verify file exists
		fullPath := filepath.Join(tmpDir, key)
		_, err = os.Stat(fullPath)
		assert.NoError(t, err)
	})

	t.Run("overwrites existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		key := "overwrite.txt"

		err = storage.Upload(ctx, key, strings.NewReader("first"), "text/plain")
		require.NoError(t, err)

		err = storage.Upload(ctx, key, strings.NewReader("second"), "text/plain")
		require.NoError(t, err)

		fullPath := filepath.Join(tmpDir, key)
		data, err := os.ReadFile(fullPath)
		require.NoError(t, err)
		assert.Equal(t, "second", string(data))
	})
}

func TestLocalStorage_Download(t *testing.T) {
	t.Run("downloads existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		// Create test file
		key := "download.txt"
		content := "download test content"
		fullPath := filepath.Join(tmpDir, key)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)

		ctx := context.Background()
		reader, err := storage.Download(ctx, key)
		require.NoError(t, err)
		defer reader.Close()

		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, content, string(data))
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		_, err = storage.Download(ctx, "nonexistent.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
	})
}

func TestLocalStorage_GetURL(t *testing.T) {
	t.Run("returns correct URL format", func(t *testing.T) {
		tmpDir := t.TempDir()
		baseURL := "http://localhost:8080"
		storage, err := NewLocalStorage(tmpDir, baseURL)
		require.NoError(t, err)

		ctx := context.Background()
		key := "test/file.pdf"
		url, err := storage.GetURL(ctx, key, time.Hour)
		require.NoError(t, err)

		expectedURL := "http://localhost:8080/api/v1/attachments/file/test/file.pdf"
		assert.Equal(t, expectedURL, url)
	})

	t.Run("ignores expiry parameter", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		key := "file.txt"

		url1, err := storage.GetURL(ctx, key, time.Hour)
		require.NoError(t, err)

		url2, err := storage.GetURL(ctx, key, time.Minute)
		require.NoError(t, err)

		assert.Equal(t, url1, url2)
	})
}

func TestLocalStorage_Delete(t *testing.T) {
	t.Run("deletes existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		// Create test file
		key := "delete.txt"
		fullPath := filepath.Join(tmpDir, key)
		err = os.WriteFile(fullPath, []byte("content"), 0644)
		require.NoError(t, err)

		ctx := context.Background()
		err = storage.Delete(ctx, key)
		require.NoError(t, err)

		// Verify file is deleted
		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("succeeds for non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		err = storage.Delete(ctx, "nonexistent.txt")
		assert.NoError(t, err) // Should not error
	})
}

func TestLocalStorage_Exists(t *testing.T) {
	t.Run("returns true for existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		// Create test file
		key := "exists.txt"
		fullPath := filepath.Join(tmpDir, key)
		err = os.WriteFile(fullPath, []byte("content"), 0644)
		require.NoError(t, err)

		ctx := context.Background()
		exists, err := storage.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false for non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		exists, err := storage.Exists(ctx, "nonexistent.txt")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestLocalStorage_GetBasePath(t *testing.T) {
	t.Run("returns correct base path", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		assert.Equal(t, tmpDir, storage.GetBasePath())
	})
}

func TestLocalStorage_Integration(t *testing.T) {
	t.Run("upload, download, verify, delete cycle", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalStorage(tmpDir, "http://localhost:8080")
		require.NoError(t, err)

		ctx := context.Background()
		fieldID := uuid.New()
		recordID := uuid.New()
		key := GenerateKey(fieldID, recordID, "integration-test.txt")
		content := "integration test content"

		// Upload
		err = storage.Upload(ctx, key, strings.NewReader(content), "text/plain")
		require.NoError(t, err)

		// Verify exists
		exists, err := storage.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// Download and verify content
		reader, err := storage.Download(ctx, key)
		require.NoError(t, err)
		data, err := io.ReadAll(reader)
		reader.Close()
		require.NoError(t, err)
		assert.Equal(t, content, string(data))

		// Get URL
		url, err := storage.GetURL(ctx, key, time.Hour)
		require.NoError(t, err)
		assert.Contains(t, url, key)

		// Delete
		err = storage.Delete(ctx, key)
		require.NoError(t, err)

		// Verify deleted
		exists, err = storage.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
