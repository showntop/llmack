package engine

import (
	"context"

	"github.com/showntop/llmack/llm"
)

// ChatEngine ...
type ChatEngine struct {
	*BotEngine
}

// NewChatEngine ...
func NewChatEngine(settings *Settings, opts ...Option) *ChatEngine {
	r := &ChatEngine{}
	r.BotEngine = NewBotEngine(opts...)
	r.Settings = settings
	return r
}

// Execute ... return channel
func (r *ChatEngine) Execute(ctx context.Context, input Input) *EventStream {
	settings := r.Settings
	estm := NewEventStream()

	inputs := input.Inputs
	query := input.Query
	contexts := ""

	messages, _ := r.RenderMessages(ctx, settings.PresetPrompt, inputs, query, contexts)

	// Invoke model
	instance := llm.NewInstance(settings.LLMModel.Provider)
	response, err := instance.Invoke(ctx, messages,
		llm.WithModel(settings.LLMModel.Name),
		llm.WithStream(true),
	)
	if err != nil {
		r.opts.logger.ErrorContextf(ctx, "invoke model failed: %s", err)
		estm.Push(ErrorEvent(err))
		estm.Close()
		return estm
	}
	go func() {
		final := ""
		for r := response.Stream().Next(); r != nil; r = response.Stream().Next() {
			final += r.Delta.Message.Content()
			estm.Push(ToastEvent(r))
		}
		estm.Push(EndEvent(final))
		estm.Close()
	}()
	return estm
}

// Invoke ... return channel
func (r *ChatEngine) Invoke(ctx context.Context, input Input) (any, error) {
	ctx = r.HookOnStart(ctx)
	defer r.HookOnFinish(ctx, nil)

	// chain call
	// agent.Context(contexts).Inputs(inputs).Query(query).LLM(settings.LLMModel.Provider).Invoke()

	settings := r.Settings

	inputs := input.Inputs
	query := input.Query
	// context := input.Context
	contexts := ""

	messages, _ := r.RenderMessages(ctx, settings.PresetPrompt, inputs, query, contexts)
	// 如果配置了知识库

	// Invoke model
	ctx = r.BeforeLLMStart(ctx)
	instance := llm.NewInstance(settings.LLMModel.Provider)
	response, err := instance.Invoke(ctx, messages, nil, llm.WithModel(settings.LLMModel.Name)) // llm.WithModel("quanyutong"),
	if err != nil {
		return nil, err
	}
	r.AfterLLMFinish(ctx, err)

	return response, nil
}
