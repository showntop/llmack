package openaic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
)

const (
	// Name name of llm
	Name = "openai-c"
)

func init() {
	llm.Register(Name, &LLM{})
}

// LLM TODO
type LLM struct {
	client *openai.Client
}

// Invoke TODO
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, options *llm.InvokeOptions) (*llm.Response, error) {
	if err := m.setupClient(); err != nil { // TODO sync.Once
		return nil, err
	}

	var messagesOpenAI []openai.ChatCompletionMessageParamUnion
	for _, m := range messages {
		if m.Role() == llm.MessageRoleSystem {
			messagesOpenAI = append(messagesOpenAI, openai.SystemMessage(m.Content()))
		} else if m.Role() == llm.MessageRoleAssistant {
			messagesOpenAI = append(messagesOpenAI, openai.AssistantMessage(m.Content()))
		} else if m.Role() == llm.MessageRoleUser {
			if m.Content() != "" {
				messagesOpenAI = append(messagesOpenAI, openai.UserMessage(m.Content()))
			}
			if len(m.MultipartContent()) > 0 {
				parts := m.MultipartContent()
				partsOpenAI := []openai.ChatCompletionContentPartUnionParam{}

				for i := 0; i < len(parts); i++ {
					if parts[i].Type == "text" {
						partsOpenAI = append(partsOpenAI, openai.TextPart(parts[i].Data))
					}
					if parts[i].Type == "image_url" {
						partsOpenAI = append(partsOpenAI, openai.ImagePart(parts[i].Data))
					}
				}
				messagesOpenAI = append(messagesOpenAI, openai.UserMessageParts(partsOpenAI...))
			}
		} else if m.Role() == llm.MessageRoleTool {
			messagesOpenAI = append(messagesOpenAI, openai.ToolMessage(m.Content(), m.ToolID()))
		} else {
			continue
		}
	}

	var toolsOpenAI []openai.ChatCompletionToolParam
	for _, t := range options.Tools {
		toolsOpenAI = append(toolsOpenAI, openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.F(t.Function.Name),
				Description: openai.F(t.Function.Description),
				Parameters:  openai.F(openai.FunctionParameters(t.Function.Parameters)),
			}),
		})
	}

	params := openai.ChatCompletionNewParams{
		Messages: openai.F(messagesOpenAI),
		Tools:    openai.F(toolsOpenAI),
		Model:    openai.F(options.Model),
	}
	rawx, _ := json.Marshal(params)
	log.InfoContextf(ctx, "Openai-c chat request payload %s", string(rawx))
	stream := m.client.Chat.Completions.NewStreaming(ctx, params)

	// 流式响应
	acc := openai.ChatCompletionAccumulator{}

	response := llm.NewStreamResponse()
	go func() {
		defer response.Stream().Close()

		for stream.Next() {
			chunk := stream.Current()
			if !acc.AddChunk(chunk) { // error dismatch
				return // error
			}
			if _, ok := acc.JustFinishedContent(); len(chunk.Choices[0].Delta.ToolCalls) > 0 && !ok {
				continue
			}

			mmm := llm.NewAssistantMessage(chunk.Choices[0].Delta.Content)

			if tool, ok := acc.JustFinishedToolCall(); ok {
				call := acc.Choices[0].Message.ToolCalls[tool.Index]
				mmm.ToolCalls = append(mmm.ToolCalls, &llm.ToolCall{
					ID:   call.ID,
					Type: string(call.Type),
					Function: llm.ToolCallFunction{
						Name:      call.Function.Name,
						Arguments: call.Function.Arguments,
					},
				})
			}

			// When this fires, the current chunk value will not contain content data
			if _, ok := acc.JustFinishedContent(); ok {
				break
			}

			// if refusal, ok := acc.JustFinishedRefusal(); ok {
			// 	println("Refusal stream finished:", refusal)
			// }

			response.Stream().Push(llm.NewChunk(0, mmm, nil))
		}
		if err := stream.Err(); err != nil {
			panic(err)
		}
	}()

	return response, nil
}

func (m *LLM) setupClient() error {
	config, _ := llm.Config.Get("openai-c").(map[string]any)
	if config == nil {
		return fmt.Errorf("openai-c config not found")
	}
	baseURL, _ := config["base_url"].(string)
	apiKey, _ := config["api_key"].(string)
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL),
	)
	m.client = client
	return nil
}
