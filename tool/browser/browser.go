package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/playwright-community/playwright-go"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/pkg/browser"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/browser/controller"
)

type Browser struct {
	controller     *controller.Controller
	llm            *llm.Instance
	BrowserSession *browser.Session
	tool.Tool
}

type ToolParams struct {
	Thought *AgentThought          `json:"thought"`
	Actions []*controller.ActModel `json:"actions" jsonschema:"minItems=1"` // List of actions to execute
}

// Current state of the agent
type AgentThought struct {
	EvaluationPreviousGoal string `json:"evaluation_previous_goal"`
	Memory                 string `json:"memory"`
	NextGoal               string `json:"next_goal"`
}

func (b *Browser) DoAction(ctx context.Context, args string) (string, error) {
	checkForNewElements := true

	results := []*controller.ActionResult{}

	var params ToolParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	cachedSelectorMap := b.BrowserSession.GetSelectorMap()
	cachedPathHashes := mapset.NewSet[string]()
	if cachedSelectorMap != nil {
		for _, e := range *cachedSelectorMap {
			cachedPathHashes.Add(e.Hash().BranchPathHash)
		}
	}

	b.BrowserSession.RemoveHighlights()

	for i, action := range params.Actions {
		if action.GetIndex() != nil && i != 0 {
			newState := b.BrowserSession.GetState(false)
			newSelectorMap := newState.SelectorMap

			// Detect index change after previous action
			index := action.GetIndex()
			if index != nil {
				origTarget := (*cachedSelectorMap)[*index]
				var origTargetHash *string = nil
				if origTarget != nil {
					origTargetHash = playwright.String(origTarget.Hash().BranchPathHash)
				}
				newTarget := (*newSelectorMap)[*index]
				var newTargetHash *string = nil
				if newTarget != nil {
					newTargetHash = playwright.String(newTarget.Hash().BranchPathHash)
				}

				if origTargetHash == nil || newTargetHash == nil || *origTargetHash != *newTargetHash {
					msg := fmt.Sprintf("Element index changed after action %d / %d, because page changed.", i, len(params.Actions))
					log.Info(msg)
					results = append(results, &controller.ActionResult{ExtractedContent: &msg, IncludeInMemory: true})
					break
				}

				newPathHashes := mapset.NewSet[string]()
				if newSelectorMap != nil {
					for _, e := range *newSelectorMap {
						newPathHashes.Add(e.Hash().BranchPathHash)
					}
				}

				if checkForNewElements && !newPathHashes.IsSubset(cachedPathHashes) {
					msg := fmt.Sprintf("Something new appeared after action %d / %d", i, len(params.Actions))
					log.Info(msg)
					results = append(results, &controller.ActionResult{ExtractedContent: &msg, IncludeInMemory: true})
					break
				}
			}
		}
		result, err := b.controller.ExecuteAction(action, b.BrowserSession, b.llm, nil, nil)
		if err != nil {
			// TODO(LOW): implement signal handler error
			// log.Infof("Action %d was cancelled due to Ctrl+C", i+1)
			// if len(results) > 0 {
			// 	results = append(results, &controller.ActionResult{Error: playwright.String("The action was cancelled due to Ctrl+C"), IncludeInMemory: true})
			// }
			// return nil, errors.New("Action cancelled by user")
			return "", err
		}
		results = append(results, result)
		lastIndex := len(results) - 1
		if (results[lastIndex].IsDone != nil && *results[lastIndex].IsDone) || results[lastIndex].Error != nil || i == len(params.Actions)-1 {
			break
		}

		time.Sleep(500 * time.Millisecond) // ag.BrowserContext.Config.WaitBetweenActions
	}
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	result := string(resultsJSON) + "\n" + b.GetCurrentState(ctx, "")
	return result, nil
}

func (b *Browser) GetCurrentState(ctx context.Context, args string) string {
	// browser state
	browserState := b.BrowserSession.GetState(true)

	// get specific attribute clickable elements in DomTree as string
	// elementText := browserState.ElementTree.ClickableElementsToString(amp.IncludeAttributes)
	elementText := browserState.ElementTree.ClickableElementsToString(nil)

	hasContentAbove := browserState.PixelAbove > 0
	hasContentBelow := browserState.PixelBelow > 0

	if elementText != "" {
		if hasContentAbove {
			elementText = fmt.Sprintf("... %d pixels above - scroll or extract content to see more ...\n%s", browserState.PixelAbove, elementText)
		} else {
			elementText = fmt.Sprintf("[Start of page]\n%s", elementText)
		}
		// Update elementText by appending the new info to the existing value
		if hasContentBelow {
			elementText = fmt.Sprintf("%s\n... %d pixels below - scroll or extract content to see more ...", elementText, browserState.PixelBelow)
		} else {
			elementText = fmt.Sprintf("%s\n[End of page]", elementText)
		}
	} else {
		elementText = "empty page"
	}

	// var stepInfoDescription string
	// if amp.StepInfo != nil {
	// 	current := int(amp.StepInfo.StepNumber) + 1
	// 	max := int(amp.StepInfo.MaxSteps)
	// 	stepInfoDescription = fmt.Sprintf("Current step: %d/%d", current, max)
	// } else {
	// 	stepInfoDescription = ""
	// }
	timeStr := time.Now().Format("2006-01-02 15:04")
	// stepInfoDescription += fmt.Sprintf("Current date and time: %s", timeStr)
	currentDateAndTime := fmt.Sprintf("Current date and time: %s", timeStr)

	stateDescription := fmt.Sprintf(`
[Browser Current state was here]
The following is one-time information - if you need to remember it write it to memory:
Current url: %s
Available tabs:
%s
Interactive elements from top layer of the current page inside the viewport:
%s
%s`,
		browserState.Url,
		browser.TabsToString(browserState.Tabs),
		elementText,
		currentDateAndTime,
	)

	// if amp.Result != nil {
	// 	for i, result := range amp.Result {
	// 		if result.ExtractedContent != nil {
	// 			stateDescription += fmt.Sprintf("\nAction result %d/%d: %s", i+1, len(amp.Result), *result.ExtractedContent)
	// 		}
	// 		if result.Error != nil {
	// 			// only use last line of error
	// 			errStr := *result.Error
	// 			splitted := strings.Split(errStr, "\n")
	// 			lastLine := splitted[len(splitted)-1]
	// 			stateDescription += fmt.Sprintf("\nAction error %d/%d: ...%s", i+1, len(amp.Result), lastLine)
	// 		}
	// 	}
	// }

	// if amp.State.Screenshot != nil && useVision {
	// 	// Format message for vision model
	// 	return &schema.Message{
	// 		Role: schema.User,
	// 		MultiContent: []schema.ChatMessagePart{
	// 			{
	// 				Type: schema.ChatMessagePartTypeText,
	// 				Text: stateDescription,
	// 			},
	// 			{
	// 				Type: schema.ChatMessagePartTypeImageURL,
	// 				ImageURL: &schema.ChatMessageImageURL{
	// 					URL: "data:image/png;base64," + *amp.State.Screenshot,
	// 				},
	// 			},
	// 		},
	// 	}
	// }

	return stateDescription
}

func Tools(browserSession *browser.Session, supportedActions *controller.ActionModel) string {
	browserTool := &Browser{
		controller: controller.NewController(),
		// llm:            llm.NewInstance("gpt-4o-mini"),
		BrowserSession: browserSession,
	}
	if supportedActions == nil {
		supportedActions = browserTool.controller.Registry.CreateActionModel(nil, browserSession.GetCurrentPage())
	}

	actionSchemas := map[string]*openapi3.SchemaRef{}
	for _, action := range supportedActions.Actions {
		// if action.Tool == nil {
		// 	panic(fmt.Sprintf("action tool is nil: %+v", action))
		// }
		actionSchema, ok := action.Tool.Parameters().(*openapi3.Schema)
		if !ok {
			panic(fmt.Sprintf("action tool parameters is not a openapi3.Schema: %+v", action.Tool))
		}
		actionSchema.Title = action.Tool.Name
		actionSchema.Description = action.Tool.Description
		actionSchemas[action.Tool.Name] = &openapi3.SchemaRef{
			Value: actionSchema,
		}
	}

	types := openapi3.Types{openapi3.TypeObject}
	agentThoughtSchema := &openapi3.Schema{
		Type: &types,
		Properties: map[string]*openapi3.SchemaRef{
			"evaluation_previous_goal": {
				Value: &openapi3.Schema{
					Type: &openapi3.Types{openapi3.TypeString},
				},
			},
			"memory": {
				Value: &openapi3.Schema{
					Type: &openapi3.Types{openapi3.TypeString},
				},
			},
			"next_goal": {
				Value: &openapi3.Schema{
					Type: &openapi3.Types{openapi3.TypeString},
				},
			},
		},
	}
	agentThoughtSchema.Description = "Current thought of the agent"

	tl := tool.New(
		tool.WithName("BrowserUse"),
		tool.WithDescription("Use Browser to do some actions(supported actions list see actions field).\n"+prompt),
		tool.WithParameters(
			&openapi3.Schema{
				Type: &types,
				Properties: map[string]*openapi3.SchemaRef{
					"actions": {
						Value: &openapi3.Schema{
							Description: "List of actions to execute",
							Type:        &openapi3.Types{openapi3.TypeArray},
							Items: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Properties: actionSchemas,
								},
							},
						},
					},
					"thought": {
						Value: agentThoughtSchema,
					},
				},
				Required: []string{"actions", "thought"},
			},
		),
		tool.WithFunction(browserTool.DoAction),
	)

	tool.Register(tl)

	return tl.Name
}

var prompt = `
You must follow the following rules:
# Rules
2. ACTIONS: You can specify multiple actions in the list to be executed in sequence. But always specify only one action name per item. Use maximum {max_actions} actions per sequence.
Common action sequences:

- Form filling: [{"input_text": {"index": 1, "text": "username"}}, {"input_text": {"index": 2, "text": "password"}}, {"click_element_by_index": {"index": 3}}]
- Navigation and extraction: [{"go_to_url": {"url": "https://example.com"}}, {"extract_content": {"goal": "extract the names"}}]
- Actions are executed in the given order
- If the page changes after an action, the sequence is interrupted and you get the new state.
- Only provide the action sequence until an action which changes the page state significantly.
- Try to be efficient, e.g. fill forms at once, or chain actions where nothing changes on the page
- only use multiple actions if it makes sense.

3. ELEMENT INTERACTION:

- Only use indexes of the interactive elements

4. NAVIGATION & ERROR HANDLING:

- If no suitable elements exist, use other functions to complete the task
- If stuck, try alternative approaches - like going back to a previous page, new search, new tab etc.
- Handle popups/cookies by accepting or closing them
- Use scroll to find elements you are looking for
- If you want to research something, open a new tab instead of using the current tab
- If captcha pops up, try to solve it - else try a different approach
- If the page is not fully loaded, use wait action


6. VISUAL CONTEXT:

- When an image is provided, use it to understand the page layout
- Bounding boxes with labels on their top right corner correspond to element indexes

7. Form filling:

- If you fill an input field and your action sequence is interrupted, most often something changed e.g. suggestions popped up under the field.


9. Extraction:

- If your task is to find information - call extract_content on the specific pages to get and store the information.
  Your responses must be always JSON with the specified format.
`
