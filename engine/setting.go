package engine

import (
	"encoding/json"

	"github.com/showntop/llmack/rag"
	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/workflow"
)

// DefaultSettings ...
func DefaultSettings() *Settings {
	return &Settings{
		ChatMode:     BotModeChat,
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
	ChatMode     int           `json:"chat_mode"` // 1. 一问一答 2. 多轮问答
	PresetPrompt string        `json:"preset_prompt"`
	Preamble     string        `json:"preamble"` // 机器人开场白
	Knowledge    []rag.Options `json:"knowledge"`
	LLMModel     struct {
		Provider string `json:"provider"`
		Name     string `json:"name"`
	} `json:"llm_model"`
	Agent struct {
		MaxIteration int `json:"max_iteration"`
	}
	Workflow *workflow.Workflow `json:"workflow"`
	Tools    []ToolSetting      `json:"tools"`
	Stream   bool               `json:"stream"`
}

// String ... return json string
func (s *Settings) String() string {
	raw, _ := json.Marshal(s)
	return string(raw)
}

// ToolSetting ...
type ToolSetting struct {
	ProviderKind string //  builtin api workflow
	ProviderID   int64  //  builtin api workflow

	Name        string
	Key         string
	Description string
	Parameters  []tool.Parameter
	Parameters2 string

	// 扩展
	Extensions map[string]any
}
