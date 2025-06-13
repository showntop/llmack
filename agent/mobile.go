package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/doubao"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/memory"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/storage"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/adb"
)

type MobileAgent struct {
	Agent
	// controller *controller.Controller
	mobileController *adb.Controller
}

// NewMobileAgent ...
func NewMobileAgent(name string, options ...Option) *MobileAgent {
	agent := &MobileAgent{
		Agent:            *NewAgent(name, options...),
		mobileController: adb.NewController("emulator-5554"),
	}
	for _, option := range options { // TODO: 避免重新赋值
		option(agent)
	}

	return agent
}

// Invoke concurrent invoke not support
func (agent *MobileAgent) Invoke(ctx context.Context, task string, opts ...InvokeOption) *AgentRunResponse {
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

func (agent *MobileAgent) invoke(ctx context.Context, task string, options *InvokeOptions) (*AgentRunResponse, error) {
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
func (agent *MobileAgent) retry(ctx context.Context, task string, stream bool) (*AgentRunResponse, error) {

	var tools []any
	for _, tool := range agent.Tools {
		tools = append(tools, tool)
	}
	// tools = append(tools, agent.execActionTool(ctx, actionModel))
	mobileToolName := adb.NewTools()
	tools = append(tools, mobileToolName...)
	// tools = append(tools, agent.getMobileState())

	prompt := ""
	if agent.Name != "" {
		prompt = strings.Replace(prompt, "{name}", agent.Name, 1)
	}
	if agent.Role != "" {
		prompt = strings.Replace(prompt, "{role}", agent.Role, 1)
	}
	// prompt += "You are designed to use mobile device to automate tasks.\n"
	// prompt += "Your goal is to accomplish the ultimate task following the rules.\n"
	prompt += androidAgentInstruction
	predictor := program.FunCall(
		program.WithLLMInstance(agent.llm),
		program.WithMaxIterationNum(50),
		program.WithResetMessages(func(ctx context.Context, messages []llm.Message) []llm.Message {
			// messages = append(messages, llm.NewUserMultipartMessage(
			// 	// llm.MultipartContentImageBase64("png", agent.mobile.GetCurrentScreenshot(ctx)),
			// 	llm.MultipartContentText(fmt.Sprintf(`{
			// 		"message": "the image given is a screenshot of the current screen with resolution %d x %d, you can use to help you complete the task",
			// 	}`, 1080, 2400)),
			// ))
			return messages
		}),
	).WithInstruction(prompt).
		// WithInputs(input).
		WithTools(tools...).
		WithStream(stream).
		WithToolChoice("auto").
		// WithToolChoice(map[string]any{
		// 	"type": "function",
		// 	"function": map[string]any{
		// 		"name": mobileToolName,
		// 	},
		// }).
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

func (agent *MobileAgent) getInitialMessages(ctx context.Context, task string) []llm.Message {

	messages := []llm.Message{llm.NewUserTextMessage(strings.Replace(userTaskPrompt, "{{task}}", task, 1))}

	// messages = append(messages, llm.NewAssistantMessage("nothing"))
	// messages = append(messages, llm.NewUserTextMessage("Example output: "))
	// args := adb.ToolParams{
	// 	Thought: &adb.AgentThought{
	// 		EvaluationPreviousGoal: `Success - I successfully launched the mobile phone.`,
	// 		Memory:                 `I successfully launched the mobile phone.`,
	// 		CurrentGoal:            `screenshot the current screen.`,
	// 	},
	// 	Actions: []map[string]any{
	// 		{
	// 			"name": "launch_mobile",
	// 		},
	// 	},
	// }
	// argsBytes, err := json.Marshal(args)
	// if err != nil {
	// 	panic(err)
	// }
	// exampleToolCallMessage := llm.NewAssistantMessage("I should first launch the mobile phone").WithToolCalls([]*llm.ToolCall{
	// 	{
	// 		ID:       "0001",
	// 		Type:     "function",
	// 		Function: llm.ToolCallFunction{Name: "MobileUse", Arguments: string(argsBytes)},
	// 	},
	// })
	// messages = append(messages, exampleToolCallMessage)

	// messages = append(messages, llm.NewToolMessage(`I successfully launched the mobile phone.`, "0001"))
	// messages = append(messages, llm.NewToolMessage(fmt.Sprintf(`{
	// 	"message": "I successfully launched the mobile phone.",
	// 	"current_state": %q,
	// 	"next_goal": %q
	// }`, agent.mobile.GetCurrentState(ctx), task), "0001"))
	// messages = append(messages, llm.NewUserMultipartMessage(
	// 	// llm.MultipartContentImageBase64("png", agent.mobile.GetCurrentScreenshot(ctx)),
	// 	// llm.MultipartContentCustom("png", agent.mobile.GetCurrentScreenshot(ctx)),
	// 	llm.MultipartContentText(fmt.Sprintf(`{
	// 		"message": "the image given is a screenshot of the current screen with resolution %d x %d, you can use to help you complete the task",
	// 	}`, 1080, 2400)),
	// 	// llm.MultipartContentText(fmt.Sprintf(`the current active element: \n %s`,
	// 	// 	agent.mobile.GetCurrentClickableElements(ctx))),
	// ))

	// messages = append(messages, llm.NewUserTextMessage("[Your task history memory starts here]"))
	return messages
}

func (agent *MobileAgent) getLayoutAndCoordinates() string {
	model := llm.NewInstance(doubao.Name, llm.WithDefaultModel("doubao-1-5-ui-tars-250428"))
	function := func(ctx context.Context, args string) (string, error) {
		// screenshot := agent.mobile.GetCurrentScreenshot(ctx)
		// TODO: get layout and coordinates
		response, err := model.Invoke(ctx, []llm.Message{
			llm.NewSystemMessage(`
You are a GUI agent that can get the layout and coordinates of an element on the screenshot by the task given to you.
You should more precisely with pixel-level analysis.
Answer in chinese.
			`),

			// llm.NewUserTextMessage(fmt.Sprintf("the image given is a screenshot of the current screen with resolution %d x %d, you can use to help you complete the task", 1080, 2400)),
			llm.NewUserMultipartMessage(
				// llm.MultipartContentImageBase64("png", screenshot),
				llm.MultipartContentText(fmt.Sprintf(`{
					"message": "the image given is a screenshot of the current screen with resolution %d x %d, you can use to help you complete the task",
				}`, 1080, 2400)),
				llm.MultipartContentText(fmt.Sprintf(`Your task is: %s`, args)),
			),
		}, llm.WithStream(true))
		if err != nil {
			return "", err
		}
		result := response.Result()
		return "reasoning: " + result.Message.ReasoningContent + "\n" + "answer: " + result.Message.Content(), nil
	}
	toolx := tool.New(
		tool.WithName("GetLayoutAndCoordinates"),
		tool.WithDescription("Get the layout and coordinates of an element on the mobile."),
		tool.WithFunction(function),
		tool.WithParameters(tool.Parameter{
			Name:          "task",
			Type:          tool.String,
			LLMDescrition: "The task to complete",
			Required:      true,
		}),
	)
	tool.Register(toolx)
	return toolx.Name
}

func (agent *MobileAgent) getCurrentMobileState() string {
	toolx := tool.New(
		tool.WithName("GetMobileCurrentState"),
		tool.WithDescription("Get the current state of the mobile."),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			// screenshot, err := agent.(ctx, "")
			// if err != nil {
			// 	return "", err
			// }
			// return screenshot, nil
			return "", nil
		}),
	)
	tool.Register(toolx)
	return toolx.Name
}

func (agent *MobileAgent) fetchOrCreateSession(ctx context.Context, sessionID string) (*storage.Session, error) {

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
	androidAgentInstruction = `
# Response Rules

1. RESPONSE FORMAT: You must ALWAYS respond with valid JSON in this exact format:
   {"thought": {"evaluation_previous_goal": "Success|Failed|Unknown - Analyze the current elements and the image to check if the previous goals/actions are successful like intended by the task. Mention if something unexpected happened. Shortly state why/why not",
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
