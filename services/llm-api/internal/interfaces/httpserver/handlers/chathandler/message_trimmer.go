package chathandler

import (
	"encoding/json"
	"unicode/utf8"

	"jan-server/services/llm-api/internal/infrastructure/logger"

	openai "github.com/sashabaranov/go-openai"
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

// estimateTokenCount provides a rough estimate of token count for content.
func estimateTokenCount(content interface{}) int {
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

// estimateMessagesTokenCount estimates total tokens across all messages.
func estimateMessagesTokenCount(messages []openai.ChatCompletionMessage) int {
	total := 0
	for _, msg := range messages {
		// Add overhead for role and structure (~10 tokens per message)
		total += 10
		total += estimateTokenCount(msg.Content)

		// Handle multipart content
		if len(msg.MultiContent) > 0 {
			for _, part := range msg.MultiContent {
				total += estimateTokenCount(part.Text)
			}
		}

		// Add tokens for tool calls
		if msg.ToolCalls != nil {
			for _, tc := range msg.ToolCalls {
				total += 20 // Overhead for tool call structure
				total += estimateTokenCount(tc.Function.Name)
				total += estimateTokenCount(tc.Function.Arguments)
			}
		}
	}
	return total
}

// TrimMessagesResult contains the result of trimming messages.
type TrimMessagesResult struct {
	Messages        []openai.ChatCompletionMessage
	TrimmedCount    int
	EstimatedTokens int
}

// TrimMessagesToFitContext removes oldest tool results and assistant messages
// to fit within the context length limit.
// Priority order for removal (oldest first):
// 1. Tool result messages (role="tool")
// 2. Assistant messages with tool calls
// 3. Regular assistant messages
// Never removes: system prompts, user messages
func TrimMessagesToFitContext(messages []openai.ChatCompletionMessage, contextLength int) TrimMessagesResult {
	if contextLength <= 0 {
		contextLength = DefaultContextLength
	}

	// Apply safety margin
	maxTokens := int(float64(contextLength) * SafetyMarginRatio)

	currentTokens := estimateMessagesTokenCount(messages)
	if currentTokens <= maxTokens {
		return TrimMessagesResult{
			Messages:        messages,
			TrimmedCount:    0,
			EstimatedTokens: currentTokens,
		}
	}

	log := logger.GetLogger()
	log.Info().
		Int("initial_messages", len(messages)).
		Int("initial_tokens", currentTokens).
		Int("max_tokens", maxTokens).
		Int("context_length", contextLength).
		Msg("starting message trimming")

	// Create a working copy
	result := make([]openai.ChatCompletionMessage, len(messages))
	copy(result, messages)
	trimmedCount := 0

	// Build a token count cache for efficient removal
	// This avoids O(n²) complexity from recalculating all tokens on each removal
	messageTokens := make([]int, len(result))
	for i := range result {
		tokens := 10 // Overhead for role and structure
		tokens += estimateTokenCount(result[i].Content)
		
		if len(result[i].MultiContent) > 0 {
			for _, part := range result[i].MultiContent {
				tokens += estimateTokenCount(part.Text)
			}
		}
		
		if result[i].ToolCalls != nil {
			for _, tc := range result[i].ToolCalls {
				tokens += 20
				tokens += estimateTokenCount(tc.Function.Name)
				tokens += estimateTokenCount(tc.Function.Arguments)
			}
		}
		messageTokens[i] = tokens
	}

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

		// Decrement token count by the removed message's tokens
		removedTokens := messageTokens[removedIdx]
		currentTokens -= removedTokens
		
		log.Debug().
			Str("role", result[removedIdx].Role).
			Int("index", removedIdx).
			Int("message_tokens", removedTokens).
			Int("remaining_tokens", currentTokens).
			Int("remaining_messages", len(result)-1).
			Msg("trimmed message")
		
		// Remove the message and its token count from caches
		result = append(result[:removedIdx], result[removedIdx+1:]...)
		messageTokens = append(messageTokens[:removedIdx], messageTokens[removedIdx+1:]...)
		trimmedCount++
	}

	log.Info().
		Int("trimmed_count", trimmedCount).
		Int("final_messages", len(result)).
		Int("final_tokens", currentTokens).
		Int("tokens_freed", estimateMessagesTokenCount(messages)-currentTokens).
		Msg("message trimming completed")

	return TrimMessagesResult{
		Messages:        result,
		TrimmedCount:    trimmedCount,
		EstimatedTokens: currentTokens,
	}
}
