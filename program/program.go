package program

import (
	"context"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/llm/deepseek"
)

// var defaultLLM = llm.NewInstance(moonshot.Name, llm.WithDefaultModel("moonshot-v1-8k"))
// var defaultLLM = llm.NewInstance(hunyuan.Name, llm.WithDefaultModel("hunyuan"))
// var defaultLLM = llm.NewInstance(zhipu.Name, llm.WithDefaultModel("GLM-4-Flash"))
var defaultLLM = llm.New(deepseek.Name, llm.WithDefaultModel("deepseek-chat"))

// SetLLM can not set concurrent
func SetLLM(provider, model string) {
	defaultLLM = llm.New(provider, llm.WithDefaultModel(model))
}

// Program defines the interface for a program
// Program 定义了一个可运行的程序，不会保存状态，随用随 new
type Program interface {
	Prompt() *Promptx
	Update(...option)
	Forward(context.Context, map[string]any) (any, error)
}

// ProgramFunc is a function that implements the Program interface
type ProgramFunc func(context.Context, string) (string, error)
