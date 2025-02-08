package deepseek

import (
	"context"
	"sync"

	"github.com/showntop/llmack/llm"
)

// Name ...
var Name = "deepseek"

func init() {
	llm.Register(Name, NewLLM())
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
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, optFuncs ...llm.InvokeOption) (*llm.Response, error) {
	m.once.Do(func() {
		url := "https://api.deepseek.com/chat/completions"

		config, _ := llm.Config.Get(Name).(map[string]any)
		if config == nil {
			// return nil, fmt.Errorf("deepseek config not found")
		}
		apiKey, _ := config["api_key"].(string)
		baseURL, _ := config["base_url"].(string)
		if baseURL != "" {
			url = baseURL + "/chat/completions"
		}
		m.engine = llm.NewOAILLM(url, apiKey)
	})
	return m.engine.Invoke(ctx, messages, optFuncs...)
}
