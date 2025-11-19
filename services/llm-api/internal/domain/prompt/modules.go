package prompt

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// MemoryModule adds user memory to system prompts
type MemoryModule struct {
	enabled bool
}

// NewMemoryModule creates a new memory module
func NewMemoryModule(enabled bool) *MemoryModule {
	return &MemoryModule{enabled: enabled}
}

// Name returns the module identifier
func (m *MemoryModule) Name() string {
	return "memory"
}

// ShouldApply checks if memory should be included
func (m *MemoryModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	return m.enabled && len(promptCtx.Memory) > 0
}

// Apply adds memory to the system prompt
func (m *MemoryModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if len(promptCtx.Memory) == 0 {
		return messages, nil
	}

	memoryText := "\n\nUse the following personal memory for this user:\n"
	for _, item := range promptCtx.Memory {
		memoryText += "- " + item + "\n"
	}

	// Find or create system message
	result := make([]openai.ChatCompletionMessage, 0, len(messages))
	systemFound := false

	for _, msg := range messages {
		if msg.Role == "system" {
			// Append memory to existing system message
			msg.Content = msg.Content + memoryText
			systemFound = true
		}
		result = append(result, msg)
	}

	// If no system message exists, prepend one
	if !systemFound {
		result = append([]openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant." + memoryText,
			},
		}, result...)
	}

	return result, nil
}

// ToolInstructionsModule adds tool usage instructions
type ToolInstructionsModule struct {
	enabled bool
}

// NewToolInstructionsModule creates a new tool instructions module
func NewToolInstructionsModule(enabled bool) *ToolInstructionsModule {
	return &ToolInstructionsModule{enabled: enabled}
}

// Name returns the module identifier
func (m *ToolInstructionsModule) Name() string {
	return "tool_instructions"
}

// ShouldApply checks if tool instructions should be added
func (m *ToolInstructionsModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	// Check if any preferences indicate tool usage
	if !m.enabled {
		return false
	}

	if prefs := promptCtx.Preferences; prefs != nil {
		if useTools, ok := prefs["use_tools"].(bool); ok && useTools {
			return true
		}
	}

	return false
}

// Apply adds tool instructions to the system prompt
func (m *ToolInstructionsModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	toolInstructions := "\n\nYou have access to various tools. Always choose the best tool for the task. When you need to search for information, use web search. When you need to execute code, use the code execution tool."

	result := make([]openai.ChatCompletionMessage, 0, len(messages))
	systemFound := false

	for _, msg := range messages {
		if msg.Role == "system" {
			msg.Content = msg.Content + toolInstructions
			systemFound = true
		}
		result = append(result, msg)
	}

	if !systemFound {
		result = append([]openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant." + toolInstructions,
			},
		}, result...)
	}

	return result, nil
}

// CodeAssistantModule adds code-specific instructions
type CodeAssistantModule struct{}

// NewCodeAssistantModule creates a new code assistant module
func NewCodeAssistantModule() *CodeAssistantModule {
	return &CodeAssistantModule{}
}

// Name returns the module identifier
func (m *CodeAssistantModule) Name() string {
	return "code_assistant"
}

// ShouldApply checks if the question is code-related
func (m *CodeAssistantModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	// Check last user message for code-related keywords
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			content := strings.ToLower(messages[i].Content)
			codeKeywords := []string{"code", "function", "implement", "debug", "error", "program", "script", "api", "bug", "syntax"}
			for _, keyword := range codeKeywords {
				if strings.Contains(content, keyword) {
					return true
				}
			}
			break
		}
	}
	return false
}

// Apply adds code assistant instructions
func (m *CodeAssistantModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	codeInstructions := "\n\nWhen providing code assistance:\n" +
		"1. Provide clear, well-commented code\n" +
		"2. Explain your approach and reasoning\n" +
		"3. Include error handling where appropriate\n" +
		"4. Follow best practices and conventions\n" +
		"5. Suggest testing approaches when relevant"

	result := make([]openai.ChatCompletionMessage, 0, len(messages))
	systemFound := false

	for _, msg := range messages {
		if msg.Role == "system" {
			msg.Content = msg.Content + codeInstructions
			systemFound = true
		}
		result = append(result, msg)
	}

	if !systemFound {
		result = append([]openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant." + codeInstructions,
			},
		}, result...)
	}

	return result, nil
}

// ChainOfThoughtModule adds chain-of-thought reasoning instructions
type ChainOfThoughtModule struct{}

// NewChainOfThoughtModule creates a new chain-of-thought module
func NewChainOfThoughtModule() *ChainOfThoughtModule {
	return &ChainOfThoughtModule{}
}

// Name returns the module identifier
func (m *ChainOfThoughtModule) Name() string {
	return "chain_of_thought"
}

// ShouldApply checks if the question requires reasoning
func (m *ChainOfThoughtModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	// Apply for complex questions (check for question marks and length)
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			content := messages[i].Content
			// Complex questions are typically longer and may contain multiple sentences
			if len(content) > 100 && strings.Contains(content, "?") {
				return true
			}
			// Look for reasoning keywords
			reasoningKeywords := []string{"why", "how", "explain", "analyze", "compare", "evaluate", "what if"}
			contentLower := strings.ToLower(content)
			for _, keyword := range reasoningKeywords {
				if strings.Contains(contentLower, keyword) {
					return true
				}
			}
			break
		}
	}
	return false
}

// Apply adds chain-of-thought instructions
func (m *ChainOfThoughtModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	cotInstructions := "\n\nFor complex questions, think step-by-step:\n" +
		"1. Break down the problem\n" +
		"2. Analyze each component\n" +
		"3. Consider different perspectives\n" +
		"4. Synthesize your conclusion\n" +
		"5. Provide a clear, structured answer"

	result := make([]openai.ChatCompletionMessage, 0, len(messages))
	systemFound := false

	for _, msg := range messages {
		if msg.Role == "system" {
			msg.Content = msg.Content + cotInstructions
			systemFound = true
		}
		result = append(result, msg)
	}

	if !systemFound {
		result = append([]openai.ChatCompletionMessage{
			{
				Role:    "system",
				Content: "You are a helpful assistant." + cotInstructions,
			},
		}, result...)
	}

	return result, nil
}
