package deepseek

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/showntop/llmack/llm"
)

// Name ...
var Name = "deepseek"

func init() {
	llm.Register(Name, NewLLM())
}

// LLM ...
type LLM struct {
	bearer string
	client *http.Client
}

// NewLLM ...
func NewLLM() *LLM {
	return &LLM{client: http.DefaultClient}
}

// Name ...
func (m *LLM) Name() string {
	return Name
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, optFuncs ...llm.InvokeOption) (*llm.Response, error) {

	options := &llm.InvokeOptions{}
	for i := 0; i < len(optFuncs); i++ {
		optFuncs[i](options)
	}
	// validate
	if options.Model == "" {
		return nil, errors.New("model is required")
	}

	// chat completions
	body, err := m.ChatCompletions(ctx, m.buildRequest(messages, options))
	if err != nil {
		return nil, err
	}

	if options.Stream {
		return m.handleStreamResponse(ctx, body)
	}
	// TODO implement non strem response
	return nil, nil

}

// ChatCompletions ...
func (m *LLM) ChatCompletions(ctx context.Context, req *ChatCompletionsRequest) (io.ReadCloser, error) {
	url := "https://api.deepseek.com/chat/completions"
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	fmt.Println("payload: ", string(payload))
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	config, _ := llm.Config.Get(Name).(map[string]any)
	if config == nil {
		return nil, fmt.Errorf("deepseek config not found")
	}
	apiKey, _ := config["api_key"].(string)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	return resp.Body, nil
}

// buildRequest ...
func (m *LLM) buildRequest(messages []llm.Message, options *llm.InvokeOptions) *ChatCompletionsRequest {
	request := &ChatCompletionsRequest{}
	request.Model = options.Model
	if options.TopP > 0 {
		request.TopP = options.TopP
	}
	if options.Temperature > 0 {
		request.Temperature = options.Temperature
	}
	request.Stream = options.Stream
	// messages
	for _, m := range messages {
		request.Messages = append(request.Messages, &Message{
			Role:       string(m.Role()),
			Content:    m.Content().Data,
			ToolCallID: m.ToolID(),
		})
	}
	// tools
	if len(options.Tools) <= 0 {
		return request
	}
	request.ToolChoice = "auto"
	request.Tools = make([]*Tool, len(options.Tools))
	for i, t := range options.Tools {
		raw, _ := json.Marshal(t.Function.Parameters)
		params := string(raw)
		request.Tools[i] = &Tool{
			Type: "function",
			Function: &ToolFunction{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  params,
			},
		}
	}
	return request
}

// handleStreamResponse ...
func (m *LLM) handleStreamResponse(ctx context.Context, body io.ReadCloser) (*llm.Response, error) {
	response := llm.NewStreamResponse()
	go func() {
		defer body.Close()
		defer response.Stream().Close()

		var toolCall *llm.ToolCall
		toolCalls := []*llm.ToolCall{}
		scanner := bufio.NewScanner(body)
		for scanner.Scan() {
			raw := scanner.Bytes()
			if len(raw) == 0 {
				continue
			}
			if arr := bytes.SplitN(raw, []byte(":"), 2); len(arr) == 2 {
				raw = arr[1]
			} else {
				continue
			}
			if bytes.Contains(raw, []byte("[DONE]")) { // 结束
				break
			}
			fmt.Println("deepseek chunk: ", string(raw))
			var chunk ChatCompletionsResponse
			if err := json.Unmarshal(raw, &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) <= 0 {
				continue
			}

			// 提取 tool call
			mmm := llm.AssistantPromptMessage(chunk.Choices[0].Delta.Content)
			if responseToolCalls := chunk.Choices[0].Delta.ToolCalls; len(responseToolCalls) > 0 {
				newToolCalls := m.extractResponseToolCalls(responseToolCalls)
				if len(newToolCalls) > 0 {
					newToolCall := newToolCalls[0]
					if toolCall == nil {
						toolCall = newToolCall
					} else if toolCall.ID != newToolCall.ID {
						toolCalls = append(toolCalls, toolCall)
						toolCall = newToolCall
					} else {
						toolCall.Function.Name += newToolCall.Function.Name
						toolCall.Function.Arguments += newToolCall.Function.Arguments
					}
				}
				if toolCall != nil && len(toolCall.Function.Name) > 0 && len(toolCall.Function.Arguments) > 0 {
					toolCalls = append(toolCalls, toolCall)
					toolCall = nil
				}
			}
			if len(toolCalls) > 0 { // 重置content
				// mmm
			}

			// 直到finish
			if chunk.Choices[0].FinishReason == "tool_calls" {
				mmm.ToolCalls = toolCalls
				toolCall = nil
				toolCalls = toolCalls[:0]
			}

			response.Stream().Push(llm.NewChunk(0, mmm, nil))
		}
	}()

	return response, nil
}

func (m *LLM) extractResponseToolCalls(responseToolCalls []*ToolCall) []*llm.ToolCall {
	toolCalls := make([]*llm.ToolCall, 0)
	for _, call := range responseToolCalls {
		toolCalls = append(toolCalls, &llm.ToolCall{
			ID:   call.ID,
			Type: "function",
			Function: llm.ToolCallFunction{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		})
	}
	return toolCalls
}
