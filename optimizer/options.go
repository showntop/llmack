package optimizer

import (
	"github.com/showntop/llmack/llm"
)

// Options ...
type Options struct {
	LLM      *llm.Instance
	LLMModel string
	trainset []*Example
	metric   Metric
}

// NewOptions ...
func NewOptions(opts ...Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// Option ...
type Option func(options *Options)

// WithLLM ...
func WithLLM(llm *llm.Instance, model string) Option {
	return func(options *Options) {
		options.LLM = llm
		options.LLMModel = model
	}
}

// WithTrainset ...
func WithTrainset(exs []*Example) Option {
	return func(options *Options) {
		options.trainset = exs
	}
}

// WithMetric ...
func WithMetric(metric Metric) Option {
	return func(options *Options) {
		options.metric = metric
	}
}

// OptimizeOptions ...
type OptimizeOptions struct {
	InitialInstruction string
	Description        string
	Inputs             string
	Outputs            string
}

// NewOptimizeOptions ...
func NewOptimizeOptions() *OptimizeOptions {
	options := &OptimizeOptions{}
	// default
	return options
}

// OptimizeOption ...
type OptimizeOption func(options *OptimizeOptions)

// WithInitialInstruction ...
func WithInitialInstruction(instruction string) OptimizeOption {
	return func(options *OptimizeOptions) {
		options.InitialInstruction = instruction
	}
}

// WithDescription ...
func WithDescription(desc string) OptimizeOption {
	return func(options *OptimizeOptions) {
		options.Description = desc
	}
}

// WithOutputs ...
func WithOutputs(out string) OptimizeOption {
	return func(options *OptimizeOptions) {
		options.Outputs = out
	}
}
