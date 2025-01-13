package ollama

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/ollama/ollama/api"
	"github.com/showntop/llmack/llm"
)

const (
	// Name name of llm
	Name = "ollama"
)

func init() {
	llm.Register(Name, &LLM{name: Name})
}

// LLM ...
type LLM struct {
	model  string
	client *api.Client
	name   string
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, tools []llm.PromptMessageTool,
	options ...llm.InvokeOption) (*llm.Response, error) {
	if len(messages) == 0 {
		return nil, errors.New("empty messages")
	}
	if err := m.setupClient(); err != nil {
		return nil, err
	}

	var opts llm.InvokeOptions
	for _, o := range options {
		o(&opts)
	}

	var response *llm.Response = new(llm.Response)
	response = response.MakeStream()
	req := &api.ChatRequest{
		Model:  opts.Model,
		Stream: &opts.Stream,
		Tools:  nil,
	}
	for _, m := range messages {
		req.Messages = append(req.Messages, api.Message{
			Role:    string(m.Role()),
			Content: m.Content().Data,
		})
	}
	// TODO tools support

	idx := 1
	err := m.client.Chat(ctx, req, func(resp api.ChatResponse) error {
		response.Stream().Push(llm.NewChunk(idx, llm.AssistantPromptMessage(resp.Message.Content), nil))
		if resp.Done {
			response.Stream().Close()
			return nil
		}
		idx++
		return nil
	})

	if err != nil {
		return nil, err
	}
	return response, nil
}

func (m *LLM) setupClient() error {
	config, _ := llm.Config.Get(Name).(map[string]any)
	if config == nil {
		return errors.New("ollama config not found")
	}
	urlx, err := url.Parse(config["base_url"].(string))
	if err != nil {
		return err
	}
	client := api.NewClient(
		urlx,
		http.DefaultClient,
	)
	m.client = client
	return nil
}
