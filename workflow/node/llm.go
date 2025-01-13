package node

import (
	"context"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/workflow"
)

// LLMNode TODO
type LLMNode struct {
	workflow.Node
}

// NewLLMNode 创建LLMNode
func NewLLMNode(n *workflow.Node) *LLMNode {
	return &LLMNode{
		Node: *n,
	}
}

// Execute 执行LLM节点，单次执行
func (n *LLMNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	// 解析metadata，获取模型配置
	model := llm.NewInstance("zhipu")
	messages := []llm.Message{llm.SystemPromptMessage(" "), llm.UserPromptMessage("content")}
	model.Invoke(ctx, messages, nil)
	return nil, nil
}
