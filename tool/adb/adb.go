package adb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/showntop/llmack/tool"
)

type AdbTool struct {
	controller *Controller
}

func NewAdbTool(ctrl *Controller) *AdbTool {
	return &AdbTool{controller: ctrl}
}

type ToolParams struct {
	Thought *AgentThought    `json:"thought"`
	Actions []map[string]any `json:"actions" jsonschema:"minItems=1"` // List of actions to execute
}

// Current state of the agent
type AgentThought struct {
	EvaluationPreviousGoal string `json:"evaluation_previous_goal"`
	Memory                 string `json:"memory"`
	NextGoal               string `json:"next_goal"`
}

func (t *AdbTool) GetMobileCurrentScreenshot(ctx context.Context) ([]byte, error) {
	localPath, err := t.controller.TakeScreenshot(ctx, TakeScreenshotParams{
		Quality: 100,
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("\n\nscreenshot path: ", localPath)
	return os.ReadFile(localPath)
}

func (t *AdbTool) GetMobileCurrentScreenshotObject(ctx context.Context) (any, error) {
	localPath, err := t.controller.TakeScreenshot(ctx, TakeScreenshotParams{
		Quality: 100,
	})
	if err != nil {
		return nil, err
	}

	resp, err := http.Post("http://v2.open.venus.oa.com/chat/temp/storage/info", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"model": "claude-3-7-sonnet-20250219",
		"filename": "%s"
	}`, localPath))))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Println("screenshot cos response: ", string(body))
	var respData struct {
		Data struct {
			BucketName  string `json:"bucketName"`
			Endpoint    string `json:"endpoint"`
			AccessKey   string `json:"accessKey"`
			SecretKey   string `json:"secretKey"`
			Token       string `json:"token"`
			Region      string `json:"region"`
			ExpiredTime string `json:"expiredTime"`
			Filepath    string `json:"filepath"`
			PutURL      string `json:"putUrl"`
			GetURL      string `json:"getUrl"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return nil, err
	}

	return map[string]any{
		"bucketName": respData.Data.BucketName,
		"filepath":   respData.Data.Filepath,
	}, nil
}

func (t *AdbTool) GetMobileCurrentClickableElements(ctx context.Context) ([]UIElement, error) {
	elements, err := t.controller.GetClickableElements(ctx, GetClickableElementsParams{
		Serial: t.controller.Serial,
	})
	if err != nil {
		return nil, err
	}
	return elements, nil
}

func (t *AdbTool) GetMobileCurrentPhoneState(ctx context.Context) (*PhoneState, error) {
	state, err := t.controller.GetPhoneState(ctx, GetPhoneStateParams{
		Serial: t.controller.Serial,
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (t *AdbTool) DoActions(ctx context.Context, args string) (string, error) {
	var params ToolParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}
	results := []string{}
	for _, action := range params.Actions {
		for name, params := range action {
			rawParams, _ := json.Marshal(params)
			result, err := registry.Tools[name].Invoke(ctx, string(rawParams))
			if err != nil {
				return "", err
			}
			time.Sleep(5 * time.Second) // sleep for wait completed.
			result = "\n<" + name + "_result>\n" + result + "\n<" + name + "_result/>\n"
			results = append(results, result)
		}
	}
	return strings.Join(results, "\n\n"), nil
}

func (t *AdbTool) AtomicActions() map[string]*tool.Tool {
	return registry.Tools
}

func (t *AdbTool) NewTools() []any {
	// return registry.AvailableTools(nil)
	// 打包成一个 tool(使用思考模式)，原始的 tool 作为action参数
	actionSchemas := map[string]*openapi3.SchemaRef{}
	for _, subtool := range registry.Tools {
		actionSchema, ok := subtool.Parameters().(*openapi3.Schema)
		if !ok {
			panic(fmt.Sprintf("action tool parameters is not a openapi3.Schema: %+v", subtool))
		}
		actionSchema.Title = subtool.Name
		actionSchema.Description = subtool.Description
		actionSchemas[subtool.Name] = &openapi3.SchemaRef{
			Value: actionSchema,
		}
	}

	thoughtSchema := &openapi3.Schema{ // 包含上一次目标的执行评估，记忆，和下一次的目标
		Type: openapi3.TypeObject,
		Properties: map[string]*openapi3.SchemaRef{
			"evaluation_previous_goal": {
				Value: &openapi3.Schema{
					Type:        openapi3.TypeString,
					Description: "对于上一次目标的评估",
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
		Description: "Current thought of the agent",
	}

	paramsSchema := &openapi3.Schema{
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
			"thought": {
				Value: thoughtSchema,
			},
		},
		Required: []string{"actions", "thought"},
	}

	tl := tool.New(
		tool.WithName("MobileUse"),
		tool.WithDescription("Use Mobile to do some actions(supported actions list see actions field).\n Detail instructions:\n"+prompt),
		tool.WithParameters(paramsSchema),
		tool.WithFunction(t.DoActions),
	)

	tool.Register(tl)

	return []any{tl.Name}
}

var prompt = `

# Response Rules

1. RESPONSE FORMAT: You must ALWAYS respond with valid JSON in this exact format:
   {"thought": {"evaluation_previous_goal": "Success|Failed|Unknown - Analyze the current elements and the image to check if the previous goals/actions are successful like intended by the task. Mention if something unexpected happened. Shortly state why/why not",
   "memory": "Description of what has been done and what you need to remember. Be very specific. Count here ALWAYS how many times you have done something and how many remain. E.g. 0 out of 10 websites analyzed. Continue with abc and xyz",
   "next_goal": "What needs to be done with the next immediate action"},
   "actions":[{"one_action_name": {// action-specific parameter}}, // ... more actions in sequence]}

2. ACTIONS: You can specify multiple actions in the list to be executed in sequence. But always specify only one action name per item. Use maximum 3 actions per sequence.
Common action sequences:
- Form filling: [{"input_text": {"index": 1, "text": "username"}}, {"input_text": {"index": 2, "text": "password"}}, {"tap_by_index": {"index": 3}}]
- Navigation and extraction: [{"go_to_url": {"url": "https://example.com"}}, {"extract_content": {"goal": "extract the names"}}]
- Actions are executed in the given order
- If the page changes after an action, the sequence is interrupted and you get the new state.
- Only provide the action sequence until an action which changes the page state significantly.
- Try to be efficient, e.g. fill forms at once, or chain actions where nothing changes on the page
- only use multiple actions if it makes sense.

3. ELEMENT INTERACTION:
- Only use indexes of the interactive elements
- You should always use tap_by_index to click an element, except you are sure the element is not in the clickable elements list and you can use tap_by_coordinates to click it.
- When use tap_by_index, make sure the index is the index field of the element in the clickable elements list

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
