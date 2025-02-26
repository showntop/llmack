package engine

import (
	"context"
	"encoding/json"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/pkg/strings"
	"github.com/showntop/llmack/program"
	"github.com/showntop/llmack/prompt"
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

// Execute ... return channel
// ReAct mode or Function Call mode.
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
	messageTools := engine.renderTools(settings.Tools...)

	go func() {
		defer result.Close()

		finalAnswer := ""
		finish := false
		for i := 0; i <= 5 && !finish; i++ { // 最大5次
			if i == 5 { // 最后一次不使用工具
				// messageTools = nil
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

func (engine *AgentEngine) iterateByReAct(ctx context.Context,
	inputs map[string]any, query string, contexts string,
	tools []*llm.Tool, eventSteam *EventStream) (string, bool, error) {

	var result struct {
		Tool *struct {
			Name string         `json:"name"`
			Args map[string]any `json:"args"`
		}
		Thoughts struct {
			Text    string `json:"text"`
			Reason  string `json:"reasoning"`
			Plan    string `json:"plan"`
			Critism string `json:"criticism"`
			Speak   string `json:"speak"`
		}
	}

	var rawResult string
	err := program.ReAct(program.WithLLM(engine.Settings.LLMModel.Provider, engine.Settings.LLMModel.Name)).
		WithInstruction(engine.Settings.PresetPrompt).
		WithTools(tools...).
		ChatWith(inputs).
		Result(ctx, &rawResult)
	if err != nil {
		return "", false, err
	}
	if err := json.Unmarshal([]byte(rawResult), &result); err != nil {
		return "", false, err
	}

	if result.Tool != nil {
		// TODO check function name valid?
		toolResult, err := tool.Spawn(result.Tool.Name).Invoke(ctx, result.Tool.Args)
		log.InfoContextf(ctx, "AgentEngine invokeTool: %s, %v response: %s error: %v \n", result.Tool.Name, result.Tool.Args, toolResult, err)
		if err != nil {
			return "", false, err
		}
		return toolResult, false, nil
	}
	return "", true, nil
}

func (engine *AgentEngine) iterateByFunCall(ctx context.Context,
	inputs map[string]any, query string, contexts string,
	tools []*llm.Tool, result *EventStream) (string, bool, error) {
	messages, _ := engine.renderPromptMessages(ctx, engine.Settings.PresetPrompt, inputs, query, contexts)
	instance := llm.NewInstance(engine.Settings.LLMModel.Provider)
	response, err := instance.Invoke(ctx, messages,
		llm.WithTools(tools...),
		llm.WithStream(engine.Settings.Stream),
		llm.WithModel(engine.Settings.LLMModel.Name),
	)
	if err != nil {
		return "", false, err
	}
	stream := response.Stream()
	toolCalls := []*llm.ToolCall{}
	final := ""
	finish := false

	// fill tool calls from chunk
	findToolCall := func(id string) *llm.ToolCall {
		if id == "" {
			return toolCalls[len(toolCalls)-1]
		}
		for _, t := range toolCalls {
			if t.ID == id {
				return t
			}
		}
		t := &llm.ToolCall{ID: id}
		toolCalls = append(toolCalls, t)
		return t
	}
	for r := stream.Next(); r != nil; r = stream.Next() {
		if len(r.Delta.Message.ToolCalls) > 0 { // tool call
			for i := 0; i < len(r.Delta.Message.ToolCalls); i++ {
				t := findToolCall(r.Delta.Message.ToolCalls[i].ID)
				t.Type = r.Delta.Message.ToolCalls[i].Type
				t.Function.Name += r.Delta.Message.ToolCalls[i].Function.Name
				t.Function.Arguments += r.Delta.Message.ToolCalls[i].Function.Arguments
			}
			continue
		}

		final += r.Delta.Message.Content()
		result.Push(ToastEvent(r))
	}

	if len(toolCalls) > 0 {
		engine.thoughs = append(engine.thoughs, llm.AssistantPromptMessage(string(final)).WithToolCalls(toolCalls))
		for i := 0; i < len(toolCalls); i++ {
			toolResult, err := engine.invokeTool(ctx, engine.Settings.Tools, toolCalls[i])
			if err != nil { // 调用工具出错
				return "", false, err
			}
			if toolResult == "" {
				engine.thoughs = append(engine.thoughs,
					llm.ToolPromptMessage("no result", toolCalls[i].ID))
				continue
			}
			// 记录工具调用
			engine.thoughs = append(engine.thoughs,
				llm.ToolPromptMessage(toolResult, toolCalls[i].ID))
		}
	} else {
		final += "\n"
		finish = true
	}
	return final, finish, nil
}

func (engine *AgentEngine) iterate(ctx context.Context,
	inputs map[string]any, query string, contexts string,
	tools []*llm.Tool, result *EventStream) (string, bool, error) {

	if engine.Settings.Agent.Mode == "ReAct" {
		return engine.iterateByReAct(ctx, inputs, query, contexts, tools, result)
	} else if engine.Settings.Agent.Mode == "FunCall" {
		return engine.iterateByFunCall(ctx, inputs, query, contexts, tools, result)
	}
	return "", false, nil
}

// "ToolCalls":[{"Id":"call_cr1kufc2c3m560b2ioe0","Type":"function","Function":{"Name":"weather","Arguments":"{\"city\":\"北京三里屯\"}"}}]}}]
func (engine *AgentEngine) invokeTool(ctx context.Context, tools []string, t *llm.ToolCall) (string, error) {
	// TODO check function name valid?
	var args map[string]any
	json.Unmarshal([]byte(t.Function.Arguments), &args)
	result, err := tool.Spawn(t.Function.Name).Invoke(ctx, args)
	log.InfoContextf(ctx, "AgentEngine invokeTool: %s, %s response: %s error: %v \n", t.Function.Name, t.Function.Arguments, result, err)
	if err != nil {
		return "", err
	}
	return result, nil
}

// renderPromptMessages ...
func (engine *AgentEngine) renderPromptMessages(ctx context.Context, preset string,
	inputs map[string]any, query string, contexts string) ([]llm.Message, []string) {
	messages := []llm.Message{}

	presetPrompt := `Use the following context as your learned knowledge, inside <context></context> XML tags.\n\n<context>\n{{contexts}}\n</context>\n\nWhen answer to user:\n- If you don't know, just say that you don't know.\n- If you don't know when you are not sure, ask for clarification.\nAvoid mentioning that you obtained the information from the context.\nAnd answer according to the language of the user's question.`
	presetPrompt += "\n" + preset
	presetPrompt, err := prompt.Render(presetPrompt, inputs)
	if err != nil {
		panic(err)
		// return nil, nil nothing here
	}

	_ = contexts

	if query != "" {
		messages = append(messages, llm.SystemPromptMessage(presetPrompt))
	} else {
		messages = append(messages, llm.UserTextPromptMessage(presetPrompt))
	}

	// if memory from history
	if engine.opts.Memory != nil {
		histories := engine.FetchHistoryMessages(ctx)
		messages = append(messages, histories...)
	}
	if query != "" {
		messages = append(messages, llm.UserTextPromptMessage(query)) // 本次 query
	}

	messages = append(messages, engine.thoughs...)
	return messages, nil
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

// renderTools ...
func (engine *AgentEngine) renderTools(tools ...string) []*llm.Tool {
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
