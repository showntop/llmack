package qwen

import (
	"context"
	"fmt"
	"sync"

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
	once   sync.Once
	engine *llm.OAILLM
}

var initOnce sync.Once

// Invoke TODO
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, optFuncs ...llm.InvokeOption) (*llm.Response, error) {
	var err error
	m.once.Do(func() {
		url := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
		config, _ := llm.Config.Get(Name).(map[string]any)
		if config == nil {
			err = fmt.Errorf(Name + " config not found")
		}
		apiKey, _ := config["api_key"].(string)
		m.engine = llm.NewOAILLM(url, apiKey)
	})
	if err != nil {
		return nil, err
	}
	return m.engine.Invoke(ctx, messages, optFuncs...)
}
