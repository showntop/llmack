package deepseek

import (
	"context"
	"fmt"
	"sync"

	"github.com/showntop/llmack/llm"
)

// Name ...
var Name = "deepseek"

func init() {
	llm.Register(Name, NewLLM)
}

// LLM ...
type LLM struct {
	once   sync.Once
	engine *llm.OAILLM
}

// NewLLM ...
func NewLLM(o *llm.ProviderOptions) llm.Provider {
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
		url := "https://api.deepseek.com/chat/completions"
		config, _ := llm.Config.Get(Name).(map[string]any)
		if config == nil {
			err = fmt.Errorf("deepseek config not found")
			// return nil, fmt.Errorf("deepseek config not found")
		}
		apiKey, _ := config["api_key"].(string)
		baseURL, _ := config["base_url"].(string)
		if baseURL != "" {
			url = baseURL + "/chat/completions"
		}
		m.engine = llm.NewOAILLM(url, apiKey)
	})
	if err != nil {
		return nil, err
	}
	return m.engine.Invoke(ctx, messages, opts)
}
