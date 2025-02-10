package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/showntop/llmack/llm"
)

const (
	// Name name of llm
	Name = "azure-openai"
)

func init() {
	llm.Register(Name, &LLM{})
}

// LLM TODO
type LLM struct {
	client *azopenai.Client
}

// Invoke TODO
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, options ...llm.InvokeOption) (*llm.Response, error) {
	if err := m.setupClient(); err != nil { // TODO sync.Once
		return nil, err
	}
	var opts llm.InvokeOptions
	for _, o := range options {
		o(&opts)
	}

	var messagesOpenAI []azopenai.ChatRequestMessageClassification
	for _, m := range messages {
		if m.Role() == llm.PromptMessageRoleSystem {
			messagesOpenAI = append(messagesOpenAI, &azopenai.ChatRequestSystemMessage{
				Content: azopenai.NewChatRequestSystemMessageContent(m.Content())})
		} else if m.Role() == llm.PromptMessageRoleAssistant {
			messagesOpenAI = append(messagesOpenAI, &azopenai.ChatRequestAssistantMessage{
				Content: azopenai.NewChatRequestAssistantMessageContent(m.Content())})
		} else if m.Role() == llm.PromptMessageRoleUser {
			messagesOpenAI = append(messagesOpenAI, &azopenai.ChatRequestUserMessage{
				Content: azopenai.NewChatRequestUserMessageContent(m.Content())})
		} else if m.Role() == llm.PromptMessageRoleTool {
			messagesOpenAI = append(messagesOpenAI, &azopenai.ChatRequestToolMessage{
				Content: azopenai.NewChatRequestToolMessageContent(m.Content())})
		} else {
			continue
		}
	}

	// @TODO tools support
	toolsOpenAI := make([]azopenai.ChatCompletionsToolDefinitionClassification, 0, len(opts.Tools))
	for _, t := range opts.Tools {
		raw, _ := json.Marshal(t.Function.Parameters)
		toolsOpenAI = append(toolsOpenAI, &azopenai.ChatCompletionsFunctionToolDefinition{
			Function: &azopenai.ChatCompletionsFunctionToolDefinitionFunction{
				Name:        &t.Function.Name,
				Description: &t.Function.Description,
				Parameters:  raw,
			},
		})
	}

	chatCompletionsResp, err := m.client.GetChatCompletionsStream(ctx, azopenai.ChatCompletionsStreamOptions{
		Messages:       messagesOpenAI,
		DeploymentName: &opts.Model,
		// Tools:    toolsOpenAI,
	}, nil)
	if err != nil {
		return nil, err
	}

	response := llm.NewStreamResponse()
	go func() {
		defer chatCompletionsResp.ChatCompletionsStream.Close()
		defer response.Stream().Close()

		for {
			chatCompletions, err := chatCompletionsResp.ChatCompletionsStream.Read()

			if errors.Is(err, io.EOF) {
				break
			}

			if err != nil {
				//  TODO: Update the following line with your application specific error handling logic
				log.Printf("ERROR: %s", err)
				return
			}

			for _, choice := range chatCompletions.Choices {

				text := ""

				if choice.Delta.Content != nil {
					text = *choice.Delta.Content
				}

				// role := ""

				// if choice.Delta.Role != nil {
				// 	role = string(*choice.Delta.Role)
				// }
				response.Stream().Push(llm.NewChunk(0, llm.AssistantPromptMessage(text), nil))
			}
		}
	}()

	return response, nil
}

func (m *LLM) setupClient() error {
	config, _ := llm.Config.Get(Name).(map[string]any)
	if config == nil {
		return fmt.Errorf("azure-openai config not found")
	}

	apiKey, _ := config["api_key"].(string)
	endpoint, _ := config["endpoint"].(string)
	apiVersion, _ := config["api_version"].(string)
	keyCredential := azcore.NewKeyCredential(apiKey)
	options := &azopenai.ClientOptions{}

	options.APIVersion = apiVersion
	client, err := azopenai.NewClientWithKeyCredential(
		endpoint,
		keyCredential,
		options,
	)
	if err != nil {
		return err
	}
	m.client = client
	return nil
}
