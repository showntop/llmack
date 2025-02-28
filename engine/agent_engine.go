package engine

import (
	"context"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/program"
)

// AgentEngine ...
type AgentEngine struct {
	ConvID    int64
	MessageID int64

	thoughs []llm.Message
	BotEngine
}

// NewAgentEngine ...
func NewAgentEngine(settings *Settings, opts ...Option) Engine {
	r := &AgentEngine{}
	r.BotEngine = *NewBotEngine(opts...)
	// load tools

	r.Settings = settings
	return r
}

// Execute ... return channel
// ReAct mode or Function Call mode.
func (engine *AgentEngine) Execute(ctx context.Context, input Input) *EventStream {
	response := NewEventStream()

	settings := engine.Settings
	inputs := input.Inputs
	// query := input.Query
	// contexts, err := engine.renderContexts(ctx, settings, query)
	// if err != nil {
	// 	// return nil, err
	// }
	// tools

	var result *program.Result
	if settings.Agent.Mode == "ReAct" {
		result = program.ReAct(program.WithLLM(engine.Settings.LLMModel.Provider, engine.Settings.LLMModel.Name)).
			WithInstruction(engine.Settings.PresetPrompt).
			WithTools(settings.Tools...).
			Invoke(ctx, inputs)
		if result.Error() != nil {
			response.Push(ErrorEvent(result.Error()))
			return response
		}
		go func() {
			defer response.Close()
			for message := range result.Stream() {
				response.Push(ToastEvent(message))
			}
			response.Push(EndEvent(result.Completion()))
		}()
	} else {
		result = program.FunCall(program.WithLLM(engine.Settings.LLMModel.Provider, engine.Settings.LLMModel.Name)).
			WithInstruction(engine.Settings.PresetPrompt).
			WithTools(settings.Tools...).
			Invoke(ctx, inputs)
		if result.Error() != nil {
			response.Push(ErrorEvent(result.Error()))
			return response
		}

		go func() {
			defer response.Close()
			for message := range result.Stream() {
				response.Push(ToastEvent(message))
			}
			response.Push(EndEvent(result.Completion()))
		}()
	}
	return response
}
