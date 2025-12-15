package llm

import (
	"encoding/json"
	"unicode/utf8"
)

const (
	// DefaultContextLength is used when model context length is unknown.
	DefaultContextLength = 128000 // 128k tokens as fallback

	// TokenEstimateRatio estimates ~4 characters per token (conservative estimate).
	TokenEstimateRatio = 4

	// MinMessagesToKeep ensures we always keep system prompt + at least one user message.
	MinMessagesToKeep = 2

	// SafetyMarginRatio reserves space for response and overhead (20% margin).
	SafetyMarginRatio = 0.80
)

// EstimateTokenCount provides a rough estimate of token count for a message.
// Uses character count / 4 as a conservative approximation.
func EstimateTokenCount(content interface{}) int {
	var text string
	switch v := content.(type) {
	case string:
		text = v
	case nil:
		return 0
	default:
		bytes, _ := json.Marshal(v)
		text = string(bytes)
	}
	return utf8.RuneCountInString(text) / TokenEstimateRatio
}

// EstimateMessagesTokenCount estimates total tokens across all messages.
func EstimateMessagesTokenCount(messages []ChatMessage) int {
	total := 0
	for _, msg := range messages {
		// Add overhead for role and structure (~10 tokens per message)
		total += 10
		total += EstimateTokenCount(msg.Content)

		// Add tokens for tool calls
		for _, tc := range msg.ToolCalls {
			total += 20 // Overhead for tool call structure
			total += EstimateTokenCount(tc.Function.Name)
			total += EstimateTokenCount(string(tc.Function.Arguments))
		}
	}
	return total
}

// TrimMessagesResult contains the result of trimming messages.
type TrimMessagesResult struct {
	Messages       []ChatMessage
	TrimmedCount   int
	EstimatedTokens int
}

// TrimMessagesToFitContext removes oldest tool results and assistant messages
// to fit within the context length limit.
// Priority order for removal (oldest first):
// 1. Tool result messages (role="tool")
// 2. Assistant messages with tool calls
// 3. Regular assistant messages
// Never removes: system prompts, user messages
func TrimMessagesToFitContext(messages []ChatMessage, contextLength int) TrimMessagesResult {
	if contextLength <= 0 {
		contextLength = DefaultContextLength
	}

	// Apply safety margin
	maxTokens := int(float64(contextLength) * SafetyMarginRatio)

	currentTokens := EstimateMessagesTokenCount(messages)
	if currentTokens <= maxTokens {
		return TrimMessagesResult{
			Messages:        messages,
			TrimmedCount:    0,
			EstimatedTokens: currentTokens,
		}
	}

	// Create a working copy
	result := make([]ChatMessage, len(messages))
	copy(result, messages)
	trimmedCount := 0

	// Find indices of messages that can be removed (in order of priority)
	// We iterate from oldest to newest (excluding system prompt at index 0)
	for currentTokens > maxTokens && len(result) > MinMessagesToKeep {
		removedIdx := -1

		// Phase 1: Remove oldest tool result message
		for i := 1; i < len(result); i++ {
			if result[i].Role == "tool" {
				removedIdx = i
				break
			}
		}

		// Phase 2: Remove oldest assistant message with tool calls (and its following tool results)
		if removedIdx == -1 {
			for i := 1; i < len(result); i++ {
				if result[i].Role == "assistant" && len(result[i].ToolCalls) > 0 {
					removedIdx = i
					break
				}
			}
		}

		// Phase 3: Remove oldest regular assistant message
		if removedIdx == -1 {
			for i := 1; i < len(result); i++ {
				if result[i].Role == "assistant" {
					removedIdx = i
					break
				}
			}
		}

		// If no removable message found, stop
		if removedIdx == -1 {
			break
		}

		// Remove the message
		result = append(result[:removedIdx], result[removedIdx+1:]...)
		trimmedCount++
		currentTokens = EstimateMessagesTokenCount(result)
	}

	return TrimMessagesResult{
		Messages:        result,
		TrimmedCount:    trimmedCount,
		EstimatedTokens: currentTokens,
	}
}

// TrimToolResultContent truncates tool result content if it exceeds maxChars.
// Returns the original content if within limits.
func TrimToolResultContent(content interface{}, maxChars int) interface{} {
	if maxChars <= 0 {
		return content
	}

	var text string
	switch v := content.(type) {
	case string:
		text = v
	case map[string]interface{}:
		if textVal, ok := v["text"].(string); ok {
			text = textVal
		} else {
			return content
		}
	default:
		return content
	}

	runes := []rune(text)
	if len(runes) <= maxChars {
		return content
	}

	truncated := string(runes[:maxChars]) + "... [truncated]"

	// Return in same format as input
	switch content.(type) {
	case string:
		return truncated
	case map[string]interface{}:
		return map[string]interface{}{
			"type": "text",
			"text": truncated,
		}
	}
	return truncated
}
