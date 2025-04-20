package program

import (
	"github.com/showntop/llmack/llm"
)

type option func(*predictor)

func WithLLM(provider string, model string, opts ...llm.Option) option {
	return func(p *predictor) {
		opts = append(opts, llm.WithDefaultModel(model))
		p.model = llm.NewInstance(provider, opts...)
	}
}

func WithLLMInstance(model *llm.Instance) option {
	return func(p *predictor) {
		p.model = model
	}
}
