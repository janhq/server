package chatrequests

import (
	"encoding/json"

	"jan-server/services/llm-api/internal/domain/conversation"

	openai "github.com/sashabaranov/go-openai"
)

// ChatCompletionRequest extends OpenAI's ChatCompletionRequest with conversation support
type ChatCompletionRequest struct {
	openai.ChatCompletionRequest

	TopK              *int     `json:"top_k,omitempty"`
	RepetitionPenalty *float32 `json:"repetition_penalty,omitempty"`

	// Conversation can be either a string (conversation ID) or a conversation object
	// Items from this conversation are prepended to Messages for this response request.
	// Input items and output items from this response are automatically added to this conversation after completion.
	Conversation *ConversationReference `json:"conversation,omitempty"`
	// Store controls whether the latest input and generated response should be persisted
	Store *bool `json:"store,omitempty"`
	// StoreReasoning controls whether reasoning content (if present) should also be persisted
	StoreReasoning *bool `json:"store_reasoning,omitempty"`
	// DeepResearch enables the Deep Research mode which uses a specialized prompt
	// for conducting in-depth investigations with tool usage.
	// Requires a model with supports_reasoning: true capability.
	DeepResearch *bool `json:"deep_research,omitempty"`
}

// ConversationReference can unmarshal from either a string (ID) or an object
type ConversationReference struct {
	ID     *string                    `json:"-"` // Conversation ID when provided as string
	Object *conversation.Conversation `json:"-"` // Conversation object when provided as object
}

// UnmarshalJSON implements custom unmarshaling to support both string and object types
// This is required because OpenAI's API allows conversation to be either:
//   - A string: "conversation": "conv_abc123"
//   - An object: "conversation": {"id": "conv_abc123", ...}
func (c *ConversationReference) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		c.ID = &str
		return nil
	}

	// If not a string, try to unmarshal as conversation object
	var obj conversation.Conversation
	if err := json.Unmarshal(data, &obj); err != nil {
		return err
	}
	c.Object = &obj
	return nil
}

// MarshalJSON implements custom marshaling
func (c *ConversationReference) MarshalJSON() ([]byte, error) {
	if c.ID != nil {
		return json.Marshal(*c.ID)
	}
	if c.Object != nil {
		return json.Marshal(*c.Object)
	}
	return json.Marshal(nil)
}

// IsEmpty returns true if the conversation reference is empty
// Note: Includes nil check for defensive programming. Callers should still check for nil
// before calling this method to avoid potential panics.
func (c *ConversationReference) IsEmpty() bool {
	return c == nil || (c.ID == nil && c.Object == nil)
}

// GetID returns the conversation ID, whether it was provided directly or from an object
// Returns empty string if the reference is nil or has no ID.
func (c *ConversationReference) GetID() string {
	if c == nil {
		return ""
	}
	if c.ID != nil {
		return *c.ID
	}
	if c.Object != nil {
		return c.Object.PublicID
	}
	return ""
}

// GetConversation returns the conversation object if provided
// Returns nil if the reference is nil or contains only an ID string.
func (c *ConversationReference) GetConversation() *conversation.Conversation {
	if c == nil {
		return nil
	}
	return c.Object
}
