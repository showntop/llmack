package agent

import (
	"context"

	"github.com/showntop/llmack/llm"
)

// EchoAgent ...
type EchoAgent struct {
	// 模型agent
}

// Run ...
// 一轮对话
// 历史消息，query
func (a *EchoAgent) Run(ctx context.Context, messages []llm.Message, query string, stream bool) (any, error) {
	messages = append(messages, llm.UserTextPromptMessage(query))
	if stream {
		stream := llm.NewStream()
		stream.Push(llm.NewChunk(0, llm.AssistantPromptMessage(query), nil))
		return stream, nil
	}
	return llm.Result{Message: llm.AssistantPromptMessage("echo message!!!")}, nil
}
