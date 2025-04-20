package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/tool"
)

type TeamMode string

const (
	TeamModeRoute       TeamMode = "route"       // 路由模式, 路由模式下，团队leader 会根据用户请求，选择合适的 agent 进行处理
	TeamModeCoordinate  TeamMode = "coordinate"  // 协调模式, 协调模式下，团队leader 会根据用户请求，拆分任务给各个 agent 进行处理，最后综合给出答案
	TeamModeCollaborate TeamMode = "collaborate" // 协作模式, 协作模式下，团队leader 会根据用户请求，
)

// Team A team of agents
type Team struct {
	Agent
	mode     TeamMode
	members  []*Agent
	response *TeamRunResponse
}

func NewTeam(mode TeamMode, opts ...Option) *Team {
	team := &Team{
		mode:     mode,
		response: &TeamRunResponse{},
	}
	for _, opt := range opts {
		opt(team)
	}
	if team.llm == nil {
		team.llm = llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-v3")) // default model
	}
	return team
}

func (t *Team) Run(ctx context.Context, query string) (string, error) {
	prog := program.FunCall(program.WithLLMInstance(t.llm))
	if t.mode == TeamModeRoute {
		prog.WithInstruction(routePrompt).WithTools(t.distributeTask())
	} else if t.mode == TeamModeCoordinate {
		prog.WithInstruction(coordinatePrompt).WithTools(t.assignTask())
	} else if t.mode == TeamModeCollaborate {
		prog.WithInstruction(collaboratePrompt)
	}

	predictor := prog.WithInputs(map[string]any{
		"name":         t.renderName(),
		"description":  "<description>\n" + t.Description + "\n</description>",
		"instructions": "<instructions>\n" + strings.Join(t.Instructions, "\n") + "\n</instructions>",
		"agents":       t.renderAgents(t.members),
	}).InvokeQuery(ctx, query)
	if predictor.Error() != nil {
		return "", predictor.Error()
	}
	// log result
	// for _, toolCall := range result.ToolCalls() {
	// 	log.InfoContextf(ctx, "team result tool call:\n - id: %s\n - name: %s\n - args: %s", toolCall.ID, toolCall.Function.Name, toolCall.Function.Arguments)
	// }

	// toolCalls := result.ToolCalls()
	// if len(toolCalls) > 0 {
	// 	results, err := t.invokeTools(ctx, toolCalls)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	return results[toolCalls[0].Function.Name], nil
	// }
	return predictor.Completion(), nil
}

func (t *Team) assignTask() string {
	fun := func(ctx context.Context, args map[string]any) (string, error) {
		// find the agent
		memberID := args["member_name"].(string)
		task := args["task"].(string)
		expectedOutput := args["expected_output"].(string)

		var agent *Agent
		for _, a := range t.members {
			if a.Name == memberID {
				agent = a
				break
			}
		}
		if agent == nil {
			return "", fmt.Errorf("agent not found")
		}

		taskInstruction := "You are a member of a team of agents. Your goal is to complete the following task:"
		taskInstruction += "\n\n<task>\n" + task + "\n</task>"
		if expectedOutput != "" {
			taskInstruction += "\n\n<expected_output>\n" + expectedOutput + "\n</expected_output>"
		}
		// run the agent
		response := agent.Invoke(ctx, taskInstruction)
		if response.Error != nil {
			return "", response.Error
		}
		t.response.AddMemberResponse(response)

		return response.Completion(), nil
	}
	tl := &tool.Tool{}
	tl.Name = "assign_task_to_member"
	tl.Description = "Use this function to transfer a task to the selected team member.\nYou must provide a clear and concise description of the task the member should achieve AND the expected output."
	tl.Parameters = append(tl.Parameters, tool.Parameter{
		Name:          "member_name",
		Type:          "string",
		LLMDescrition: "The name of the member agent who will be assigned the task.",
		Required:      true,
	}, tool.Parameter{
		Name:          "task",
		Type:          "string",
		LLMDescrition: "A clear and concise description of the task the member agent should achieve.",
		Required:      true,
	}, tool.Parameter{
		Name:          "expected_output",
		Type:          "string",
		LLMDescrition: "The expected output from the member agent (optional).",
		Required:      true,
	})
	tl.Invokex = fun
	tool.Register(tl)

	return tl.Name
}

func (t *Team) distributeTask() string {
	fun := func(ctx context.Context, args map[string]any) (string, error) {
		// find the agent
		agentName := args["agent"].(string)
		expectedOutput := args["expected_output"].(string)

		var agent *Agent
		for _, a := range t.members {
			if a.Name == agentName {
				agent = a
				break
			}
		}
		if agent == nil {
			return "", fmt.Errorf("agent not found")
		}
		// run the agent
		response := agent.Invoke(ctx, expectedOutput)
		if response.Error != nil {
			return "", response.Error
		}
		t.response.AddMemberResponse(response)
		return "", nil
	}
	tl := &tool.Tool{}
	tl.Name = "distribute_task"
	tl.Kind = "code"
	tl.Description = "Use this function to forward the request to the nominated agent."
	tl.Parameters = append(tl.Parameters, tool.Parameter{
		Name:          "agent",
		Type:          "string",
		LLMDescrition: "The name of the agent to transfer the task to.",
		Required:      true,
	}, tool.Parameter{
		Name:          "expected_output",
		Type:          "string",
		LLMDescrition: "The expected output from the agent.",
		Required:      true,
	})
	tl.Invokex = fun
	tool.Register(tl)

	return tl.Name
}

func (t *Team) renderName() string {
	if t.Name == "" {
		return ""
	}
	return "Your Name is " + t.Name
}

func (t *Team) renderAgents(members []*Agent) string {
	builder := strings.Builder{}
	builder.WriteString("<team_members>\n")
	for idx, member := range members {
		builder.WriteString(fmt.Sprintf("- Agent %d:\n", idx))
		builder.WriteString(fmt.Sprintf("\t- Name: %s\n", member.Name))
		builder.WriteString(fmt.Sprintf("\t- Description: %s\n", member.Description))
		builder.WriteString("\t- Available Tools: \n")
		for _, tool := range member.Tools {
			builder.WriteString(fmt.Sprintf("\t\t- %s\n", tool))
		}
	}
	builder.WriteString("</team_members>\n")
	return builder.String()
}

var coordinatePrompt = `
You are the leader of a team and sub-teams of AI Agents.

Your task is to coordinate the team to complete the user's request.

Here are the agents in your team:
{{agents}}

<how_to_respond>
- You can either respond directly or assign tasks to other Agents in your team depending on the tools available to them and their roles.
- If you assign a task to another Agent, make sure to include:
  - member_name (str): The name of the Agent to assign the task to.
  - task (str): A clear description of the task.
  - expected_output (str): The expected output.
- You can pass tasks to multiple members at once.
- You must always validate the output of the other Agents before responding to the user.
- Evaluate the response from other agents. If you feel the task has been completed, you can stop and respond to the user.
- You can re-assign the task if you are not satisfied with the result.
</how_to_respond>

{{name}}

{{description}}

{{instructions}}
`

var collaboratePrompt = `
You are the leader of a team and sub-teams of AI Agents.

Your task is to coordinate the team to complete the user's request.

Here are the agents in your team:
{{agents}}

<how_to_respond>
- Only call run_member_agent once for all agents in the team.
- Take all the responses from the other Agents into account and evaluate whether the task has been completed.
- If you feel the task has been completed, you can stop and respond to the user.
</how_to_respond>

{{name}}

{{description}}
`

var routePrompt = `
You are the leader of a team and sub-teams of AI Agents.

Your task is to coordinate the team to complete the user's request.

Here are the agents in your team:
{{agents}}

<how_to_respond>
- You act as a router for the user's request. You have to choose the correct agent(s) to forward the user's request to. This should be the agent that has the highest likelihood of completing the task.
- When you forward a task to another Agent, make sure to include:
	- agent_name (str): The name of the Agent to transfer the task to.
	- expected_output (str): The expected output.
- You should do your best to forward the task to a single agent.
- If the user request requires it (e.g. if they are asking for multiple things), you can forward to multiple agents at once.
</how_to_respond>

{{name}}

{{description}}

{{instructions}}
`
