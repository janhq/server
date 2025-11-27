package requests

import (
	"jan-server/services/media-api/internal/domain/media"
)

// IngestRequest represents media ingestion request
type IngestRequest struct {
	Source   Source `json:"source" binding:"required"`
	Filename string `json:"filename"`
	UserID   string `json:"user_id"`
}

// Source describes the media source
type Source struct {
	Type    string `json:"type" binding:"required"`
	DataURL string `json:"data_url"`
	URL     string `json:"url"`
}

// ToDomain converts request to domain model
func (r *IngestRequest) ToDomain() *media.IngestRequest {
	return &media.IngestRequest{
		Source: media.Source{
			Type:    r.Source.Type,
			DataURL: r.Source.DataURL,
			URL:     r.Source.URL,
		},
		Filename: r.Filename,
		UserID:   r.UserID,
	}
}

// ResolveRequest represents media resolution request
type ResolveRequest struct {
	Payload interface{} `json:"payload" binding:"required"`
}

// PrepareUploadRequest represents presigned upload preparation
type PrepareUploadRequest struct {
	Filename string `json:"filename" binding:"required"`
	MimeType string `json:"mime_type" binding:"required"`
	SizeKB   int64  `json:"size_kb"`
}

// DirectUploadRequest represents direct file upload metadata
type DirectUploadRequest struct {
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
}
