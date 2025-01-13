package node

import (
	"context"

	"github.com/showntop/llmack/workflow"
)

// VDBNode ...
type VDBNode struct {
	// Node
}

// NewVDBNode 创建VDBNode
func NewVDBNode(n *workflow.Node) *VDBNode {
	return &VDBNode{
		// Node: *n,
	}
}

// Execute 执行VDBNode
func (n *VDBNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {

	return nil, nil
}
