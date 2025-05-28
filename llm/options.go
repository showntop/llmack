package llm

import (
	"context"
	"encoding/json"
)

// InvokeOptions ...
type InvokeOptions struct {
	Stop           []string                      `json:"stop,omitempty"`
	ResponseFormat *ChatCompletionResponseFormat `json:"response_format,omitempty"`
	// LogitBias is must be a token id string (specified by their token ID in the tokenizer), not a word string.
	// incorrect: `"logit_bias":{"You": 6}`, correct: `"logit_bias":{"1639": 6}`
	// refs: https://platform.openai.com/docs/api-reference/chat/create#chat/create-logit_bias
	LogitBias map[string]int `json:"logit_bias,omitempty"`
	// LogProbs indicates whether to return log probabilities of the output tokens or not.
	// If true, returns the log probabilities of each output token returned in the content of message.
	// This option is currently not available on the gpt-4-vision-preview model.
	LogProbs bool `json:"logprobs,omitempty"`
	// TopLogProbs is an integer between 0 and 5 specifying the number of most likely tokens to return at each
	// token position, each with an associated log probability.
	// logprobs must be set to true if this parameter is used.
	TopLogProbs int `json:"top_logprobs,omitempty"`
	// Options for streaming response. Only set this when you set stream: true.
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`

	// Model is the model to use.
	Model string `json:"model,omitempty"`
	// Stream is the stream output.
	Stream bool `json:"stream,omitempty"`
	// CandidateCount is the number of response candidates to generate.
	CandidateCount int `json:"candidate_count,omitempty"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature is the temperature for sampling, between 0 and 1.
	Temperature float64 `json:"temperature,omitempty"`
	// StopWords is a list of words to stop on.
	StopWords []string `json:"stop_words,omitempty"`
	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
	// TopK is the number of tokens to consider for top-k sampling.
	TopK int `json:"top_k,omitempty"`
	// TopP is the cumulative probability for top-p sampling.
	TopP float64 `json:"top_p,omitempty"`
	// Seed is a seed for deterministic sampling.
	Seed int `json:"seed,omitempty"`
	// MinLength is the minimum length of the generated text.
	MinLength int `json:"min_length,omitempty"`
	// MaxLength is the maximum length of the generated text.
	MaxLength int `json:"max_length,omitempty"`
	// N is how many chat completion choices to generate for each input message.
	N int `json:"n,omitempty"`
	// RepetitionPenalty is the repetition penalty for sampling.
	RepetitionPenalty float64 `json:"repetition_penalty,omitempty"`
	// FrequencyPenalty is the frequency penalty for sampling.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
	// PresencePenalty is the presence penalty for sampling.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`

	// JSONMode is a flag to enable JSON mode.
	JSONMode bool `json:"json_mode,omitempty"`

	// Tools is a list of tools to use. Each tool can be a specific tool or a function.
	Tools []*Tool `json:"tools,omitempty"`
	// ToolChoice is the choice of tool to use, it can either be "none", "auto" (the default behavior), or a specific tool as described in the ToolChoice type.
	ToolChoice any `json:"tool_choice,omitempty"`

	// Metadata is a map of metadata to include in the request.
	// The meaning of this field is specific to the backend in use.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type StreamOptions struct {
	// If set, an additional chunk will be streamed before the data: [DONE] message.
	// The usage field on this chunk shows the token usage statistics for the entire request,
	// and the choices field will always be an empty array.
	// All other chunks will also include a usage field, but with a null value.
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type ChatCompletionResponseFormatType string

const (
	ChatCompletionResponseFormatTypeJSONObject ChatCompletionResponseFormatType = "json_object"
	ChatCompletionResponseFormatTypeJSONSchema ChatCompletionResponseFormatType = "json_schema"
	ChatCompletionResponseFormatTypeText       ChatCompletionResponseFormatType = "text"
)

type ChatCompletionResponseFormat struct {
	Type       ChatCompletionResponseFormatType        `json:"type,omitempty"`
	JSONSchema *ChatCompletionResponseFormatJSONSchema `json:"json_schema,omitempty"`
}

type ChatCompletionResponseFormatJSONSchema struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Schema      json.Marshaler `json:"schema"`
	Strict      bool           `json:"strict"`
}

// Tool is a tool that can be used by the model.
type Tool struct {
	// Type is the type of the tool.
	Type string `json:"type"`
	// Function is the function to call.
	Function *FunctionDefinition `json:"function,omitempty"`
}

// FunctionDefinition is a definition of a function that can be called by the model.
type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description"`
	// Parameters is a list of parameters for the function.
	Parameters any `json:"parameters,omitempty"`
}

// ToolChoice is a specific tool to use.
type ToolChoice struct {
	// Type is the type of the tool.
	Type string `json:"type"`
	// Function is the function to call (if the tool is a function).
	Function *FunctionReference `json:"function,omitempty"`
}

// FunctionReference is a reference to a function.
type FunctionReference struct {
	// Name is the name of the function.
	Name string `json:"name"`
}

// FunctionCallBehavior is the behavior to use when calling functions.
type FunctionCallBehavior string

const (
	// FunctionCallBehaviorNone will not call any functions.
	FunctionCallBehaviorNone FunctionCallBehavior = "none"
	// FunctionCallBehaviorAuto will call functions automatically.
	FunctionCallBehaviorAuto FunctionCallBehavior = "auto"
)

// InvokeOption is a function that configures a InvokeOptions.
type InvokeOption func(*InvokeOptions)

// // WithSuperParams specifies the super params for the model.
// func WithSuperParams(o1 *InvokeOptions) InvokeOption {
// 	return func(o *InvokeOptions) {
// 		o.Temperature = o1.Temperature
// 		o.MaxTokens = o1.MaxTokens
// 		o.TopP = o1.TopP
// 		o.TopK = o1.TopK
// 		o.FrequencyPenalty = o1.FrequencyPenalty
// 		o.PresencePenalty = o1.PresencePenalty
// 		o.RepetitionPenalty = o1.RepetitionPenalty
// 	}
// }

func WithTemperature(temperature float64) InvokeOption {
	return func(o *InvokeOptions) {
		o.Temperature = temperature
	}
}

func WithMaxTokens(maxTokens int) InvokeOption {
	return func(o *InvokeOptions) {
		o.MaxTokens = maxTokens
	}
}

func WithTopP(topP float64) InvokeOption {
	return func(o *InvokeOptions) {
		o.TopP = topP
	}
}

func WithTopK(topK int) InvokeOption {
	return func(o *InvokeOptions) {
		o.TopK = topK
	}
}

func WithFrequencyPenalty(frequencyPenalty float64) InvokeOption {
	return func(o *InvokeOptions) {
		o.FrequencyPenalty = frequencyPenalty
	}
}

// WithModel specifies which model name to use.
func WithModel(model string) InvokeOption {
	return func(o *InvokeOptions) {
		o.Model = model
	}
}

// WithTools specifies which tools to use.
func WithTools(tools ...*Tool) InvokeOption {
	return func(o *InvokeOptions) {
		o.Tools = tools
	}
}

// WithStream specifies stream output.
func WithStream(stream bool) InvokeOption {
	return func(o *InvokeOptions) {
		o.Stream = stream
	}
}
