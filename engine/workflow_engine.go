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

// Execute ... return channel， 不支持streaming
func (engine *WorkflowEngine) Execute(ctx context.Context, input Input) *EventStream {
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
			resultChan.Push(WorkflowEvent(s.Data))
		}
		resultChan.Push(WorkflowResultEvent(result))
	}()
	return resultChan
}
