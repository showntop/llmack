package doubao

import (
	"context"
	"sync"

	"github.com/showntop/llmack/llm"
)

// Name ...
var Name = "doubao"

func init() {
	llm.Register(Name, NewLLM)
}

// LLM ...
type LLM struct {
	opts   *llm.ProviderOptions
	once   sync.Once
	engine *llm.OAILLM
}

// NewLLM ...
func NewLLM(o *llm.ProviderOptions) llm.Provider {
	return &LLM{opts: o}
}

// Name ...
func (m *LLM) Name() string {
	return Name
}

// Invoke ...
func (m *LLM) Invoke(ctx context.Context, messages []llm.Message, opts *llm.InvokeOptions) (*llm.Response, error) {
	// 优先使用 provider 的配置，如果 provider 的配置为空，则使用 config 的配置， 最后使用默认配置
	var url string = "https://ark.cn-beijing.volces.com/api/v3"
	if m.opts.BaseURL != "" {
		url = m.opts.BaseURL + "/chat/completions"
	} else if config, _ := llm.Config.Get(Name).(map[string]any); config != nil {
		if value, ok := config["base_url"].(string); ok {
			url = value + "/chat/completions"
		}
	}

	var apiKey string
	if m.opts.ApiKey != "" {
		apiKey = m.opts.ApiKey
	} else if config, _ := llm.Config.Get(Name).(map[string]any); config != nil {
		if value, ok := config["api_key"].(string); ok {
			apiKey = value
		}
	}
	m.once.Do(func() {
		m.engine = llm.NewOAILLM(url, apiKey)
	})
	return m.engine.Invoke(ctx, messages, opts)
}
