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
	llm.Register(Name, &LLM{})
}

// LLM ...
type LLM struct {
	client *openai.Client
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, tools []llm.PromptMessageTool,
	options ...llm.InvokeOption) (*llm.Response, error) {
	if err := m.setupClient(); err != nil { // TODO sync.Once
		return nil, err
	}
	var opts llm.InvokeOptions
	for _, o := range options {
		o(&opts)
	}

	var messagesOpenAI []openai.ChatCompletionMessageParamUnion
	for _, m := range messages {
		if m.Role() == llm.PromptMessageRoleSystem {
			messagesOpenAI = append(messagesOpenAI, openai.SystemMessage(m.Content().Data))
		} else if m.Role() == llm.PromptMessageRoleAssistant {
			messagesOpenAI = append(messagesOpenAI, openai.AssistantMessage(m.Content().Data))
		} else if m.Role() == llm.PromptMessageRoleUser {
			messagesOpenAI = append(messagesOpenAI, openai.UserMessage(m.Content().Data))
		} else if m.Role() == llm.PromptMessageRoleTool {
			messagesOpenAI = append(messagesOpenAI, openai.ToolMessage(m.Content().Data, m.ToolID()))
		} else {
			continue
		}
	}

	var toolsOpenAI []openai.ChatCompletionToolParam
	for _, t := range tools {
		toolsOpenAI = append(toolsOpenAI, openai.ChatCompletionToolParam{
			Type: openai.F(openai.ChatCompletionToolTypeFunction),
			Function: openai.F(openai.FunctionDefinitionParam{
				Name:        openai.F(t.Name),
				Description: openai.F(t.Description),
				Parameters:  openai.F(openai.FunctionParameters(t.Parameters)),
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
			mmm := llm.AssistantPromptMessage(chunk.Choices[0].Delta.Content)
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