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
	for _, ts := range engine.Settings.Tools {
		toolIns := tool.NewAPITool(tool.APIToolBundle{
			ServerURL:  ts.Extensions["serverURL"].(string),
			Parameters: ts.Parameters,
			Method:     ts.Extensions["method"].(string),
			PostCode:   ts.Extensions["postCode"].(string),
		})
		output, err := toolIns.Invoke(ctx, inputs)
		if err != nil {
			return nil, err
		}
		answer += output
	}

	return answer, nil
}

// Stream ... return channel， not support streaming
func (engine *ProxyEngine) Stream(ctx context.Context, input Input) *EventStream {
	resultChan := NewEventStream()

	inputs := input.Inputs

	if len(engine.Settings.Tools) != 1 {
		resultChan.Push(ErrorEvent(errors.New("proxy engine only support one tool")))
		return resultChan
	}
	setting := engine.Settings.Tools[0]
	// query := input.Query
	toolIns := tool.NewAPITool(tool.APIToolBundle{
		ServerURL:  setting.Extensions["serverURL"].(string),
		Parameters: setting.Parameters,
		Method:     setting.Extensions["method"].(string),
		PostCode:   setting.Extensions["postCode"].(string),
	})
	go func() {
		output, err := toolIns.Stream(ctx, inputs)
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
