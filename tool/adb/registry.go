package adb

import (
	"context"
	"errors"

	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/tool"
)

var (
	registry *Registry = NewRegistry()
)

// The main service class that manages action registration and execution
type Registry struct {
	Tools map[string]*tool.Tool
}

func NewRegistry() *Registry {
	return &Registry{
		Tools: make(map[string]*tool.Tool),
	}
}

// RegisterTool should be called after registry initialization
// registry.Tool("click_element_by_index", ClickElementFunc, "click action", paramModel, domains, pageFilter)
func RegisterTool[T, D any](
	r *Registry,
	name string,
	description string,
	function tool.ToolFunc[T, D],
) error {

	toolx, err := tool.NewWithToolFunc(name, description, function)
	if err != nil {
		return err
	}
	r.Tools[name] = toolx
	tool.Register(toolx)
	log.Info("注册工具", "name", name, "description", description)
	return nil
}

// ExecuteTool a registered action
// TODO(LOW): support Context
func (r *Registry) ExecuteTool(
	ctx context.Context,
	toolName string,
	arguments string,
	sensitiveData map[string]string,
) (string, error) {
	tool, ok := r.Tools[toolName]
	if !ok {
		return "", errors.New("tool not found")
	}

	result, err := tool.Invoke(ctx, arguments)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (r *Registry) AvailableTools(includeTools []string) []any {
	// Create model from registered actions, used by LLM APIs that support tool calling

	// Filter tools based on includeTools if provided:
	//   if includeTools is nil, only include tools with no filters
	//   if includeTools is provided, only include tools that match the includeTools

	availableTools := make([]any, 0)
	for name := range r.Tools {
		availableTools = append(availableTools, name)
	}

	return availableTools
}
