package prompt

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"jan-server/services/llm-api/internal/domain/prompttemplate"
)

const (
	deepResearchModuleName = "deep_research"
)

// DeepResearchModule injects the Deep Research system prompt when enabled
type DeepResearchModule struct {
	templateService *prompttemplate.Service
}

// NewDeepResearchModule creates a new deep research module
func NewDeepResearchModule(templateService *prompttemplate.Service) *DeepResearchModule {
	return &DeepResearchModule{
		templateService: templateService,
	}
}

// Name returns the module identifier
func (m *DeepResearchModule) Name() string {
	return deepResearchModuleName
}

// ShouldApply determines if the Deep Research prompt should be injected
// This module applies when:
// 1. Deep research is explicitly enabled in preferences
// 2. Module is not disabled
func (m *DeepResearchModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if promptCtx == nil || promptCtx.Preferences == nil {
		return false
	}

	// Check if module is disabled
	if isModuleDisabled(promptCtx.Preferences, m.Name()) {
		return false
	}

	// Check if deep_research is enabled in preferences
	deepResearch, ok := promptCtx.Preferences["deep_research"]
	if !ok {
		return false
	}

	// Handle different types
	switch v := deepResearch.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true"
	default:
		return false
	}
}

// Apply injects the Deep Research system prompt at the beginning of the messages
func (m *DeepResearchModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}

	var promptContent string

	// Try to fetch the Deep Research prompt template from the database/service
	if m.templateService != nil {
		template, err := m.templateService.GetDeepResearchPrompt(ctx)
		if err == nil && template != nil && template.IsActive {
			promptContent = template.Content
		}
	}

	// Fallback to the default constant if no template was found
	if promptContent == "" {
		promptContent = prompttemplate.DefaultDeepResearchPrompt
	}

	// Prepend the Deep Research system prompt
	return prependDeepResearchPrompt(messages, promptContent), nil
}

// prependDeepResearchPrompt prepends the deep research system prompt to messages
func prependDeepResearchPrompt(messages []openai.ChatCompletionMessage, content string) []openai.ChatCompletionMessage {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return messages
	}

	result := make([]openai.ChatCompletionMessage, 0, len(messages)+1)

	// Add Deep Research system message first
	result = append(result, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: trimmed,
	})

	// Append all existing messages
	for _, msg := range messages {
		result = append(result, cloneMessage(msg))
	}

	return result
}

// DeepResearchConfig holds configuration extracted from template metadata
type DeepResearchConfig struct {
	MaxTokens            int      `json:"max_tokens,omitempty"`
	Temperature          float32  `json:"temperature,omitempty"`
	RequiresTools        bool     `json:"requires_tools,omitempty"`
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`
}

// GetDeepResearchConfig extracts configuration from template metadata
func GetDeepResearchConfig(metadata map[string]any) *DeepResearchConfig {
	if metadata == nil {
		return &DeepResearchConfig{
			MaxTokens:            16384,
			Temperature:          0.7,
			RequiresTools:        true,
			RequiredCapabilities: []string{"reasoning"},
		}
	}

	config := &DeepResearchConfig{
		MaxTokens:            16384,
		Temperature:          0.7,
		RequiresTools:        true,
		RequiredCapabilities: []string{"reasoning"},
	}

	if maxTokens, ok := metadata["max_tokens"].(float64); ok {
		config.MaxTokens = int(maxTokens)
	}
	if temperature, ok := metadata["temperature"].(float64); ok {
		config.Temperature = float32(temperature)
	}
	if requiresTools, ok := metadata["requires_tools"].(bool); ok {
		config.RequiresTools = requiresTools
	}
	if capabilities, ok := metadata["required_capabilities"].([]interface{}); ok {
		config.RequiredCapabilities = make([]string, 0, len(capabilities))
		for _, cap := range capabilities {
			if s, ok := cap.(string); ok {
				config.RequiredCapabilities = append(config.RequiredCapabilities, s)
			}
		}
	}

	return config
}
