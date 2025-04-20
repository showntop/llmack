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
	llm.Register(Name, &LLM{})
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
					part.ImageURL = &zhipu.URLItem{
						URL: m.Data,
					}
				}
				if m.Type == "text" {
					part.Text = m.Data
				}
				contents = append(contents, part)
			}
			internalMessages = append(internalMessages, zhipu.ChatCompletionMultiMessage{
				Role:    string(m.Role()),
				Content: contents,
			})
		}
	}
	raw, _ := json.Marshal(internalMessages)
	fmt.Println(string(raw))
	var internalTools []zhipu.ChatCompletionTool
	for _, t := range opts.Tools {
		internalTools = append(internalTools, zhipu.ChatCompletionToolFunction{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  openai.FunctionParameters(t.Function.Parameters),
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
