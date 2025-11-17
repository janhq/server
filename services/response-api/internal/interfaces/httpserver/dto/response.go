package dto

import "jan-server/services/response-api/internal/domain/response"

// ResponsePayload is returned to clients.
type ResponsePayload struct {
	ID                 string                 `json:"id"`
	Object             string                 `json:"object"`
	Created            int64                  `json:"created"`
	CreatedAt          int64                  `json:"created_at"` // Same as Created, for compatibility
	Model              string                 `json:"model"`
	Status             string                 `json:"status"`
	Input              interface{}            `json:"input"`
	Output             interface{}            `json:"output,omitempty"`
	Usage              interface{}            `json:"usage,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
	ConversationID     *string                `json:"conversation_id,omitempty"`
	PreviousResponseID *string                `json:"previous_response_id,omitempty"`
	SystemPrompt       *string                `json:"system_prompt,omitempty"`
	Stream             bool                   `json:"stream"`
	Background         bool                   `json:"background"`
	Store              bool                   `json:"store"`
	Error              interface{}            `json:"error,omitempty"`
}

// FromDomain maps the domain response to DTO.
func FromDomain(r *response.Response) ResponsePayload {
	createdUnix := r.CreatedAt.Unix()
	return ResponsePayload{
		ID:                 r.PublicID,
		Object:             r.Object,
		Created:            createdUnix,
		CreatedAt:          createdUnix, // Duplicate for compatibility
		Model:              r.Model,
		Status:             string(r.Status),
		Input:              r.Input,
		Output:             r.Output,
		Usage:              r.Usage,
		Metadata:           r.Metadata,
		ConversationID:     r.ConversationPublicID,
		PreviousResponseID: r.PreviousResponseID,
		SystemPrompt:       r.SystemPrompt,
		Stream:             r.Stream,
		Background:         r.Background,
		Store:              r.Store,
		Error:              r.Error,
	}
}
