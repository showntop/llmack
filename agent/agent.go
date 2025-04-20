package agent

import (
	"context"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/program"
)

type Agent struct {
	llm *llm.Instance `json:"-"` // 模型

	// session
	SessionID string `json:"session_id"` // 会话ID, for 持久化信息
	// context

	// memory

	// basic
	ID   string `json:"id"` // 唯一标识
	Name string `json:"name"`
	Role string `json:"role"` // 角色

	// for prompt
	Description  string   `json:"description"`
	Goals        []string `json:"goals"`
	Instructions []string `json:"instructions"`
	Outputs      []string `json:"outputs"`
	Tools        []any    `json:"tools"`

	// for team
	TeamID string `json:"team_id"` // 团队ID

	response *AgentRunResponse
}

func NewAgent(name string, options ...Option) *Agent {
	agent := &Agent{
		Name: name,
	}
	for _, option := range options {
		option(agent)
	}
	if agent.llm == nil {
		agent.llm = llm.NewInstance(deepseek.Name, llm.WithDefaultModel("deepseek-v3")) // default model
	}
	return agent
}

type InvokeOptions struct {
	Retries int `json:"retries"` // 重试次数
}

type InvokeOption func(*InvokeOptions)

func WithRetries(retries int) InvokeOption {
	return func(o *InvokeOptions) {
		o.Retries = retries
	}
}

func (agent *Agent) Invoke(ctx context.Context, task string, opts ...InvokeOption) *AgentRunResponse {
	options := &InvokeOptions{
		Retries: 1,
	}
	for _, opt := range opts {
		opt(options)
	}

	for range options.Retries {
		response, err := agent.invoke(ctx, task)
		if err != nil {
			response.Error = err
			return response
		}
		agent.response = response
	}

	return agent.response
}

// 执行 agent 返回 stream
func (agent *Agent) invoke(ctx context.Context, task string) (*AgentRunResponse, error) {
	var response *AgentRunResponse = &AgentRunResponse{}
	go func() {
		input := map[string]any{}
		if agent.Name != "" {
			input["name"] = "<name>\n" + agent.Name + "\n</name>"
		}
		if agent.Role != "" {
			input["role"] = "<role>\n" + agent.Role + "\n</role>"
		}
		if agent.Description != "" {
			input["description"] = "<description>\n" + agent.Description + "\n</description>"
		}
		if len(agent.Instructions) > 0 {
			input["instructions"] = "<instructions>\n" + strings.Join(agent.Instructions, "\n") + "\n</instructions>"
		}
		if len(agent.Goals) > 0 {
			input["goals"] = "<goals>\n" + strings.Join(agent.Goals, "\n") + "\n</goals>"
		}
		if len(agent.Outputs) > 0 {
			input["outputs"] = "<outputs>\n" + strings.Join(agent.Outputs, "\n") + "\n</outputs>"
		}
		// Steps:
		// 1. Prepare the Agent for the run

		// 2. Update the Model and resolve context
		// 3. Read existing session from storage
		// 4. Prepare run messages
		// 5. Reason about the task if reasoning is enabled
		// 6. Start the Run by yielding a RunStarted event
		// 7. Generate a response from the Model (includes running function calls)
		predictor := program.FunCall(
			program.WithLLMInstance(agent.llm),
		).WithInstruction(agentPrompt).
			WithInputs(input).
			WithTools(agent.Tools...).
			InvokeQuery(ctx, task)

		for chunk := range predictor.Stream() {
			response.Answer += chunk.Choices[0].Message.Content()
			response.streamx <- chunk
			response.Stream <- chunk
		}
	}()

	// 8. Update RunResponse
	// 9. Update Agent Memory
	// 10. Calculate session metrics
	// 11. Save session to storage
	// 12. Save output to file if save_response_to_file is set
	// if len(result.ToolCalls()) > 0 {
	// 	toolResults, err := agent.invokeTools(ctx, result.ToolCalls())
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// }

	return response, nil
}

var (
	agentPrompt = `
{{role}}

{{description}}

{{goals}}

{{instructions}}

{{outputs}}

{{tools}}
	`
)
