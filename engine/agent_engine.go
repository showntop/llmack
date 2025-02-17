package engine

import (
	"context"
	"encoding/json"
	stdstrs "strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/pkg/strings"
	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/tool"
)

// AgentEngine ...
type AgentEngine struct {
	ConvID    int64
	MessageID int64

	thoughs []llm.Message
	BotEngine
}

// NewAgentEngine ...
func NewAgentEngine(settings *Settings, opts ...Option) Engine {
	r := &AgentEngine{}
	r.BotEngine = *NewBotEngine(opts...)
	// load tools

	r.Settings = settings
	return r
}

// RenderPromptMessages ...
func (engine *AgentEngine) RenderPromptMessages(ctx context.Context, preset string,
	inputs map[string]any, query string, contexts string) ([]llm.Message, []string) {
	messages := []llm.Message{}

	systemPrompt := `Use the following context as your learned knowledge, inside <context></context> XML tags.\n\n<context>\n{{contexts}}\n</context>\n\nWhen answer to user:\n- If you don't know, just say that you don't know.\n- If you don't know when you are not sure, ask for clarification.\nAvoid mentioning that you obtained the information from the context.\nAnd answer according to the language of the user's question.`
	systemPrompt = stdstrs.ReplaceAll(systemPrompt, "{{contexts}}", contexts)
	systemPrompt += preset
	messages = append(messages, llm.SystemPromptMessage(systemPrompt))
	// messages = append(messages, llm.UserPromptMessage(preset)) // user preset prompt format inputs

	// formatter := prompt.SimplePromptFormatter{}
	// messages, _ := formatter.Format(preset, inputs, query, contexts)

	// if preset != "" {
	// 	messages = append(messages, llm.SystemPromptMessage(preset))
	// }
	// if memory from history
	if engine.opts.Memory != nil {
		histories := engine.FetchHistoryMessages(ctx)
		messages = append(messages, histories...)
	}

	messages = append(messages, llm.UserTextPromptMessage(query)) // 本次 query

	messages = append(messages, engine.thoughs...)
	return messages, nil
}

// Execute ... return channel
// ReAct 模式
func (engine *AgentEngine) Execute(ctx context.Context, input Input) *EventStream {
	result := NewEventStream()

	settings := engine.Settings
	inputs := input.Inputs
	query := input.Query
	contexts, err := engine.renderContexts(ctx, settings, query)
	if err != nil {
		// return nil, err
	}
	// tools
	messageTools := engine.RenderTools(settings.Tools...)

	go func() {
		defer result.Close()

		finalAnswer := ""
		finish := false
		for i := 0; i <= 5 && !finish; i++ { // 最大5次
			if i == 5 { // 最后一次不使用工具
				messageTools = nil
			}
			response, finish, err := engine.iterate(ctx, inputs, query, contexts, messageTools, result)
			if err != nil { // send error
				result.Push(ErrorEvent(err))
				return
			}

			finalAnswer += response
			if finish {
				result.Push(EndEvent(response))
				break
			}
		}

		// send end event
		result.Push(EndEvent(finalAnswer))
	}()

	return result
}

func (engine *AgentEngine) iterate(ctx context.Context,
	inputs map[string]any, query string, contexts string,
	tools []*llm.Tool, result *EventStream) (string, bool, error) {
	messages, _ := engine.RenderPromptMessages(ctx, engine.Settings.PresetPrompt, inputs, query, contexts)

	instance := llm.NewInstance(engine.Settings.LLMModel.Provider)
	reponse, err := instance.Invoke(ctx, messages,
		llm.WithTools(tools...),
		llm.WithStream(engine.Settings.Stream),
		llm.WithModel(engine.Settings.LLMModel.Name),
	)
	if err != nil {
		return "", false, err
	}
	stream := reponse.Stream()
	toolCalls := []llm.ToolCall{}
	response := ""
	finish := false
	for r := stream.Next(); r != nil; r = stream.Next() {
		if len(r.Delta.Message.ToolCalls) > 0 { // tool call
			for i := 0; i < len(r.Delta.Message.ToolCalls); i++ {
				toolCalls = append(toolCalls, *r.Delta.Message.ToolCalls[i])
			}
		}

		response += r.Delta.Message.Content()
		result.Push(ToastEvent(r))
	}

	if len(toolCalls) > 0 {
		engine.thoughs = append(engine.thoughs, llm.AssistantPromptMessage(response))
		for i := 0; i < len(toolCalls); i++ {
			toolResult, err := engine.invokeTool(ctx, engine.Settings.Tools, toolCalls[i])
			if err != nil { // 调用工具出错
				return "", false, err
			}
			if toolResult == "" {
				engine.thoughs = append(engine.thoughs,
					llm.ToolPromptMessage("未获取任何信息", toolCalls[i].ID))
				continue
			}
			// 记录工具调用
			engine.thoughs = append(engine.thoughs,
				llm.ToolPromptMessage(toolResult, toolCalls[i].ID))
		}
	} else {
		response += "\n"
		finish = true
	}
	return response, finish, nil
}

// "ToolCalls":[{"Id":"call_cr1kufc2c3m560b2ioe0","Type":"function","Function":{"Name":"weather","Arguments":"{\"city\":\"北京三里屯\"}"}}]}}]
func (engine *AgentEngine) invokeTool(ctx context.Context, tools []string, t llm.ToolCall) (string, error) {
	// TODO check function name valid?
	var args map[string]any
	json.Unmarshal([]byte(t.Function.Arguments), &args)
	return tool.Spawn(t.Function.Name).Invoke(ctx, args)
}

// renderContexts 从知识库中检索相关信息
func (engine *AgentEngine) renderContexts(ctx context.Context, settings *Settings, query string) (string, error) {
	var contexts string
	if len(settings.Knowledge) > 0 {
		query = strings.TrimSpecial(query)
		// query rewrite TODO，暂时实现历史消息合并
		histories := engine.FetchHistoryMessages(ctx)
		for _, history := range histories {
			if history.Role() == llm.PromptMessageRoleUser {
				query += history.Content()
			}
		}
		kns, err := engine.opts.Rag.Retrieve(ctx, query, &rag.Options{
			LibraryID:      settings.Knowledge[0].LibraryID,
			Kind:           settings.Knowledge[0].Kind,
			IndexID:        settings.Knowledge[0].IndexID,
			TopK:           settings.Knowledge[0].TopK,
			ScoreThreshold: settings.Knowledge[0].ScoreThreshold,
		})
		if err != nil {
			return "", err
		}
		for _, kn := range kns {
			contexts += kn.Answer
		}
	}
	return contexts, nil
}

// RenderTools ...
func (engine *AgentEngine) RenderTools(tools ...string) []*llm.Tool {
	messageTools := make([]*llm.Tool, 0)
	for _, toolName := range tools {
		tool := tool.Spawn(toolName)
		messageTool := &llm.Tool{
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
