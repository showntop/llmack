package node

import (
	"context"
	"encoding/json"
	"fmt"

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
	toolName, _ := n.Node.Metadata["tool_name"].(string)
	stream, _ := n.Node.Metadata["stream"].(bool)

	// if providerKind == "" || toolName == "" {
	if toolName == "" {
		return nil, fmt.Errorf("provider_kind or tool_name is empty")
	}

	toolRun := tool.Spawn(toolName)

	inputsJSON, err := json.Marshal(r.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inputs: %v", err)
	}

	if stream {
		result, err := toolRun.Invoke(ctx, string(inputsJSON))
		return map[string]any{"result": result}, err
	}
	result, err := toolRun.Invoke(ctx, string(inputsJSON))
	return map[string]any{"result": result}, err
}
