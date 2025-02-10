package moonshot

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/northes/go-moonshot"
	"github.com/showntop/llmack/llm"
)

const (
	// Name name of llm
	Name = "moonshot"
)

func init() {
	llm.Register(Name, &LLM{name: Name})
}

// LLM ...
type LLM struct {
	client *moonshot.Client
	name   string
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, options ...llm.InvokeOption) (*llm.Response, error) {
	if err := m.setupClient(); err != nil { // TODO sync.Once
		return nil, err
	}
	var opts llm.InvokeOptions
	for _, o := range options {
		o(&opts)
	}

	req := &moonshot.ChatCompletionsRequest{
		Model:       moonshot.ModelMoonshotV18K,
		Messages:    []*moonshot.ChatCompletionsMessage{},
		Temperature: opts.Temperature,
		Stream:      true,
	}
	// messages
	for _, m := range messages {
		req.Messages = append(req.Messages, &moonshot.ChatCompletionsMessage{
			Role:       moonshot.ChatCompletionsMessageRole(m.Role()),
			Content:    m.Content(),
			ToolCallID: m.ToolID(),
		})
	}

	// tools
	if len(opts.Tools) > 0 {
		req.Tools = make([]*moonshot.ChatCompletionsTool, len(opts.Tools))
		for i, t := range opts.Tools {
			req.Tools[i] = &moonshot.ChatCompletionsTool{
				Type: "function",
				Function: &moonshot.ChatCompletionsToolFunction{
					Name:        t.Function.Name,
					Description: t.Function.Description,
					// Parameters:  t.Function.Parameters,
				},
			}
		}
	}

	resp, err := m.client.Chat().CompletionsStream(ctx, req)
	if err != nil {
		return nil, err
	}
	// 流式响应
	response := llm.NewStreamResponse()

	go func() {
		defer response.Stream().Close()
		for receive := range resp.Receive() {
			msg, err := receive.GetMessage()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				break
			}
			response.Stream().Push(llm.NewChunk(0, llm.AssistantPromptMessage(msg.Content), nil))
		}
	}()
	return response, nil
}

func (m *LLM) setupClient() error {
	config, _ := llm.Config.Get(Name).(map[string]any)
	if config == nil {
		return fmt.Errorf("%s config not found", Name)
	}
	apiKey, _ := config["api_key"].(string)

	cli, err := moonshot.NewClientWithConfig(
		moonshot.NewConfig(
			moonshot.WithAPIKey(apiKey),
		),
	)
	if err != nil {
		return err
	}

	m.client = cli
	return nil
}
