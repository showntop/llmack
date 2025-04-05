package program

import (
	"reflect"

	"github.com/showntop/llmack/llm"
)

type option func(*predictor)

func WithLLM(provider string, model string, opts ...llm.Option) option {
	return func(p *predictor) {
		opts = append(opts, llm.WithDefaultModel(model))
		p.model = llm.NewInstance(provider, opts...)
	}
}

// WithInstruction ...
func WithInstruction(info string) option {
	return func(p *predictor) {
		p.Promptx.Instruction = info
	}
}

// WithOutput ...
func WithOutput(tuple ...any) option {
	return func(p *predictor) {
		if len(tuple) <= 0 {
			return
		}
		out := &Field{Name: tuple[0].(string)}
		if len(tuple) >= 2 {
			out.Description = tuple[1].(string)
		}
		if len(tuple) >= 3 {
			out.Marker = tuple[2].(string)
		}
		if len(tuple) >= 4 {
			out.Type = tuple[3].(reflect.Kind)
		}
		// 重复 ？
		p.Promptx.OutputFields[out.Name] = out
	}
}
