package node

import (
	"context"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/workflow"
	wf "github.com/showntop/llmack/workflow"
)

// llmNode TODO
type llmNode struct {
	Node *wf.Node

	executeable
	Identifier
}

// LLMNode 创建LLMNode
func LLMNode(n *workflow.Node) *llmNode {
	return &llmNode{
		Node: n,
	}
}

// Execute 执行LLM节点，单次执行
func (n *llmNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	// 解析metadata，获取模型配置
	provider, _ := n.Node.Metadata["provider"].(string)
	modelName, _ := n.Node.Metadata["model"].(string)
	model := llm.NewInstance(provider, llm.WithDefaultModel(modelName))

	// 处理 inputs @TODO default from query field
	messages := []llm.Message{llm.UserTextPromptMessage(r.Inputs["query"].(string))}
	return model.Invoke(ctx, messages, llm.WithStream(true))
}
