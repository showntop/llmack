package agent

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"slices"

	"github.com/google/uuid"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/storage"
)

type Agent struct {
	storage storage.Storage `json:"-"` // 存储
	ragrtv  *rag.Indexer    `json:"-"` // rag indexer
	llm     *llm.Instance   `json:"-"` // 模型
	stream  bool            `json:"-"` // 是否流式输出

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
	SessionID string `json:"session_id"` // 会话ID, for 持久化信息
	Retries   int    `json:"retries"`    // 重试次数
	Stream    bool   `json:"stream"`     // 是否流式输出
}

type InvokeOption func(*InvokeOptions)

func WithSessionID(sessionID string) InvokeOption {
	return func(o *InvokeOptions) {
		o.SessionID = sessionID
	}
}

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
		Goals:        slices.Clone(agent.Goals),
		Instructions: append([]string{}, agent.Instructions...),
		Outputs:      append([]string{}, agent.Outputs...),
		Tools:        append([]any{}, agent.Tools...),
		TeamID:       agent.TeamID,
		llm:          agent.llm,
		storage:      agent.storage,
	}
	return newAgent

}

// concurrent invoke not support
func (agent *Agent) Invoke(ctx context.Context, task string, opts ...InvokeOption) *AgentRunResponse {
	options := &InvokeOptions{
		Retries: 1,
		Stream:  true,
	}
	for _, opt := range opts {
		opt(options)
	}
	agent.response = &AgentRunResponse{
		Stream: make(chan *llm.Chunk, 10),
	}
	// fetch or create a new session
	session, err := agent.fetchOrCreateSession(ctx, options.SessionID)
	if err != nil {
		agent.response.Error = err
		return agent.response
	}
	_ = session
	if options.Stream {
		go func() {
			defer func() {
				close(agent.response.Stream)
			}()
			agent.invoke(ctx, task, options)
		}()
		return agent.response
	} else {
		agent.invoke(ctx, task, options)
		return agent.response
	}
}

func (agent *Agent) invoke(ctx context.Context, task string, options *InvokeOptions) (*AgentRunResponse, error) {

	// defer agent.storage.UpdateSession(ctx, &storage.Session{
	// 	ID:         session.ID,
	// 	EngineID:   agent.ID,
	// 	EngineType: "agent" + "(" + agent.Name + ")",
	// 	EngineData: map[string]any{
	// 		"task": task,
	// 	},
	// 	CreatedAt: time.Now(),
	// 	UpdatedAt: time.Now(),
	// })
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
func (agent *Agent) retry(ctx context.Context, task string, stream bool) (*AgentRunResponse, error) {
	input := map[string]any{}
	agentPrompt := ""
	if agent.Name != "" {
		agentPrompt += "\n<name>\n" + agent.Name + "\n</name>\n"
	}
	if agent.Role != "" {
		agentPrompt += "\n<role>\n" + agent.Role + "\n</role>\n"
	}
	if agent.Description != "" {
		agentPrompt += "\n<description>\n" + agent.Description + "\n</description>\n"
	}
	if len(agent.Instructions) > 0 {
		agentPrompt += "\n<instructions>\n" + strings.Join(agent.Instructions, "\n") + "\n</instructions>\n"
	}
	if len(agent.Goals) > 0 {
		agentPrompt += "\n<goals>\n" + strings.Join(agent.Goals, "\n") + "\n</goals>\n"
	}
	if len(agent.Outputs) > 0 {
		agentPrompt += "\n<outputs>\n" + strings.Join(agent.Outputs, "\n") + "\n</outputs>\n"
	}

	if agent.ragrtv != nil {
		knowledges, err := agent.ragrtv.Retrieve(ctx, task, rag.WithTopK(10))
		if err != nil {
			agent.response.Error = err
			return agent.response, err
		}
		if len(knowledges) > 0 {
			jsonKnowledges, err := json.Marshal(knowledges)
			if err != nil {
				agent.response.Error = err
				return agent.response, err
			}
			task += "\n\nReference the following knowledges from the knowledge base if it helps:\n"
			task += "<knowledges>\n" + string(jsonKnowledges) + "\n</knowledges>\n"
		}
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

func (agent *Agent) fetchOrCreateSession(ctx context.Context, sessionID string) (*storage.Session, error) {
	if agent.storage == nil { // no storage just in memory
		return nil, nil
	}
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
		if err := agent.storage.SaveSession(ctx, session); err != nil {
			return nil, err
		}
		return session, nil
	}
	session, err := agent.storage.FetchSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
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
