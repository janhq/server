package prompt

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"jan-server/services/llm-api/internal/domain/usersettings"
)

const (
	moduleMarkerFormat           = "[[prompt-module:%s]]"
	projectInstructionModuleName = "project_instruction"
	userProfileModuleName        = "user_profile"
)

func moduleMarker(name string) string {
	return fmt.Sprintf(moduleMarkerFormat, strings.ToLower(name))
}

func hasMarker(content, marker string) bool {
	return strings.Contains(strings.ToLower(content), strings.ToLower(marker))
}

func cloneMessage(msg openai.ChatCompletionMessage) openai.ChatCompletionMessage {
	clone := msg

	if len(msg.MultiContent) > 0 {
		clone.MultiContent = make([]openai.ChatMessagePart, len(msg.MultiContent))
		for i, part := range msg.MultiContent {
			clone.MultiContent[i] = part
			if part.ImageURL != nil {
				img := *part.ImageURL
				clone.MultiContent[i].ImageURL = &img
			}
		}
	}

	if len(msg.ToolCalls) > 0 {
		clone.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
		copy(clone.ToolCalls, msg.ToolCalls)
	}

	if msg.FunctionCall != nil {
		fn := *msg.FunctionCall
		clone.FunctionCall = &fn
	}

	return clone
}

// prependInstructionSystemMessage returns a copy of messages with the instruction system message prepended.
// A marker is appended to the content to avoid duplicate injections.
func prependInstructionSystemMessage(messages []openai.ChatCompletionMessage, instruction, moduleName string) []openai.ChatCompletionMessage {
	trimmed := strings.TrimSpace(instruction)
	if trimmed == "" {
		return messages
	}

	marker := moduleMarker(moduleName)
	for _, msg := range messages {
		if msg.Role == openai.ChatMessageRoleSystem && hasMarker(msg.Content, marker) {
			return messages
		}
	}

	result := make([]openai.ChatCompletionMessage, 0, len(messages)+1)
	result = append(result, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: fmt.Sprintf("%s\n%s", trimmed, marker),
	})

	for _, msg := range messages {
		result = append(result, cloneMessage(msg))
	}

	return result
}

// PrependProjectInstruction injects the project instruction as the first system message.
func PrependProjectInstruction(messages []openai.ChatCompletionMessage, instruction string) []openai.ChatCompletionMessage {
	return prependInstructionSystemMessage(messages, instruction, projectInstructionModuleName)
}

func appendSystemContent(messages []openai.ChatCompletionMessage, additional, moduleName, defaultPersona string) []openai.ChatCompletionMessage {
	marker := moduleMarker(moduleName)
	result := make([]openai.ChatCompletionMessage, 0, len(messages)+1)
	systemFound := false

	for _, m := range messages {
		msg := cloneMessage(m)
		if msg.Role == openai.ChatMessageRoleSystem {
			if !hasMarker(msg.Content, marker) && strings.TrimSpace(additional) != "" {
				var builder strings.Builder
				builder.WriteString(msg.Content)
				builder.WriteString("\n\n")
				builder.WriteString(additional)
				builder.WriteString("\n")
				builder.WriteString(marker)
				msg.Content = builder.String()
			}
			systemFound = true
		}
		result = append(result, msg)
	}

	if !systemFound {
		var builder strings.Builder
		personaText := "You are a helpful assistant."
		if strings.TrimSpace(defaultPersona) != "" {
			personaText = fmt.Sprintf("You are a %s. Follow the rules strictly.", defaultPersona)
		}
		builder.WriteString(personaText)
		if strings.TrimSpace(additional) != "" {
			builder.WriteString("\n\n")
			builder.WriteString(additional)
		}
		builder.WriteString("\n")
		builder.WriteString(marker)
		systemMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: builder.String(),
		}
		result = append([]openai.ChatCompletionMessage{systemMsg}, result...)
	}

	return result
}

func hasModuleMarker(messages []openai.ChatCompletionMessage, moduleName string) bool {
	marker := moduleMarker(moduleName)
	for _, msg := range messages {
		if msg.Role == openai.ChatMessageRoleSystem && hasMarker(msg.Content, marker) {
			return true
		}
	}
	return false
}

func personaFromPreferences(preferences map[string]interface{}) string {
	if preferences == nil {
		return ""
	}
	if persona, ok := preferences["persona"]; ok {
		switch val := persona.(type) {
		case string:
			return strings.TrimSpace(val)
		case []byte:
			return strings.TrimSpace(string(val))
		default:
			return strings.TrimSpace(fmt.Sprint(val))
		}
	}
	return ""
}

func disabledModules(preferences map[string]interface{}) map[string]struct{} {
	disabled := map[string]struct{}{}
	if preferences == nil {
		return disabled
	}
	raw, ok := preferences["disable_modules"]
	if !ok {
		return disabled
	}
	switch v := raw.(type) {
	case string:
		for _, part := range strings.Split(v, ",") {
			if trimmed := strings.ToLower(strings.TrimSpace(part)); trimmed != "" {
				disabled[trimmed] = struct{}{}
			}
		}
	case []string:
		for _, part := range v {
			if trimmed := strings.ToLower(strings.TrimSpace(part)); trimmed != "" {
				disabled[trimmed] = struct{}{}
			}
		}
	case []interface{}:
		for _, part := range v {
			if str, ok := part.(string); ok {
				if trimmed := strings.ToLower(strings.TrimSpace(str)); trimmed != "" {
					disabled[trimmed] = struct{}{}
				}
			}
		}
	}
	return disabled
}

func isModuleDisabled(preferences map[string]interface{}, moduleName string) bool {
	disabled := disabledModules(preferences)
	_, found := disabled[strings.ToLower(moduleName)]
	return found
}

// ProjectInstructionModule injects project-specific instructions at the start of the conversation.
type ProjectInstructionModule struct{}

// NewProjectInstructionModule creates a new project instruction module.
func NewProjectInstructionModule() *ProjectInstructionModule {
	return &ProjectInstructionModule{}
}

// Name returns the module identifier.
func (m *ProjectInstructionModule) Name() string {
	return projectInstructionModuleName
}

// ShouldApply determines if project instructions should be injected.
func (m *ProjectInstructionModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if promptCtx == nil {
		return false
	}
	return strings.TrimSpace(promptCtx.ProjectInstruction) != ""
}

// Apply prepends the project instruction as a system message.
func (m *ProjectInstructionModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if promptCtx == nil || strings.TrimSpace(promptCtx.ProjectInstruction) == "" {
		return messages, nil
	}

	return PrependProjectInstruction(messages, promptCtx.ProjectInstruction), nil
}

// PersonaModule ensures a consistent system prompt/persona is applied
type PersonaModule struct {
	defaultPersona string
}

// NewPersonaModule creates a new persona module
func NewPersonaModule(defaultPersona string) *PersonaModule {
	return &PersonaModule{defaultPersona: strings.TrimSpace(defaultPersona)}
}

// Name returns the module identifier
func (m *PersonaModule) Name() string {
	return "persona"
}

// resolvePersona picks persona from user preferences or default
func (m *PersonaModule) resolvePersona(promptCtx *Context) string {
	if promptCtx != nil {
		if persona := personaFromPreferences(promptCtx.Preferences); persona != "" {
			return persona
		}
	}
	if m.defaultPersona != "" {
		return m.defaultPersona
	}
	return "helpful assistant"
}

// ShouldApply always applies when a persona is available
func (m *PersonaModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if promptCtx == nil {
		return false
	}
	persona := m.resolvePersona(promptCtx)
	return strings.TrimSpace(persona) != ""
}

// Apply injects or prefixes the system prompt with persona instructions
func (m *PersonaModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if promptCtx == nil {
		return messages, nil
	}

	persona := m.resolvePersona(promptCtx)
	if persona == "" {
		return messages, nil
	}

	personaText := fmt.Sprintf("You are a %s. Follow the rules strictly.", persona)
	result := appendSystemContent(messages, personaText, m.Name(), persona)
	return result, nil
}

// UserProfileModule injects user profile personalization into the system prompt.
type UserProfileModule struct{}

// NewUserProfileModule creates a new user profile module.
func NewUserProfileModule() *UserProfileModule {
	return &UserProfileModule{}
}

// Name returns the module identifier.
func (m *UserProfileModule) Name() string {
	return userProfileModuleName
}

// ShouldApply determines if user profile information should be injected.
func (m *UserProfileModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if promptCtx == nil || promptCtx.Profile == nil {
		return false
	}
	profile := promptCtx.Profile

	// Apply when any personalization field is present (base style defaults to Friendly so non-empty).
	return profile.BaseStyle != "" ||
		strings.TrimSpace(profile.CustomInstructions) != "" ||
		strings.TrimSpace(profile.NickName) != "" ||
		strings.TrimSpace(profile.Occupation) != "" ||
		strings.TrimSpace(profile.MoreAboutYou) != ""
}

func baseStyleInstruction(style usersettings.BaseStyle) string {
	switch style {
	case usersettings.BaseStyleConcise:
		return "Use a concise style: brief, direct answers with minimal filler."
	case usersettings.BaseStyleFriendly:
		return "Use a friendly, warm, and encouraging tone while staying helpful."
	case usersettings.BaseStyleProfessional:
		return "Use a professional, clear, and structured tone appropriate for business settings."
	default:
		if strings.TrimSpace(string(style)) != "" {
			return fmt.Sprintf("Use the user's preferred style: %s.", style)
		}
		return ""
	}
}

// Apply injects user profile guidance and persona instructions.
func (m *UserProfileModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if promptCtx == nil || promptCtx.Profile == nil {
		return messages, nil
	}

	profile := promptCtx.Profile
	var sections []string

	if styleText := baseStyleInstruction(profile.BaseStyle); styleText != "" {
		sections = append(sections, styleText)
	}

	if custom := strings.TrimSpace(profile.CustomInstructions); custom != "" {
		sections = append(sections, fmt.Sprintf("Custom instructions from the user:\n%s", custom))
	}

	var details []string
	if nick := strings.TrimSpace(profile.NickName); nick != "" {
		details = append(details, fmt.Sprintf("Address the user as \"%s\".", nick))
	}
	if occupation := strings.TrimSpace(profile.Occupation); occupation != "" {
		details = append(details, fmt.Sprintf("Occupation: %s.", occupation))
	}
	if more := strings.TrimSpace(profile.MoreAboutYou); more != "" {
		details = append(details, fmt.Sprintf("About the user: %s.", more))
	}
	if len(details) > 0 {
		var builder strings.Builder
		builder.WriteString("User context:\n")
		for _, detail := range details {
			builder.WriteString("- ")
			builder.WriteString(detail)
			builder.WriteString("\n")
		}
		sections = append(sections, strings.TrimSpace(builder.String()))
	}

	instruction := strings.TrimSpace(strings.Join(sections, "\n\n"))
	if instruction == "" {
		return messages, nil
	}

	result := appendSystemContent(messages, instruction, m.Name(), "")
	return result, nil
}

// WithDisabledModules returns a shallow copy of Context with module disable list merged
func WithDisabledModules(ctx *Context, disable []string) *Context {
	if ctx == nil {
		return &Context{
			Preferences: map[string]interface{}{
				"disable_modules": disable,
			},
		}
	}
	prefs := ctx.Preferences
	if prefs == nil {
		prefs = map[string]interface{}{}
	}
	prefs["disable_modules"] = disable
	ctx.Preferences = prefs
	return ctx
}

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
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if !m.enabled || promptCtx == nil {
		return false
	}
	return len(promptCtx.Memory) > 0
}

// Apply adds memory to the system prompt
func (m *MemoryModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if promptCtx == nil || len(promptCtx.Memory) == 0 {
		return messages, nil
	}

	var builder strings.Builder
	builder.WriteString("Use the following personal memory for this user:\n")
	for _, item := range promptCtx.Memory {
		builder.WriteString("- ")
		builder.WriteString(item)
		builder.WriteString("\n")
	}

	result := appendSystemContent(messages, strings.TrimSpace(builder.String()), m.Name(), "")
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
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	// Check if any preferences indicate tool usage
	if !m.enabled {
		return false
	}

	if promptCtx == nil {
		return false
	}

	if detectToolUsage(promptCtx, messages) {
		return true
	}

	return false
}

// Apply adds tool instructions to the system prompt
func (m *ToolInstructionsModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if promptCtx == nil || !detectToolUsage(promptCtx, messages) {
		return messages, nil
	}

	var builder strings.Builder
	builder.WriteString("You have access to various tools. Always choose the best tool for the task.\n")
	builder.WriteString("When you need to search for information, use web search. When you need to execute code, use the code execution tool.")

	if promptCtx != nil && promptCtx.Preferences != nil {
		if desc, ok := promptCtx.Preferences["tool_descriptions"].(string); ok && strings.TrimSpace(desc) != "" {
			builder.WriteString("\nAvailable tools: ")
			builder.WriteString(strings.TrimSpace(desc))
		}
		if list, ok := promptCtx.Preferences["tool_descriptions"].([]string); ok && len(list) > 0 {
			builder.WriteString("\nAvailable tools:\n")
			for _, item := range list {
				builder.WriteString("- ")
				builder.WriteString(strings.TrimSpace(item))
				builder.WriteString("\n")
			}
		}
	}

	result := appendSystemContent(messages, strings.TrimSpace(builder.String()), m.Name(), "")
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
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	// Check last user message for code-related keywords
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == openai.ChatMessageRoleUser {
			content := strings.ToLower(messages[i].Content)
			if isLikelyCodeQuery(content) {
				return true
			}
			break
		}
	}
	return false
}

// Apply adds code assistant instructions
func (m *CodeAssistantModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if hasModuleMarker(messages, m.Name()) {
		return messages, nil
	}

	var builder strings.Builder
	builder.WriteString("When providing code assistance:\n")
	builder.WriteString("1. Provide clear, well-commented code\n")
	builder.WriteString("2. Explain your approach and reasoning\n")
	builder.WriteString("3. Include error handling where appropriate\n")
	builder.WriteString("4. Follow best practices and conventions\n")
	builder.WriteString("5. Suggest testing approaches when relevant")

	result := appendSystemContent(messages, builder.String(), m.Name(), "")
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
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	// Apply for complex questions
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == openai.ChatMessageRoleUser {
			content := messages[i].Content
			if isComplexQuestion(content) {
				return true
			}
			break
		}
	}
	return false
}

// Apply adds chain-of-thought instructions
func (m *ChainOfThoughtModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if hasModuleMarker(messages, m.Name()) {
		return messages, nil
	}

	var builder strings.Builder
	builder.WriteString("For complex questions, think step-by-step:\n")
	builder.WriteString("1. Break down the problem\n")
	builder.WriteString("2. Analyze each component\n")
	builder.WriteString("3. Consider different perspectives\n")
	builder.WriteString("4. Synthesize your conclusion\n")
	builder.WriteString("5. Provide a clear, structured answer")

	result := appendSystemContent(messages, builder.String(), m.Name(), "")
	return result, nil
}

func detectToolUsage(promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if promptCtx != nil && promptCtx.Preferences != nil {
		if useTools, ok := promptCtx.Preferences["use_tools"].(bool); ok && useTools {
			return true
		}
	}

	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == openai.ChatMessageRoleTool {
			return true
		}
		if len(messages[i].ToolCalls) > 0 || messages[i].FunctionCall != nil {
			return true
		}
	}
	return false
}

func isLikelyCodeQuery(content string) bool {
	if content == "" {
		return false
	}
	if strings.Contains(content, "```") {
		return true
	}
	strongSignals := []string{"func ", "function(", "class ", "package ", "import ", "console.log", "panic(", "error ", "exception", "stack trace", "traceback", "sql", "json", "yaml", "schema"}
	for _, sig := range strongSignals {
		if strings.Contains(content, sig) {
			return true
		}
	}

	if strings.Contains(content, "code of conduct") {
		return false
	}

	codeKeywords := []string{"code", "function", "implement", "debug", "bug", "syntax", "compile", "script", "api", "snippet", "library"}
	actionKeywords := []string{"write", "example", "implement", "show", "fix", "break down", "refactor", "debug", "troubleshoot"}
	keywordHit := false
	for _, keyword := range codeKeywords {
		if strings.Contains(content, keyword) {
			keywordHit = true
			break
		}
	}
	actionHit := false
	for _, act := range actionKeywords {
		if strings.Contains(content, act) {
			actionHit = true
			break
		}
	}
	return keywordHit && actionHit
}

func isComplexQuestion(content string) bool {
	if strings.TrimSpace(content) == "" {
		return false
	}
	lower := strings.ToLower(content)
	reasoningKeywords := []string{"why", "how", "explain", "analyze", "compare", "evaluate", "what if", "step by step"}
	for _, keyword := range reasoningKeywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}

	wordCount := len(strings.Fields(content))
	if wordCount >= 20 && strings.Contains(content, "?") {
		return true
	}
	if wordCount >= 30 {
		return true
	}
	return false
}
