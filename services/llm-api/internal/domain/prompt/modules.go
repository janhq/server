package prompt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"

	"jan-server/services/llm-api/internal/domain/prompttemplate"
	"jan-server/services/llm-api/internal/domain/usersettings"
)

const (
	projectInstructionModuleName = "project_instruction"
	userProfileModuleName        = "user_profile"
)

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

// prependInstructionSystemMessage returns a copy of messages with the instruction system
// message prepended.
func prependInstructionSystemMessage(messages []openai.ChatCompletionMessage, instruction, moduleName string) []openai.ChatCompletionMessage {
	trimmed := strings.TrimSpace(instruction)
	if trimmed == "" {
		return messages
	}

	var builder strings.Builder
	builder.WriteString(trimmed)

	// Special handling for project instructions: explicitly state priority
	if moduleName == projectInstructionModuleName {
		builder.WriteString("\n\n")
		builder.WriteString("Project priority: These project-specific instructions have the highest priority. ")
		builder.WriteString("If any user settings, style preferences, or other guidance conflict with these project instructions, ")
		builder.WriteString("you must follow the project instructions.")
	}

	result := make([]openai.ChatCompletionMessage, 0, len(messages)+1)
	result = append(result, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: builder.String(),
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

// appendSystemContent attaches "additional" instructions into a suitable system message
// or creates a new system message if needed.
//
// Rules:
//   - We NEVER modify the first system message if it contains project instructions.
//   - When we create a new system message, we insert it right after the first
//     system message (if present) so project instructions stay topologically first.
//   - baseContent is the base system content (if non-empty) when creating a new message.
func appendSystemContent(
	messages []openai.ChatCompletionMessage,
	additional, moduleName, baseContent string,
) []openai.ChatCompletionMessage {
	additional = strings.TrimSpace(additional)

	result := make([]openai.ChatCompletionMessage, 0, len(messages)+1)
	applied := false
	firstSystemIdx := -1

	for i, m := range messages {
		msg := cloneMessage(m)
		if msg.Role == openai.ChatMessageRoleSystem {
			if firstSystemIdx == -1 {
				firstSystemIdx = i
				// Don't modify the first system message (may contain project instructions)
			} else if !applied && additional != "" {
				// Append to subsequent system messages
				var b strings.Builder
				b.WriteString(strings.TrimSpace(msg.Content))
				b.WriteString("\n\n")
				b.WriteString(additional)
				msg.Content = b.String()
				applied = true
			}
		}
		result = append(result, msg)
	}

	// If we successfully appended to an existing system message, we're done.
	if applied {
		return result
	}

	// If we have nothing to say and no base content, just return.
	if additional == "" && strings.TrimSpace(baseContent) == "" {
		return result
	}

	// Create a new system message with base content + additional content.
	var builder strings.Builder
	baseText := strings.TrimSpace(baseContent)
	if baseText == "" {
		baseText = "You are Jan, a helpful AI assistant.  Jan is trained by Menlo Research (https://www.menlo.ai) - a fame research lab. "
	}
	builder.WriteString(baseText)

	if additional != "" {
		builder.WriteString("\n\n")
		builder.WriteString(additional)
	}

	systemMsg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: builder.String(),
	}

	// Insert AFTER first system message if it exists,
	// so project instructions stay top priority.
	insertIdx := 0
	if firstSystemIdx >= 0 {
		insertIdx = firstSystemIdx + 1
	}

	// Insert at insertIdx
	result = append(result, openai.ChatCompletionMessage{}) // grow slice
	copy(result[insertIdx+1:], result[insertIdx:])
	result[insertIdx] = systemMsg

	return result
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
	if promptCtx.Preferences != nil && isModuleDisabled(promptCtx.Preferences, m.Name()) {
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

// TimingModule injects the AI assistant intro and current date into the system prompt.
type TimingModule struct {
	templateService *prompttemplate.Service
}

// NewTimingModule creates a new timing module without template service (uses fallback).
func NewTimingModule() *TimingModule {
	return &TimingModule{}
}

// NewTimingModuleWithService creates a new timing module with template service.
func NewTimingModuleWithService(service *prompttemplate.Service) *TimingModule {
	return &TimingModule{
		templateService: service,
	}
}

// Name returns the module identifier.
func (m *TimingModule) Name() string {
	return "timing"
}

// ShouldApply always applies when prompt orchestration is enabled and module not disabled.
func (m *TimingModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if promptCtx != nil && promptCtx.Preferences != nil && isModuleDisabled(promptCtx.Preferences, m.Name()) {
		return false
	}
	return true
}

// Apply injects the AI assistant intro and current date.
func (m *TimingModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}

	// Format current date as "Month Day, Year" (e.g., "November 28, 2025")
	currentDate := time.Now().Format("January 2, 2006")

	var timingText string

	// Try to fetch timing template from database and render with current date
	if m.templateService != nil {
		template, err := m.templateService.GetByKey(ctx, prompttemplate.TemplateKeyTiming)
		if err == nil && template != nil && template.IsActive {
			// Render template with current date variable
			rendered, renderErr := m.templateService.RenderTemplate(ctx, prompttemplate.TemplateKeyTiming, map[string]any{
				"CurrentDate": currentDate,
			})
			if renderErr == nil {
				timingText = rendered
				log.Info().
					Str("template_key", template.TemplateKey).
					Str("current_date", currentDate).
					Msg("TimingModule: Loaded and rendered template from database")
			} else {
				log.Warn().Err(renderErr).Msg("TimingModule: Failed to render template, using fallback")
			}
		} else {
			if err != nil {
				log.Warn().Err(err).Msg("TimingModule: Failed to load template from database, using fallback")
			}
		}
	}

	// Fallback to hardcoded timing text if template not loaded
	if timingText == "" {
		timingText = fmt.Sprintf(
			"You are Jan, a helpful AI assistant who helps the user with their requests. Jan is trained by Menlo Research (https://www.menlo.ai) - a fame research lab.\n"+
				"Today is: %s.\n"+
				"Always treat this as the current date.",
			currentDate,
		)
		log.Info().Msg("TimingModule: Using fallback hardcoded prompt")
	}

	result := appendSystemContent(messages, "", m.Name(), timingText)
	return result, nil
}

// UserProfileModule injects user profile personalization into the system prompt.
type UserProfileModule struct{
	templateService *prompttemplate.Service
}

// NewUserProfileModule creates a new user profile module without template service (uses fallback).
func NewUserProfileModule() *UserProfileModule {
	return &UserProfileModule{}
}

// NewUserProfileModuleWithService creates a new user profile module with template service.
func NewUserProfileModuleWithService(service *prompttemplate.Service) *UserProfileModule {
	return &UserProfileModule{
		templateService: service,
	}
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
	if promptCtx.Preferences != nil && isModuleDisabled(promptCtx.Preferences, m.Name()) {
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

	var instruction string

	// Try to fetch user_profile template from database first
	if m.templateService != nil {
		template, err := m.templateService.GetByKey(ctx, prompttemplate.TemplateKeyUserProfile)
		if err == nil && template != nil && template.IsActive {
			// Build variables for template rendering
			vars := map[string]any{
				"BaseStyle":          string(promptCtx.Profile.BaseStyle),
				"CustomInstructions": promptCtx.Profile.CustomInstructions,
				"NickName":           promptCtx.Profile.NickName,
				"Occupation":         promptCtx.Profile.Occupation,
				"MoreAboutYou":       promptCtx.Profile.MoreAboutYou,
			}

			// Render template with user profile variables
			rendered, renderErr := m.templateService.RenderTemplate(ctx, prompttemplate.TemplateKeyUserProfile, vars)
			if renderErr == nil {
				instruction = rendered
				log.Info().
					Str("template_key", template.TemplateKey).
					Msg("UserProfileModule: Loaded and rendered template from database")
			} else {
				log.Warn().Err(renderErr).Msg("UserProfileModule: Failed to render template, using fallback")
			}
		} else {
			if err != nil {
				log.Warn().Err(err).Msg("UserProfileModule: Failed to load template from database, using fallback")
			}
		}
	}

	// Fallback to hardcoded logic if template not loaded
	if instruction == "" {
		profile := promptCtx.Profile
		var sections []string

		// General note about precedence vs project instructions.
		sections = append(sections,
			"User-level settings are preferences for style and context. "+
				"If they ever conflict with explicit project or system instructions, always follow the project or system instructions.")

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

		instruction = strings.TrimSpace(strings.Join(sections, "\n\n"))
		log.Info().Msg("UserProfileModule: Using fallback hardcoded prompt")
	}

	if instruction == "" {
		return messages, nil
	}

	result := appendSystemContent(messages, instruction, m.Name(), "")
	return result, nil
}

// WithDisabledModules returns a shallow copy of Context with module disable list merged.
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

// MemoryModule adds user memory to system prompts.
type MemoryModule struct {
	enabled         bool
	templateService *prompttemplate.Service
}

// NewMemoryModule creates a new memory module without template service (uses fallback).
func NewMemoryModule(enabled bool) *MemoryModule {
	return &MemoryModule{enabled: enabled}
}

// NewMemoryModuleWithService creates a new memory module with template service.
func NewMemoryModuleWithService(enabled bool, service *prompttemplate.Service) *MemoryModule {
	return &MemoryModule{
		enabled:         enabled,
		templateService: service,
	}
}

// Name returns the module identifier.
func (m *MemoryModule) Name() string {
	return "memory"
}

// ShouldApply checks if memory should be included.
func (m *MemoryModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if !m.enabled || promptCtx == nil {
		return false
	}
	if promptCtx.Preferences != nil && isModuleDisabled(promptCtx.Preferences, m.Name()) {
		return false
	}
	return len(promptCtx.Memory) > 0
}

// Apply adds memory to the system prompt.
func (m *MemoryModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if promptCtx == nil || len(promptCtx.Memory) == 0 {
		return messages, nil
	}

	var memoryText string

	// Try to fetch memory template from database first
	if m.templateService != nil {
		template, err := m.templateService.GetByKey(ctx, prompttemplate.TemplateKeyMemory)
		if err == nil && template != nil && template.IsActive {
			// Build memory items for template rendering
			vars := map[string]any{
				"MemoryItems": promptCtx.Memory,
			}

			// Render template with memory variables
			rendered, renderErr := m.templateService.RenderTemplate(ctx, prompttemplate.TemplateKeyMemory, vars)
			if renderErr == nil {
				memoryText = rendered
				log.Info().
					Str("template_key", template.TemplateKey).
					Int("memory_count", len(promptCtx.Memory)).
					Msg("MemoryModule: Loaded and rendered template from database")
			} else {
				log.Warn().Err(renderErr).Msg("MemoryModule: Failed to render template, using fallback")
			}
		} else {
			if err != nil {
				log.Warn().Err(err).Msg("MemoryModule: Failed to load template from database, using fallback")
			}
		}
	}

	// Fallback to hardcoded memory text if template not loaded
	if memoryText == "" {
		var builder strings.Builder
		builder.WriteString("Use the following personal memory for this user when helpful, without overriding project or system instructions:\n")
		for _, item := range promptCtx.Memory {
			builder.WriteString("- ")
			builder.WriteString(item)
			builder.WriteString("\n")
		}
		memoryText = strings.TrimSpace(builder.String())
		log.Info().Msg("MemoryModule: Using fallback hardcoded prompt")
	}

	result := appendSystemContent(messages, memoryText, m.Name(), "")
	return result, nil
}

// ToolInstructionsModule adds tool usage instructions.
type ToolInstructionsModule struct {
	enabled         bool
	templateService *prompttemplate.Service
}

// NewToolInstructionsModule creates a new tool instructions module without template service (uses fallback).
func NewToolInstructionsModule(enabled bool) *ToolInstructionsModule {
	return &ToolInstructionsModule{enabled: enabled}
}

// NewToolInstructionsModuleWithService creates a new tool instructions module with template service.
func NewToolInstructionsModuleWithService(enabled bool, service *prompttemplate.Service) *ToolInstructionsModule {
	return &ToolInstructionsModule{
		enabled:         enabled,
		templateService: service,
	}
}

// Name returns the module identifier.
func (m *ToolInstructionsModule) Name() string {
	return "tool_instructions"
}

// ShouldApply checks if tool instructions should be added.
func (m *ToolInstructionsModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if !m.enabled {
		return false
	}
	if promptCtx == nil {
		return false
	}
	if promptCtx.Preferences != nil && isModuleDisabled(promptCtx.Preferences, m.Name()) {
		return false
	}

	return detectToolUsage(promptCtx, messages)
}

// Apply adds tool instructions to the system prompt.
func (m *ToolInstructionsModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}
	if promptCtx == nil || !detectToolUsage(promptCtx, messages) {
		return messages, nil
	}

	var toolText string

	// Try to fetch tool_instructions template from database first
	if m.templateService != nil {
		template, err := m.templateService.GetByKey(ctx, prompttemplate.TemplateKeyToolInstructions)
		if err == nil && template != nil && template.IsActive {
			// Build variables for template rendering
			vars := map[string]any{
				"ToolDescriptions": promptCtx.Preferences["tool_descriptions"],
			}

			// Render template with tool variables
			rendered, renderErr := m.templateService.RenderTemplate(ctx, prompttemplate.TemplateKeyToolInstructions, vars)
			if renderErr == nil {
				toolText = rendered
				log.Info().
					Str("template_key", template.TemplateKey).
					Msg("ToolInstructionsModule: Loaded and rendered template from database")
			} else {
				log.Warn().Err(renderErr).Msg("ToolInstructionsModule: Failed to render template, using fallback")
			}
		} else {
			if err != nil {
				log.Warn().Err(err).Msg("ToolInstructionsModule: Failed to load template from database, using fallback")
			}
		}
	}

	// Fallback to hardcoded tool text if template not loaded
	if toolText == "" {
		var builder strings.Builder
		builder.WriteString("You have access to various tools. Always choose the best tool for the task.\n")
		builder.WriteString("When you need to search for information, use web search. When you need to execute code, use the code execution tool.\n")
		builder.WriteString("Tool usage must respect project instructions and system-level constraints at all times.")

		if promptCtx.Preferences != nil {
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
		toolText = strings.TrimSpace(builder.String())
		log.Info().Msg("ToolInstructionsModule: Using fallback hardcoded prompt")
	}

	result := appendSystemContent(messages, toolText, m.Name(), "")
	return result, nil
}

// CodeAssistantModule adds code-specific instructions.
type CodeAssistantModule struct{
	templateService *prompttemplate.Service
}

// NewCodeAssistantModule creates a new code assistant module without template service (uses fallback).
func NewCodeAssistantModule() *CodeAssistantModule {
	return &CodeAssistantModule{}
}

// NewCodeAssistantModuleWithService creates a new code assistant module with template service.
func NewCodeAssistantModuleWithService(service *prompttemplate.Service) *CodeAssistantModule {
	return &CodeAssistantModule{
		templateService: service,
	}
}

// Name returns the module identifier.
func (m *CodeAssistantModule) Name() string {
	return "code_assistant"
}

// ShouldApply checks if the question is code-related.
func (m *CodeAssistantModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if promptCtx != nil && promptCtx.Preferences != nil && isModuleDisabled(promptCtx.Preferences, m.Name()) {
		return false
	}
	// Check last user message for code-related keywords.
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

// Apply adds code assistant instructions.
func (m *CodeAssistantModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}

	var codeText string

	// Try to fetch code_assistant template from database first
	if m.templateService != nil {
		template, err := m.templateService.GetByKey(ctx, prompttemplate.TemplateKeyCodeAssistant)
		if err == nil && template != nil && template.IsActive {
			// Render template (no variables needed for code assistant)
			rendered, renderErr := m.templateService.RenderTemplate(ctx, prompttemplate.TemplateKeyCodeAssistant, map[string]any{})
			if renderErr == nil {
				codeText = rendered
				log.Info().
					Str("template_key", template.TemplateKey).
					Msg("CodeAssistantModule: Loaded and rendered template from database")
			} else {
				log.Warn().Err(renderErr).Msg("CodeAssistantModule: Failed to render template, using fallback")
			}
		} else {
			if err != nil {
				log.Warn().Err(err).Msg("CodeAssistantModule: Failed to load template from database, using fallback")
			}
		}
	}

	// Fallback to hardcoded code assistant text if template not loaded
	if codeText == "" {
		var builder strings.Builder
		builder.WriteString("When providing code assistance:\n")
		builder.WriteString("1. Provide clear, well-commented code.\n")
		builder.WriteString("2. Explain your approach and reasoning.\n")
		builder.WriteString("3. Include error handling where appropriate.\n")
		builder.WriteString("4. Follow best practices and conventions.\n")
		builder.WriteString("5. Suggest testing approaches when relevant.\n")
		builder.WriteString("6. Respect project instructions and user constraints; never violate them to simplify code.")
		codeText = builder.String()
		log.Info().Msg("CodeAssistantModule: Using fallback hardcoded prompt")
	}

	result := appendSystemContent(messages, codeText, m.Name(), "")
	return result, nil
}

// ChainOfThoughtModule adds chain-of-thought reasoning instructions.
type ChainOfThoughtModule struct{
	templateService *prompttemplate.Service
}

// NewChainOfThoughtModule creates a new chain-of-thought module without template service (uses fallback).
func NewChainOfThoughtModule() *ChainOfThoughtModule {
	return &ChainOfThoughtModule{}
}

// NewChainOfThoughtModuleWithService creates a new chain-of-thought module with template service.
func NewChainOfThoughtModuleWithService(service *prompttemplate.Service) *ChainOfThoughtModule {
	return &ChainOfThoughtModule{
		templateService: service,
	}
}

// Name returns the module identifier.
func (m *ChainOfThoughtModule) Name() string {
	return "chain_of_thought"
}

// ShouldApply checks if the question requires reasoning.
func (m *ChainOfThoughtModule) ShouldApply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) bool {
	if ctx == nil || ctx.Err() != nil {
		return false
	}
	if promptCtx != nil && promptCtx.Preferences != nil && isModuleDisabled(promptCtx.Preferences, m.Name()) {
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

// Apply adds chain-of-thought instructions.
func (m *ChainOfThoughtModule) Apply(ctx context.Context, promptCtx *Context, messages []openai.ChatCompletionMessage) ([]openai.ChatCompletionMessage, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return messages, err
		}
	}

	var cotText string

	// Try to fetch chain_of_thought template from database first
	if m.templateService != nil {
		template, err := m.templateService.GetByKey(ctx, prompttemplate.TemplateKeyChainOfThought)
		if err == nil && template != nil && template.IsActive {
			// Render template (no variables needed for chain of thought)
			rendered, renderErr := m.templateService.RenderTemplate(ctx, prompttemplate.TemplateKeyChainOfThought, map[string]any{})
			if renderErr == nil {
				cotText = rendered
				log.Info().
					Str("template_key", template.TemplateKey).
					Msg("ChainOfThoughtModule: Loaded and rendered template from database")
			} else {
				log.Warn().Err(renderErr).Msg("ChainOfThoughtModule: Failed to render template, using fallback")
			}
		} else {
			if err != nil {
				log.Warn().Err(err).Msg("ChainOfThoughtModule: Failed to load template from database, using fallback")
			}
		}
	}

	// Fallback to hardcoded chain-of-thought text if template not loaded
	if cotText == "" {
		var builder strings.Builder
		builder.WriteString("For complex questions, think step-by-step:\n")
		builder.WriteString("1. Break down the problem\n")
		builder.WriteString("2. Analyze each component\n")
		builder.WriteString("3. Consider different perspectives\n")
		builder.WriteString("4. Synthesize your conclusion\n")
		builder.WriteString("5. Provide a clear, structured answer")
		cotText = builder.String()
		log.Info().Msg("ChainOfThoughtModule: Using fallback hardcoded prompt")
	}

	result := appendSystemContent(messages, cotText, m.Name(), "")
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
