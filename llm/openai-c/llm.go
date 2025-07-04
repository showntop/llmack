package deepseek

import (
	"context"
	"fmt"
	"sync"

	"github.com/showntop/llmack/llm"
)

// Name ...
var Name = "openai-c"

func init() {
	llm.Register(Name, func(o *llm.ProviderOptions) llm.Provider { return &LLM{} })
}

// LLM ...
type LLM struct {
	once   sync.Once
	engine *llm.OAILLM
}

// NewLLM ...
func NewLLM() *LLM {
	return &LLM{}
}

// Name ...
func (m *LLM) Name() string {
	return Name
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, opts *llm.InvokeOptions) (*llm.Response, error) {
	var err error
	m.once.Do(func() {
		config, _ := llm.Config.Get(Name).(map[string]any)
		if config == nil {
			err = fmt.Errorf("openai-c config not found")
		}
		apiKey, _ := config["api_key"].(string)
		baseURL, _ := config["base_url"].(string)
		url := baseURL + "/chat/completions"
		m.engine = llm.NewOAILLM(url, apiKey)
	})
	if err != nil {
		return nil, err
	}
	return m.engine.Invoke(ctx, messages, opts)
}
