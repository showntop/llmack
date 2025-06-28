package zhipu

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/showntop/llmack/llm"
	"github.com/yankeguo/zhipu"
)

const (
	// Name name of llm
	Name = "zhipu"
)

func init() {
	llm.Register(Name, func(o *llm.ProviderOptions) llm.Provider { return &LLM{} })
}

// LLM TODO
type LLM struct {
	client *zhipu.Client
}

// Invoke TODO
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, opts *llm.InvokeOptions) (*llm.Response, error) {
	response := llm.NewStreamResponse()

	if err := m.setupClient(); err != nil { // TODO sync.Once
		return nil, err
	}

	internalMessages := []zhipu.ChatCompletionMessageType{}
	for _, m := range messages {
		if m.Content() != "" {
			internalMessages = append(internalMessages, zhipu.ChatCompletionMessage{
				Role:       string(m.Role()),
				Content:    m.Content(),
				ToolCallID: m.ToolID(),
			})
		}
		if len(m.MultipartContent()) > 0 {
			contents := []zhipu.ChatCompletionMultiContent{}
			for _, m := range m.MultipartContent() {
				part := zhipu.ChatCompletionMultiContent{
					Type: m.Type,
				}
				if m.Type == "image_url" {
					if url, ok := m.Data.(string); ok {
						part.ImageURL = &zhipu.URLItem{
							URL: url,
						}
					}
				}
				if m.Type == "text" {
					if text, ok := m.Data.(string); ok {
						part.Text = text
					}
				}
				contents = append(contents, part)
			}
			internalMessages = append(internalMessages, zhipu.ChatCompletionMultiMessage{
				Role:    string(m.Role()),
				Content: contents,
			})
		}
	}
	var internalTools []zhipu.ChatCompletionTool
	for _, t := range opts.Tools {
		params, ok := t.Function.Parameters.(openai.FunctionParameters)
		if !ok {
			// 如果类型断言失败，尝试通过JSON序列化/反序列化转换
			paramsBytes, err := json.Marshal(t.Function.Parameters)
			if err != nil {
				continue
			}
			var convertedParams openai.FunctionParameters
			if err := json.Unmarshal(paramsBytes, &convertedParams); err != nil {
				continue
			}
			params = convertedParams
		}

		internalTools = append(internalTools, zhipu.ChatCompletionToolFunction{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  params,
		})
	}
	service := m.client.
		ChatCompletion(opts.Model).
		AddMessage(internalMessages...).
		AddTool(internalTools...).
		SetStreamHandler(func(chunk zhipu.ChatCompletionResponse) error {
			// println("chunk content: ", chunk.Choices[0].Delta.Content)
			mmm := llm.NewAssistantMessage(chunk.Choices[0].Delta.Content)
			response.Stream().Push(llm.NewChunk(0, mmm, nil))
			return nil
		})

	go func() {
		defer response.Stream().Close()
		_, err := service.Do(ctx)
		if err != nil {
			panic(err)
			// return nil, err
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
	// or you can specify the API key
	client, err := zhipu.NewClient(zhipu.WithAPIKey(apiKey))
	if err != nil {
		return err
	}
	m.client = client
	return nil
}
