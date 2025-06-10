package android

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/android/controller"
)

type Android struct {
	controller *controller.Controller
	tool.Tool
}

func (b *Android) DoAction(ctx context.Context, args string) (string, error) {

	return "", nil
}

func Tools() string {
	ctrl := controller.NewController()
	androidTool := &Android{
		controller: ctrl,
		// llm:            llm.NewInstance("gpt-4o-mini"),
	}

	actionSchemas := map[string]*openapi3.SchemaRef{}
	for _, action := range ctrl.Registry().Actions {
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
		tool.WithName("AndroidUse"),
		tool.WithDescription("Use this tool to do some actions(supported actions list see actions field) on android device."),
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
					"thought": {
						Value: agentThoughtSchema,
					},
				},
				Required: []string{"actions", "thought"},
			},
		),
		tool.WithFunction(androidTool.DoAction),
	)

	tool.Register(tl)

	return tl.Name
}
