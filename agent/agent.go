package agent

import (
	"context"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/program"
)

type Agent struct {
	llm    *llm.Instance `json:"-"` // 模型
	stream bool          `json:"-"` // 是否流式输出

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
	Retries int  `json:"retries"` // 重试次数
	Stream  bool `json:"stream"`  // 是否流式输出
}

type InvokeOption func(*InvokeOptions)

func WithRetries(retries int) InvokeOption {
	return func(o *InvokeOptions) {
		o.Retries = retries
	}
}

func WithStream(stream bool) InvokeOption {
	return func(o *InvokeOptions) {
		o.Stream = stream
	}
}

func (agent *Agent) Copy() *Agent {
	newAgent := &Agent{
		ID:           agent.ID,
		Name:         agent.Name,
		Role:         agent.Role,
		Description:  agent.Description,
		Goals:        append([]string{}, agent.Goals...),
		Instructions: append([]string{}, agent.Instructions...),
		Outputs:      append([]string{}, agent.Outputs...),
		Tools:        append([]any{}, agent.Tools...),
		TeamID:       agent.TeamID,
		llm:          agent.llm,
	}
	return newAgent

}

// concurrent invoke not support
func (agent *Agent) Invoke(ctx context.Context, task string, opts ...InvokeOption) *AgentRunResponse {
	options := &InvokeOptions{
		Retries: 1,
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
			agent.invoke(ctx, task, options.Retries, true)
		}()
		return agent.response
	} else {
		agent.invoke(ctx, task, options.Retries, false)
		return agent.response
	}
}

func (agent *Agent) invoke(ctx context.Context, task string, retries int, stream bool) (*AgentRunResponse, error) {
	for range retries {
		response, err := agent.retry(ctx, task, stream)
		if err != nil {
			response.Error = err
			return agent.response, err
		}
		agent.response = response
	}
	return agent.response, nil
}

// 迭代一次
func (agent *Agent) retry(ctx context.Context, task string, stream bool) (*AgentRunResponse, error) {
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

	predictor := program.FunCall(
		program.WithLLMInstance(agent.llm),
	).WithInstruction(agentPrompt).
		WithInputs(input).
		WithTools(agent.Tools...).
		WithStream(stream).
		InvokeQuery(ctx, task)
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
