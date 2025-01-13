package program

import (
	"context"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/zhipu"
)

// var defaultLLM = llm.NewInstance(moonshot.Name, llm.WithDefaultModel("moonshot-v1-8k"))
// var defaultLLM = llm.NewInstance(hunyuan.Name, llm.WithDefaultModel("hunyuan"))
var defaultLLM = llm.NewInstance(zhipu.Name, llm.WithDefaultModel("GLM-4-Flash"))

// SetLLM ...
func SetLLM(provider, model string) {
	defaultLLM = llm.NewInstance(provider, llm.WithDefaultModel(model))
}

// Program defines the interface for a program
type Program interface {
	Prompt() *Promptx
	Update(...option)
	Forward(context.Context, map[string]any) (any, error)
}

// ProgramFunc is a function that implements the Program interface
type ProgramFunc func(context.Context, string) (string, error)
