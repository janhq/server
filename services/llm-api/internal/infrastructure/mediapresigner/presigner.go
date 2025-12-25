package mediapresigner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"jan-server/services/llm-api/internal/config"
)

// Presigner resolves jan_* file IDs to presigned URLs.
type Presigner interface {
	// PresignFileID returns a presigned URL for the given file ID.
	// Returns empty string and nil error if the file ID is invalid or not found.
	PresignFileID(ctx context.Context, fileID string) (string, error)
}

type httpPresigner struct {
	baseURL string
	client  *http.Client
	log     zerolog.Logger
}

// NewPresigner creates a new HTTP-based presigner that calls media-api.
// Returns nil if MediaInternalURL is not configured.
func NewPresigner(cfg *config.Config, log zerolog.Logger) Presigner {
	if cfg == nil {
		return nil
	}

	// Use internal URL for service-to-service communication (bypasses Kong auth)
	baseURL := strings.TrimSpace(cfg.MediaInternalURL)
	if baseURL == "" {
		// Fall back to resolve URL and extract base
		resolveURL := strings.TrimSpace(cfg.MediaResolveURL)
		if resolveURL != "" {
			// Extract base URL from resolve URL (e.g., "http://media-api:8285/v1/media/resolve" -> "http://media-api:8285/v1/media")
			baseURL = strings.TrimSuffix(resolveURL, "/resolve")
		}
	}

	if baseURL == "" {
		log.Warn().Msg("media presigner disabled: no MediaInternalURL or MediaResolveURL configured")
		return nil
	}

	timeout := 5 * time.Second

	return &httpPresigner{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{Timeout: timeout},
		log:     log.With().Str("component", "media-presigner").Logger(),
	}
}

func (p *httpPresigner) PresignFileID(ctx context.Context, fileID string) (string, error) {
	if fileID == "" {
		return "", nil
	}

	// Call GET /v1/media/{id}/presign
	url := fmt.Sprintf("%s/%s/presign", p.baseURL, fileID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build presign request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		p.log.Warn().Err(err).Str("file_id", fileID).Msg("presign request failed")
		return "", fmt.Errorf("presign request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		p.log.Debug().Str("file_id", fileID).Msg("file not found for presigning")
		return "", nil
	}

	if resp.StatusCode != http.StatusOK {
		p.log.Warn().Str("file_id", fileID).Int("status", resp.StatusCode).Msg("presign returned non-200")
		return "", fmt.Errorf("presign returned status %d", resp.StatusCode)
	}

	var result struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode presign response: %w", err)
	}

	p.log.Debug().Str("file_id", fileID).Msg("successfully presigned file")
	return result.URL, nil
}

// NoOpPresigner is a presigner that does nothing (used when presigning is disabled).
type NoOpPresigner struct{}

func (n *NoOpPresigner) PresignFileID(ctx context.Context, fileID string) (string, error) {
	return "", nil
}
