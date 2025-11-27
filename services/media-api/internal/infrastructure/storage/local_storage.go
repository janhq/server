package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"jan-server/services/media-api/internal/config"
)

var errLocalStorageDisabled = errors.New("local storage is not configured; set MEDIA_LOCAL_STORAGE_PATH to enable")

// LocalStorage handles uploads and downloads to local filesystem.
type LocalStorage struct {
	basePath string
	baseURL  string
	log      zerolog.Logger
	disabled bool
}

// NewLocalStorage creates a new local filesystem storage backend.
func NewLocalStorage(cfg *config.Config, log zerolog.Logger) (*LocalStorage, error) {
	logger := log.With().Str("component", "local-storage").Logger()

	basePath := strings.TrimSpace(cfg.LocalStoragePath)
	if basePath == "" {
		logger.Warn().Msg("MEDIA_LOCAL_STORAGE_PATH is not set; local storage will be disabled")
		return &LocalStorage{
			log:      logger,
			disabled: true,
		}, nil
	}

	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create local storage directory: %w", err)
	}

	storage := &LocalStorage{
		basePath: basePath,
		baseURL:  strings.TrimSpace(cfg.LocalStorageBaseURL),
		log:      logger,
		disabled: false,
	}

	logger.Info().
		Str("path", basePath).
		Str("base_url", storage.baseURL).
		Msg("local storage initialized")

	return storage, nil
}

func (l *LocalStorage) ensureEnabled() error {
	if l.disabled {
		return errLocalStorageDisabled
	}
	return nil
}

// Upload stores a file to the local filesystem.
func (l *LocalStorage) Upload(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if err := l.ensureEnabled(); err != nil {
		return err
	}

	fullPath := filepath.Join(l.basePath, filepath.FromSlash(key))
	dir := filepath.Dir(fullPath)

	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data to file
	written, err := io.Copy(file, body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	l.log.Debug().
		Str("key", key).
		Int64("bytes", written).
		Msg("file uploaded to local storage")

	return nil
}

// PresignGet returns a direct URL to the file (no presigning needed for local storage).
// If LocalStorageBaseURL is set, it returns a URL, otherwise returns the file path.
func (l *LocalStorage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := l.ensureEnabled(); err != nil {
		return "", err
	}

	// Check if file exists
	fullPath := filepath.Join(l.basePath, filepath.FromSlash(key))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", key)
	}

	// If base URL is configured, return a URL
	if l.baseURL != "" {
		// Normalize the key to use forward slashes for URLs
		urlKey := filepath.ToSlash(key)
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(l.baseURL, "/"), urlKey), nil
	}

	// Otherwise return a file:// URL
	return fmt.Sprintf("file://%s", fullPath), nil
}

// PresignPut is not supported for local storage (direct upload only).
// Returns an error indicating presigned uploads are not available.
func (l *LocalStorage) PresignPut(ctx context.Context, key string, contentType string, ttl time.Duration) (string, error) {
	if err := l.ensureEnabled(); err != nil {
		return "", err
	}
	return "", errors.New("presigned PUT not supported for local storage; use direct upload endpoint")
}

// Download reads a file from the local filesystem.
func (l *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, string, error) {
	if err := l.ensureEnabled(); err != nil {
		return nil, "", err
	}

	fullPath := filepath.Join(l.basePath, filepath.FromSlash(key))

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", fmt.Errorf("file not found: %s", key)
		}
		return nil, "", fmt.Errorf("failed to open file: %w", err)
	}

	// Try to detect content type from extension
	contentType := detectContentTypeFromPath(fullPath)

	l.log.Debug().
		Str("key", key).
		Str("content_type", contentType).
		Msg("file downloaded from local storage")

	return file, contentType, nil
}

// Health checks if the storage directory is accessible.
func (l *LocalStorage) Health(ctx context.Context) error {
	if l.disabled {
		return nil
	}

	// Check if we can write to the storage directory
	testFile := filepath.Join(l.basePath, ".health_check")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err != nil {
		return fmt.Errorf("storage directory not writable: %w", err)
	}

	// Clean up test file
	_ = os.Remove(testFile)

	return nil
}

// SupportsPresignedUploads returns false for local storage.
func (l *LocalStorage) SupportsPresignedUploads() bool {
	return false
}

// detectContentTypeFromPath attempts to determine content type from file extension.
func detectContentTypeFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".tiff", ".tif":
		return "image/tiff"
	default:
		return "application/octet-stream"
	}
}
