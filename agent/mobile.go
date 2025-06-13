package agent

import (
	"context"
	"encoding/json"
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
	adbTool          *adb.AdbTool
}

// NewMobileAgent ...
func NewMobileAgent(name string, options ...Option) *MobileAgent {
	ctrl := adb.NewController("emulator-5554")
	agent := &MobileAgent{
		Agent:            *NewAgent(name, options...),
		mobileController: ctrl,
		adbTool:          adb.NewAdbTool(ctrl),
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
	mobileToolName := agent.adbTool.NewTools()
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
			screenshot, err := agent.adbTool.GetMobileCurrentScreenshot(ctx)
			if err != nil {
				return messages
			}
			elements, err := agent.adbTool.GetMobileCurrentClickableElements(ctx)
			if err != nil {
				return messages
			}
			elementsJSON, err := json.Marshal(elements)
			if err != nil {
				return messages
			}
			state, err := agent.adbTool.GetMobileCurrentPhoneState(ctx)
			if err != nil {
				return messages
			}
			stateJSON, err := json.Marshal(state)
			if err != nil {
				return messages
			}
			messages = append(messages, llm.NewUserMultipartMessage(
				llm.MultipartContentImageBase64("png", screenshot),
				llm.MultipartContentText(fmt.Sprintf(`the current clickable elements: \n %s`, elementsJSON)),
				llm.MultipartContentText(fmt.Sprintf(`the current phone state: \n %s`, stateJSON)),
			))
			return messages
		}),
	).WithInstruction(prompt).
		// WithInputs(input).
		WithTools(tools...).
		WithStream(stream).
		// WithToolChoice("auto").
		WithToolChoice(map[string]any{
			"type": "function",
			"function": map[string]any{
				"name": mobileToolName,
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
You are a GUI agent that can use mobile device to automate tasks.
	`
)
