package node

import (
	"context"
	"fmt"
	"strings"

	"github.com/showntop/flatmap"

	"github.com/showntop/llmack/tool"
	"github.com/showntop/llmack/workflow"
	wf "github.com/showntop/llmack/workflow"
)

type toolNode struct {
	executeable
	Identifier
	Node *wf.Node
}

// ToolNode initializes new function node with function definition and configured arguments and results
func ToolNode(node *workflow.Node) (*toolNode, error) {
	// if def.Kind != wf.NodeKindFunction.Name() {
	// 	return nil, fmt.Errorf("expecting function kind")
	// }

	return &toolNode{Node: node}, nil
}

// Exec executes function node
//
// Configured arguments are evaluated with the node's scope and input and passed
// to the function
//
// Configured results are evaluated with the results from the function and final scope is then returned
func (n *toolNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	providerKind, _ := n.Node.Metadata["provider_kind"].(string)
	toolName, _ := n.Node.Metadata["tool_name"].(string)
	providerID, _ := n.Node.Metadata["provider_id"].(float64)
	stream, _ := n.Node.Metadata["stream"].(bool)

	if providerKind == "" || toolName == "" {
		return nil, fmt.Errorf("provider_kind or tool_name is empty")
	}

	mp, err := flatmap.Flatten(r.Scope, flatmap.DefaultTokenizer)
	if err != nil {
		return nil, err
	}
	// 提取input
	inputs := make(map[string]any)
	for _, input := range n.Node.Inputs {
		pointer := strings.TrimPrefix(input.Value, "{{")
		pointer = strings.TrimSuffix(pointer, "}}")
		value := mp.Get(pointer)
		inputs[input.Name] = value
	}

	toolRun := tool.Instantiate(int64(providerID), providerKind, toolName)
	if stream {
		results, err := toolRun.Stream(ctx, inputs)
		return map[string]any{"result": results}, err
	}
	result, err := toolRun.Invoke(ctx, inputs)
	return map[string]any{"result": result}, err
}
