package qwen

import (
	"context"
	"fmt"
	"sync"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/showntop/llmack/llm"
)

const (
	// Name name of llm
	Name = "qwen"
)

func init() {
	llm.Register(Name, &LLM{})
}

// LLM TODO
type LLM struct {
	client *openai.Client
}

var initOnce sync.Once

// Invoke TODO
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, options ...llm.InvokeOption) (*llm.Response, error) {
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
	for _, t := range opts.Tools {
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
		Model:    openai.F(opts.Model),
	}

	stream := m.client.Chat.Completions.NewStreaming(ctx, params)

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
		if err := stream.Err(); err != nil {
			panic(err)
		}
	}()

	return response, nil
}

func (m *LLM) setupClient() error {
	var err error
	initOnce.Do(func() {
		config, _ := llm.Config.Get(Name).(map[string]any)
		if config == nil {
			err = fmt.Errorf("%s config not found", Name)
		}
		apiKey, _ := config["api_key"].(string)
		client := openai.NewClient(
			option.WithAPIKey(apiKey),
			option.WithBaseURL("https://dashscope.aliyuncs.com/compatible-mode/v1/"),
		)
		m.client = client
	})

	return err
}