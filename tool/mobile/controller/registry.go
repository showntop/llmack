package controller

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
)

// The main service class that manages action registration and execution
type Registry struct {
	Actions        map[string]*Action
	ExcludeActions []string
}

func NewRegistry() *Registry {
	return &Registry{
		Actions:        make(map[string]*Action),
		ExcludeActions: []string{},
	}
}

// Action registers a new action into the registry.
// should be called after registry initialization
// registry.Action("click_element_by_index", ClickElementFunc, "click action", paramModel, domains, pageFilter)
func RegisterAction[T, D any](
	r *Registry,
	name string,
	description string,
	function ActionFunc[T, D],
) error {
	// if ExcludeActions contains name, return
	if slices.Contains(r.ExcludeActions, name) {
		return errors.New("action " + name + " is already registered")
	}

	action, err := NewAction(name, description, function)
	if err != nil {
		return err
	}
	r.Actions[name] = action
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

func (r *Registry) replaceSensitiveData(arguments string, sensitiveData map[string]string) string {
	secretPattern := regexp.MustCompile(`<secret>(.*?)</secret>`)

	replaceSecrets := func(value string) string {
		if strings.Contains(value, "<secret>") {
			matches := secretPattern.FindAllStringSubmatch(value, -1)
			for _, match := range matches {
				placeholder := match[1]
				if _, ok := sensitiveData[placeholder]; ok {
					value = strings.ReplaceAll(value, fmt.Sprintf("<secret>%s</secret>", placeholder), sensitiveData[placeholder])
				}
			}
		}
		return value
	}
	return replaceSecrets(arguments)
}

func (r *Registry) AvailableActions(includeActions []string) map[string]*Action {
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
