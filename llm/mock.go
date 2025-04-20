package llm

import (
	"context"

	"github.com/google/uuid"
)

var MockLLMModelName = "mock"

type MockLLMModel struct {
}

func (m *MockLLMModel) Invoke(ctx context.Context, messages []Message, opts *InvokeOptions) (*Response, error) {
	if opts.Stream {
		response := NewStreamResponse()
		go func() {
			response.stream.Push(NewChunk(0, NewAssistantMessage("我是"), &Usage{
				PromptTokens:     1,
				CompletionTokens: 1,
			}))
			response.stream.Push(NewChunk(0, NewAssistantMessage("mock的"), &Usage{
				PromptTokens:     1,
				CompletionTokens: 1,
			}))
			response.stream.Push(NewChunk(0, NewAssistantMessage("内容"), &Usage{
				PromptTokens:     1,
				CompletionTokens: 1,
			}))
			response.stream.Push(NewChunk(0, NewAssistantMessage("结束"), &Usage{
				PromptTokens:     1,
				CompletionTokens: 1,
			}))
			if len(opts.Tools) > 0 {
				toolCalls := []*ToolCall{}
				for _, tool := range opts.Tools {
					toolCalls = append(toolCalls, &ToolCall{
						ID: uuid.New().String(),
						Function: ToolCallFunction{
							Name: tool.Function.Name,
						},
					})
				}
				response.stream.Push(NewChunk(0, NewAssistantMessage("").WithToolCalls(toolCalls), &Usage{
					PromptTokens:     1,
					CompletionTokens: 1,
				}))
			}

		}()
		return response, nil
	} else {
		return &Response{}, nil
	}
}

func (m *MockLLMModel) Name() string {
	return MockLLMModelName
}

func init() {
	Register(MockLLMModelName, &MockLLMModel{})
}
