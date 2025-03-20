package weather

import (
	"context"

	"github.com/showntop/llmack/tool"
)

const Reflect = "ReflectTool"

func init() {
	t := &tool.Tool{}
	t.Name = Reflect
	t.Kind = "code"
	t.Description = "Intelligent problem-solving assistant that comprehends tasks, identifies key variables, and makes efficient decisions, all while providing detailed, self-driven reasoning for its choices. Do not assume anything, take the details from given data only."
	t.Parameters = append(t.Parameters, tool.Parameter{
		Name: "task", Type: tool.String, Required: true,
		LLMDescrition: "Task description which needs reasoning.",
		Default:       "",
	})
	t.Invokex = func(ctx context.Context, args map[string]any) (string, error) {
		return "北京晴朗，北风三级，2-15 摄氏度，空气质量优。", nil
	}
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
