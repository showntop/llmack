package node

import (
	"context"
	"strings"

	"github.com/showntop/llmack/llm"
	"github.com/showntop/llmack/prompt"
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
	stream, ok := n.Node.Metadata["stream"].(bool)
	if !ok {
		stream = true
	}
	provider, _ := n.Node.Metadata["provider"].(string)
	modelName, _ := n.Node.Metadata["model"].(string)
	systemPrompt, _ := n.Node.Metadata["system_prompt"].(string)
	userPrompt, _ := n.Node.Metadata["user_prompt"].(string)
	model := llm.New(provider, llm.WithDefaultModel(modelName))

	// 处理 inputs @TODO default from query field
	messages := []llm.Message{}
	if systemPrompt != "" {
		newSystemPrompt, err := prompt.Render(systemPrompt, r.Inputs)
		if err != nil {
			return nil, err
		}
		messages = append(messages, llm.NewSystemMessage(newSystemPrompt))
	}
	newUserPrompt, err := prompt.Render(userPrompt, r.Inputs)
	if err != nil {
		return nil, err
	}
	messages = append(messages, llm.NewUserTextMessage(newUserPrompt))
	response, err := model.Invoke(ctx, messages, llm.WithStream(true))
	if err != nil {
		return nil, err
	}
	if stream {
		return response, nil
	}
	result := response.Result()
	// result to map
	var mmm = map[string]any{
		"message": map[string]any{
			"content": strings.TrimLeft(strings.Trim(result.Message.Content(), "```"), "json"),
		},
		"model": result.Model,
		"usage": result.Usage,
	}
	return mmm, nil
}
