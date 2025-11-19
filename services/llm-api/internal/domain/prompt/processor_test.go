package prompt

import (
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	openai "github.com/sashabaranov/go-openai"
)

func TestMemoryModule(t *testing.T) {
	module := NewMemoryModule(true)

	ctx := context.Background()
	promptCtx := &Context{
		UserID:         1,
		ConversationID: "test-conv",
		Memory: []string{
			"User prefers concise answers",
			"User is a software engineer",
		},
	}

	messages := []openai.ChatCompletionMessage{
		{Role: "user", Content: "What is Go?"},
	}

	if !module.ShouldApply(ctx, promptCtx, messages) {
		t.Error("Memory module should apply when memory exists")
	}

	result, err := module.Apply(ctx, promptCtx, messages)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Should have prepended a system message
	if len(result) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(result))
	}

	if result[0].Role != "system" {
		t.Error("First message should be system role")
	}

	content := result[0].Content
	if !strings.Contains(content, "User prefers concise answers") {
		t.Error("Memory not included in system prompt")
	}
}

func TestCodeAssistantModule(t *testing.T) {
	module := NewCodeAssistantModule()

	ctx := context.Background()
	promptCtx := &Context{
		UserID:         1,
		ConversationID: "test-conv",
	}

	tests := []struct {
		name        string
		messages    []openai.ChatCompletionMessage
		shouldApply bool
	}{
		{
			name: "code-related question",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "How do I implement a binary search function in Go?"},
			},
			shouldApply: true,
		},
		{
			name: "non-code question",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "What's the weather like today?"},
			},
			shouldApply: false,
		},
		{
			name: "debug question",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "I'm getting a syntax error in my code"},
			},
			shouldApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldApply := module.ShouldApply(ctx, promptCtx, tt.messages)
			if shouldApply != tt.shouldApply {
				t.Errorf("Expected shouldApply=%v, got %v", tt.shouldApply, shouldApply)
			}

			if shouldApply {
				result, err := module.Apply(ctx, promptCtx, tt.messages)
				if err != nil {
					t.Fatalf("Apply failed: %v", err)
				}

				if len(result) == 0 {
					t.Error("Result should not be empty")
				}

				// Check if code instructions were added
				hasCodeInstructions := false
				for _, msg := range result {
					if msg.Role == "system" && strings.Contains(msg.Content, "code assistance") {
						hasCodeInstructions = true
						break
					}
				}

				if !hasCodeInstructions {
					t.Error("Code instructions not found in system prompt")
				}
			}
		})
	}
}

func TestChainOfThoughtModule(t *testing.T) {
	module := NewChainOfThoughtModule()

	ctx := context.Background()
	promptCtx := &Context{
		UserID:         1,
		ConversationID: "test-conv",
	}

	tests := []struct {
		name        string
		messages    []openai.ChatCompletionMessage
		shouldApply bool
	}{
		{
			name: "complex question",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "How does machine learning differ from traditional programming, and why would you choose one approach over the other?"},
			},
			shouldApply: true,
		},
		{
			name: "simple question",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Hello"},
			},
			shouldApply: false,
		},
		{
			name: "analytical question",
			messages: []openai.ChatCompletionMessage{
				{Role: "user", Content: "Analyze the pros and cons of microservices architecture"},
			},
			shouldApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldApply := module.ShouldApply(ctx, promptCtx, tt.messages)
			if shouldApply != tt.shouldApply {
				t.Errorf("Expected shouldApply=%v, got %v", tt.shouldApply, shouldApply)
			}
		})
	}
}

func TestProcessor(t *testing.T) {
	log := zerolog.Nop()
	config := ProcessorConfig{
		EnableMemory:    true,
		EnableTemplates: true,
		EnableTools:     false,
		DefaultPersona:  "helpful assistant",
	}

	processor := NewProcessor(config, log)

	ctx := context.Background()
	promptCtx := &Context{
		UserID:         1,
		ConversationID: "test-conv",
		Memory: []string{
			"User likes detailed explanations",
		},
		Preferences: map[string]interface{}{
			"language": "en",
		},
	}

	messages := []openai.ChatCompletionMessage{
		{Role: "user", Content: "How do I implement error handling in Go?"},
	}

	result, err := processor.Process(ctx, promptCtx, messages)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Result should not be empty")
	}

	// Should have applied memory and code assistant modules
	hasSystemMessage := false
	for _, msg := range result {
		if msg.Role == "system" {
			hasSystemMessage = true
			// Check for memory
			if !strings.Contains(msg.Content, "User likes detailed explanations") {
				t.Error("Memory not found in system message")
			}
			// Check for code instructions
			if !strings.Contains(msg.Content, "code assistance") {
				t.Error("Code assistance instructions not found")
			}
		}
	}

	if !hasSystemMessage {
		t.Error("System message should have been added")
	}
}

func TestProcessorWithoutMemory(t *testing.T) {
	log := zerolog.Nop()
	config := ProcessorConfig{
		EnableMemory:    false,
		EnableTemplates: true,
		EnableTools:     false,
		DefaultPersona:  "helpful assistant",
	}

	processor := NewProcessor(config, log)

	ctx := context.Background()
	promptCtx := &Context{
		UserID:         1,
		ConversationID: "test-conv",
		Memory: []string{
			"This should be ignored",
		},
	}

	messages := []openai.ChatCompletionMessage{
		{Role: "user", Content: "Hello"},
	}

	result, err := processor.Process(ctx, promptCtx, messages)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}

	// Memory module should not be registered
	for _, msg := range result {
		if msg.Role == "system" && strings.Contains(msg.Content, "This should be ignored") {
			t.Error("Memory should not be applied when disabled")
		}
	}
}
