package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/memory"
	"github.com/showntop/llmack/pkg/browser"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/storage"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/tool/browser/controller"
)

type BrowserAgent struct {
	Agent
	controller *controller.Controller

	BrowserContext *browser.BrowserContext
	Browser        *browser.Browser
}

func NewBrowserAgent(name string, options ...Option) *BrowserAgent {
	return &BrowserAgent{
		Agent: *NewAgent(name, options...),
	}
}

// concurrent invoke not support
func (agent *BrowserAgent) Invoke(ctx context.Context, task string, opts ...InvokeOption) *AgentRunResponse {
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
	} else {
		agent.invoke(ctx, task, options)
		return agent.response
	}
}

func (agent *BrowserAgent) invoke(ctx context.Context, task string, options *InvokeOptions) (*AgentRunResponse, error) {
	// fetch or create a new session
	session, err := agent.fetchOrCreateSession(ctx, options.SessionID)
	if err != nil {
		agent.response.Error = err
		return agent.response, err
	}

	agent.SessionID = session.ID

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
func (agent *BrowserAgent) retry(ctx context.Context, task string, stream bool) (*AgentRunResponse, error) {
	// browser state
	agent.BrowserContext.GetState(true)
	currentPage := agent.BrowserContext.GetCurrentPage()
	if currentPage == nil {
		return nil, errors.New("no active page")
	}
	// tool
	actionModel := agent.controller.Registry.CreateActionModel(nil, currentPage)
	if actionModel == nil {
		return nil, errors.New("no action model")
	}
	//
	return nil, nil
}

func (agent *BrowserAgent) execActionTool(ctx context.Context, actionModel *controller.ActionModel) string {
	fun := func(ctx context.Context, args string) (string, error) {
		return "agent.response.Answer", nil
	}
	tl := tool.New(
		tool.WithName("AgentOutput"),
		tool.WithDescription("AgentOutput model with custom actions."),
		tool.WithParameters(nil),
		tool.WithFunction(fun),
	)

	tool.Register(tl)

	return tl.Name
}

func (agent *BrowserAgent) retry2(ctx context.Context, task string, stream bool) (*AgentRunResponse, error) {
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

	if agent.memory != nil {
		items, err := agent.memory.Get(ctx, agent.SessionID)
		if err != nil {
			agent.response.Error = err
			return agent.response, err
		}
		if len(items) > 0 {
			agentPrompt += "You have access to memories from previous interactions with the user that you can use:\n\n"
			agentPrompt += "<memories_from_previous_interactions>"
			for _, item := range items {
				agentPrompt += "\n- " + item.Content
			}
			agentPrompt += "\n</memories_from_previous_interactions>\n\n"
			agentPrompt += "Note: this information is from previous interactions and may be updated in this conversation. "
			agentPrompt += "You should always prefer information from this conversation over the past memories.\n"
		} else {
			agentPrompt += "You have the capability to retain memories from previous interactions with the user, "
			agentPrompt += "but have not had any interactions with the user yet.\n"
		}
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

func (agent *BrowserAgent) fetchOrCreateSession(ctx context.Context, sessionID string) (*storage.Session, error) {

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
	browserAgentPrompt = `
{{role}}

{{description}}

{{goals}}

{{instructions}}

{{outputs}}

{{tools}}
	`
)
