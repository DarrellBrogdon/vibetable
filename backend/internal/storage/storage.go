package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Storage defines the interface for file storage backends
type Storage interface {
	// Upload stores a file and returns the storage key
	Upload(ctx context.Context, key string, data io.Reader, contentType string) error

	// Download retrieves a file by its storage key
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// GetURL returns a URL to access the file (may be signed/temporary)
	GetURL(ctx context.Context, key string, expiry time.Duration) (string, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, key string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)
}

// LocalStorage implements Storage using the local filesystem
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage backend
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// GenerateKey generates a unique storage key for a file
func GenerateKey(fieldID, recordID uuid.UUID, filename string) string {
	ext := filepath.Ext(filename)
	uniqueID := uuid.New().String()
	return fmt.Sprintf("%s/%s/%s%s", fieldID.String(), recordID.String(), uniqueID, ext)
}

// Upload stores a file on the local filesystem
func (s *LocalStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) error {
	fullPath := filepath.Join(s.basePath, key)

	// Create directory structure (0700 = owner rwx only, principle of least privilege)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	if _, err := io.Copy(file, data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Download retrieves a file from the local filesystem
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// GetURL returns a URL to access the file
// For local storage, this returns a static URL path
func (s *LocalStorage) GetURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// For local storage, we serve files through an API endpoint
	// The expiry parameter is ignored for local storage
	return fmt.Sprintf("%s/api/v1/attachments/file/%s", s.baseURL, key), nil
}

// Delete removes a file from the local filesystem
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(s.basePath, key)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil // File already doesn't exist
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if a file exists on the local filesystem
func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(s.basePath, key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file: %w", err)
	}

	return true, nil
}

// GetBasePath returns the base path for the storage
func (s *LocalStorage) GetBasePath() string {
	return s.basePath
}
