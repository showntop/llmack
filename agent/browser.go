package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/memory"
	"github.com/showntop/llmack/pkg/browser"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/storage"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/browser/controller"
)

type BrowserAgent struct {
	Agent
	controller *controller.Controller

	BrowserSession *browser.Session
	Browser        *browser.Browser
}

// NewBrowserAgent ...
func NewBrowserAgent(name string, options ...Option) *BrowserAgent {
	browserInstance := browser.NewBrowser(browser.NewBrowserConfig())
	return &BrowserAgent{
		Agent:          *NewAgent(name, options...),
		Browser:        browserInstance,
		controller:     controller.NewController(),
		BrowserSession: browserInstance.NewContext(),
	}
}

// Invoke concurrent invoke not support
func (agent *BrowserAgent) Invoke(ctx context.Context, task string, opts ...InvokeOption) *AgentRunResponse {
	options := &InvokeOptions{
		Retries: 1,
		Stream:  false,
	}
	for _, opt := range opts {
		opt(options)
	}
	agent.response = &AgentRunResponse{
		Stream: make(chan *llm.Chunk, 10),
	}

	if options.Stream {
		go func() {
			defer func() {
				close(agent.response.Stream)
			}()
			agent.invoke(ctx, task, options)
		}()
		return agent.response
	}
	agent.invoke(ctx, task, options)
	return agent.response
}

func (agent *BrowserAgent) invoke(ctx context.Context, task string, options *InvokeOptions) (*AgentRunResponse, error) {
	// fetch or create a new session
	session, err := agent.fetchOrCreateSession(ctx, options.SessionID)
	if err != nil {
		agent.response.Error = err
		return agent.response, err
	}

	agent.SessionID = session.ID

	defer func() { //  Update Agent Memory

		log.DebugContextf(ctx, "agent response:\n")
		log.DebugContextf(ctx, "===============================\n %s", agent.response.Answer)
		log.DebugContextf(ctx, "===============================")
		if agent.memory != nil {
			agent.memory.Add(ctx, session.ID, memory.NewMemoryItem(session.ID, task, nil))
		}
		if agent.storage != nil {
			agent.storage.UpdateSession(ctx, session)
		}
	}()

	for range options.Retries {
		response, err := agent.retry(ctx, task, options.Stream)
		if err != nil {
			response.Error = err
			return agent.response, err
		}
		agent.response = response
	}
	return agent.response, nil
}

// 迭代一次
func (agent *BrowserAgent) retry(ctx context.Context, task string, stream bool) (*AgentRunResponse, error) {

	currentPage := agent.BrowserSession.GetCurrentPage()
	if currentPage == nil {
		return nil, errors.New("no active page")
	}
	// tool
	actionModel := agent.controller.Registry.CreateActionModel(nil, currentPage)
	if actionModel == nil {
		return nil, errors.New("no action model")
	}
	//
	var tools []any
	for _, tool := range agent.Tools {
		tools = append(tools, tool)
	}
	tools = append(tools, agent.execActionTool(ctx, actionModel))
	tools = append(tools, agent.getBrowserState())

	prompt := ""
	if agent.Name != "" {
		prompt = strings.Replace(prompt, "{name}", agent.Name, 1)
	}
	if agent.Role != "" {
		prompt = strings.Replace(prompt, "{role}", agent.Role, 1)
	}
	prompt += "You are designed to use browser to automate tasks.\n"
	prompt += "Your goal is to accomplish the ultimate task following the rules.\n"
	prompt += browserAgentPrompt
	predictor := program.FunCall(
		program.WithLLMInstance(agent.llm),
	).WithInstruction(prompt).
		// WithInputs(input).
		WithTools(tools...).
		WithStream(stream).
		WithToolChoice(map[string]any{
			"type": "function",
			"function": map[string]any{
				"name": "AgentOutput",
			},
		}).
		InvokeWithMessages(ctx, agent.getInitialMessages(ctx, task))
	if predictor.Error() != nil {
		agent.response.Error = predictor.Error()
		return agent.response, predictor.Error()
	}
	if stream {
		for chunk := range predictor.Stream() {
			agent.response.Stream <- chunk
		}
	}
	agent.response.Answer = predictor.Response().Completion()
	return agent.response, nil
}

func (agent *BrowserAgent) getInitialMessages(_ context.Context, task string) []llm.Message {

	messages := []llm.Message{llm.NewUserTextMessage(strings.Replace(userTaskPrompt, "{{task}}", task, 1))}

	messages = append(messages, llm.NewAssistantMessage(""))
	messages = append(messages, llm.NewUserTextMessage("Example output: "))
	args := AgentOutput{
		CurrentState: &AgentBrain{
			EvaluationPreviousGoal: `Success - I successfully clicked on the 'Apple' link from the Google Search results page, 
				which directed me to the 'Apple' company homepage. This is a good start toward finding 
				the best place to buy a new iPhone as the Apple website often list iPhones for sale.`,
			Memory: `I searched for 'iPhone retailers' on Google. From the Google Search results page, 
				I used the 'click_element_by_index' tool to click on a element labelled 'Best Buy' but calling 
				the tool did not direct me to a new page. I then used the 'click_element_by_index' tool to click 
				on a element labelled 'Apple' which redirected me to the 'Apple' company homepage. 
				Currently at step 3/15.`,
			NextGoal: `Looking at reported structure of the current page, I can see the item '[127]<h3 iPhone/>' 
				in the content. I think this button will lead to more information and potentially prices 
				for iPhones. I'll click on the link to 'iPhone' at index [127] using the 'click_element_by_index' 
				tool and hope to see prices on the next page.`,
		},
		Actions: []*controller.ActModel{
			{
				"click_element_by_index": map[string]interface{}{
					"index": 127,
				},
			},
		},
	}
	argsBytes, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}
	exampleToolCallMessage := llm.NewAssistantMessage("").WithToolCalls([]*llm.ToolCall{
		{
			ID:       "0001",
			Type:     "tool_call",
			Function: llm.ToolCallFunction{Name: "AgentOutput", Arguments: string(argsBytes)},
		},
	})
	messages = append(messages, exampleToolCallMessage)

	messages = append(messages, llm.NewToolMessage("Browser started", "0001"))

	messages = append(messages, llm.NewUserTextMessage("[Your task history memory starts here]"))
	return messages
}

// Current state of the agent
type AgentBrain struct {
	EvaluationPreviousGoal string `json:"evaluation_previous_goal"`
	Memory                 string `json:"memory"`
	NextGoal               string `json:"next_goal"`
}

// @dev note: this model is extended with custom actions in AgentService.
// You can also use some fields that are not in this model as provided by the linter, as long as they are registered in the DynamicActions model.
type AgentOutput struct {
	CurrentState *AgentBrain            `json:"current_state"`
	Actions      []*controller.ActModel `json:"actions" jsonschema:"minItems=1"` // List of actions to execute
}

func (agent *BrowserAgent) getBrowserState() string {
	toolx := tool.New(
		tool.WithName("GetBrowserCurrentState"),
		tool.WithDescription("Get the current state of the browser."),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			// browser state
			browserState := agent.BrowserSession.GetState(true)

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

			return stateDescription, nil
		}),
	)
	tool.Register(toolx)
	return toolx.Name
}

func (agent *BrowserAgent) execActionTool(_ context.Context, customActions *controller.ActionModel) string {
	checkForNewElements := true
	fun := func(ctx context.Context, args string) (string, error) {
		results := []*controller.ActionResult{}

		var params AgentOutput
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return "", err
		}

		cachedSelectorMap := agent.BrowserSession.GetSelectorMap()
		cachedPathHashes := mapset.NewSet[string]()
		if cachedSelectorMap != nil {
			for _, e := range *cachedSelectorMap {
				cachedPathHashes.Add(e.Hash().BranchPathHash)
			}
		}

		agent.BrowserSession.RemoveHighlights()

		for i, action := range params.Actions {
			if action.GetIndex() != nil && i != 0 {
				newState := agent.BrowserSession.GetState(false)
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
			model := agent.llm
			result, err := agent.controller.ExecuteAction(action, agent.BrowserSession, model, nil, nil)
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
		return string(resultsJSON), nil
	}
	actionSchemas := map[string]*openapi3.SchemaRef{}
	for _, action := range customActions.Actions {
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
	// agentBrain, err := einoUtils.GoStruct2ParamsOneOf[AgentBrain]()
	// if err != nil {
	// 	return "", err
	// }
	// EvaluationPreviousGoal string `json:"evaluation_previous_goal"`
	// Memory                 string `json:"memory"`
	// NextGoal               string `json:"next_goal"`
	agentBrainSchema := &openapi3.Schema{
		Type: openapi3.TypeObject,
		Properties: map[string]*openapi3.SchemaRef{
			"evaluation_previous_goal": {
				Value: &openapi3.Schema{
					Type: openapi3.TypeString,
				},
			},
			"memory": {
				Value: &openapi3.Schema{
					Type: openapi3.TypeString,
				},
			},
			"next_goal": {
				Value: &openapi3.Schema{
					Type: openapi3.TypeString,
				},
			},
		},
	}
	agentBrainSchema.Description = "Current state of the agent"

	tl := tool.New(
		tool.WithName("AgentOutput"),
		tool.WithDescription("AgentOutput model with custom actions."),
		tool.WithParameters(
			&openapi3.Schema{
				Type: openapi3.TypeObject,
				Properties: map[string]*openapi3.SchemaRef{
					"actions": {
						Value: &openapi3.Schema{
							Description: "List of actions to execute",
							Type:        openapi3.TypeArray,
							Items: &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Properties: actionSchemas,
								},
							},
						},
					},
					"current_state": {
						Value: agentBrainSchema,
					},
				},
				Required: []string{"actions", "current_state"},
			},
		),
		tool.WithFunction(fun),
	)

	tool.Register(tl)

	return tl.Name
}

func (agent *BrowserAgent) fetchOrCreateSession(ctx context.Context, sessionID string) (*storage.Session, error) {

	if sessionID == "" {
		sessionID = uuid.NewString()
		session := &storage.Session{
			ID:         sessionID,
			EngineID:   agent.ID,
			EngineType: "agent" + "(" + agent.Name + ")",
			EngineData: map[string]any{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if agent.storage == nil { // no storage just in memory
			return session, nil
		}
		if err := agent.storage.SaveSession(ctx, session); err != nil {
			return nil, err
		}
		return session, nil
	}

	if agent.storage == nil { // no storage just in memory
		return &storage.Session{
			ID:         sessionID,
			EngineID:   agent.ID,
			EngineType: "agent" + "(" + agent.Name + ")",
			EngineData: map[string]any{},
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}, nil
	}

	session, err := agent.storage.FetchSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

var (
	browserAgentPrompt = `
# Input Format

Task
Previous steps
Current URL
Open Tabs
Interactive Elements
[index]<type>text</type>

- index: Numeric identifier for interaction
- type: HTML element type (button, input, etc.)
- text: Element description
  Example:
  [33]<div>User form</div>
  \t*[35]*<button aria-label='Submit form'>Submit</button>

- Only elements with numeric indexes in [] are interactive
- (stacked) indentation (with \t) is important and means that the element is a (html) child of the element above (with a lower index)
- Elements with \* are new elements that were added after the previous step (if url has not changed)

# Response Rules

1. RESPONSE FORMAT: You must ALWAYS respond with valid JSON in this exact format:
   {"current_state": {"evaluation_previous_goal": "Success|Failed|Unknown - Analyze the current elements and the image to check if the previous goals/actions are successful like intended by the task. Mention if something unexpected happened. Shortly state why/why not",
   "memory": "Description of what has been done and what you need to remember. Be very specific. Count here ALWAYS how many times you have done something and how many remain. E.g. 0 out of 10 websites analyzed. Continue with abc and xyz",
   "next_goal": "What needs to be done with the next immediate action"},
   "actions":[{"one_action_name": {// action-specific parameter}}, // ... more actions in sequence]}

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

5. TASK COMPLETION:

- Use the done action as the last action as soon as the ultimate task is complete
- Dont use "done" before you are done with everything the user asked you, except you reach the last step of max_steps.
- If you reach your last step, use the done action even if the task is not fully finished. Provide all the information you have gathered so far. If the ultimate task is completely finished set success to true. If not everything the user asked for is completed set success in done to false!
- If you have to do something repeatedly for example the task says for "each", or "for all", or "x times", count always inside "memory" how many times you have done it and how many remain. Don't stop until you have completed like the task asked you. Only call done after the last step.
- Don't hallucinate actions
- Make sure you include everything you found out for the ultimate task in the done text parameter. Do not just say you are done, but include the requested information of the task.

6. VISUAL CONTEXT:

- When an image is provided, use it to understand the page layout
- Bounding boxes with labels on their top right corner correspond to element indexes

7. Form filling:

- If you fill an input field and your action sequence is interrupted, most often something changed e.g. suggestions popped up under the field.

8. Long tasks:

- Keep track of the status and subresults in the memory.
- You are provided with procedural memory summaries that condense previous task history (every N steps). Use these summaries to maintain context about completed actions, current progress, and next steps. The summaries appear in chronological order and contain key information about navigation history, findings, errors encountered, and current state. Refer to these summaries to avoid repeating actions and to ensure consistent progress toward the task goal.

9. Extraction:

- If your task is to find information - call extract_content on the specific pages to get and store the information.
  Your responses must be always JSON with the specified format.
	`
)

var userTaskPrompt = `
Your ultimate task is: "{{task}}",If you achieved your ultimate task, stop everything and use the done action in the next step to complete the task. If not, continue as usual.

`
