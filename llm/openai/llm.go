package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/log"
)

const (
	// Name name of llm
	Name = "openai"
)

func init() {
	llm.Register(Name, func(o *llm.ProviderOptions) llm.Provider { return &LLM{} })
}

// LLM ...
type LLM struct {
	client *openai.Client
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, opts *llm.InvokeOptions) (*llm.Response, error) {
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
			messagesOpenAI = append(messagesOpenAI, openai.UserMessage(m.Content()))
		} else if m.Role() == llm.MessageRoleTool {
			messagesOpenAI = append(messagesOpenAI, openai.ToolMessage(m.Content(), m.ToolID()))
		} else {
			continue
		}
	}

	var toolsOpenAI []openai.ChatCompletionToolParam
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

		toolsOpenAI = append(toolsOpenAI, openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.F(t.Function.Name),
				Description: openai.F(t.Function.Description),
				Parameters:  openai.F(params),
			}),
		})
	}

	params := openai.ChatCompletionNewParams{
		Messages: openai.F(messagesOpenAI),
		Tools:    openai.F(toolsOpenAI),
		// Seed:     openai.Int(0),
		Model: openai.F(opts.Model),
	}
	raw, _ := json.Marshal(params)
	log.InfoContextf(ctx, "openai params: %s", string(raw))
	stream := m.client.Chat.Completions.NewStreaming(ctx, params)

	// 流式响应
	acc := openai.ChatCompletionAccumulator{}

	response := llm.NewStreamResponse()
	go func() {
		defer response.Stream().Close()
		for stream.Next() {
			chunk := stream.Current()
			mmm := llm.NewAssistantMessage(chunk.Choices[0].Delta.Content)
			acc.AddChunk(chunk)
			// When this fires, the current chunk value will not contain content data
			if _, ok := acc.JustFinishedContent(); ok {
				break
			}

			if tool, ok := acc.JustFinishedToolCall(); ok {
				// println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
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

			// if refusal, ok := acc.JustFinishedRefusal(); ok {
			// 	println("Refusal stream finished:", refusal)
			// }

			// It's best to use chunks after handling JustFinished events
			// if len(chunk.Choices) > 0 {
			// 	println(chunk.Choices[0].Delta.JSON.RawJSON())
			// }
			response.Stream().Push(llm.NewChunk(0, mmm, nil))

		}
		if stream.Err() != nil {
			panic(stream.Err())
		}
	}()

	return response, nil
}

func (m *LLM) setupClient() error {
	config, _ := llm.Config.Get("openai").(map[string]any)
	if config == nil {
		return fmt.Errorf("openai config not found")
	}
	apiKey, _ := config["api_key"].(string)
	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)
	m.client = client
	return nil
}
