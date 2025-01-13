package engine

import (
	"context"

	"github.com/showntop/llmack/workflow"
	"github.com/showntop/llmack/workflow/memory"
)

// WorkflowEngine ...
type WorkflowEngine struct {
	BotEngine
}

// NewWorkflowEngine ...
func NewWorkflowEngine(settings *Settings, opts ...Option) *WorkflowEngine {
	engine := &WorkflowEngine{}
	engine.Settings = settings
	return engine
}

// Invoke ... return channel， 不支持streaming
func (engine *WorkflowEngine) Invoke(ctx context.Context, input Input) (any, error) {

	inputs := input.Inputs
	// query := input.Query
	// 调用 工具
	// new workflow exeutor
	excutor := memory.NewExecutor(engine.Settings.Workflow)
	result, err := excutor.Execute(ctx, inputs)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Stream ... return channel， 不支持streaming
func (engine *WorkflowEngine) Stream(ctx context.Context, input Input) *EventStream {
	resultChan := NewEventStream()
	if input.Inputs == nil {
		input.Inputs = make(map[string]any)
	}
	input.Inputs["query"] = input.Query
	// new workflow exeutor
	excutor := memory.NewExecutor(engine.Settings.Workflow)

	var result *workflow.Result
	go func() {
		result, _ = excutor.Execute(ctx, input.Inputs)
	}()
	go func() {
		defer resultChan.Close()

		for s := range excutor.Events() {
			resultChan.Push(WorkflowEvent(s))
		}
		resultChan.Push(WorkflowResultEvent(result))
	}()
	return resultChan
}
