package prompt

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	openai "github.com/sashabaranov/go-openai"
)

func testLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

func TestNewProcessorDisabledReturnsNil(t *testing.T) {
	cfg := ProcessorConfig{
		Enabled: false,
	}

	proc := NewProcessor(cfg, testLogger())
	if proc == nil {
		t.Fatalf("expected noop processor when disabled, got nil")
	}
	msgs := []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: "hi"}}
	out, err := proc.Process(context.Background(), &Context{}, msgs)
	if err != nil {
		t.Fatalf("noop processor returned error: %v", err)
	}
	if len(out) != len(msgs) || out[0].Content != "hi" {
		t.Fatalf("noop processor should passthrough messages")
	}
}

func TestPersonaModuleAddsSystemMessage(t *testing.T) {
	cfg := ProcessorConfig{
		Enabled:        true,
		DefaultPersona: "friendly guide",
	}

	proc := NewProcessor(cfg, testLogger())
	ctx := &Context{
		Preferences: map[string]interface{}{
			"persona": "strict analyst",
		},
	}

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "hello"},
	}

	processed, err := proc.Process(context.Background(), ctx, msgs)
	if err != nil {
		t.Fatalf("process returned error: %v", err)
	}

	system := firstSystemMessage(processed)
	if system == nil {
		t.Fatalf("expected system message with persona applied")
	}

	if !strings.Contains(strings.ToLower(system.Content), "strict analyst") {
		t.Fatalf("expected persona text in system message, got: %s", system.Content)
	}
}

func TestMemoryModuleAppendsMemory(t *testing.T) {
	cfg := ProcessorConfig{
		Enabled:        true,
		EnableMemory:   true,
		DefaultPersona: "helpful assistant",
	}

	proc := NewProcessor(cfg, testLogger())
	ctx := &Context{
		Memory: []string{"prefers concise answers", "call them Sam"},
	}

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "How are you?"},
	}

	processed, err := proc.Process(context.Background(), ctx, msgs)
	if err != nil {
		t.Fatalf("process returned error: %v", err)
	}

	system := firstSystemMessage(processed)
	if system == nil {
		t.Fatalf("expected system message with memory applied")
	}

	for _, expected := range ctx.Memory {
		if !strings.Contains(system.Content, expected) {
			t.Fatalf("expected memory %q in system content: %s", expected, system.Content)
		}
	}
}

func TestToolInstructionsModuleApplies(t *testing.T) {
	cfg := ProcessorConfig{
		Enabled:     true,
		EnableTools: true,
	}

	proc := NewProcessor(cfg, testLogger())
	ctx := &Context{
		Preferences: map[string]interface{}{
			"use_tools": true,
		},
	}

	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "Search the docs"},
	}

	processed, err := proc.Process(context.Background(), ctx, msgs)
	if err != nil {
		t.Fatalf("process returned error: %v", err)
	}

	system := firstSystemMessage(processed)
	if system == nil || !strings.Contains(system.Content, "You have access to various tools") {
		t.Fatalf("expected tool instructions in system message, got: %v", system)
	}
}

func TestToolInstructionsDetectsToolCallsWithoutPreference(t *testing.T) {
	cfg := ProcessorConfig{
		Enabled:     true,
		EnableTools: true,
	}
	proc := NewProcessor(cfg, testLogger())
	msgs := []openai.ChatCompletionMessage{
		{
			Role: openai.ChatMessageRoleAssistant,
			ToolCalls: []openai.ToolCall{
				{ID: "tool1", Type: openai.ToolTypeFunction},
			},
		},
	}
	processed, err := proc.Process(context.Background(), &Context{}, msgs)
	if err != nil {
		t.Fatalf("process returned error: %v", err)
	}
	system := firstSystemMessage(processed)
	if system == nil || !strings.Contains(system.Content, "module:tool_instructions") {
		t.Fatalf("expected tool instructions marker in system content")
	}
}

func TestTemplateModulesRespectFlag(t *testing.T) {
	templatePrompt := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "How would you debug a function that fails intermittently?"},
	}

	disabledCfg := ProcessorConfig{
		Enabled:         true,
		EnableTemplates: false,
		DefaultPersona:  "helper",
	}
	disabledProc := NewProcessor(disabledCfg, testLogger())
	disabledOut, err := disabledProc.Process(context.Background(), &Context{}, templatePrompt)
	if err != nil {
		t.Fatalf("process returned error (templates disabled): %v", err)
	}
	disabledSystem := firstSystemMessage(disabledOut)
	if disabledSystem != nil {
		if strings.Contains(disabledSystem.Content, "think step-by-step") || strings.Contains(disabledSystem.Content, "When providing code assistance") {
			t.Fatalf("template instructions should not be applied when templates are disabled")
		}
	}

	enabledCfg := ProcessorConfig{
		Enabled:         true,
		EnableTemplates: true,
		DefaultPersona:  "helper",
	}
	enabledProc := NewProcessor(enabledCfg, testLogger())
	enabledOut, err := enabledProc.Process(context.Background(), &Context{}, templatePrompt)
	if err != nil {
		t.Fatalf("process returned error (templates enabled): %v", err)
	}
	enabledSystem := firstSystemMessage(enabledOut)
	if enabledSystem == nil {
		t.Fatalf("expected system message when templates are enabled")
	}
	if !strings.Contains(enabledSystem.Content, "think step-by-step") {
		t.Fatalf("expected chain-of-thought instructions to be applied")
	}
	if !strings.Contains(enabledSystem.Content, "When providing code assistance") {
		t.Fatalf("expected code assistant instructions to be applied")
	}
}

func TestProcessorDoesNotDuplicateMarkersOnReentry(t *testing.T) {
	cfg := ProcessorConfig{
		Enabled:         true,
		EnableTemplates: true,
		DefaultPersona:  "helper",
	}
	proc := NewProcessor(cfg, testLogger())
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "Explain how to sort numbers?"},
	}
	promptCtx := &Context{}
	processedOnce, err := proc.Process(context.Background(), promptCtx, msgs)
	if err != nil {
		t.Fatalf("first process returned error: %v", err)
	}
	processedTwice, err := proc.Process(context.Background(), promptCtx, processedOnce)
	if err != nil {
		t.Fatalf("second process returned error: %v", err)
	}
	system := firstSystemMessage(processedTwice)
	if system == nil {
		t.Fatalf("expected system message")
	}
	if count := strings.Count(system.Content, "[module:persona]"); count > 1 {
		t.Fatalf("persona marker duplicated: %d", count)
	}
	if count := strings.Count(system.Content, "[module:chain_of_thought]"); count > 1 {
		t.Fatalf("chain_of_thought marker duplicated: %d", count)
	}
}

func TestProcessorRespectsPriorityPersonaBeforeMemory(t *testing.T) {
	cfg := ProcessorConfig{
		Enabled:        true,
		EnableMemory:   true,
		DefaultPersona: "helper",
	}
	proc := NewProcessor(cfg, testLogger())
	msgs := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: "Hello"},
	}
	ctx := &Context{
		Memory: []string{"prefers concise replies"},
	}
	processed, err := proc.Process(context.Background(), ctx, msgs)
	if err != nil {
		t.Fatalf("process returned error: %v", err)
	}
	system := firstSystemMessage(processed)
	if system == nil {
		t.Fatalf("expected system message")
	}
	idxPersona := strings.Index(strings.ToLower(system.Content), "you are a")
	idxMemory := strings.Index(strings.ToLower(system.Content), "use the following personal memory")
	if idxPersona == -1 || idxMemory == -1 || idxPersona > idxMemory {
		t.Fatalf("expected persona instructions to appear before memory instructions")
	}
}

func firstSystemMessage(messages []openai.ChatCompletionMessage) *openai.ChatCompletionMessage {
	for i := range messages {
		if messages[i].Role == openai.ChatMessageRoleSystem {
			return &messages[i]
		}
	}
	return nil
}
