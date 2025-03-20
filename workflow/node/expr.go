package node

import (
	"context"
	"errors"

	"github.com/expr-lang/expr"

	"github.com/showntop/llmack/workflow"
	wf "github.com/showntop/llmack/workflow"
)

// exprNode TODO
type exprNode struct {
	executeable
	Identifier

	Node *wf.Node
}

// ExprNode 创建expr node
func ExprNode(n *workflow.Node) *exprNode {
	return &exprNode{
		Node: n,
	}
}

// Execute 执行JSON节点，单次执行
func (n *exprNode) Execute(ctx context.Context, r *ExecRequest) (ExecResponse, error) {

	expression, _ := n.Node.Metadata["expr"].(string)
	if expression == "" {
		return nil, nil
	}

	program, err := expr.Compile(expression)
	if err != nil {
		return nil, errors.Join(err)
	}
	result, err := expr.Run(program, r.Inputs)
	if err != nil {
		return nil, errors.Join(err)
	}

	return result, nil
}
