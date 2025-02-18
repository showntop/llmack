package engine

import (
	"context"
	"errors"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/tool"
)

// ProxyEngine ...
type ProxyEngine struct {
	*BotEngine
}

// NewProxyEngine ...
func NewProxyEngine(settings *Settings, opts ...Option) Engine {
	engine := &ProxyEngine{}
	engine.Settings = settings
	return engine
}

// Invoke ... return channel， 不支持streaming
func (engine *ProxyEngine) Invoke(ctx context.Context, input Input) (any, error) {

	inputs := input.Inputs
	// query := input.Query
	// 调用 工具
	answer := ""
	for _, toolName := range engine.Settings.Tools {
		toolIns := tool.Spawn(toolName)
		output, err := toolIns.Invoke(ctx, inputs)
		if err != nil {
			return nil, err
		}
		answer += output
	}

	return answer, nil
}

// Execute ... return channel， not support streaming
func (engine *ProxyEngine) Execute(ctx context.Context, input Input) *EventStream {
	resultChan := NewEventStream()

	inputs := input.Inputs

	if len(engine.Settings.Tools) != 1 {
		resultChan.Push(ErrorEvent(errors.New("proxy engine only support one tool")))
		return resultChan
	}
	toolName := engine.Settings.Tools[0]
	// query := input.Query
	toolIns := tool.Spawn(toolName)
	go func() {
		output, err := toolIns.Invoke(ctx, inputs)
		if err != nil {
			resultChan.Push(ErrorEvent(err))
			return
		}
		_ = output
		msg := llm.AssistantPromptMessage("answer")
		resultChan.Push(EndEvent(msg.Content()))
	}()

	return resultChan
}
