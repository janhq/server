package observability

import (
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/janhq/jan-server/packages/go-common/telemetry"
)

// Standard attribute keys
const (
	AttrConversationID   = "conversation_id"
	AttrTenantID         = "tenant_id"
	AttrUserID           = "user_id"
	AttrRequestID        = "request_id"
	AttrModel            = "llm.model"
	AttrTokensPrompt     = "llm.tokens.prompt"
	AttrTokensCompletion = "llm.tokens.completion"
	AttrToolName         = "mcp.tool.name"
	AttrPromptCategory   = "prompt.category"
	AttrPromptPersona    = "prompt.persona"
	AttrPromptLanguage   = "prompt.language"
)

// WithConversationAttrs returns standard attributes for correlation
func WithConversationAttrs(conversationID, tenantID, userID string, sanitizer *telemetry.Sanitizer) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String(AttrConversationID, conversationID),
	}

	if tenantID != "" {
		attrs = append(attrs, attribute.String(AttrTenantID, tenantID))
	}

	if userID != "" && sanitizer != nil {
		attrs = append(attrs, attribute.String(AttrUserID, sanitizer.SanitizeUserID(userID)))
	}

	return attrs
}

// AddConversationAttrsToSpan adds correlation attributes to current span
func AddConversationAttrsToSpan(span trace.Span, conversationID, tenantID, userID string, sanitizer *telemetry.Sanitizer) {
	if span == nil {
		return
	}
	span.SetAttributes(WithConversationAttrs(conversationID, tenantID, userID, sanitizer)...)
}

// WithLLMAttrs returns LLM-specific attributes
func WithLLMAttrs(model string, promptTokens, completionTokens int64) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	if model != "" {
		attrs = append(attrs, attribute.String(AttrModel, model))
	}

	if promptTokens > 0 {
		attrs = append(attrs, attribute.Int64(AttrTokensPrompt, promptTokens))
	}

	if completionTokens > 0 {
		attrs = append(attrs, attribute.Int64(AttrTokensCompletion, completionTokens))
	}

	return attrs
}

// WithPromptMetadata returns prompt classification attributes
func WithPromptMetadata(category, persona, language string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	if category != "" {
		attrs = append(attrs, attribute.String(AttrPromptCategory, category))
	}

	if persona != "" {
		attrs = append(attrs, attribute.String(AttrPromptPersona, persona))
	}

	if language != "" {
		attrs = append(attrs, attribute.String(AttrPromptLanguage, language))
	}

	return attrs
}

// WithRequestID returns a request ID attribute
func WithRequestID(requestID string) attribute.KeyValue {
	return attribute.String(AttrRequestID, requestID)
}

// WithToolName returns a tool name attribute
func WithToolName(toolName string) attribute.KeyValue {
	return attribute.String(AttrToolName, toolName)
}
