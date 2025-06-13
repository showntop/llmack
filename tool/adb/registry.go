package adb

import (
	"context"
	"errors"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/showntop/llmack/tool"
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

// Tool registers a new tool into the registry.
// should be called after registry initialization
// registry.Tool("click_element_by_index", ClickElementFunc, "click action", paramModel, domains, pageFilter)
func RegisterTool[T, D any](
	r *Registry,
	name string,
	description string,
	function tool.ToolFunc[T, D],
) error {

	tool, err := tool.NewWithToolFunc(name, description, function)
	if err != nil {
		return err
	}
	r.Tools[name] = tool
	return nil
}

// Execute a registered action
// TODO(LOW): support Context
func (r *Registry) ExecuteAction(
	ctx context.Context,
	actionName string,
	arguments string,
	sensitiveData map[string]string,
) (string, error) {
	action, ok := r.Actions[actionName]
	if !ok {
		return "", errors.New("action not found")
	}

	if len(sensitiveData) > 0 {
		arguments = r.replaceSensitiveData(arguments, sensitiveData)
		log.Debug(arguments)
	}

	result, err := action.Tool.Invoke(ctx, arguments)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (r *Registry) AvailableTools(includeTools []string) map[string]*tool.Tool {
	// Create model from registered actions, used by LLM APIs that support tool calling

	// Filter actions based on page if provided:
	//   if page is None, only include actions with no filters
	//   if page is provided, only include actions that match the page

	availableActions := make(map[string]*Action)
	for name, action := range r.Actions {
		if includeActions != nil && !slices.Contains(includeActions, name) {
			continue
		}
		availableActions[name] = action
	}

	return availableActions
}
