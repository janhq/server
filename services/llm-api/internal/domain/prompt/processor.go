package prompt

import (
	"context"

	"github.com/rs/zerolog"
	openai "github.com/sashabaranov/go-openai"
)

// ProcessorImpl implements the Processor interface
type ProcessorImpl struct {
	config  ProcessorConfig
	modules []Module
	log     zerolog.Logger
}

// NewProcessor creates a new prompt processor with the given configuration
func NewProcessor(config ProcessorConfig, log zerolog.Logger) *ProcessorImpl {
	processor := &ProcessorImpl{
		config:  config,
		modules: make([]Module, 0),
		log:     log.With().Str("component", "prompt-processor").Logger(),
	}

	// Register modules based on configuration
	if config.EnableMemory {
		processor.RegisterModule(NewMemoryModule(true))
	}

	if config.EnableTools {
		processor.RegisterModule(NewToolInstructionsModule(true))
	}

	// Always register conditional modules
	processor.RegisterModule(NewCodeAssistantModule())
	processor.RegisterModule(NewChainOfThoughtModule())

	return processor
}

// RegisterModule adds a module to the processor
func (p *ProcessorImpl) RegisterModule(module Module) {
	p.modules = append(p.modules, module)
	p.log.Debug().Str("module", module.Name()).Msg("registered prompt module")
}

// Process applies all relevant modules to the messages
func (p *ProcessorImpl) Process(
	ctx context.Context,
	promptCtx *Context,
	messages []openai.ChatCompletionMessage,
) ([]openai.ChatCompletionMessage, error) {
	result := messages
	appliedModules := make([]string, 0)

	for _, module := range p.modules {
		if module.ShouldApply(ctx, promptCtx, result) {
			var err error
			result, err = module.Apply(ctx, promptCtx, result)
			if err != nil {
				p.log.Error().
					Err(err).
					Str("module", module.Name()).
					Msg("failed to apply prompt module")
				return messages, err
			}
			appliedModules = append(appliedModules, module.Name())
		}
	}

	if len(appliedModules) > 0 {
		p.log.Debug().
			Strs("applied_modules", appliedModules).
			Str("conversation_id", promptCtx.ConversationID).
			Msg("applied prompt orchestration modules")
	}

	return result, nil
}
