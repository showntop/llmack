package llm

import "context"

// InvokeOptions ...
type InvokeOptions struct {
	// Model is the model to use.
	Model string `json:"model"`
	// Stream is the stream output.
	Stream bool `json:"stream"`
	// CandidateCount is the number of response candidates to generate.
	CandidateCount int `json:"candidate_count"`
	// MaxTokens is the maximum number of tokens to generate.
	MaxTokens int `json:"max_tokens"`
	// Temperature is the temperature for sampling, between 0 and 1.
	Temperature float64 `json:"temperature"`
	// StopWords is a list of words to stop on.
	StopWords []string `json:"stop_words"`
	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc func(ctx context.Context, chunk []byte) error `json:"-"`
	// TopK is the number of tokens to consider for top-k sampling.
	TopK int `json:"top_k"`
	// TopP is the cumulative probability for top-p sampling.
	TopP float64 `json:"top_p"`
	// Seed is a seed for deterministic sampling.
	Seed int `json:"seed"`
	// MinLength is the minimum length of the generated text.
	MinLength int `json:"min_length"`
	// MaxLength is the maximum length of the generated text.
	MaxLength int `json:"max_length"`
	// N is how many chat completion choices to generate for each input message.
	N int `json:"n"`
	// RepetitionPenalty is the repetition penalty for sampling.
	RepetitionPenalty float64 `json:"repetition_penalty"`
	// FrequencyPenalty is the frequency penalty for sampling.
	FrequencyPenalty float64 `json:"frequency_penalty"`
	// PresencePenalty is the presence penalty for sampling.
	PresencePenalty float64 `json:"presence_penalty"`

	// JSONMode is a flag to enable JSON mode.
	JSONMode bool `json:"json"`

	// Tools is a list of tools to use. Each tool can be a specific tool or a function.
	Tools []Tool `json:"tools,omitempty"`
	// ToolChoice is the choice of tool to use, it can either be "none", "auto" (the default behavior), or a specific tool as described in the ToolChoice type.
	ToolChoice any `json:"tool_choice"`

	// Function defitions to include in the request.
	// Deprecated: Use Tools instead.
	Functions []FunctionDefinition `json:"functions,omitempty"`
	// FunctionCallBehavior is the behavior to use when calling functions.
	//
	// If a specific function should be invoked, use the format:
	// `{"name": "my_function"}`
	// Deprecated: Use ToolChoice instead.
	FunctionCallBehavior FunctionCallBehavior `json:"function_call,omitempty"`
	// Metadata is a map of metadata to include in the request.
	// The meaning of this field is specific to the backend in use.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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
	Parameters map[string]any `json:"parameters,omitempty"`
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

// WithModel specifies which model name to use.
func WithModel(model string) InvokeOption {
	return func(o *InvokeOptions) {
		o.Model = model
	}
}

// WithTools specifies which tools to use.
func WithTools(tools ...Tool) InvokeOption {
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
