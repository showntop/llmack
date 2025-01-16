package engine

import (
	"context"

	"github.com/google/uuid"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/prompt"
)

// Input ...
type Input struct {
	Inputs    map[string]any `json:"inputs"`
	Query     string         `json:"query"`
	TurnID    string         `json:"turn_id"`
	SceneID   string         `json:"scene_id"`
	Turnround int            `json:"turnround"`
	ChunkID   string         `json:"chunk_id"`
}

// Engine ...
type Engine interface {
	Invoke(ctx context.Context, input Input) (any, error)
	Stream(ctx context.Context, input Input) *EventStream
}

// BotEngine ...
type BotEngine struct {
	context   context.Context
	SessionID string
	MessageID int64

	HistoryMessages []llm.Message

	opts     *Options
	Settings *Settings
	Hooks    []Hook
}

// HookOnStart ...
func (r *BotEngine) HookOnStart(ctx context.Context) context.Context {
	for _, h := range r.Hooks {
		ctx = h.OnStart(ctx)
	}
	return ctx
}

// HookOnFinish ...
func (r *BotEngine) HookOnFinish(ctx context.Context, err error) {
	for _, h := range r.Hooks {
		h.OnFinish(ctx, err)
	}
}

// BeforeLLMStart ...
func (r *BotEngine) BeforeLLMStart(ctx context.Context) context.Context {
	for _, h := range r.Hooks {
		ctx = h.BeforeLLMStart(ctx)
	}
	return ctx
}

// AfterLLMFinish ...
func (r *BotEngine) AfterLLMFinish(ctx context.Context, err error) {
	for _, h := range r.Hooks {
		h.AfterLLMFinish(ctx, err)
	}
}

// Invoke ...
func (r *BotEngine) Invoke(ctx context.Context, input Input) (any, error) {
	return nil, nil
}

// Stream ... return channel
func (r *BotEngine) Stream(ctx context.Context, input Input) <-chan Event {
	return nil
}

// NewBotEngine ...
func NewBotEngine(opts ...Option) *BotEngine {
	options := &Options{
		logger: &log.NoneLogger{},
	}
	for i := 0; i < len(opts); i++ {
		opts[i](options)
	}
	return &BotEngine{opts: options, Hooks: options.Hooks, SessionID: uuid.NewString()}
}

// WithContext ...
func (r *BotEngine) WithContext(ctx context.Context) *BotEngine {
	r.context = ctx
	return r
}

// Context ...
func (r *BotEngine) Context() context.Context {
	return r.context
}

// FetchHistoryMessages ...
func (r *BotEngine) FetchHistoryMessages(ctx context.Context) []llm.Message {
	if r.opts.Memory == nil {
		return nil
	}
	if r.HistoryMessages != nil {
		return r.HistoryMessages
	}

	messages, err := r.opts.Memory.FetchMemories(ctx, r.opts.ConversationID)
	if err != nil {
		return nil
	}

	for _, m := range messages {
		if m.Content().Data == "" {
			continue
		}
		r.HistoryMessages = append(r.HistoryMessages, m)
	}
	return r.HistoryMessages
}

// RenderMessages ...
func (r *BotEngine) RenderMessages(ctx context.Context, preset string,
	inputs map[string]any, query string, contexts string) ([]llm.Message, []string) {
	formatter := prompt.SimplePromptFormatter{}
	messages, stops := formatter.Format(preset, inputs, query, contexts)

	// history
	if r.opts.Memory != nil {
		histories := r.FetchHistoryMessages(ctx)
		// messages 反转
		for i := 0; i < len(histories); i++ {
			messages = append(messages, histories[len(histories)-1-i])
		}
	}

	// query
	messages = append(messages, llm.UserPromptMessage(query))
	return messages, stops
}

// RenderTools ...
func (r *BotEngine) RenderTools(tools []ToolSetting) []llm.Tool {
	messageTools := make([]llm.Tool, 0)
	for _, tool := range tools {
		messageTool := llm.Tool{
			Type: "function",
			Function: &llm.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: map[string]any{
					"type":       "object",
					"properties": map[string]any{},
					"required":   []string{},
				},
			},
		}

		for _, p := range tool.Parameters {
			properties := messageTool.Function.Parameters["properties"].(map[string]any)
			properties[p.Name] = map[string]any{
				"description": p.LLMDescrition,
				"type":        p.Type,
				"enum":        nil,
			}
			if p.Required {
				messageTool.Function.Parameters["required"] = append(messageTool.Function.Parameters["required"].([]string), p.Name)
			}
		}

		messageTools = append(messageTools, messageTool)
	}
	return messageTools
}
