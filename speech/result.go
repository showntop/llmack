package speech

import "github.com/showntop/llmack/llm"

// AudioChunk ...
type AudioChunk struct {
	Audio     string
	Text      string
	ToolCalls []llm.ToolCall
}
