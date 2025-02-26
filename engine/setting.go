package engine

import (
	"encoding/json"

	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/workflow"
)

// DefaultSettings ...
func DefaultSettings() *Settings {
	return &Settings{
		PresetPrompt: "",
		Preamble:     "",
		Knowledge:    make([]rag.Options, 0),
		LLMModel: struct {
			Provider string `json:"provider"`
			Name     string `json:"name"`
		}{
			Provider: "openai",
			Name:     "gpt-3.5-turbo",
		},
		Workflow: nil,
		Tools:    nil,
		Stream:   true,
	}
}

// Settings ...
type Settings struct {
	PresetPrompt string        `json:"preset_prompt"`
	Preamble     string        `json:"preamble"` // 机器人开场白
	Knowledge    []rag.Options `json:"knowledge"`
	LLMModel     struct {
		Provider string `json:"provider"`
		Name     string `json:"name"`
	} `json:"llm_model"`
	Agent struct {
		Mode         string `json:"mode"`
		MaxIteration int    `json:"max_iteration"`
	}
	Workflow *workflow.Workflow `json:"workflow"`
	Tools    []string           `json:"tools"`
	Stream   bool               `json:"stream"`
}

// String ... return json string
func (s *Settings) String() string {
	raw, _ := json.Marshal(s)
	return string(raw)
}
