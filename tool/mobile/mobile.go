package android

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	appiumgo "github.com/showntop/llmack/pkg/appium"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/mobile/controller"
)

type Mobile struct {
	controller *controller.Controller
	driver     *appiumgo.WebDriver
	tool.Tool
}

func NewMobile() *Mobile {
	options := appiumgo.NewAppiumOptions().
		SetPlatformName("Android").
		SetDeviceName("emulator-5554").
		SetAutomationName("UIAutomator2")

	d, err := appiumgo.NewWebDriver("", options, nil, nil)
	if err != nil {
		panic(err)
	}

	ctrl := controller.NewController(d)
	mobileCtrl := &Mobile{
		controller: ctrl,
		driver:     d,
		// llm:            llm.NewInstance("gpt-4o-mini"),
	}
	return mobileCtrl
}

type ToolParams struct {
	Thought *AgentThought    `json:"thought"`
	Actions []map[string]any `json:"actions" jsonschema:"minItems=1"` // List of actions to execute
}

// Current state of the agent
type AgentThought struct {
	EvaluationPreviousGoal string `json:"evaluation_previous_goal"`
	Memory                 string `json:"memory"`
	CurrentGoal            string `json:"current_goal"`
}

func (b *Mobile) DoAction(ctx context.Context, args string) (string, error) {
	results := []*controller.ActionResult{}

	var params ToolParams
	if err := json.Unmarshal([]byte(args), &params); err != nil {
		return "", err
	}

	for i, action := range params.Actions {
		result, err := b.controller.ExecuteAction(ctx, action, b.driver, nil, nil, nil)
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
	result := string(resultsJSON)
	return result, nil
}

func (b *Mobile) GetCurrentScreenshot(ctx context.Context) []byte {
	// 截图
	screenshot, err := b.driver.Screenshot()
	if err != nil {
		return nil
	}
	os.WriteFile("screenshot.png", screenshot, 0644)
	return screenshot
}

func (b *Mobile) GetCurrentClickableElements(ctx context.Context) string {
	// 截图
	// elements, err := b.driver.GetPageSource()
	// if err != nil {
	// 	return ""
	// }

	// jsonBytes, err := json.Marshal(elements)
	// if err != nil {
	// 	return ""
	// }
	// return string(jsonBytes)
	return ""
}

func (b *Mobile) Tools() string {

	actionSchemas := map[string]*openapi3.SchemaRef{}
	for _, action := range b.controller.Registry().Actions {
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

	agentThoughtSchema := &openapi3.Schema{
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
	agentThoughtSchema.Description = "Current thought of the agent"

	tl := tool.New(
		tool.WithName("MobileUse"),
		tool.WithDescription("Use this tool to do some actions(supported actions list see actions field) on mobile device.\nYou can only use actions listed in the actions field."),
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
					"actions_params_thought": {
						Value: &openapi3.Schema{
							Type:        openapi3.TypeString,
							Description: "actions的参数值的计算/确定详细逻辑",
						},
					},
					"thought": {
						Value: agentThoughtSchema,
					},
				},
				Required: []string{"actions", "thought"},
			},
		),
		tool.WithFunction(b.DoAction),
	)

	tool.Register(tl)

	return tl.Name
}
