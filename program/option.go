package program

import (
	"context"

	"github.com/showntop/llmack/llm"
)

type option func(*predictor)

func WithAdapter(adapter Adapter) option {
	return func(p *predictor) {
		p.adapter = adapter
	}
}

func WithLLM(provider string, model string, opts ...llm.Option) option {
	return func(p *predictor) {
		opts = append(opts, llm.WithDefaultModel(model))
		p.model = llm.New(provider, opts...)
	}
}

func WithLLMInstance(model *llm.Instance) option {
	return func(p *predictor) {
		p.model = model
	}
}

func WithMaxIterationNum(num int) option {
	return func(p *predictor) {
		p.maxIterationNum = num
	}
}

func WithResetMessages(resetMessages func(ctx context.Context, messages []llm.Message) []llm.Message) option {
	return func(p *predictor) {
		p.resetMessages = resetMessages
	}
}
