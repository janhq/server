package responses

import (
	"jan-server/services/media-api/internal/domain/media"
)

// IngestResponse represents successful media ingestion
type IngestResponse struct {
	ID           string `json:"id"`
	Mime         string `json:"mime"`
	Bytes        int64  `json:"bytes"`
	Deduped      bool   `json:"deduped"`
	PresignedURL string `json:"presigned_url,omitempty"`
}

// BuildIngestResponse creates response from domain object
func BuildIngestResponse(obj *media.MediaObject, deduped bool, presignedURL string) *IngestResponse {
	return &IngestResponse{
		ID:           obj.ID,
		Mime:         obj.MimeType,
		Bytes:        obj.Bytes,
		Deduped:      deduped,
		PresignedURL: presignedURL,
	}
}

// ResolveResponse represents media resolution result
type ResolveResponse struct {
	Payload interface{} `json:"payload"`
}

// BuildResolveResponse creates response from resolved payload
func BuildResolveResponse(payload interface{}) *ResolveResponse {
	return &ResolveResponse{
		Payload: payload,
	}
}

// PrepareUploadResponse contains presigned upload information
type PrepareUploadResponse struct {
	UploadURL  string            `json:"upload_url"`
	UploadID   string            `json:"upload_id"`
	FormFields map[string]string `json:"form_fields,omitempty"`
}

// BuildPrepareUploadResponse creates upload preparation response
func BuildPrepareUploadResponse(uploadURL, uploadID string, formFields map[string]string) *PrepareUploadResponse {
	return &PrepareUploadResponse{
		UploadURL:  uploadURL,
		UploadID:   uploadID,
		FormFields: formFields,
	}
}

// DirectUploadResponse contains upload result
type DirectUploadResponse struct {
	ID           string `json:"id"`
	PresignedURL string `json:"presigned_url,omitempty"`
}

// BuildDirectUploadResponse creates direct upload response
func BuildDirectUploadResponse(id, presignedURL string) *DirectUploadResponse {
	return &DirectUploadResponse{
		ID:           id,
		PresignedURL: presignedURL,
	}
}

// PresignedURLResponse contains presigned URL
type PresignedURLResponse struct {
	URL string `json:"url"`
}

// BuildPresignedURLResponse creates presigned URL response
func BuildPresignedURLResponse(url string) *PresignedURLResponse {
	return &PresignedURLResponse{
		URL: url,
	}
}
