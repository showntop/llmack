package program

import (
	"context"
	"errors"
	"sync"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
	"github.com/showntop/llmack/prompt"
	"github.com/showntop/llmack/tool"
)

var MaxIterationNum = 2000

// FunCall ...
func FunCall(opts ...option) *predictor {
	p := NewPredictor(opts...)
	p.Mode = "funcall"
	p.invoker = &funcall{p}
	return p
}

type funcall struct {
	*predictor
}

func (rp *funcall) Invoke(ctx context.Context, messages []llm.Message, query string, inputs map[string]any) *predictor {
	// at end recycle response stream
	defer close(rp.reponse.stream)

	maxIterationNum := min(rp.maxIterationNum, MaxIterationNum)
	// 迭代次数
	for i := range maxIterationNum {
		if i == maxIterationNum-1 { // remove tool
			rp.tools = []any{}
		}
		p, finish := rp.invoke(ctx, messages, query, inputs)
		if p.reponse.err != nil {
			return p
		}
		if finish {
			return p
		}
	}
	rp.reponse.err = errors.New("failed to invoke query until max iteration")
	return rp.predictor
}

func (rp *funcall) invoke(ctx context.Context, messages []llm.Message, query string, inputs map[string]any) (*predictor, bool) {
	systemMessages, err := rp.adapter.Format(rp.predictor, rp.inputs, nil) // system message
	if err != nil {
		rp.reponse.err = err
		return rp.predictor, false
	}
	messages = append(systemMessages, messages...)

	if len(query) > 0 {
		messages = append(messages, llm.NewUserTextMessage(query))
	}

	llmResponse, err := rp.invokeLLM(ctx, messages)
	if err != nil {
		rp.reponse.err = err
		return rp.predictor, false
	}

	stream := llmResponse.Stream()
	finish := false

	toolCalls := []*llm.ToolCall{}
	answer := ""
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
	for chunk := range stream.Next() {
		deltaMessage := chunk.Choices[0].Delta
		if len(deltaMessage.ToolCalls) > 0 { // tool call
			for i := range deltaMessage.ToolCalls {
				t := findToolCall(deltaMessage.ToolCalls[i].ID)
				t.Index = deltaMessage.ToolCalls[i].Index
				t.Type = "function" //deltaMessage.ToolCalls[i].Type
				t.Function.Name += deltaMessage.ToolCalls[i].Function.Name
				t.Function.Arguments += deltaMessage.ToolCalls[i].Function.Arguments
			}
		}
		answer += deltaMessage.Content()
		if rp.stream {
			rp.reponse.stream <- chunk
		}
	}
	rp.reponse.message = llm.NewAssistantMessage(answer)
	if len(toolCalls) > 0 {
		log.InfoContextf(ctx, "program funcall invoke tools toolcalls: %v", toolCalls)
		rp.observers = append(rp.observers, llm.NewAssistantMessage(answer).WithToolCalls(toolCalls))
		toolResults, err := rp.invokeTools(ctx, toolCalls)
		if err != nil { // 调用工具出错
			if rp.stream {
				rp.reponse.stream <- llm.NewChunk(0, llm.NewAssistantMessage(err.Error()), nil)
			}
			rp.reponse.err = err
			return rp.predictor, finish
		}
		log.InfoContextf(ctx, "program funcall invoke tools result %v", toolResults)
		// 记录工具调用
		for i := range toolCalls {
			rp.observers = append(rp.observers, llm.NewToolMessage(toolResults[toolCalls[i].ID], toolCalls[i].ID))
		}
	} else {
		answer += "\n"
		finish = true
	}

	return rp.predictor, finish
}

// invokeFuncall ...
func (rp *funcall) invokeLLM(ctx context.Context, messages []llm.Message) (*llm.Response, error) {
	messageTools := rp.buildTools(rp.tools...)
	if len(messageTools) <= 0 {
		rp.toolChoice = "none"
	}
	// append observer message
	messages = append(messages, rp.observers...)
	if rp.resetMessages != nil {
		messages = rp.resetMessages(ctx, messages)
	}
	response, err := rp.model.Invoke(ctx, messages,
		llm.WithTools(messageTools...),
		llm.WithStream(true),
		llm.WithToolChoice(rp.toolChoice),
	)
	if err != nil {
		return nil, err
	}

	return response, nil
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

	messages = append(messages, llm.NewUserTextMessage(presetPrompt))

	messages = append(messages, rp.observers...)
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

// common method
func (rp *funcall) buildTools(tools ...any) []*llm.Tool {
	messageTools := make([]*llm.Tool, 0)
	for _, tl := range tools {
		if tl == nil {
			continue
		}
		tool := tool.Spawn(tl.(string)) // 暂时只支持 string
		messageTool := &llm.Tool{
			Type: "function",
			Function: &llm.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters(),
			},
		}
		messageTools = append(messageTools, messageTool)
	}
	return messageTools
}

func (rp *funcall) invokeTools(ctx context.Context, toolCalls []*llm.ToolCall) (map[string]string, error) {
	// 并发调用
	ch := make(chan [2]string, len(toolCalls))
	// wg := errgroup.Group{}
	wg := sync.WaitGroup{}
	for _, toolCall := range toolCalls {
		wg.Add(1)
		go func() error {
			defer wg.Done()
			toolResult, err := tool.Spawn(toolCall.Function.Name).Invoke(ctx, toolCall.Function.Arguments)
			if err != nil {
				ch <- [2]string{toolCall.ID, "error with " + err.Error()}
				return err
			}
			// log.InfoContextf(ctx, "program funcall invoke tool: %s, %s response: %s error: %v \n", toolCall.ID, toolCall.Function.Arguments, toolResult, err)
			ch <- [2]string{toolCall.ID, toolResult}
			return nil
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make(map[string]string) // 使用 chan fix 并发冲突
	for result := range ch {
		results[result[0]] = result[1]
	}
	return results, nil
}
