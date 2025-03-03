package program

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/prompt"
	"github.com/showntop/llmack/tool"
)

type funcall struct {
	*predictor
	memory  Memory
	tools   []string
	thoughs []llm.Message
	buffer  chan any
}

// FunCall ...
func FunCall(opts ...option) *funcall {
	funcall := &funcall{buffer: make(chan any, 5)}

	p := &predictor{
		adapter: &RawAdapter{},
		Promptx: Promptx{
			Name:         "FunCallAgent",
			Instruction:  FunCallPrompt,
			Description:  "FunCall mode Agent for General tasks Solve.",
			InputFields:  make(map[string]*Field),
			OutputFields: make(map[string]*Field),
		},
	}
	for i := 0; i < len(opts); i++ {
		opts[i](p)
	}
	if p.model == nil {
		p.model = defaultLLM
	}
	funcall.predictor = p
	return funcall
}

func (rp *funcall) WithTools(tools ...string) *funcall {
	rp.tools = tools
	return rp
}

func (rp *funcall) WithInstruction(i string) *funcall {
	instruction := rp.Instruction
	instruction = strings.ReplaceAll(instruction, "{{instruction}}", i)
	rp.predictor.Promptx.Instruction = instruction
	return rp
}

// Invoke invoke forward for predicte
func (rp *funcall) Invoke(ctx context.Context, inputs map[string]any) *Result {
	var value Result = Result{p: rp.predictor, stream: rp.buffer}

	go func() {
		defer close(rp.buffer)
		answer := ""
		for i := 0; i < 20; i++ {
			result, finish, err := rp.invoke(ctx, inputs)
			if err != nil {
				value.err = err
				log.ErrorContextf(ctx, "agent funcall invoke error: %v", err)
				continue
			}
			_ = result
			if finish {
				answer = result
				break
			}
		}
		value.completion = answer
	}()
	return &value
}

func (rp *funcall) invoke(ctx context.Context, inputs map[string]any) (string, bool, error) {
	messageTools := rp.renderTools(rp.tools...)

	messages, _ := rp.renderPromptMessages(ctx, rp.Instruction, inputs, "contexts")
	response, err := rp.model.Invoke(ctx, messages,
		llm.WithTools(messageTools...),
		llm.WithStream(true),
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
			final += r.Delta.Message.Content()
			continue
		}
		rp.buffer <- r // when text only
		final += r.Delta.Message.Content()
	}
	if len(toolCalls) > 0 {
		rp.thoughs = append(rp.thoughs, llm.AssistantPromptMessage(string(final)).WithToolCalls(toolCalls))
		for i := 0; i < len(toolCalls); i++ {
			toolResult, err := rp.invokeTool(ctx, rp.tools, toolCalls[i])
			if err != nil { // 调用工具出错
				return "", false, err
			}
			if toolResult == "" {
				rp.thoughs = append(rp.thoughs,
					llm.ToolPromptMessage("no result", toolCalls[i].ID))
				continue
			}
			// 记录工具调用
			rp.thoughs = append(rp.thoughs,
				llm.ToolPromptMessage(toolResult, toolCalls[i].ID))
		}
	} else {
		final += "\n"
		finish = true
	}
	return final, finish, nil

}

func (rp *funcall) invokeTool(ctx context.Context, tools []string, t *llm.ToolCall) (string, error) {
	// TODO check function name valid?
	var args map[string]any
	if err := json.Unmarshal([]byte(t.Function.Arguments), &args); err != nil {
		return "", err
	}
	result, err := tool.Spawn(t.Function.Name).Invoke(ctx, args)
	log.InfoContextf(ctx, "program funcall invoke tool: %s, %s response: %s error: %v \n", t.Function.Name, t.Function.Arguments, result, err)
	if err != nil {
		return "", err
	}
	return result, nil
}

// renderPromptMessages ...
func (rp *funcall) renderPromptMessages(ctx context.Context, preset string,
	inputs map[string]any, contexts string) ([]llm.Message, []string) {
	messages := []llm.Message{}
	presetPrompt := preset
	presetPrompt, err := prompt.Render(presetPrompt, inputs)
	if err != nil {
		panic(err)
		// return nil, nil nothing here
	}

	_ = contexts

	messages = append(messages, llm.UserTextPromptMessage(presetPrompt))

	// if memory from history
	if rp.memory != nil {
		histories := rp.FetchHistoryMessages(ctx)
		messages = append(messages, histories...)
	}

	messages = append(messages, rp.thoughs...)
	return messages, nil
}

// renderContexts 从知识库中检索相关信息
func (rp *funcall) renderContexts(ctx context.Context, settings any, query string) (string, error) {
	var contexts string
	// if len(settings.Knowledge) > 0 {
	// 	// query = strings.TrimSpecial(query)
	// 	// query rewrite TODO，暂时实现历史消息合并
	// 	histories := rp.FetchHistoryMessages(ctx)
	// 	for _, history := range histories {
	// 		if history.Role() == llm.PromptMessageRoleUser {
	// 			query += history.Content()
	// 		}
	// 	}
	// 	kns, err := rp.opts.Rag.Retrieve(ctx, query, &rag.Options{
	// 		LibraryID:      settings.Knowledge[0].LibraryID,
	// 		Kind:           settings.Knowledge[0].Kind,
	// 		IndexID:        settings.Knowledge[0].IndexID,
	// 		TopK:           settings.Knowledge[0].TopK,
	// 		ScoreThreshold: settings.Knowledge[0].ScoreThreshold,
	// 	})
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	for _, kn := range kns {
	// 		contexts += kn.Answer
	// 	}
	// }
	return contexts, nil
}

// renderTools ...
func (rp *funcall) renderTools(tools ...string) []*llm.Tool {
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

var FunCallPrompt = `
Use the following context as your learned knowledge, inside <context></context> XML tags.

<context>
{{contexts}}
</context>

When answer to user:
- If you don't know, just say that you don't know.
- If you don't know when you are not sure, ask for clarification.\nAvoid mentioning that you obtained the information from the context.\nAnd answer according to the language of the user's question.

{{instruction}}
`
