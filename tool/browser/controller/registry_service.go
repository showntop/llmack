package controller

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/pkg/browser"

	"github.com/playwright-community/playwright-go"
)

// The main service class that manages action registration and execution
type Registry struct {
	Registry       *ActionRegistry
	ExcludeActions []string
}

func NewRegistry() *Registry {
	return &Registry{
		Registry:       NewActionRegistry(),
		ExcludeActions: []string{},
	}
}

// Action registers a new action into the registry.
// should be called after registry initialization
// registry.Action("click_element_by_index", ClickElementFunc, "click action", paramModel, domains, pageFilter)
func registerAction[T, D any](
	r *Registry,
	name string,
	description string,
	function ActionFunc[T, D],
	domains []string,
	pageFilter func(playwright.Page) bool,
) error {
	// if ExcludeActions contains name, return
	if slices.Contains(r.ExcludeActions, name) {
		return errors.New("action " + name + " is already registered")
	}

	action, err := NewRegisteredAction(name, description, function, domains, pageFilter)
	if err != nil {
		return err
	}
	r.Registry.Actions[name] = action
	return nil
}

type contextKey string

const (
	browserKey            contextKey = "browser"
	pageExtractionLlmKey  contextKey = "page_extraction_llm"
	availableFilePathsKey contextKey = "available_file_paths"
)

// Execute a registered action
// TODO(LOW): support Context
func (r *Registry) ExecuteAction(
	actionName string,
	argumentsInJson string,
	browser *browser.BrowserContext,
	model *llm.Instance,
	sensitiveData map[string]string,
	availableFilePaths []string,
	/*context Context*/) (string, error) {

	// ex) actionName: "ClickElementAction"
	action, ok := r.Registry.Actions[actionName]
	if !ok {
		return "", errors.New("action not found")
	}

	ctx := context.Background()
	if browser != nil {
		ctx = context.WithValue(ctx, browserKey, browser)
	}
	if model != nil {
		ctx = context.WithValue(ctx, pageExtractionLlmKey, model)
	}
	if availableFilePaths != nil {
		ctx = context.WithValue(ctx, availableFilePathsKey, availableFilePaths)
	}

	if len(sensitiveData) > 0 {
		argumentsInJson = r.replaceSensitiveData(argumentsInJson, sensitiveData)
		log.Debug(argumentsInJson)
	}

	result, err := (*action.Tool).Invoke(ctx, argumentsInJson)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (r *Registry) replaceSensitiveData(argumentsInJson string, sensitiveData map[string]string) string {
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
	return replaceSecrets(argumentsInJson)
}

func (r *Registry) CreateActionModel(includeActions []string, page playwright.Page) *ActionModel {
	// Create model from registered actions, used by LLM APIs that support tool calling

	// Filter actions based on page if provided:
	//   if page is None, only include actions with no filters
	//   if page is provided, only include actions that match the page

	availableActions := make(map[string]*RegisteredAction)
	for name, action := range r.Registry.Actions {
		if includeActions != nil && !slices.Contains(includeActions, name) {
			continue
		}

		// If no page provided, only include actions with no filters
		if page == nil {
			if action.PageFilter == nil && len(action.Domains) == 0 {
				availableActions[name] = action
			}
			continue
		}

		// Check page_filter if present
		domainIsAllowed := r.Registry.matchDomains(action.Domains, page.URL())
		pageIsAllowed := r.Registry.matchPageFilter(action.PageFilter, page)

		// Include action if both filters match (or if either is not present)
		if domainIsAllowed && pageIsAllowed {
			availableActions[name] = action
		}
	}

	return &ActionModel{
		Actions: availableActions,
	}
}

func (r *Registry) GetPromptDescription(page playwright.Page) string {
	return r.Registry.GetPromptDescription(page)
}
