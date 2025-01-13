package node

import (
	"context"

	"github.com/showntop/llmack/workflow"
)

type startNode struct {
	executeable
	Identifier
}

// StartNode initializes new start node with start definition
func StartNode(n *workflow.Node) *startNode {
	return &startNode{
		Identifier: Identifier{
			id: n.ID,
		},
	}
}

// Exec executes start node
func (f startNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {
	return nil, nil
}
