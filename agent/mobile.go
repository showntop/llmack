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
	mobileController  *adb.Controller
	adbTool           *adb.AdbTool
	TakeScreenshotURL func(ctx context.Context, adbTool *adb.AdbTool) (string, error)
}

// NewMobileAgent ...
func NewMobileAgent(name string, deviceID string, options ...Option) *MobileAgent {
	ctrl := adb.NewController(deviceID)
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
	agent.session = session

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
	var initialMessages []llm.Message
	if len(agent.session.Messages) > 0 {
		initialMessages = agent.session.Messages
	} else {
		initialMessages = agent.getInitialMessages(ctx, task)
	}
	predictor := program.FunCall(
		program.WithLLMInstance(agent.llm),
		program.WithMaxIterationNum(500),
		program.WithResetMessages(func(ctx context.Context, messages []llm.Message) []llm.Message {
			// update session messages
			agent.storage.UpdateSession(ctx, &storage.Session{
				Messages: messages,
			})
			newMessages := []llm.Message{}
			if len(messages) > 15 { // 轮次过多，summary
				// 重新组织 messags, 删除过早的 assistant 和 tool 的消息
				newMessages = append(newMessages, messages[0]) // system message
				newMessages = append(newMessages, messages[1]) // user message

				// dialog summary
				detail := ""
				for i := 2; i < len(messages)-2; i++ {
					if messages[i].Role() == llm.MessageRoleAssistant {
						detail += "assistant: " + messages[i].Content() + "\n"
					}
					if messages[i].Role() == llm.MessageRoleTool {
						detail += "tool: " + messages[i].Content() + "\n"
					}
				}
				// 总结对话
				// response, err := agent.llm.Invoke(ctx, []llm.Message{
				// 	llm.NewSystemMessage(`
				// 	You are a helpful assistant that can summarize the dialog messages.
				// 	`),
				// 	llm.NewUserTextMessage(fmt.Sprintf(`the dialog detail: \n %s`, detail)),
				// })
				// if err != nil {
				// 	return messages
				// }
				// summary := response.Result().Message.Content()
				// newMessages = append(newMessages, llm.NewUserTextMessage(fmt.Sprintf(`execute proceed summary: \n %s`, summary)))
				newMessages = append(newMessages, messages[len(messages)-6])
				newMessages = append(newMessages, messages[len(messages)-5])
				newMessages = append(newMessages, messages[len(messages)-4])
				newMessages = append(newMessages, messages[len(messages)-3])
				newMessages = append(newMessages, messages[len(messages)-2])
				newMessages = append(newMessages, messages[len(messages)-1])
			} else {
				newMessages = messages
			}

			screenshotURL, err := agent.TakeScreenshotURL(ctx, agent.adbTool)
			// screenshotURL, err := agent.adbTool.GetMobileCurrentScreenshotURL(ctx)
			// screenshot, err := agent.adbTool.GetMobileCurrentScreenshotObject(ctx)
			if err != nil {
				log.ErrorContextf(ctx, "get mobile current screenshot error: %v", err)
				// return newMessages
			}
			elements, err := agent.adbTool.GetMobileCurrentClickableElements(ctx)
			if err != nil {
				log.ErrorContextf(ctx, "get mobile current clickable elements error: %v", err)
				// return newMessages
			}
			elementsJSON, err := json.Marshal(elements)
			if err != nil {
				log.ErrorContextf(ctx, "marshal mobile current clickable elements error: %v", err)
				// return newMessages
			}
			state, err := agent.adbTool.GetMobileCurrentPhoneState(ctx)
			if err != nil {
				log.ErrorContextf(ctx, "get mobile current phone state error: %v", err)
				// return newMessages
			}
			stateJSON, err := json.Marshal(state)
			if err != nil {
				log.ErrorContextf(ctx, "marshal mobile current phone state error: %v", err)
				// return newMessages
			}
			newMessages = append(newMessages, llm.NewUserMultipartMessage(
				// llm.MultipartContentImageBase64("png", screenshot),
				// llm.MultipartContentCustom("venus_image_url", map[string]any{
				// 	"venus_image_url": screenshot,
				// }),
				llm.MultipartContentImageURL(screenshotURL),
				llm.MultipartContentText(fmt.Sprintf(`the image given above is the current screenshot of mobile with resolution 720x1280, you can use it to help you complete the task`)),
				llm.MultipartContentText(fmt.Sprintf(`the current clickable elements: \n %s`, elementsJSON)),
				llm.MultipartContentText(fmt.Sprintf(`the current phone state: \n %s`, stateJSON)),
			))
			return newMessages
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
		InvokeWithMessages(ctx, initialMessages)
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
	agent.response.Usage.PromptTokens += predictor.Usage.PromptTokens
	agent.response.Usage.CompletionTokens += predictor.Usage.CompletionTokens
	agent.response.Usage.TotalTokens += predictor.Usage.TotalTokens
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
	model := llm.New(doubao.Name, llm.WithDefaultModel("doubao-1-5-ui-tars-250428"))
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
You are GUI agent for operating mobile phones. Your goal is to choose the correct actions to complete the user's ultimate task. Think as if you are a human user operating the phone.
如果给你了当前屏幕的截图，你需要仔细审视是否符合当前目标的执行。

使用中文回答。

# action usage attention
1. 如果 tap_by_index 没有达成任务，请使用 tap_by_coordinates 来完成任务。
2. 用 tap_by_coordinates 来进行区域筛选。
`
)
