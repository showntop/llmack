package think

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/tool"
)

const Think = "ThinkTool"

func init() {
	t := tool.New(
		tool.WithName(Think),
		tool.WithKind("code"),
		tool.WithDescription("Intelligent problem-solving assistant that comprehends tasks, identifies key variables, and makes efficient decisions, all while providing detailed, self-driven reasoning for its choices. Do not assume anything, take the details from given data only."),
		tool.WithParameters(tool.Parameter{
			Name:          "task",
			Type:          tool.String,
			Required:      true,
			LLMDescrition: "Task description which needs reasoning.",
			Default:       "",
		}),
		tool.WithFunction(func(ctx context.Context, args string) (string, error) {
			var params struct {
				Task string `json:"task"`
			}
			if err := json.Unmarshal([]byte(args), &params); err != nil {
				return "", err
			}
			// TODO: Implement actual thinking logic
			return "任务分析：" + params.Task + "。基于现有信息进行推理分析...", nil
		}),
	)
	tool.Register(t)
}

var prompt = `
Given the following overall objective
Objective:
{goals}

and the following task, {{task_description}}.

Below is last tool response:
{{last_tool_response}}

Below is the relevant tool response:
{{relevant_tool_response}}

Perform the task by understanding the problem, extracting variables, and being smart
and efficient. Provide a descriptive response, make decisions yourself when
confronted with choices and provide reasoning for ideas / decisions.
`
